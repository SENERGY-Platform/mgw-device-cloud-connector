package cloud_hdl

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/go-service-base/context-hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"os"
	"path"
	"sync"
	"time"
)

const logPrefix = "[cloud-hdl]"

type Handler struct {
	cloudClient     cloud_client.ClientItf
	subjectProvider handler.SubjectProvider
	wrkSpacePath    string
	attrOrigin      string
	userID          string
	data            data
	lastSync        time.Time
	syncInterval    time.Duration
	noNetwork       bool
	mu              sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, subjectProvider handler.SubjectProvider, syncInterval time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:     cloudClient,
		subjectProvider: subjectProvider,
		syncInterval:    syncInterval,
		wrkSpacePath:    wrkSpacePath,
		attrOrigin:      attrOrigin,
	}
}

func (h *Handler) Init(ctx context.Context, networkID, networkName string, delay time.Duration) (string, string, error) {
	if !path.IsAbs(h.wrkSpacePath) {
		return "", "", fmt.Errorf("workspace path must be absolute")
	}
	util.Logger.Info(logPrefix, " begin init")
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return "", "", err
	}
	d, err := readData(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	if networkID == "" {
		networkID = d.NetworkID
	}
	ch := context_hdl.New()
	defer ch.CancelAll()
	timer := time.NewTimer(time.Millisecond * 10)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	stop := false
	for !stop {
		select {
		case <-timer.C:
			util.Logger.Debugf("%s get user ID", logPrefix)
			h.userID, err = h.subjectProvider.GetUserID(ch.Add(context.WithCancel(ctx)))
			if err != nil {
				util.Logger.Errorf("%s get user ID: %s", logPrefix, err)
				timer.Reset(delay)
				break
			}
			if h.userID == "" {
				util.Logger.Errorf("%s get user ID: invalid", logPrefix)
				timer.Reset(delay)
				break
			}
			stop = true
			break
		case <-ctx.Done():
			return "", "", fmt.Errorf("get user ID: %s", ctx.Err())
		}
	}
	timer.Reset(time.Millisecond * 10)
	stop = false
	for !stop {
		select {
		case <-timer.C:
			if networkID != "" {
				util.Logger.Debugf("%s get network (%s)", logPrefix, networkID)
				if hb, err := h.cloudClient.GetHub(ch.Add(context.WithCancel(ctx)), networkID); err != nil {
					var nfe *cloud_client.NotFoundError
					if !errors.As(err, &nfe) {
						util.Logger.Errorf("%s get network (%s): %s", logPrefix, networkID, err)
						timer.Reset(delay)
						break
					}
					util.Logger.Warningf("%s get network (%s): %s", logPrefix, networkID, err)
					if networkID != d.NetworkID && d.NetworkID != "" {
						networkID = d.NetworkID
						timer.Reset(time.Millisecond * 10)
						break
					} else {
						networkID = ""
					}
				} else {
					if hb.OwnerId != h.userID {
						util.Logger.Warningf("%s get network (%s): invalid user ID", logPrefix, networkID)
						if networkID != d.NetworkID && d.NetworkID != "" {
							networkID = d.NetworkID
							timer.Reset(time.Millisecond * 10)
							break
						} else {
							networkID = ""
						}
					} else {
						stop = true
						break
					}
				}
			}
			if networkID == "" {
				util.Logger.Info(logPrefix, " create network")
				hID, err := h.cloudClient.CreateHub(ch.Add(context.WithCancel(ctx)), models.Hub{Name: networkName})
				if err != nil {
					util.Logger.Errorf("%s create network: %s", logPrefix, err)
					timer.Reset(delay)
					break
				}
				networkID = hID
				util.Logger.Infof("%s created network (%s)", logPrefix, networkID)
				stop = true
				break
			}
		case <-ctx.Done():
			return "", "", fmt.Errorf("init network: %s", ctx.Err())
		}
	}
	d.NetworkID = networkID
	h.data = d
	return d.NetworkID, h.userID, writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) ([]string, []string, []string, []string, error) {
	for _, lID := range missingIDs {
		delete(h.data.DeviceIDMap, lID)
	}
	if len(newIDs)+len(changedIDs) == 0 && time.Since(h.lastSync) < h.syncInterval {
		return nil, nil, nil, nil, nil
	}
	if len(newIDs)+len(changedIDs) > 0 {
		util.Logger.Info(logPrefix, " begin devices and network sync")
	} else {
		util.Logger.Debug(logPrefix, " begin periodic devices and network sync")
	}
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s get network (%s)", logPrefix, h.data.NetworkID)
	network, err := h.cloudClient.GetHub(ctxWc, h.data.NetworkID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noNetwork = true
			h.mu.Unlock()
		}
		return nil, nil, nil, nil, fmt.Errorf("get network (%s): %s", h.data.NetworkID, err)
	}
	var createFailed []string
	syncResults := make(map[string]bool)
	for _, lID := range newIDs {
		err = h.syncDevice(ctx, devices[lID])
		if err != nil {
			createFailed = append(createFailed, lID)
			syncResults[lID] = false
			continue
		}
		syncResults[lID] = true
	}
	var updateFailed []string
	for _, lID := range changedIDs {
		err = h.syncDevice(ctx, devices[lID])
		if err != nil {
			updateFailed = append(updateFailed, lID)
			syncResults[lID] = false
			continue
		}
		syncResults[lID] = true
	}
	var recreated []string
	networkLocalIDSet := make(map[string]struct{})
	for _, lID := range network.DeviceLocalIds {
		networkLocalIDSet[lID] = struct{}{}
	}
	for lID, device := range devices {
		if _, ok := syncResults[lID]; !ok {
			if _, ok := networkLocalIDSet[lID]; !ok {
				err = h.syncDevice(ctx, device)
				if err != nil {
					createFailed = append(createFailed, lID)
					syncResults[lID] = false
					continue
				}
				recreated = append(recreated, lID)
				syncResults[lID] = true
			}
		}
	}
	updateNetwork := false
	for lID, synced := range syncResults {
		if _, ok := networkLocalIDSet[lID]; !ok && synced {
			network.DeviceLocalIds = append(network.DeviceLocalIds, lID)
			updateNetwork = true
		}
	}
	network.DeviceIds = nil
	ctxWc2, cf2 := context.WithCancel(ctx)
	defer cf2()
	if updateNetwork {
		util.Logger.Infof("%s update network (%s)", logPrefix, h.data.NetworkID)
	} else {
		util.Logger.Debugf("%s update network (%s)", logPrefix, h.data.NetworkID)
	}
	if err = h.cloudClient.UpdateHub(ctxWc2, network); err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noNetwork = true
			h.mu.Unlock()
		}
		return nil, nil, nil, nil, fmt.Errorf("update network (%s): %s", h.data.NetworkID, err)
	}
	h.lastSync = time.Now()
	if err = writeData(h.wrkSpacePath, h.data); err != nil {
		util.Logger.Errorf("%s write data: %s", logPrefix, err)
	}
	return recreated, createFailed, updateFailed, nil, nil
}

func (h *Handler) HasNetwork() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noNetwork
}

func (h *Handler) syncDevice(ctx context.Context, device model.Device) (err error) {
	rID, ok := h.data.DeviceIDMap[device.ID]
	var rIDNew string
	if !ok {
		rIDNew, err = h.createOrUpdateDevice(ctx, device)
	} else {
		rIDNew, err = h.updateOrCreateDevice(ctx, rID, device)
	}
	if err != nil {
		util.Logger.Errorf("%s %s", logPrefix, err)
		return
	}
	if rIDNew != rID {
		util.Logger.Debugf("%s update device id cache: %s -> %s", logPrefix, device.ID, rIDNew)
		h.data.DeviceIDMap[device.ID] = rIDNew
	}
	return
}

func (h *Handler) createOrUpdateDevice(ctx context.Context, device model.Device) (string, error) {
	ch := context_hdl.New()
	defer ch.CancelAll()
	nd := newDevice(device, "", h.attrOrigin)
	util.Logger.Debugf("%s create device (%s)", logPrefix, device.ID)
	rID, err := h.cloudClient.CreateDevice(ch.Add(context.WithCancel(ctx)), nd)
	if err != nil {
		var bre *cloud_client.BadRequestError
		if !errors.As(err, &bre) {
			return "", fmt.Errorf("create device (%s): %s", device.ID, err)
		}
		util.Logger.Warningf("%s create device (%s): %s", logPrefix, device.ID, err)
		util.Logger.Debugf("%s get device (%s)", logPrefix, device.ID)
		d, err := h.cloudClient.GetDeviceL(ch.Add(context.WithCancel(ctx)), device.ID)
		if err != nil {
			return "", fmt.Errorf("get device (%s): %s", device.ID, err)
		}
		rID = d.Id
		nd.Id = d.Id
		util.Logger.Debugf("%s update device (%s)", logPrefix, device.ID)
		if err = h.cloudClient.UpdateDevice(ch.Add(context.WithCancel(ctx)), nd, h.attrOrigin); err != nil {
			return "", fmt.Errorf("update device (%s): %s", device.ID, err)
		}
		util.Logger.Infof("%s updated device (%s)", logPrefix, device.ID)
	} else {
		util.Logger.Infof("%s created device (%s)", logPrefix, device.ID)
	}
	return rID, nil
}

func (h *Handler) updateOrCreateDevice(ctx context.Context, rID string, device model.Device) (string, error) {
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s update device (%s)", logPrefix, device.ID)
	err := h.cloudClient.UpdateDevice(ctxWc, newDevice(device, rID, h.attrOrigin), h.attrOrigin)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		var fe *cloud_client.ForbiddenError
		if !errors.As(err, &nfe) && !errors.As(err, &fe) {
			return "", fmt.Errorf("update device (%s): %s", device.ID, err)
		}
		util.Logger.Warningf("%s update device (%s): %s", logPrefix, device.ID, err)
		return h.createOrUpdateDevice(ctx, device)
	}
	util.Logger.Infof("%s updated device (%s)", logPrefix, device.ID)
	return rID, err
}

func (h *Handler) getDeviceIDMap(ctx context.Context, oldMap map[string]string, deviceIDs []string) (map[string]string, error) {
	deviceIDMap := make(map[string]string)
	if len(deviceIDs) > 0 {
		ch := context_hdl.New()
		defer ch.CancelAll()
		rDeviceIDMap := make(map[string]string)
		for lID, rID := range oldMap {
			rDeviceIDMap[rID] = lID
		}
		for _, rID := range deviceIDs {
			lID, ok := rDeviceIDMap[rID]
			if !ok {
				util.Logger.Debugf("%s get device (%s)", logPrefix, rID)
				device, err := h.cloudClient.GetDevice(ch.Add(context.WithCancel(ctx)), rID)
				if err != nil {
					var nfe *cloud_client.NotFoundError
					if !errors.As(err, &nfe) {
						return nil, fmt.Errorf("get device (%s): %s", rID, err)
					}
					util.Logger.Warningf("%s get device (%s): %s", logPrefix, rID, err)
					continue
				}
				lID = device.LocalId
				util.Logger.Debugf("%s update device id cache: %s -> %s", logPrefix, lID, rID)
			}
			deviceIDMap[lID] = rID
		}
	}
	return deviceIDMap, nil
}

func newDevice(device model.Device, rID, attrOrigin string) models.Device {
	var attributes []models.Attribute
	for _, attribute := range device.Attributes {
		attributes = append(attributes, models.Attribute{
			Key:    attribute.Key,
			Value:  attribute.Value,
			Origin: attrOrigin,
		})
	}
	return models.Device{
		Id:           rID,
		LocalId:      device.ID,
		Name:         device.Name,
		Attributes:   attributes,
		DeviceTypeId: device.Type,
	}
}

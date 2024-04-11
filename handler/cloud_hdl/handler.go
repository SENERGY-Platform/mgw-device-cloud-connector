package cloud_hdl

import (
	"context"
	"errors"
	"fmt"
	context_hdl "github.com/SENERGY-Platform/go-service-base/context-hdl"
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
	cloudClient  cloud_client.ClientItf
	wrkSpacePath string
	attrOrigin   string
	data         data
	lastSync     time.Time
	syncInterval time.Duration
	noHub        bool
	mu           sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, syncInterval time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		syncInterval: syncInterval,
		wrkSpacePath: wrkSpacePath,
		attrOrigin:   attrOrigin,
	}
}

func (h *Handler) Init(ctx context.Context, hubID, hubName string, delay time.Duration) (string, error) {
	if !path.IsAbs(h.wrkSpacePath) {
		return "", fmt.Errorf("workspace path must be absolute")
	}
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return "", err
	}
	d, err := readData(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if hubID != "" {
		d.NetworkID = hubID
	}
	d.DefaultNetworkName = hubName
	timer := time.NewTimer(time.Millisecond * 10)
	stop := false
	util.Logger.Info(logPrefix, " begin hub init")
	for !stop {
		select {
		case <-timer.C:
			if d.NetworkID != "" {
				ctxWc, cf := context.WithCancel(ctx)
				defer cf()
				util.Logger.Debugf("%s get hub (%s)", logPrefix, d.NetworkID)
				if hb, err := h.cloudClient.GetHub(ctxWc, d.NetworkID); err != nil {
					var nfe *cloud_client.NotFoundError
					if !errors.As(err, &nfe) {
						util.Logger.Errorf("%s get hub (%s): %s", logPrefix, d.NetworkID, err)
						timer.Reset(delay)
						break
					}
					util.Logger.Warningf("%s get hub (%s): %s", logPrefix, d.NetworkID, err)
					d.NetworkID = ""
				} else {
					if deviceIDMap, err := h.getDeviceIDMap(ctx, d.DeviceIDMap, hb.DeviceIds); err == nil {
						d.DeviceIDMap = deviceIDMap
					} else {
						util.Logger.Errorf("%s refresh device id cache: %s", logPrefix, err)
					}
					stop = true
					break
				}
			}
			if d.NetworkID == "" {
				ctxWc, cf := context.WithCancel(ctx)
				defer cf()
				util.Logger.Info(logPrefix, " create hub")
				hID, err := h.cloudClient.CreateHub(ctxWc, models.Hub{Name: hubName})
				if err != nil {
					util.Logger.Errorf("%s create hub: %s", logPrefix, err)
					timer.Reset(delay)
					break
				}
				d.NetworkID = hID
				util.Logger.Infof("%s created hub (%s)", logPrefix, hID)
				stop = true
				break
			}
		case <-ctx.Done():
			return "", fmt.Errorf("init hub: %s", ctx.Err())
		}
	}
	if d.DeviceIDMap == nil {
		d.DeviceIDMap = make(map[string]string)
	}
	h.data = d
	return d.NetworkID, writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) ([]string, []string, []string, []string, error) {
	for _, lID := range missingIDs {
		delete(h.data.DeviceIDMap, lID)
	}
	if len(newIDs)+len(changedIDs) == 0 && time.Since(h.lastSync) < h.syncInterval {
		return nil, nil, nil, nil, nil
	}
	if len(newIDs)+len(changedIDs) > 0 {
		util.Logger.Info(logPrefix, " begin devices and hub sync")
	} else {
		util.Logger.Debug(logPrefix, " begin periodic devices and hub sync")
	}
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s get hub (%s)", logPrefix, h.data.NetworkID)
	hb, err := h.cloudClient.GetHub(ctxWc, h.data.NetworkID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noHub = true
			h.mu.Unlock()
		}
		return nil, nil, nil, nil, fmt.Errorf("get hub (%s): %s", h.data.NetworkID, err)
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
	hubLocalIDSet := make(map[string]struct{})
	for _, lID := range hb.DeviceLocalIds {
		hubLocalIDSet[lID] = struct{}{}
	}
	for lID, device := range devices {
		if _, ok := syncResults[lID]; !ok {
			if _, ok := hubLocalIDSet[lID]; !ok {
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
	updateHub := false
	for lID, synced := range syncResults {
		if _, ok := hubLocalIDSet[lID]; !ok && synced {
			hb.DeviceLocalIds = append(hb.DeviceLocalIds, lID)
			updateHub = true
		}
	}
	hb.DeviceIds = nil
	ctxWc2, cf2 := context.WithCancel(ctx)
	defer cf2()
	if updateHub {
		util.Logger.Infof("%s update hub (%s)", logPrefix, h.data.NetworkID)
	} else {
		util.Logger.Debugf("%s update hub (%s)", logPrefix, h.data.NetworkID)
	}
	if err = h.cloudClient.UpdateHub(ctxWc2, hb); err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noHub = true
			h.mu.Unlock()
		}
		return nil, nil, nil, nil, fmt.Errorf("update hub (%s): %s", h.data.NetworkID, err)
	}
	h.lastSync = time.Now()
	if err = writeData(h.wrkSpacePath, h.data); err != nil {
		util.Logger.Errorf("%s write data: %s", logPrefix, err)
	}
	return recreated, createFailed, updateFailed, nil, nil
}

func (h *Handler) HasHub() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noHub
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
		util.Logger.Error("%s %s", logPrefix, err)
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

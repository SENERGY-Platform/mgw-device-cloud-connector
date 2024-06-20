package cloud_hdl

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
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
	"slices"
	"strings"
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
	if network.OwnerId != h.userID {
		h.mu.Lock()
		h.noNetwork = true
		h.mu.Unlock()
		return nil, nil, nil, nil, fmt.Errorf("get network (%s): invalid user ID", h.data.NetworkID)
	}
	cloudDevices := make(map[string]models.Device)
	if len(network.DeviceIds) > 0 {
		ctxWc2, cf2 := context.WithCancel(ctx)
		defer cf2()
		devicesList, err := h.cloudClient.GetDevices(ctxWc2, network.DeviceIds)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("get devices: %s", err)
		}
		for _, device := range devicesList {
			if device.OwnerId != h.userID {
				util.Logger.Warningf("%s get devices: device (%s) invalid user ID", logPrefix, device.Id)
				continue
			}
			cloudDevices[device.LocalId] = device
		}
	}
	syncedIDs := make(map[string]string)
	var createFailed []string
	for _, lID := range newIDs {
		cID, err := h.syncDevice(ctx, cloudDevices, devices[lID])
		if err != nil {
			createFailed = append(createFailed, lID)
			continue
		}
		syncedIDs[lID] = cID
	}
	var updateFailed []string
	for _, lID := range changedIDs {
		cID, err := h.syncDevice(ctx, cloudDevices, devices[lID])
		if err != nil {
			updateFailed = append(updateFailed, lID)
			continue
		}
		syncedIDs[lID] = cID
	}
	var recreated []string
	for lID, lDevice := range devices {
		if _, ok := syncedIDs[lID]; !ok {
			_, inCloud := cloudDevices[lID]
			cID, err := h.syncDevice(ctx, cloudDevices, lDevice)
			if err != nil {
				if !inCloud {
					createFailed = append(createFailed, lID)
				}
				continue
			}
			if !inCloud {
				recreated = append(recreated, lID)
			}
			syncedIDs[lID] = cID
		}
	}
	networkDeviceIDSet := make(map[string]struct{})
	for _, id := range network.DeviceIds {
		networkDeviceIDSet[id] = struct{}{}
	}
	lenOld := len(networkDeviceIDSet)
	for _, cID := range syncedIDs {
		networkDeviceIDSet[cID] = struct{}{}
	}
	if lenOld != len(networkDeviceIDSet) {
		util.Logger.Infof("%s update network (%s)", logPrefix, h.data.NetworkID)
		var deviceIDs []string
		for id := range networkDeviceIDSet {
			deviceIDs = append(deviceIDs, id)
		}
		network.DeviceIds = deviceIDs
		ctxWc3, cf3 := context.WithCancel(ctx)
		defer cf3()
		if err = h.cloudClient.UpdateHub(ctxWc3, network); err != nil {
			var nfe *cloud_client.NotFoundError
			if errors.As(err, &nfe) {
				h.mu.Lock()
				h.noNetwork = true
				h.mu.Unlock()
			}
			return nil, nil, nil, nil, fmt.Errorf("update network (%s): %s", h.data.NetworkID, err)
		}
	}
	h.lastSync = time.Now()
	return recreated, createFailed, updateFailed, nil, nil
}

func (h *Handler) HasNetwork() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noNetwork
}

func (h *Handler) syncDevice(ctx context.Context, cDevices map[string]models.Device, lDevice model.Device) (string, error) {
	ch := context_hdl.New()
	defer ch.CancelAll()
	cDevice, ok := cDevices[lDevice.ID]
	if !ok {
		util.Logger.Debugf("%s create device (%s)", logPrefix, lDevice.ID)
		ncd := newCloudDevice(lDevice, "", h.attrOrigin)
		cID, err := h.cloudClient.CreateDevice(ch.Add(context.WithCancel(ctx)), ncd)
		if err == nil {
			util.Logger.Infof("%s created device (%s)", logPrefix, lDevice.ID)
			return cID, nil
		} else {
			var bre *cloud_client.BadRequestError
			if !errors.As(err, &bre) {
				return "", fmt.Errorf("create device (%s): %s", lDevice.ID, err)
			}
			util.Logger.Warningf("%s create device (%s): %s", logPrefix, lDevice.ID, err)
			util.Logger.Debugf("%s get device (%s)", logPrefix, lDevice.ID)
			cDevice, err = h.cloudClient.GetDeviceL(ch.Add(context.WithCancel(ctx)), lDevice.ID)
			if err != nil {
				return "", fmt.Errorf("get device (%s): %s", lDevice.ID, err)
			}
		}
	}
	if notEqual(cDevice, lDevice, h.attrOrigin) {
		util.Logger.Debugf("%s update device (%s)", logPrefix, lDevice.ID)
		ncd := newCloudDevice(lDevice, cDevice.Id, h.attrOrigin)
		if err := h.cloudClient.UpdateDevice(ch.Add(context.WithCancel(ctx)), ncd, h.attrOrigin); err != nil {
			return "", fmt.Errorf("update device (%s): %s", lDevice.ID, err)
		}
		util.Logger.Infof("%s updated device (%s)", logPrefix, lDevice.ID)
		return cDevice.Id, nil
	}
	return cDevice.Id, nil
}

func newCloudDevice(device model.Device, rID, attrOrigin string) models.Device {
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

func notEqual(cDevice models.Device, lDevice model.Device, attrOrigin string) bool {
	var cAttrPairs []string
	for _, attr := range cDevice.Attributes {
		if attr.Origin == attrOrigin {
			cAttrPairs = append(cAttrPairs, attr.Key+attr.Value)
		}
	}
	slices.Sort(cAttrPairs)
	var lAttrPairs []string
	for _, attr := range lDevice.Attributes {
		lAttrPairs = append(lAttrPairs, attr.Key+attr.Value)
	}
	slices.Sort(lAttrPairs)
	return genHash(cDevice.DeviceTypeId, cDevice.Name, strings.Join(cAttrPairs, "")) != genHash(lDevice.Type, lDevice.Name, strings.Join(lAttrPairs, ""))
}

func genHash(str ...string) string {
	hash := sha1.New()
	for _, s := range str {
		hash.Write([]byte(s))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

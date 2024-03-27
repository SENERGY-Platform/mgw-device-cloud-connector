package cloud_device_hdl

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

type Handler struct {
	cloudClient  cloud_client.ClientItf
	timeout      time.Duration
	wrkSpacePath string
	attrOrigin   string
	data         data
	lastSync     time.Time
	syncInterval time.Duration
	noHub        bool
	mu           sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, timeout, syncInterval time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		timeout:      timeout,
		syncInterval: syncInterval,
		wrkSpacePath: wrkSpacePath,
		attrOrigin:   attrOrigin,
	}
}

func (h *Handler) Init(ctx context.Context, hubID, hubName string) (string, error) {
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
		d.HubID = hubID
	}
	d.DefaultHubName = hubName
	if d.HubID != "" {
		ctxWt, cf := context.WithTimeout(ctx, h.timeout)
		defer cf()
		if hb, err := h.cloudClient.GetHub(ctxWt, d.HubID); err != nil {
			var nfe *cloud_client.NotFoundError
			if !errors.As(err, &nfe) {
				return "", fmt.Errorf("retreiving hub '%s' from cloud failed: %s", d.HubID, err)
			}
			util.Logger.Warningf("hub '%s' not found in cloud", d.HubID)
			d.HubID = ""
		} else {
			if deviceIDMap, err := h.getDeviceIDMap(ctx, d.DeviceIDMap, hb.DeviceIds); err == nil {
				d.DeviceIDMap = deviceIDMap
			} else {
				util.Logger.Errorf("refreshing device ID cache failed: %s", err)
			}
		}
	}
	if d.HubID == "" {
		ctxWt, cf := context.WithTimeout(ctx, h.timeout)
		defer cf()
		hID, err := h.cloudClient.CreateHub(ctxWt, models.Hub{Name: hubName})
		if err != nil {
			return "", fmt.Errorf("creating hub in cloud failed: %s", err)
		}
		d.HubID = hID
		util.Logger.Infof("created hub '%s' in cloud", hID)
	}
	if d.DeviceIDMap == nil {
		d.DeviceIDMap = make(map[string]string)
	}
	h.data = d
	return d.HubID, writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) ([]string, []string, []string, error) {
	for _, lID := range missingIDs {
		delete(h.data.DeviceIDMap, lID)
	}
	if len(newIDs)+len(changedIDs) == 0 && time.Since(h.lastSync) < h.syncInterval {
		return nil, nil, nil, nil
	}
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	hb, err := h.cloudClient.GetHub(ctxWt, h.data.HubID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noHub = true
			h.mu.Unlock()
		}
		return nil, nil, nil, fmt.Errorf("retireving hub '%s' from cloud failed: %s", h.data.HubID, err)
	}
	var createFailed []string
	syncedIDs := make(map[string]struct{})
	for _, lID := range newIDs {
		err = h.syncDevice(ctx, devices[lID])
		if err != nil {
			createFailed = append(createFailed, lID)
			continue
		}
		syncedIDs[lID] = struct{}{}
	}
	var updateFailed []string
	for _, lID := range changedIDs {
		err = h.syncDevice(ctx, devices[lID])
		if err != nil {
			updateFailed = append(updateFailed, lID)
			continue
		}
		syncedIDs[lID] = struct{}{}
	}
	hubLocalIDSet := make(map[string]struct{})
	for _, lID := range hb.DeviceLocalIds {
		hubLocalIDSet[lID] = struct{}{}
	}
	for lID, device := range devices {
		if _, ok := syncedIDs[lID]; !ok {
			if _, ok := hubLocalIDSet[lID]; !ok {
				err = h.syncDevice(ctx, device)
				if err != nil {
					createFailed = append(createFailed, lID)
					continue
				}
				syncedIDs[lID] = struct{}{}
			}
		}
	}
	for lID := range syncedIDs {
		if _, ok := hubLocalIDSet[lID]; !ok {
			hb.DeviceLocalIds = append(hb.DeviceLocalIds, lID)
		}
	}
	hb.DeviceIds = nil
	ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
	defer cf2()
	if err = h.cloudClient.UpdateHub(ctxWt2, hb); err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			h.noHub = true
			h.mu.Unlock()
		}
		return nil, nil, nil, fmt.Errorf("updating hub '%s' in cloud failed: %s", h.data.HubID, err)
	}
	h.lastSync = time.Now()
	if err = writeData(h.wrkSpacePath, h.data); err != nil {
		util.Logger.Error(err)
	}
	return createFailed, updateFailed, nil, nil
}

func (h *Handler) HasHub() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noHub
}

func (h *Handler) syncDevice(ctx context.Context, device model.Device) (err error) {
	util.Logger.Debugf("synchronising device '%s' in cloud", device.ID)
	rID, ok := h.data.DeviceIDMap[device.ID]
	var rIDNew string
	if !ok {
		rIDNew, err = h.createOrUpdateDevice(ctx, device)
	} else {
		rIDNew, err = h.updateOrCreateDevice(ctx, rID, device)
	}
	if err != nil {
		util.Logger.Error(err)
		return
	}
	if rIDNew != rID {
		util.Logger.Debugf("updating device ID cache '%s' -> '%s'", device.ID, rIDNew)
		h.data.DeviceIDMap[device.ID] = rIDNew
	}
	return
}

func (h *Handler) createOrUpdateDevice(ctx context.Context, device model.Device) (string, error) {
	ch := context_hdl.New()
	defer ch.CancelAll()
	nd := newDevice(device, "", h.attrOrigin)
	rID, err := h.cloudClient.CreateDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), nd)
	if err != nil {
		var bre *cloud_client.BadRequestError
		if !errors.As(err, &bre) {
			return "", fmt.Errorf("creating device '%s' in cloud failed: %s", device.ID, err)
		}
		d, err := h.cloudClient.GetDeviceL(ch.Add(context.WithTimeout(ctx, h.timeout)), device.ID)
		if err != nil {
			return "", fmt.Errorf("retrieving device '%s' from cloud failed: %s", device.ID, err)
		}
		rID = d.Id
		nd.Id = d.Id
		if err = h.cloudClient.UpdateDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), nd, h.attrOrigin); err != nil {
			return "", fmt.Errorf("updating device '%s' in cloud failed: %s", device.ID, err)
		}
	}
	return rID, nil
}

func (h *Handler) updateOrCreateDevice(ctx context.Context, rID string, device model.Device) (string, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	err := h.cloudClient.UpdateDevice(ctxWt, newDevice(device, rID, h.attrOrigin), h.attrOrigin)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		var fe *cloud_client.ForbiddenError
		if !errors.As(err, &nfe) && !errors.As(err, &fe) {
			return "", fmt.Errorf("updating device '%s' in cloud failed: %s", device.ID, err)
		}
		return h.createOrUpdateDevice(ctx, device)
	}
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
				device, err := h.cloudClient.GetDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), rID)
				if err != nil {
					var nfe *cloud_client.NotFoundError
					if !errors.As(err, &nfe) {
						return nil, fmt.Errorf("retrieving device '%s' from cloud failed: %s", rID, err)
					}
					continue
				}
				lID = device.LocalId
				util.Logger.Debugf("adding '%s' -> '%s' to device ID cache", lID, rID)
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

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
	hubSyncFunc  func(ctx context.Context, oldID, newID string) error
	mu           sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, timeout time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		timeout:      timeout,
		wrkSpacePath: wrkSpacePath,
		attrOrigin:   attrOrigin,
	}
}

func (h *Handler) SetHubSyncFunc(f func(ctx context.Context, oldID, newID string) error) {
	h.hubSyncFunc = f
}

func (h *Handler) GetHubID() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.data.HubID
}

func (h *Handler) Init(ctx context.Context, hubID, hubName string) error {
	if !path.IsAbs(h.wrkSpacePath) {
		return fmt.Errorf("workspace path must be absolute")
	}
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return err
	}
	d, err := readData(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if hubID != "" {
		d.HubID = hubID
	}
	d.DefaultHubName = hubName
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	var deviceIDs []string
	if d.HubID != "" {
		util.Logger.Debugf("initialising hub '%s'", d.HubID)
		hub, err := h.cloudClient.GetHub(ctxWt, d.HubID)
		if err != nil {
			var nfe *cloud_client.NotFoundError
			if !errors.As(err, &nfe) {
				return fmt.Errorf("retreiving hub failed: %s", err)
			}
			util.Logger.Warningf("hub '%s' not found", d.HubID)
			ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
			defer cf2()
			hID, err := h.cloudClient.CreateHub(ctxWt2, models.Hub{Name: hubName})
			if err != nil {
				return fmt.Errorf("creating hub failed: %s", err)
			}
			d.HubID = hID
		}
		deviceIDs = hub.DeviceIds
	} else {
		util.Logger.Debug("creating new hub")
		hID, err := h.cloudClient.CreateHub(ctxWt, models.Hub{Name: hubName})
		if err != nil {
			return fmt.Errorf("creating hub failed: %s", err)
		}
		d.HubID = hID
	}
	deviceIDMap, err := h.getDeviceIDMap(ctx, d.DeviceIDMap, deviceIDs)
	if err != nil {
		return fmt.Errorf("refreshing device ID cache failed: %s", err)
	}
	d.DeviceIDMap = deviceIDMap
	h.data = d
	return writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, changedIDs []string) ([]string, error) {
	util.Logger.Debug("synchronising devices and hub")
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	hubExists := true
	hb, err := h.cloudClient.GetHub(ctxWt, h.data.HubID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if !errors.As(err, &nfe) {
			return nil, fmt.Errorf("retireving hub '%s' from cloud failed: %s", h.data.HubID, err)
		}
		hubExists = false
		util.Logger.Warningf("hub '%s' not found", h.data.HubID)
	}
	deviceIDMap, err := h.getDeviceIDMap(ctx, h.data.DeviceIDMap, hb.DeviceIds)
	if err != nil {
		return nil, fmt.Errorf("refreshing device ID cache failed: %s", err)
	}
	h.data.DeviceIDMap = deviceIDMap
	hubLocalIDSet := make(map[string]struct{})
	for _, lID := range hb.DeviceLocalIds {
		hubLocalIDSet[lID] = struct{}{}
	}
	var failed []string
	synced := make(map[string]uint8)
	for lID, device := range devices {
		var state uint8
		if _, ok := hubLocalIDSet[lID]; !ok {
			err = h.syncDevice(ctx, device)
			if err != nil {
				failed = append(failed, lID)
				continue
			}
			state = 1
		}
		synced[lID] = state
	}
	for _, lID := range changedIDs {
		if state, ok := synced[lID]; ok && state == 0 {
			err = h.syncDevice(ctx, devices[lID])
			if err != nil {
				failed = append(failed, lID)
				continue
			}
			synced[lID] = 1
		}
	}
	for lID := range synced {
		hubLocalIDSet[lID] = struct{}{}
	}
	var hubLocalIDs []string
	for lID := range hubLocalIDSet {
		hubLocalIDs = append(hubLocalIDs, lID)
	}
	ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
	defer cf2()
	if hubExists {
		util.Logger.Debugf("synchronising hub '%s'", hb.Id)
		hb.DeviceLocalIds = hubLocalIDs
		hb.DeviceIds = nil
		if err = h.cloudClient.UpdateHub(ctxWt2, hb); err != nil {
			util.Logger.Errorf("updating hub '%s' failed: %s", hb.Id, err)
		}
	} else {
		util.Logger.Debug("creating new hub")
		oldHubID := h.data.HubID
		newHubID, err := h.cloudClient.CreateHub(ctxWt2, models.Hub{
			Name:           h.data.DefaultHubName,
			DeviceLocalIds: hubLocalIDs,
		})
		if err != nil {
			util.Logger.Errorf("creating hub failed: %s", err)
		} else {
			h.mu.Lock()
			h.data.HubID = newHubID
			h.mu.Unlock()
			if h.hubSyncFunc != nil {
				if err = h.hubSyncFunc(ctx, oldHubID, newHubID); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	if err = writeData(h.wrkSpacePath, h.data); err != nil {
		util.Logger.Error(err)
	}
	return failed, nil
}

func (h *Handler) syncDevice(ctx context.Context, device model.Device) (err error) {
	util.Logger.Debugf("synchronising device '%s'", device.ID)
	rID, ok := h.data.DeviceIDMap[device.ID]
	if !ok {
		rID, err = h.createOrUpdateDevice(ctx, device)
	} else {
		rID, err = h.updateOrCreateDevice(ctx, rID, device)
	}
	if err != nil {
		return
	}
	h.data.DeviceIDMap[device.ID] = rID
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
		if !errors.As(err, &nfe) {
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

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
	hubSyncFunc  func(oldID, newID string) error
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

func (h *Handler) SetHubSyncFunc(f func(oldID, newID string) error) {
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
		hub, err := h.cloudClient.GetHub(ctxWt, d.HubID)
		if err != nil {
			var nfe *cloud_client.NotFoundError
			if !errors.As(err, &nfe) {
				return err
			}
			ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
			defer cf2()
			d.HubID, err = h.cloudClient.CreateHub(ctxWt2, models.Hub{Name: hubName})
			if err != nil {
				return err
			}
		}
		deviceIDs = hub.DeviceIds
	} else {
		d.HubID, err = h.cloudClient.CreateHub(ctxWt, models.Hub{Name: hubName})
		if err != nil {
			return err
		}
	}
	d.DeviceIDMap, err = h.getDeviceIDMap(ctx, d.DeviceIDMap, deviceIDs)
	if err != nil {
		return err
	}
	h.data = d
	return writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, changed []string) ([]string, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	hubExists := true
	hb, err := h.cloudClient.GetHub(ctxWt, h.data.HubID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if !errors.As(err, &nfe) {
			return nil, err
		}
		hubExists = false
	}
	h.data.DeviceIDMap, err = h.getDeviceIDMap(ctx, h.data.DeviceIDMap, hb.DeviceIds)
	if err != nil {
		return nil, err
	}
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
	for _, lID := range changed {
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
		hb.DeviceLocalIds = hubLocalIDs
		hb.DeviceIds = nil
		if err = h.cloudClient.UpdateHub(ctxWt2, hb); err != nil {
			util.Logger.Error(err)
		}
	} else {
		oldHubID := h.data.HubID
		newHubID, err := h.cloudClient.CreateHub(ctxWt2, models.Hub{
			Name:           h.data.DefaultHubName,
			DeviceLocalIds: hubLocalIDs,
		})
		if err != nil {
			util.Logger.Error(err)
		} else {
			h.mu.Lock()
			h.data.HubID = newHubID
			h.mu.Unlock()
			if h.hubSyncFunc != nil {
				if err = h.hubSyncFunc(oldHubID, newHubID); err != nil {
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
			return "", err
		}
		d, err := h.cloudClient.GetDeviceL(ch.Add(context.WithTimeout(ctx, h.timeout)), device.ID)
		if err != nil {
			return "", err
		}
		rID = d.Id
		nd.Id = d.Id
		if err = h.cloudClient.UpdateDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), nd, h.attrOrigin); err != nil {
			return "", err
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
			return "", err
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
						return nil, err
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

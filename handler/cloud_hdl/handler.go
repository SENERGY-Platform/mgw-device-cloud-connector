package cloud_hdl

import (
	"context"
	"errors"
	context_hdl "github.com/SENERGY-Platform/go-service-base/context-hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Handler struct {
	cloudClient  cloud_client.ClientItf
	mqttClient   mqtt.Client
	timeout      time.Duration
	wrkSpacePath string
	attrOrigin   string
	data         data
}

func New(cloudClient cloud_client.ClientItf, mqttClient mqtt.Client, timeout time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		mqttClient:   mqttClient,
		timeout:      timeout,
		wrkSpacePath: wrkSpacePath,
		attrOrigin:   attrOrigin,
	}
}

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, changed, missing []string) ([]string, error) {
	//for _, lID := range missing {
	//	res := h.mqttClient.Unsubscribe("command/{device_id}/+")
	//	var err error
	//	if res.WaitTimeout(h.timeout) {
	//		err = res.Error()
	//	} else {
	//		err = errors.New("timeout occurred, operation may still succeed")
	//	}
	//	if err != nil {
	//		return nil, err
	//	}
	//}
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
	hubLocalIDSet := make(map[string]struct{})
	for _, lID := range hb.DeviceLocalIds {
		hubLocalIDSet[lID] = struct{}{}
	}
	var failed []string
	synced := make(map[string]uint8)
	for lID, device := range devices {
		var state uint8
		if _, ok := hubLocalIDSet[lID]; !ok {
			err = h.syncDevice(ctx, lID, device)
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
			err = h.syncDevice(ctx, lID, devices[lID])
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
		err = h.cloudClient.UpdateHub(ctxWt2, hb)
	} else {
		h.data.HubID, err = h.cloudClient.CreateHub(ctxWt2, models.Hub{
			Name:           h.data.DefaultHubName,
			DeviceLocalIds: hubLocalIDs,
		})
	}
	if err != nil {
		util.Logger.Error(err)
	}
	if err = writeData(h.wrkSpacePath, h.data); err != nil {
		util.Logger.Error(err)
	}
	return failed, nil
}

//func (h *Handler) UpdateStates(_ context.Context, deviceStates map[string]string) ([]string, error) {
//	var failed []string
//	var err error
//	for ldID, state := range deviceStates {
//		if _, ok := h.data.DeviceIDMap[ldID]; ok {
//			var res mqtt.Token
//			switch state {
//			case model.Online:
//				res = h.mqttClient.Subscribe("command/{device_id}/+", 0, nil)
//			case model.Offline, "":
//				res = h.mqttClient.Unsubscribe("command/{device_id}/+")
//			}
//			if res.WaitTimeout(h.timeout) {
//				err = res.Error()
//			} else {
//				err = errors.New("timeout occurred")
//			}
//			if err != nil {
//				failed = append(failed, ldID)
//			}
//		}
//	}
//	return failed, nil
//}

func (h *Handler) syncDevice(ctx context.Context, lID string, device model.Device) (err error) {
	rID, ok := h.data.DeviceIDMap[lID]
	if !ok {
		rID, err = h.createOrUpdate(ctx, device)
	} else {
		rID, err = h.updateOrCreate(ctx, rID, device)
	}
	if err != nil {
		return
	}
	h.data.DeviceIDMap[lID] = rID
	return
}

func (h *Handler) createOrUpdate(ctx context.Context, device model.Device) (string, error) {
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

func (h *Handler) updateOrCreate(ctx context.Context, rID string, device model.Device) (string, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	err := h.cloudClient.UpdateDevice(ctxWt, newDevice(device, rID, h.attrOrigin), h.attrOrigin)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if !errors.As(err, &nfe) {
			return "", err
		}
		return h.createOrUpdate(ctx, device)
	}
	return rID, err
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

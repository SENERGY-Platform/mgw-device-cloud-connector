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
	"time"
)

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) ([]string, []string, []string, []string, error) {
	if len(newIDs)+len(changedIDs) == 0 && time.Since(h.lastSync) < h.syncInterval {
		return nil, nil, nil, nil, nil
	}
	if len(newIDs)+len(changedIDs) > 0 {
		util.Logger.Info(logPrefix, " begin devices and network sync")
	} else {
		util.Logger.Debug(logPrefix, " begin periodic devices and network sync")
	}
	network, err := h.getNetwork(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cloudDevices, err := h.getCloudDevices(ctx, network.DeviceIds)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	syncedIDs := make(map[string]string)
	var createFailed []string
	for _, lID := range newIDs {
		cID, err := h.syncDevice(ctx, cloudDevices, devices[lID])
		if err != nil {
			createFailed = append(createFailed, lID)
		}
		syncedIDs[lID] = cID
	}
	var updateFailed []string
	for _, lID := range changedIDs {
		cID, err := h.syncDevice(ctx, cloudDevices, devices[lID])
		if err != nil {
			updateFailed = append(updateFailed, lID)
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
		if cID != "" {
			networkDeviceIDSet[cID] = struct{}{}
		}
	}
	if lenOld != len(networkDeviceIDSet) {
		var deviceIDs []string
		for id := range networkDeviceIDSet {
			deviceIDs = append(deviceIDs, id)
		}
		network.DeviceIds = deviceIDs
		if err = h.updateNetwork(ctx, network); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	h.lastSync = time.Now()
	return recreated, createFailed, updateFailed, nil, nil
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
				util.Logger.Errorf("%s create device (%s): %s", logPrefix, lDevice.ID, err)
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
			util.Logger.Errorf("%s update device (%s): %s", logPrefix, lDevice.ID, err)
			return "", fmt.Errorf("update device (%s): %s", lDevice.ID, err)
		}
		util.Logger.Infof("%s updated device (%s)", logPrefix, lDevice.ID)
		return cDevice.Id, nil
	}
	return cDevice.Id, nil
}

func (h *Handler) getNetwork(ctx context.Context) (models.Hub, error) {
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s get network (%s)", logPrefix, h.data.NetworkID)
	network, err := h.cloudClient.GetHub(ctxWc, h.data.NetworkID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.noNetwork = true
		}
		return models.Hub{}, fmt.Errorf("get network (%s): %s", h.data.NetworkID, err)
	}
	if network.OwnerId != h.userID {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.noNetwork = true
		return models.Hub{}, fmt.Errorf("get network (%s): invalid user ID", h.data.NetworkID)
	}
	return network, nil
}

func (h *Handler) updateNetwork(ctx context.Context, network models.Hub) error {
	util.Logger.Infof("%s update network (%s)", logPrefix, h.data.NetworkID)
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	if err := h.cloudClient.UpdateHub(ctxWc, network); err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.noNetwork = true
		}
		return fmt.Errorf("update network (%s): %s", h.data.NetworkID, err)
	}
	return nil
}

func (h *Handler) getCloudDevices(ctx context.Context, cDeviceIDs []string) (map[string]models.Device, error) {
	cloudDevices := make(map[string]models.Device)
	if len(cDeviceIDs) > 0 {
		ctxWc, cf := context.WithCancel(ctx)
		defer cf()
		devicesList, err := h.cloudClient.GetDevices(ctxWc, cDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("get devices: %s", err)
		}
		for _, device := range devicesList {
			if device.OwnerId != h.userID {
				util.Logger.Warningf("%s get devices: device (%s) invalid user ID", logPrefix, device.Id)
				continue
			}
			cloudDevices[device.LocalId] = device
		}
	}
	return cloudDevices, nil
}

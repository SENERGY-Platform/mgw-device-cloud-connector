package cloud_hdl

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	context_hdl "github.com/SENERGY-Platform/mgw-go-service-base/context-hdl"
	"github.com/SENERGY-Platform/models/go/models"
	"strings"
	"time"
)

func (h *Handler) Sync(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs []string) {
	if h.syncRequired(newIDs, changedIDs) {
		err := h.syncDevAndNet(ctx, devices)
		if err != nil {
			util.Logger.Errorf("%s synchronisation: %s", logPrefix, err)
			return
		}
	}
	h.syncStates(ctx, devices, missingIDs, onlineIDs, offlineIDs)
}

func (h *Handler) syncStates(ctx context.Context, devices map[string]model.Device, missingIDs, onlineIDs, offlineIDs []string) {
	if h.stateSyncFunc != nil {
		devices2 := make(map[string]model.Device)
		for lID, lDevice := range devices {
			if _, k := h.syncedIDs[lID]; k {
				devices2[lID] = lDevice
			}
		}
		h.stateSyncFunc(ctx, devices2, missingIDs, onlineIDs, offlineIDs)
	}
}

func (h *Handler) syncRequired(newIDs, changedIDs []string) bool {
	if len(newIDs)+len(changedIDs) > 0 || !h.syncOK {
		util.Logger.Info(logPrefix, " synchronisation")
		return true
	} else {
		if time.Since(h.lastSync) > h.syncInterval {
			util.Logger.Debug(logPrefix, " periodic synchronisation")
			return true
		}
	}
	return false
}

func (h *Handler) syncDevAndNet(ctx context.Context, devices map[string]model.Device) error {
	var err error
	defer func() {
		if err != nil {
			h.syncOK = false
		}
	}()
	devSyncAllowed, err := h.checkAccPols(ctx)
	if err != nil {
		return err
	}
	network, err := h.getNetwork(ctx)
	if err != nil {
		return err
	}
	cloudDevices, err := h.getCloudDevices(ctx, network.DeviceIds)
	if err != nil {
		return err
	}
	var syncedIDs map[string]string
	var ok bool
	if devSyncAllowed {
		syncedIDs, ok, err = h.syncDevices(ctx, devices, cloudDevices)
		if err != nil {
			return err
		}
	} else {
		syncedIDs, ok, err = h.syncDeviceIDs(ctx, devices, cloudDevices)
		if err != nil {
			return err
		}
	}
	if err = h.syncNetwork(ctx, network, syncedIDs); err != nil {
		return err
	}
	h.syncOK = ok
	h.syncedIDs = syncedIDs
	h.lastSync = time.Now()
	return nil
}

func (h *Handler) syncNetwork(ctx context.Context, network models.Hub, syncedIDs map[string]string) error {
	networkDeviceIDSet := make(map[string]struct{})
	for _, id := range network.DeviceIds {
		networkDeviceIDSet[id] = struct{}{}
	}
	lenOld := len(networkDeviceIDSet)
	for _, cID := range syncedIDs {
		networkDeviceIDSet[cID] = struct{}{}
	}
	if lenOld != len(networkDeviceIDSet) {
		util.Logger.Infof("%s update network (%s)", logPrefix, h.networkID)
		var deviceIDs []string
		for id := range networkDeviceIDSet {
			deviceIDs = append(deviceIDs, id)
		}
		network.DeviceIds = deviceIDs
		ctxWc, cf := context.WithCancel(ctx)
		defer cf()
		if err := h.cloudClient.UpdateHub(ctxWc, network); err != nil {
			var nfe *cloud_client.NotFoundError
			if errors.As(err, &nfe) {
				h.mu.Lock()
				defer h.mu.Unlock()
				h.noNetwork = true
			}
			return fmt.Errorf("update network (%s): %s", h.networkID, err)
		}
	}
	return nil
}

func (h *Handler) syncDevices(ctx context.Context, lDevices map[string]model.Device, cDevices map[string]models.Device) (map[string]string, bool, error) {
	ch := context_hdl.New()
	defer ch.CancelAll()
	var missingIDs []string
	for lID := range lDevices {
		if _, ok := cDevices[lID]; !ok {
			missingIDs = append(missingIDs, lID)
		}
	}
	cDevices2, err := h.getCloudDevicesL(ctx, missingIDs)
	if err != nil {
		return nil, false, err
	}
	for lID, cDevice := range cDevices2 {
		cDevices[lID] = cDevice
	}
	syncedIDs := make(map[string]string)
	syncOk := true
	for lID, lDevice := range lDevices {
		cDevice, ok := cDevices[lID]
		if !ok {
			util.Logger.Debugf("%s create device (%s)", logPrefix, lID)
			cID, err := h.cloudClient.CreateDevice(ch.Add(context.WithCancel(ctx)), newCloudDevice(lDevice, "", h.attrOrigin))
			if err != nil {
				util.Logger.Errorf("%s create device (%s): %s", logPrefix, lID, err)
				syncOk = false
				continue
			}
			util.Logger.Infof("%s created device (%s)", logPrefix, lID)
			syncedIDs[lID] = cID
		} else {
			if notEqual(cDevice, lDevice, h.attrOrigin) {
				util.Logger.Debugf("%s update device (%s)", logPrefix, lID)
				if err = h.cloudClient.UpdateDevice(ch.Add(context.WithCancel(ctx)), newCloudDevice(lDevice, cDevice.Id, h.attrOrigin), h.attrOrigin); err != nil {
					util.Logger.Errorf("%s update device (%s): %s", logPrefix, lID, err)
					syncOk = false
				} else {
					util.Logger.Infof("%s updated device (%s)", logPrefix, lID)
				}
			}
			syncedIDs[lID] = cDevice.Id
		}
	}
	return syncedIDs, syncOk, nil
}

func (h *Handler) syncDeviceIDs(ctx context.Context, lDevices map[string]model.Device, cDevices map[string]models.Device) (map[string]string, bool, error) {
	syncedIDs := make(map[string]string)
	var missingIDs []string
	for lID := range lDevices {
		if cDevice, ok := cDevices[lID]; ok {
			syncedIDs[lID] = cDevice.Id
		} else {
			missingIDs = append(missingIDs, lID)
		}
	}
	cDevices2, err := h.getCloudDevicesL(ctx, missingIDs)
	if err != nil {
		return nil, false, err
	}
	for _, lID := range missingIDs {
		if cDevice, ok := cDevices2[lID]; ok {
			syncedIDs[lID] = cDevice.Id
		}
	}
	return syncedIDs, true, nil
}

func (h *Handler) getNetwork(ctx context.Context) (models.Hub, error) {
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s get network (%s)", logPrefix, h.networkID)
	network, err := h.cloudClient.GetHub(ctxWc, h.networkID)
	if err != nil {
		var nfe *cloud_client.NotFoundError
		if errors.As(err, &nfe) {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.noNetwork = true
		}
		return models.Hub{}, fmt.Errorf("get network (%s): %s", h.networkID, err)
	}
	if network.OwnerId != h.userID {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.noNetwork = true
		return models.Hub{}, fmt.Errorf("get network (%s): invalid user ID", h.networkID)
	}
	return network, nil
}

func (h *Handler) getCloudDevices(ctx context.Context, cDeviceIDs []string) (map[string]models.Device, error) {
	return h.getCloudDevs(ctx, cDeviceIDs, h.cloudClient.GetDevices)
}

func (h *Handler) getCloudDevicesL(ctx context.Context, lDeviceIDs []string) (map[string]models.Device, error) {
	return h.getCloudDevs(ctx, lDeviceIDs, h.cloudClient.GetDevicesL)
}

func (h *Handler) getCloudDevs(ctx context.Context, ids []string, gf func(context.Context, []string) ([]models.Device, error)) (map[string]models.Device, error) {
	cloudDevices := make(map[string]models.Device)
	if len(ids) > 0 {
		util.Logger.Debugf("%s get devices (%s)", logPrefix, strings.Join(ids, ", "))
		ctxWc, cf := context.WithCancel(ctx)
		defer cf()
		devicesList, err := gf(ctxWc, ids)
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

func (h *Handler) checkAccPols(ctx context.Context) (bool, error) {
	msgStr := "check access policies"
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	util.Logger.Debugf("%s %s", logPrefix, msgStr)
	accPol, err := h.cloudClient.GetAccessPolicies(ctxWc)
	if err != nil {
		return false, fmt.Errorf("%s: %s", msgStr, err)
	}
	if !accPol.Hubs().Read() {
		return false, errors.New(msgStr + " (network): read not allowed")
	}
	if !accPol.Hubs().Update() {
		return false, errors.New(msgStr + " (network): update not allowed")
	}
	if !accPol.Devices().Read() {
		return false, errors.New(msgStr + " (devices cloud id): read not allowed")
	}
	if !accPol.DevicesL().Read() {
		return false, errors.New(msgStr + " (devices local id): read not allowed")
	}
	if !accPol.Devices().Create() {
		util.Logger.Debugf("%s %s (devices cloud id): create not allowed", logPrefix, msgStr)
		return false, nil
	}
	if !accPol.Devices().Update() {
		util.Logger.Debugf("%s %s (devices cloud id): update not allowed", logPrefix, msgStr)
		return false, nil
	}
	return true, nil
}

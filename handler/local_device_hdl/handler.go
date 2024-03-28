package local_device_hdl

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"slices"
	"sync"
	"time"
)

type device struct {
	model.Device
	Hash string
}

type Handler struct {
	dmClient            dm_client.ClientItf
	timeout             time.Duration
	queryInterval       time.Duration
	idPrefix            string
	sChan               chan bool
	dChan               chan struct{}
	devices             map[string]device
	running             bool
	loopMu              sync.RWMutex
	ctx                 context.Context
	cf                  context.CancelFunc
	deviceSyncFunc      func(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) (recreated, createFailed, updateFailed, deleteFailed []string, err error)
	deviceStateSyncFunc func(ctx context.Context, devices map[string]model.Device, isOnlineIDs, isOfflineIDs, isOnlineAgainIDs []string) (failed []string, err error)
}

func New(ctx context.Context, dmClient dm_client.ClientItf, timeout, queryInterval time.Duration, idPrefix string) *Handler {
	ctx2, cf := context.WithCancel(ctx)
	return &Handler{
		dmClient:      dmClient,
		timeout:       timeout,
		queryInterval: queryInterval,
		idPrefix:      idPrefix,
		sChan:         make(chan bool),
		dChan:         make(chan struct{}),
		ctx:           ctx2,
		cf:            cf,
	}
}

func (h *Handler) Start() {
	go h.run()
}

func (h *Handler) Running() bool {
	h.loopMu.RLock()
	defer h.loopMu.RUnlock()
	return h.running
}

func (h *Handler) Stop() {
	h.cf()
	<-h.dChan
}

func (h *Handler) SetDeviceSyncFunc(f func(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs []string) (recreated, createFailed, updateFailed, deleteFailed []string, err error)) {
	h.deviceSyncFunc = f
}

func (h *Handler) SetDeviceStateSyncFunc(f func(ctx context.Context, devices map[string]model.Device, isOnlineIDs, isOfflineIDs, isOnlineAgainIDs []string) (failed []string, err error)) {
	h.deviceStateSyncFunc = f
}

func (h *Handler) run() {
	h.loopMu.Lock()
	h.running = true
	h.loopMu.Unlock()
	timer := time.NewTimer(h.queryInterval)
	loop := true
	var err error
	for loop {
		select {
		case <-timer.C:
			err = h.RefreshDevices(h.ctx)
			if err != nil {
				util.Logger.Error(err)
			}
			timer.Reset(h.queryInterval)
		case <-h.ctx.Done():
			loop = false
			break
		}
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	h.loopMu.Lock()
	h.running = false
	h.loopMu.Unlock()
	h.dChan <- struct{}{}
}

func (h *Handler) RefreshDevices(ctx context.Context) error {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWt)
	if err != nil {
		return fmt.Errorf("retreiving local devices failed: %s", err)
	}
	devices := make(map[string]device)
	for id, dmDevice := range dmDevices {
		devices[h.idPrefix+id] = newDevice(h.idPrefix+id, dmDevice)
	}
	var recreatedIDs []string
	if h.deviceSyncFunc != nil {
		newIDs, changedIDs, missingIDs, deviceMap := h.diffDevices(devices)
		recreated, createFailed, updateFailed, deleteFailed, err := h.deviceSyncFunc(ctx, deviceMap, newIDs, changedIDs, missingIDs)
		if err != nil {
			return fmt.Errorf("synchronising devices failed: %s", err)
		}
		recreatedIDs = recreated
		for _, id := range createFailed {
			delete(devices, id)
		}
		for _, id := range updateFailed {
			if d, ok := h.devices[id]; ok {
				devices[id] = d
			}
		}
		for _, id := range deleteFailed {
			if d, ok := h.devices[id]; ok {
				devices[id] = d
			}
		}
	}
	if h.deviceStateSyncFunc != nil {
		isOnlineIDs, isOfflineIDs, isOnlineAgainIDs, deviceMap := h.diffDeviceStates(devices, recreatedIDs)
		failed, err := h.deviceStateSyncFunc(ctx, deviceMap, isOnlineIDs, isOfflineIDs, isOnlineAgainIDs)
		if err != nil {
			return fmt.Errorf("synchronising device states failed: %s", err)
		}
		for _, id := range failed {
			d := devices[id]
			od, ok := h.devices[id]
			if ok {
				if d.State != od.State {
					d.State = od.State
				} else {
					d.State = model.Offline
				}
			} else {
				d.State = ""
			}
			devices[id] = d
		}
	}
	h.devices = devices
	return nil
}

func (h *Handler) diffDevices(devices map[string]device) (newIDs, changedIDs, missingIDs []string, deviceMap map[string]model.Device) {
	deviceMap = make(map[string]model.Device)
	for id, queriedDevice := range devices {
		deviceMap[id] = queriedDevice.Device
		storedDevice, ok := h.devices[id]
		if ok {
			if storedDevice.Hash != queriedDevice.Hash {
				changedIDs = append(changedIDs, id)
			}
		} else {
			newIDs = append(newIDs, id)
		}
	}
	for id := range h.devices {
		if _, ok := devices[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}
	return
}

func (h *Handler) diffDeviceStates(devices map[string]device, recreated []string) (isOnlineIDs, isOfflineIDs, isOnlineAgainIDs []string, deviceMap map[string]model.Device) {
	deviceMap = make(map[string]model.Device)
	for id, queriedDevice := range devices {
		deviceMap[id] = queriedDevice.Device
		storedDevice, ok := h.devices[id]
		if ok {
			if storedDevice.State != queriedDevice.State {
				switch queriedDevice.State {
				case model.Online:
					isOnlineIDs = append(isOnlineIDs, id)
				case model.Offline, "":
					isOfflineIDs = append(isOfflineIDs, id)
				}
			}
		} else {
			if queriedDevice.State == model.Online {
				isOnlineIDs = append(isOnlineIDs, id)
			}
		}
	}
	for id := range h.devices {
		if _, ok := devices[id]; !ok {
			isOfflineIDs = append(isOfflineIDs, id)
		}
	}
	for _, id := range recreated {
		if d, ok := devices[id]; ok && d.State == model.Online {
			isOnlineAgainIDs = append(isOnlineAgainIDs, id)
		}
	}
	return
}

func newDevice(id string, d dm_client.Device) device {
	var attrPairs []string
	var attributes []model.Attribute
	for _, attr := range d.Attributes {
		attrPairs = append(attrPairs, attr.Key+attr.Value)
		attributes = append(attributes, model.Attribute{
			Key:   attr.Key,
			Value: attr.Value,
		})
	}
	slices.Sort(attrPairs)
	return device{
		Device: model.Device{
			ID:         id,
			Name:       d.Name,
			State:      d.State,
			Type:       d.Type,
			Attributes: attributes,
		},
		Hash: genHash(d.Type, d.Name, genHash(attrPairs...)),
	}
}

func genHash(str ...string) string {
	hash := sha1.New()
	for _, s := range str {
		hash.Write([]byte(s))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

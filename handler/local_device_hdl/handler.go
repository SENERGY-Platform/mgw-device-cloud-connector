package local_device_hdl

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
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
	dmClient      dm_client.ClientItf
	timeout       time.Duration
	queryInterval time.Duration
	idPrefix      string
	sChan         chan bool
	devices       map[string]device
	running       bool
	loopMu        sync.RWMutex
	mu            sync.RWMutex
	syncFunc      func(ctx context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error)
	missingFunc   func(ctx context.Context, missingIDs []string) error
	stateFunc     func(ctx context.Context, deviceStates map[string]string) (failed []string, err error)
}

func New(dmClient dm_client.ClientItf, timeout, queryInterval time.Duration, idPrefix string) *Handler {
	return &Handler{
		dmClient:      dmClient,
		timeout:       timeout,
		queryInterval: queryInterval,
		idPrefix:      idPrefix,
		sChan:         make(chan bool, 1),
	}
}

func (h *Handler) Start() {
	h.loopMu.Lock()
	if !h.running {
		go h.run()
		h.running = true
	}
	h.loopMu.Unlock()
}

func (h *Handler) Running() bool {
	h.loopMu.RLock()
	defer h.loopMu.RUnlock()
	return h.running
}

func (h *Handler) Stop() {
	h.loopMu.Lock()
	if h.running {
		h.sChan <- true
	}
	h.loopMu.Unlock()
}

func (h *Handler) SetSyncFunc(f func(ctx context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error)) {
	h.syncFunc = f
}

func (h *Handler) SetMissingFunc(f func(ctx context.Context, missingIDs []string) error) {
	h.missingFunc = f
}

func (h *Handler) SetStateFunc(f func(ctx context.Context, deviceStates map[string]string) (failed []string, err error)) {
	h.stateFunc = f
}

func (h *Handler) GetDevices() map[string]model.Device {
	h.mu.RLock()
	defer h.mu.RUnlock()
	devices := make(map[string]model.Device)
	for id, d := range h.devices {
		devices[id] = d.Device
	}
	return devices
}

func (h *Handler) run() {
	ticker := time.NewTicker(h.queryInterval)
	defer func() {
		ticker.Stop()
		h.loopMu.Lock()
		h.running = false
		h.loopMu.Unlock()
	}()
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	var err error
	for {
		select {
		case <-h.sChan:
			return
		case <-ticker.C:
			err = h.RefreshDevices(ctx)
			if err != nil {
				util.Logger.Errorf("refreshing local devices failed: %s", err)
			}
		}
	}
}

func (h *Handler) RefreshDevices(ctx context.Context) error {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWt)
	if err != nil {
		return err
	}
	devices := make(map[string]device)
	for id, dmDevice := range dmDevices {
		devices[h.idPrefix+id] = newDevice(h.idPrefix+id, dmDevice)
	}
	changedIDs, missingIDs, deviceMap, deviceStates := h.diffDevices(devices)
	if h.syncFunc != nil {
		failed, err := h.syncFunc(ctx, deviceMap, changedIDs)
		if err != nil {
			return err
		}
		for _, id := range failed {
			delete(devices, id)
		}
	}
	if h.missingFunc != nil && len(missingIDs) > 0 {
		err = h.missingFunc(ctx, missingIDs)
		if err != nil {
			return err
		}
	}
	if h.stateFunc != nil && len(deviceStates) > 0 {
		ds := make(map[string]string)
		for id, s := range deviceStates {
			ds[id] = s[1]
		}
		failed, err := h.stateFunc(ctx, ds)
		if err != nil {
			return err
		}
		for _, id := range failed {
			if s, ok := deviceStates[id]; ok {
				d := devices[id]
				d.State = s[0]
				devices[id] = d
			}
		}
	}
	h.mu.Lock()
	h.devices = devices
	h.mu.Unlock()
	return nil
}

func (h *Handler) diffDevices(devices map[string]device) (changedIDs, missingIDs []string, deviceMap map[string]model.Device, states map[string][2]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	deviceMap = make(map[string]model.Device)
	states = make(map[string][2]string)
	for id, queriedDevice := range devices {
		deviceMap[id] = queriedDevice.Device
		storedDevice, ok := h.devices[id]
		if ok {
			if storedDevice.Hash != queriedDevice.Hash {
				changedIDs = append(changedIDs, id)
			}
		}
		if storedDevice.State != queriedDevice.State {
			states[id] = [2]string{storedDevice.State, queriedDevice.State}
		}
	}
	for id := range h.devices {
		if _, ok := devices[id]; !ok {
			missingIDs = append(missingIDs, id)
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

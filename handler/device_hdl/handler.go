package device_hdl

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
	devices       map[string]device
	sChan         chan bool
	running       bool
	loopMu        sync.RWMutex
	mu            sync.RWMutex
	syncCbk       func(newDevices, changedDevices []model.Device, missingDevices []string, deviceStates map[string]string) error
}

func New(dmClient dm_client.ClientItf, timeout, queryInterval time.Duration) *Handler {
	return &Handler{
		dmClient:      dmClient,
		timeout:       timeout,
		queryInterval: queryInterval,
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
		h.running = false
	}
	h.loopMu.Unlock()
}

func (h *Handler) SetSyncCallback(f func(newDevices, changedDevices []model.Device, missingDevices []string, deviceStates map[string]string) error) {
	h.syncCbk = f
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
	defer ticker.Stop()
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	var err error
	for {
		select {
		case <-h.sChan:
			return
		case <-ticker.C:
			err = h.refreshDevices(ctx)
			if err != nil {
				util.Logger.Errorf("refreshing devices failed: %s", err)
			}
		}
	}
}

func (h *Handler) refreshDevices(ctx context.Context) error {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWt)
	if err != nil {
		return err
	}
	devices := make(map[string]device)
	for id, dmDevice := range dmDevices {
		devices[id] = newDevice(id, dmDevice)
	}
	newDevices, changedDevices, missingDevices, deviceStates := h.diffDevices(devices)
	if h.syncCbk != nil && len(newDevices)+len(changedDevices)+len(missingDevices)+len(deviceStates) > 0 {
		err = h.syncCbk(newDevices, changedDevices, missingDevices, deviceStates)
		if err != nil {
			return err
		}
	}
	h.mu.Lock()
	h.devices = devices
	h.mu.Unlock()
	return nil
}

func (h *Handler) diffDevices(devices map[string]device) (newDevices, changedDevices []model.Device, missingDevices []string, states map[string]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, d1 := range devices {
		d2, ok := h.devices[id]
		if !ok {
			newDevices = append(newDevices, d1.Device)
		} else {
			if d1.Hash != d2.Hash {
				changedDevices = append(changedDevices, d1.Device)
			}
		}
		if d1.State != d2.State {
			states[id] = d1.State
		}
	}
	for id := range h.devices {
		if _, ok := devices[id]; !ok {
			missingDevices = append(missingDevices, id)
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

package device_hdl

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"slices"
	"sync"
	"time"
)

type Handler struct {
	dmClient          dm_client.ClientItf
	timeout           time.Duration
	queryInterval     time.Duration
	devices           map[string]model.Device
	sChan             chan bool
	running           bool
	loopMu            sync.RWMutex
	apiMu             sync.RWMutex
	newDevicesCbk     func(devices []model.Device) error
	changedDevicesCbk func(devices []model.Device) error
	missingDevicesCbk func(ids []string) error
}

func New(dmClient dm_client.ClientItf, timeout, queryInterval time.Duration) *Handler {
	return &Handler{
		dmClient:      dmClient,
		timeout:       timeout,
		queryInterval: queryInterval,
		sChan:         make(chan bool, 1),
	}
}

func (h *Handler) Start() error {
	h.loopMu.Lock()
	defer h.loopMu.Unlock()
	if h.running {
		return errors.New("already running")
	}
	go h.run()
	h.running = true
	return nil
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

func (h *Handler) SetNewDevicesCallback(f func(devices []model.Device) error) {
	h.newDevicesCbk = f
}

func (h *Handler) SetChangedDevicesCallback(f func(devices []model.Device) error) {
	h.changedDevicesCbk = f
}

func (h *Handler) SetMissingDevicesCallback(f func(ids []string) error) {
	h.missingDevicesCbk = f
}

func (h *Handler) GetDevices() map[string]model.Device {
	h.apiMu.RLock()
	defer h.apiMu.RUnlock()
	devices := make(map[string]model.Device)
	for id, device := range h.devices {
		devices[id] = device
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
	h.apiMu.Lock()
	defer h.apiMu.Unlock()
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWt)
	if err != nil {
		return err
	}
	devices := make(map[string]model.Device)
	for id, dmDevice := range dmDevices {
		devices[id] = newDevice(id, dmDevice)
	}
	newDevices, changedDevices, missingDevices := h.diffDevices(devices)
	if h.newDevicesCbk != nil && len(newDevices) > 0 {
		err = h.newDevicesCbk(newDevices)
		if err != nil {
			return err
		}
	}
	if h.changedDevicesCbk != nil && len(changedDevices) > 0 {
		err = h.changedDevicesCbk(changedDevices)
		if err != nil {
			return err
		}
	}
	if h.missingDevicesCbk != nil && len(missingDevices) > 0 {
		err = h.missingDevicesCbk(missingDevices)
		if err != nil {
			return err
		}
	}
	h.devices = devices
	return nil
}

func (h *Handler) diffDevices(devices map[string]model.Device) (newDevices, changedDevices []model.Device, missingDevices []string) {
	for id, d1 := range devices {
		d2, ok := h.devices[id]
		if !ok {
			newDevices = append(newDevices, d1)
		} else {
			if d1.Hash != d2.Hash {
				d1.LastState = d2.State
				changedDevices = append(changedDevices, d1)
			}
		}
	}
	for id := range h.devices {
		if _, ok := devices[id]; !ok {
			missingDevices = append(missingDevices, id)
		}
	}
	return
}

func newDevice(id string, d dm_client.Device) model.Device {
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
	return model.Device{
		ID:         id,
		Name:       d.Name,
		State:      d.State,
		Type:       d.Type,
		Hash:       genHash(d.Type, d.State, d.Name, d.ModuleID, genHash(attrPairs...)),
		Attributes: attributes,
	}
}

func genHash(str ...string) string {
	hash := sha1.New()
	for _, s := range str {
		hash.Write([]byte(s))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

/*

mqtt connect to cloud
	subscribe command topics if device online



*/

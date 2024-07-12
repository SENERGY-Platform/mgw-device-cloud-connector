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
	"strings"
	"sync"
	"time"
)

const logPrefix = "[local-device-hdl]"

type device struct {
	model.Device
	Hash string
}

type Handler struct {
	dmClient       dm_client.ClientItf
	queryInterval  time.Duration
	idPrefix       string
	dChan          chan struct{}
	devices        map[string]device
	running        bool
	loopMu         sync.RWMutex
	ctx            context.Context
	cf             context.CancelFunc
	deviceSyncFunc func(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs []string)
}

func New(ctx context.Context, dmClient dm_client.ClientItf, queryInterval time.Duration, idPrefix string) *Handler {
	ctx2, cf := context.WithCancel(ctx)
	return &Handler{
		dmClient:      dmClient,
		queryInterval: queryInterval,
		idPrefix:      idPrefix,
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

func (h *Handler) SetDeviceSyncFunc(f func(ctx context.Context, devices map[string]model.Device, newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs []string)) {
	h.deviceSyncFunc = f
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
			err = h.refreshDevices(h.ctx)
			if err != nil {
				util.Logger.Errorf("%s %s", logPrefix, err)
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

func (h *Handler) refreshDevices(ctx context.Context) error {
	util.Logger.Debug(logPrefix, " get devices")
	ctxWc, cf := context.WithCancel(ctx)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWc)
	if err != nil {
		return fmt.Errorf("get devices: %s", err)
	}
	util.Logger.Debugf("%s got devices (%d)", logPrefix, len(dmDevices))
	devices := make(map[string]device)
	for id, dmDevice := range dmDevices {
		devices[h.idPrefix+id] = newDevice(h.idPrefix+id, dmDevice)
	}
	newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs := h.diffDevices(devices)
	if len(newIDs) > 0 {
		util.Logger.Infof("%s new devices (%s)", logPrefix, strings.Join(newIDs, ", "))
	}
	if len(changedIDs) > 0 {
		util.Logger.Infof("%s changed devices (%s)", logPrefix, strings.Join(changedIDs, ", "))
	}
	if len(missingIDs) > 0 {
		util.Logger.Infof("%s missing devices (%s)", logPrefix, strings.Join(missingIDs, ", "))
	}
	if len(onlineIDs) > 0 {
		util.Logger.Infof("%s online devices (%s)", logPrefix, strings.Join(onlineIDs, ", "))
	}
	if len(offlineIDs) > 0 {
		util.Logger.Infof("%s offline devices (%s)", logPrefix, strings.Join(offlineIDs, ", "))
	}
	deviceMap := make(map[string]model.Device)
	for id, dev := range devices {
		deviceMap[id] = dev.Device
	}
	if h.deviceSyncFunc != nil {
		h.deviceSyncFunc(ctx, deviceMap, newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs)
	}
	h.devices = devices
	return nil
}

func (h *Handler) diffDevices(devices map[string]device) (newIDs, changedIDs, missingIDs, onlineIDs, offlineIDs []string) {
	for id, queriedDevice := range devices {
		storedDevice, ok := h.devices[id]
		if ok {
			if storedDevice.Hash != queriedDevice.Hash {
				changedIDs = append(changedIDs, id)
			}
			if storedDevice.State != queriedDevice.State {
				switch queriedDevice.State {
				case model.Online:
					onlineIDs = append(onlineIDs, id)
				case model.Offline, "":
					offlineIDs = append(offlineIDs, id)
				}
			}
		} else {
			newIDs = append(newIDs, id)
			if queriedDevice.State == model.Online {
				onlineIDs = append(onlineIDs, id)
			}
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

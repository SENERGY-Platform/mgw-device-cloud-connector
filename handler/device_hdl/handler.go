package device_hdl

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
	"sync"
	"time"
)

type Handler struct {
	dmClient      dm_client.ClientItf
	timeout       time.Duration
	queryInterval time.Duration
	devices       map[string]device
	sChan         chan bool
	running       bool
	mu            sync.RWMutex
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
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		return errors.New("already running")
	}
	go h.run()
	h.running = true
	return nil
}

func (h *Handler) Running() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

func (h *Handler) Stop() {
	h.mu.Lock()
	if h.running {
		h.sChan <- true
		h.running = false
	}
	h.mu.Unlock()
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
			fmt.Println("stop")
			return
		case <-ticker.C:
			err = h.handleDevices(ctx)
			if err != nil {
				util.Logger.Error(err)
			}
		}
	}
}

func (h *Handler) handleDevices(ctx context.Context) error {
	panic("not implemented")
}

func (h *Handler) getDevices(ctx context.Context) (map[string]device, error) {
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	dmDevices, err := h.dmClient.GetDevices(ctxWt)
	if err != nil {
		return nil, err
	}
	devices := make(map[string]device)
	for id, d := range dmDevices {
		devices[id] = newDevice(id, d)
	}
	return devices, nil
}

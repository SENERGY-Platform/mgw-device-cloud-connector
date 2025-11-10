package cloud_hdl

import (
	"context"
	"sync"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
)

const logPrefix = "[cloud-hdl]"

type Handler struct {
	cloudClient       cloud_client.ClientItf
	certManagerClient handler.CertManagerClient
	attrOrigin        string
	userID            string
	networkID         string
	syncOK            bool
	syncedIDs         map[string]string
	lastSync          time.Time
	syncInterval      time.Duration
	noNetwork         bool
	stateSyncFunc     func(ctx context.Context, devices map[string]model.Device, missingIDs, onlineIDs, offlineIDs []string)
	mu                sync.RWMutex
	cloudDevices      map[string]models.Device
	muCloudDevices    sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, certManagerClient handler.CertManagerClient, syncInterval time.Duration, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:       cloudClient,
		certManagerClient: certManagerClient,
		syncInterval:      syncInterval,
		attrOrigin:        attrOrigin,
	}
}

func (h *Handler) SetDeviceStateSyncFunc(f func(ctx context.Context, devices map[string]model.Device, missingIDs, onlineIDs, offlineIDs []string)) {
	h.stateSyncFunc = f
}

func (h *Handler) HasNetwork() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noNetwork
}

// GetCloudDevice returns the cloud device for the given local ID.
// The function operates on a best-effort basis. It returns nil if the device is not found,
// which might occur when no sync has happened yet.
func (h *Handler) GetCloudDevice(localId string) *models.Device {
	h.muCloudDevices.RLock()
	defer h.muCloudDevices.RUnlock()
	device, ok := h.cloudDevices[localId]
	if !ok {
		return nil
	}
	return &device
}

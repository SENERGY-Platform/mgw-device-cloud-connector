package cloud_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"sync"
	"time"
)

const logPrefix = "[cloud-hdl]"

type Handler struct {
	cloudClient     cloud_client.ClientItf
	subjectProvider handler.SubjectProvider
	wrkSpacePath    string
	attrOrigin      string
	userID          string
	data            data
	syncOK          bool
	syncedIDs       map[string]string
	lastSync        time.Time
	syncInterval    time.Duration
	noNetwork       bool
	stateSyncFunc   func(ctx context.Context, devices map[string]model.Device, missingIDs, onlineIDs, offlineIDs []string)
	mu              sync.RWMutex
}

func New(cloudClient cloud_client.ClientItf, subjectProvider handler.SubjectProvider, syncInterval time.Duration, wrkSpacePath, attrOrigin string) *Handler {
	return &Handler{
		cloudClient:     cloudClient,
		subjectProvider: subjectProvider,
		syncInterval:    syncInterval,
		wrkSpacePath:    wrkSpacePath,
		attrOrigin:      attrOrigin,
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

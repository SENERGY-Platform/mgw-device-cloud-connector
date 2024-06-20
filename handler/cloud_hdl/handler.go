package cloud_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
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
	lastSync        time.Time
	syncInterval    time.Duration
	noNetwork       bool
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

func (h *Handler) HasNetwork() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.noNetwork
}

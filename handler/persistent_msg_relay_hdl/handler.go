package persistent_msg_relay_hdl

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"sync"
	"time"
)

const logPrefix = "[persistent-relay-hdl]"
const cleanupFile = "cleanup.json"

type Handler struct {
	messages      chan handler.Message
	storageHdl    storageHandler
	handleFunc    handler.MessageHandler
	sendFunc      func(topic string, data []byte) error
	msgDayHour    int64
	msgNumber     int64
	mu            sync.Mutex
	limit         int
	readerDone    chan struct{}
	writerDone    chan struct{}
	writerCtx     context.Context
	writerCtxCf   context.CancelFunc
	errorState    bool
	errorStateMu  sync.RWMutex
	workSpacePath string
}

func New(workSpacePath string, buffer int, storageHdl storageHandler, handleFunc handler.MessageHandler, sendFunc func(topic string, data []byte) error, limit int) *Handler {
	ctx, cf := context.WithCancel(context.Background())
	return &Handler{
		messages:      make(chan handler.Message, buffer),
		storageHdl:    storageHdl,
		handleFunc:    handleFunc,
		sendFunc:      sendFunc,
		limit:         limit,
		readerDone:    make(chan struct{}),
		writerDone:    make(chan struct{}),
		writerCtx:     ctx,
		writerCtxCf:   cf,
		workSpacePath: workSpacePath,
	}
}

func (h *Handler) Init(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	err := h.cleanup(ctx)
	if err != nil {
		return err
	}
	dayHour, msgNum, err := h.storageHdl.LastPosition(ctx)
	if err != nil && !errors.Is(err, NoResultsErr) {
		return err
	}
	h.msgDayHour = dayHour
	h.msgNumber = msgNum
	return nil
}

func (h *Handler) Put(m handler.Message) error {
	select {
	case h.messages <- m:
	default:
		return model.BufferFullErr
	}
	return nil
}

func (h *Handler) Start(ctx context.Context) {
	go h.writer(ctx)
	go h.reader(ctx)
}

func (h *Handler) Stop(timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		h.writerCtxCf()
		util.Logger.Warningf("%s writer timout exceeded", logPrefix)
		break
	case <-h.writerDone:
		break
	}
	select {
	case <-h.writerDone:
	default:
	}
	<-h.readerDone
}

func (h *Handler) Running() bool {
	h.errorStateMu.RLock()
	defer h.errorStateMu.RUnlock()
	return !h.errorState
}

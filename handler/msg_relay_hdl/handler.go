package msg_relay_hdl

import (
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"sync"
)

type Message interface {
	Topic() string
	Payload() []byte
}

type Handler struct {
	messages   chan Message
	handleFunc func(m Message) (topic string, data []byte, err error)
	sendFunc   func(topic string, data []byte) error
	started    bool
	closed     bool
	mu         sync.RWMutex
}

func New(buffer int, handleFunc func(m Message) (topic string, data []byte, err error), sendFunc func(topic string, data []byte) error) *Handler {
	return &Handler{
		messages:   make(chan Message, buffer),
		handleFunc: handleFunc,
		sendFunc:   sendFunc,
	}
}

func (h *Handler) Put(m Message) error {
	select {
	case h.messages <- m:
	default:
		return errors.New("buffer full")
	}
	return nil
}

func (h *Handler) Start() {
	h.mu.Lock()
	if !h.started {
		go h.run()
		h.started = true
	}
	h.mu.Unlock()
}

func (h *Handler) Stop() {
	h.mu.Lock()
	if !h.closed {
		close(h.messages)
		h.closed = true
	}
	h.mu.Unlock()
}

func (h *Handler) run() {
	for message := range h.messages {
		topic, data, err := h.handleFunc(message)
		if err != nil {
			util.Logger.Error(err)
		} else {
			if err = h.sendFunc(topic, data); err != nil {
				util.Logger.Error(err)
			}
		}
	}
}

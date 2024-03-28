package msg_relay_hdl

import (
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"sync"
)

type Handler struct {
	messages   chan handler.Message
	handleFunc handler.MessageHandler
	sendFunc   func(topic string, data []byte) error
	dChan      chan struct{}
	mu         sync.RWMutex
}

func New(buffer int, handleFunc handler.MessageHandler, sendFunc func(topic string, data []byte) error) *Handler {
	return &Handler{
		messages:   make(chan handler.Message, buffer),
		handleFunc: handleFunc,
		sendFunc:   sendFunc,
		dChan:      make(chan struct{}),
	}
}

func (h *Handler) Put(m handler.Message) error {
	select {
	case h.messages <- m:
	default:
		return errors.New("buffer full")
	}
	return nil
}

func (h *Handler) Start() {
	go h.run()
}

func (h *Handler) Stop() {
	close(h.messages)
	<-h.dChan
}

func (h *Handler) run() {
	for message := range h.messages {
		topic, data, err := h.handleFunc(message)
		if err != nil {
			if err != model.NoMsgErr {
				util.Logger.Error(err)
			}
		} else {
			if err = h.sendFunc(topic, data); err != nil {
				util.Logger.Error(err)
			}
		}
	}
	h.dChan <- struct{}{}
}

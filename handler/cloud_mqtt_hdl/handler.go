package cloud_mqtt_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"sync"
)

const LogPrefix = "[cloud-mqtt]"

type Handler struct {
	client                  handler.MqttClient
	deviceCmdMsgRelayHdl    handler.MessageRelayHandler
	processesCmdMsgRelayHdl handler.MessageRelayHandler
	subscriptions           map[string]struct{}
	qos                     byte
	mu                      sync.RWMutex
}

func New(qos byte) *Handler {
	return &Handler{
		qos:           qos,
		subscriptions: make(map[string]struct{}),
	}
}

func (h *Handler) SetMqttClient(c handler.MqttClient) {
	h.client = c
}

func (h *Handler) SetMessageRelayHdl(deviceCmdMsgRelayHdl, processesCmdMsgRelayHdl handler.MessageRelayHandler) {
	h.deviceCmdMsgRelayHdl = deviceCmdMsgRelayHdl
	h.processesCmdMsgRelayHdl = processesCmdMsgRelayHdl
}

func (h *Handler) HandleOnDisconnect() {
	h.mu.Lock()
	defer h.mu.Unlock()
	clear(h.subscriptions)
	util.Logger.Debugf(LogPrefix + " subscriptions cleared")
}

func (h *Handler) HandleSubscriptions(_ context.Context, devices map[string]model.Device, _, _, _ []string) {
	err := h.subscribe(topic.Handler.CloudProcessesCmdSub(), func(m handler.Message) {
		if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err == model.NotConnectedErr {
		return
	}
	topics := map[string]struct{}{
		topic.Handler.CloudProcessesCmdSub(): {},
	}
	for id, device := range devices {
		if device.State == model.Online {
			topics[topic.Handler.CloudDeviceServiceCmdSub(id)] = struct{}{}
		}
	}
	missingTopics, newTopics := h.diffSubs(topics)
	for _, t := range missingTopics {
		if err = h.unsubscribe(t); err != nil {
			if err == model.NotConnectedErr {
				return
			}
			continue
		}
	}
	for _, t := range newTopics {
		err = h.subscribe(t, func(m handler.Message) {
			if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
			}
		})
		if err != nil {
			if err == model.NotConnectedErr {
				return
			}
			continue
		}
	}
}

func (h *Handler) subscribe(t string, mhf func(m handler.Message)) error {
	if !h.inSub(t) {
		util.Logger.Debugf(model.SubscribeString, LogPrefix, t)
		if err := h.client.Subscribe(t, h.qos, mhf); err != nil {
			util.Logger.Errorf(model.SubscribeErrString, LogPrefix, t, err)
			return err
		}
		h.addSub(t)
		util.Logger.Infof(model.SubscribedString, LogPrefix, t)
	}
	return nil
}

func (h *Handler) unsubscribe(t string) error {
	if h.inSub(t) {
		util.Logger.Debugf(model.UnsubscribeString, LogPrefix, t)
		if err := h.client.Unsubscribe(t); err != nil {
			util.Logger.Errorf(model.UnsubscribeErrString, LogPrefix, t, err)
			return err
		}
		h.deleteSub(t)
		util.Logger.Infof(model.UnsubscribedString, LogPrefix, t)
	}
	return nil
}

func (h *Handler) diffSubs(topics map[string]struct{}) (missingTopics, newTopics []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for t := range h.subscriptions {
		if _, k := topics[t]; !k {
			missingTopics = append(missingTopics, t)
		}
	}
	for t := range topics {
		if _, ok := h.subscriptions[t]; !ok {
			newTopics = append(newTopics, t)
		}
	}
	return
}

func (h *Handler) inSub(t string) (ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok = h.subscriptions[t]
	return
}

func (h *Handler) addSub(t string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscriptions[t] = struct{}{}
}

func (h *Handler) deleteSub(t string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subscriptions, t)
}

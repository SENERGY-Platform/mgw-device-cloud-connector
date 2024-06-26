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
	clear(h.subscriptions)
	h.mu.Unlock()
	util.Logger.Debugf(LogPrefix + " subscriptions cleared")
}

func (h *Handler) HandleSubscriptions(_ context.Context, devices map[string]model.Device, isOnlineIDs, isOfflineIDs, isOnlineAgainIDs []string) ([]string, error) {
	err := h.subscribe(topic.Handler.CloudProcessesCmdSub(), func(m handler.Message) {
		if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err == model.NotConnectedErr {
		return nil, err
	}
	var failed []string
	syncResults := make(map[string]bool)
	for _, id := range isOnlineIDs {
		err = h.subscribe(topic.Handler.CloudDeviceServiceCmdSub(id), func(m handler.Message) {
			if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
			}
		})
		if err != nil {
			if err == model.NotConnectedErr {
				return nil, err
			}
			failed = append(failed, id)
			syncResults[id] = false
			continue
		}
		syncResults[id] = true
	}
	for _, id := range isOnlineAgainIDs {
		err = h.resubscribe(topic.Handler.CloudDeviceServiceCmdSub(id), func(m handler.Message) {
			if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
			}
		})
		if err != nil {
			if err == model.NotConnectedErr {
				return nil, err
			}
			failed = append(failed, id)
			syncResults[id] = false
			continue
		}
		syncResults[id] = true
	}
	for _, id := range isOfflineIDs {
		if err = h.unsubscribe(topic.Handler.CloudDeviceServiceCmdSub(id)); err != nil {
			failed = append(failed, id)
		}
		if err != nil {
			if err == model.NotConnectedErr {
				return nil, err
			}
			failed = append(failed, id)
			syncResults[id] = false
			continue
		}
		syncResults[id] = true
	}
	for id, device := range devices {
		if _, ok := syncResults[id]; !ok {
			t := topic.Handler.CloudDeviceServiceCmdSub(id)
			if device.State == model.Online && !h.isSubscribed(t) {
				err = h.subscribe(t, func(m handler.Message) {
					if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
						util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
					}
				})
				if err != nil {
					if err == model.NotConnectedErr {
						return nil, err
					}
					failed = append(failed, id)
				}
			}
		}
	}
	return failed, nil
}

func (h *Handler) subscribe(t string, mhf func(m handler.Message)) error {
	if h.isSubscribed(t) {
		return nil
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, t)
	if err := h.client.Subscribe(t, h.qos, mhf); err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, t, err)
		return err
	}
	h.mu.Lock()
	h.subscriptions[t] = struct{}{}
	h.mu.Unlock()
	util.Logger.Infof(model.SubscribedString, LogPrefix, t)
	return nil
}

func (h *Handler) unsubscribe(t string) error {
	if !h.isSubscribed(t) {
		return nil
	}
	util.Logger.Debugf(model.UnsubscribeString, LogPrefix, t)
	if err := h.client.Unsubscribe(t); err != nil {
		util.Logger.Errorf(model.UnsubscribeErrString, LogPrefix, t, err)
		return err
	}
	h.mu.Lock()
	delete(h.subscriptions, t)
	h.mu.Unlock()
	util.Logger.Infof(model.UnsubscribedString, LogPrefix, t)
	return nil
}

func (h *Handler) resubscribe(t string, mhf func(m handler.Message)) error {
	util.Logger.Debugf(model.ResubscribeString, LogPrefix, t)
	if err := h.client.Unsubscribe(t); err != nil {
		util.Logger.Errorf(model.UnsubscribeErrString, LogPrefix, t, err)
		return err
	}
	if err := h.client.Subscribe(t, h.qos, mhf); err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, t, err)
		return err
	}
	h.mu.Lock()
	h.subscriptions[t] = struct{}{}
	h.mu.Unlock()
	util.Logger.Infof(model.ResubscribedString, LogPrefix, t)
	return nil
}

func (h *Handler) isSubscribed(t string) (ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok = h.subscriptions[t]
	return
}

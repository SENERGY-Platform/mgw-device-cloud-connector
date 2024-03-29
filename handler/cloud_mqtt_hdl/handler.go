package cloud_mqtt_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"sync"
)

const (
	RelayMsgErrString    = "relaying message failed: topic=%s err=%s"
	SubscribeString      = "subscribing to cloud topic '%s'"
	SubscribeErrString   = "subscribing to cloud topic '%s' failed: %s"
	UnsubscribeString    = "unsubscribing from cloud topic '%s'"
	UnsubscribeErrString = "unsubscribing from cloud topic '%s' failed: %s"
	ResubscribeString    = "resubscribing to cloud topic '%s'"
)

type Handler struct {
	client                  handler.MqttClient
	deviceCmdMsgRelayHdl    handler.MessageRelayHandler
	processesCmdMsgRelayHdl handler.MessageRelayHandler
	subscriptions           map[string]struct{}
	qos                     byte
	hubID                   string
	mu                      sync.RWMutex
}

func New(qos byte, hubID string) *Handler {
	return &Handler{
		qos:           qos,
		hubID:         hubID,
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
	util.Logger.Debug("cloud mqtt client disconnected, clearing subscriptions")
	h.mu.Lock()
	clear(h.subscriptions)
	h.mu.Unlock()
}

func (h *Handler) HandleSubscriptions(_ context.Context, devices map[string]model.Device, isOnlineIDs, isOfflineIDs, isOnlineAgainIDs []string) ([]string, error) {
	err := h.subscribe("processes/"+h.hubID+"/cmd/#", func(m handler.Message) {
		if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(RelayMsgErrString, m.Topic(), err)
		}
	})
	if err == model.NotConnectedErr {
		return nil, err
	}
	var failed []string
	syncResults := make(map[string]bool)
	for _, id := range isOnlineIDs {
		err = h.subscribe("command/"+id+"/+", func(m handler.Message) {
			if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(RelayMsgErrString, m.Topic(), err)
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
		err = h.resubscribe("command/"+id+"/+", func(m handler.Message) {
			if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(RelayMsgErrString, m.Topic(), err)
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
		if err = h.unsubscribe("command/" + id + "/+"); err != nil {
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
			t := "command/" + id + "/+"
			if device.State == model.Online && !h.isSubscribed(t) {
				err = h.subscribe(t, func(m handler.Message) {
					if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
						util.Logger.Errorf(RelayMsgErrString, m.Topic(), err)
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
	util.Logger.Debugf(SubscribeString, t)
	if err := h.client.Subscribe(t, h.qos, mhf); err != nil {
		util.Logger.Errorf(SubscribeErrString, t, err)
		return err
	}
	h.mu.Lock()
	h.subscriptions[t] = struct{}{}
	h.mu.Unlock()
	return nil
}

func (h *Handler) unsubscribe(t string) error {
	if !h.isSubscribed(t) {
		return nil
	}
	util.Logger.Debugf(UnsubscribeString, t)
	if err := h.client.Unsubscribe(t); err != nil {
		util.Logger.Errorf(UnsubscribeErrString, t, err)
		return err
	}
	h.mu.Lock()
	delete(h.subscriptions, t)
	h.mu.Unlock()
	return nil
}

func (h *Handler) resubscribe(t string, mhf func(m handler.Message)) error {
	util.Logger.Debugf(ResubscribeString, t)
	if err := h.client.Unsubscribe(t); err != nil {
		util.Logger.Errorf(UnsubscribeErrString, t, err)
		return err
	}
	if err := h.client.Subscribe(t, h.qos, mhf); err != nil {
		util.Logger.Errorf(SubscribeErrString, t, err)
		return err
	}
	h.mu.Lock()
	h.subscriptions[t] = struct{}{}
	h.mu.Unlock()
	return nil
}

func (h *Handler) isSubscribed(t string) (ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok = h.subscriptions[t]
	return
}

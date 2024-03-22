package local_mqtt_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

type Handler struct {
	client                        handler.MqttClient
	deviceEventMsgRelayHdl        handler.MessageRelayHandler
	deviceCmdRespMsgRelayHdl      handler.MessageRelayHandler
	processesStateMsgRelayHdl     handler.MessageRelayHandler
	deviceConnectorErrMsgRelayHdl handler.MessageRelayHandler
	deviceErrMsgRelayHdl          handler.MessageRelayHandler
	deviceCmdErrMsgRelayHdl       handler.MessageRelayHandler
	qos                           byte
}

func New(qos byte) *Handler {
	return &Handler{
		qos: qos,
	}
}

func (h *Handler) SetMqttClient(c handler.MqttClient) {
	h.client = c
}

func (h *Handler) HandleSubscriptions() {
	err := h.client.Subscribe(topic.LocalDeviceEventSub, h.qos, func(m handler.Message) {
		if err := h.deviceEventMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceEventSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdResponseSub, h.qos, func(m handler.Message) {
		if err := h.deviceCmdRespMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceCmdResponseSub, err)
	}
	err = h.client.Subscribe(topic.LocalProcessesStateSub, h.qos, func(m handler.Message) {
		if err := h.processesStateMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalProcessesStateSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceConnectorErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceConnectorErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceConnectorErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceCmdErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceCmdErrSub, err)
	}
}

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
}

func New(deviceEventMsgRelayHdl, deviceCmdRespMsgRelayHdl, processesStateMsgRelayHdl, deviceConnectorErrMsgRelayHdl, deviceErrMsgRelayHdl, deviceCmdErrMsgRelayHdl handler.MessageRelayHandler) *Handler {
	return &Handler{
		deviceEventMsgRelayHdl:        deviceEventMsgRelayHdl,
		deviceCmdRespMsgRelayHdl:      deviceCmdRespMsgRelayHdl,
		processesStateMsgRelayHdl:     processesStateMsgRelayHdl,
		deviceConnectorErrMsgRelayHdl: deviceConnectorErrMsgRelayHdl,
		deviceErrMsgRelayHdl:          deviceErrMsgRelayHdl,
		deviceCmdErrMsgRelayHdl:       deviceCmdErrMsgRelayHdl,
	}
}

func (h *Handler) SetMqttClient(c handler.MqttClient) {
	h.client = c
}

func (h *Handler) HandleSubscriptions() {
	err := h.client.Subscribe(topic.LocalDeviceEventSub, func(m handler.Message) {
		if err := h.deviceEventMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceEventSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdResponseSub, func(m handler.Message) {
		if err := h.deviceCmdRespMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceCmdResponseSub, err)
	}
	err = h.client.Subscribe(topic.LocalProcessesStateSub, func(m handler.Message) {
		if err := h.processesStateMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalProcessesStateSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceConnectorErrSub, func(m handler.Message) {
		if err := h.deviceConnectorErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceConnectorErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceErrSub, func(m handler.Message) {
		if err := h.deviceErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdErrSub, func(m handler.Message) {
		if err := h.deviceCmdErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, topic.LocalDeviceCmdErrSub, err)
	}
}
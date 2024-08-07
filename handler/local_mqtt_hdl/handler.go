package local_mqtt_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

const LogPrefix = "[local-mqtt]"

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

func (h *Handler) SetMessageRelayHdl(deviceEventMsgRelayHdl, deviceCmdRespMsgRelayHdl, processesStateMsgRelayHdl, deviceConnectorErrMsgRelayHdl, deviceErrMsgRelayHdl, deviceCmdErrMsgRelayHdl handler.MessageRelayHandler) {
	h.deviceEventMsgRelayHdl = deviceEventMsgRelayHdl
	h.deviceCmdRespMsgRelayHdl = deviceCmdRespMsgRelayHdl
	h.processesStateMsgRelayHdl = processesStateMsgRelayHdl
	h.deviceConnectorErrMsgRelayHdl = deviceConnectorErrMsgRelayHdl
	h.deviceErrMsgRelayHdl = deviceErrMsgRelayHdl
	h.deviceCmdErrMsgRelayHdl = deviceCmdErrMsgRelayHdl
}

func (h *Handler) HandleSubscriptions() {
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalDeviceEventSub)
	err := h.client.Subscribe(topic.LocalDeviceEventSub, h.qos, func(m handler.Message) {
		if err := h.deviceEventMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalDeviceEventSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalDeviceEventSub)
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalDeviceCmdResponseSub)
	err = h.client.Subscribe(topic.LocalDeviceCmdResponseSub, h.qos, func(m handler.Message) {
		if err := h.deviceCmdRespMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalDeviceCmdResponseSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalDeviceCmdResponseSub)
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalProcessesStateSub)
	err = h.client.Subscribe(topic.LocalProcessesStateSub, h.qos, func(m handler.Message) {
		if err := h.processesStateMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalProcessesStateSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalProcessesStateSub)
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalDeviceConnectorErrSub)
	err = h.client.Subscribe(topic.LocalDeviceConnectorErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceConnectorErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalDeviceConnectorErrSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalDeviceConnectorErrSub)
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalDeviceErrSub)
	err = h.client.Subscribe(topic.LocalDeviceErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalDeviceErrSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalDeviceErrSub)
	}
	util.Logger.Debugf(model.SubscribeString, LogPrefix, topic.LocalDeviceCmdErrSub)
	err = h.client.Subscribe(topic.LocalDeviceCmdErrSub, h.qos, func(m handler.Message) {
		if err := h.deviceCmdErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, LogPrefix, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, LogPrefix, topic.LocalDeviceCmdErrSub, err)
	} else {
		util.Logger.Infof(model.SubscribedString, LogPrefix, topic.LocalDeviceCmdErrSub)
	}
}

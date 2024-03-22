package paho_mqtt

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type LocalMqttHandler struct {
	client                        *Wrapper
	deviceEventMsgRelayHdl        handler.MessageRelayHandler
	deviceCmdRespMsgRelayHdl      handler.MessageRelayHandler
	processesStateMsgRelayHdl     handler.MessageRelayHandler
	deviceConnectorErrMsgRelayHdl handler.MessageRelayHandler
	deviceErrMsgRelayHdl          handler.MessageRelayHandler
	deviceCmdErrMsgRelayHdl       handler.MessageRelayHandler
}

func NewLocalMqttHandler(deviceEventMsgRelayHdl, deviceCmdRespMsgRelayHdl, processesStateMsgRelayHdl, deviceConnectorErrMsgRelayHdl, deviceErrMsgRelayHdl, deviceCmdErrMsgRelayHdl handler.MessageRelayHandler) *LocalMqttHandler {
	return &LocalMqttHandler{
		deviceEventMsgRelayHdl:        deviceEventMsgRelayHdl,
		deviceCmdRespMsgRelayHdl:      deviceCmdRespMsgRelayHdl,
		processesStateMsgRelayHdl:     processesStateMsgRelayHdl,
		deviceConnectorErrMsgRelayHdl: deviceConnectorErrMsgRelayHdl,
		deviceErrMsgRelayHdl:          deviceErrMsgRelayHdl,
		deviceCmdErrMsgRelayHdl:       deviceCmdErrMsgRelayHdl,
	}
}

func (h *LocalMqttHandler) SetMqttClient(w *Wrapper) {
	h.client = w
}

func (h *LocalMqttHandler) HandleSubscriptions(_ mqtt.Client) {
	err := h.client.Subscribe(topic.LocalDeviceEventSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceEventMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceEventSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdResponseSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceCmdRespMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceCmdResponseSub, err)
	}
	err = h.client.Subscribe(topic.LocalProcessesStateSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.processesStateMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalProcessesStateSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceConnectorErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceConnectorErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceConnectorErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceErrSub, err)
	}
	err = h.client.Subscribe(topic.LocalDeviceCmdErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceCmdErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceCmdErrSub, err)
	}
}

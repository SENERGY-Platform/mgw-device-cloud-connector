package paho_mqtt

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type LocalHandler struct {
	wrapper
	deviceEventMsgRelayHdl        handler.MessageRelayHandler
	deviceCmdRespMsgRelayHdl      handler.MessageRelayHandler
	processesStateMsgRelayHdl     handler.MessageRelayHandler
	deviceConnectorErrMsgRelayHdl handler.MessageRelayHandler
	deviceErrMsgRelayHdl          handler.MessageRelayHandler
	deviceCmdErrMsgRelayHdl       handler.MessageRelayHandler
}

func NewLocalHandler(client mqtt.Client, qos byte, timeout time.Duration, deviceEventMsgRelayHdl, deviceCmdRespMsgRelayHdl, processesStateMsgRelayHdl, deviceConnectorErrMsgRelayHdl, deviceErrMsgRelayHdl, deviceCmdErrMsgRelayHdl handler.MessageRelayHandler) *LocalHandler {
	return &LocalHandler{
		wrapper: wrapper{
			client:  client,
			qos:     qos,
			timeout: timeout,
		},
		deviceEventMsgRelayHdl:        deviceEventMsgRelayHdl,
		deviceCmdRespMsgRelayHdl:      deviceCmdRespMsgRelayHdl,
		processesStateMsgRelayHdl:     processesStateMsgRelayHdl,
		deviceConnectorErrMsgRelayHdl: deviceConnectorErrMsgRelayHdl,
		deviceErrMsgRelayHdl:          deviceErrMsgRelayHdl,
		deviceCmdErrMsgRelayHdl:       deviceCmdErrMsgRelayHdl,
	}
}

func (h *LocalHandler) HandleSubscriptions(_ mqtt.Client) {
	err := h.Subscribe(topic.LocalDeviceEventSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceEventMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceEventSub, err)
	}
	err = h.Subscribe(topic.LocalDeviceCmdResponseSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceCmdRespMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceCmdResponseSub, err)
	}
	err = h.Subscribe(topic.LocalProcessesStateSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.processesStateMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalProcessesStateSub, err)
	}
	err = h.Subscribe(topic.LocalDeviceConnectorErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceConnectorErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceConnectorErrSub, err)
	}
	err = h.Subscribe(topic.LocalDeviceErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceErrSub, err)
	}
	err = h.Subscribe(topic.LocalDeviceCmdErrSub, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.deviceCmdErrMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, topic.LocalDeviceCmdErrSub, err)
	}
}

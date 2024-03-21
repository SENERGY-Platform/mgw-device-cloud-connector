package paho_mqtt

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

type UpstreamHandler struct {
	wrapper
	cloudDeviceHdl          handler.CloudDeviceHandler
	localDeviceHdl          handler.LocalDeviceHandler
	deviceCmdMsgRelayHdl    handler.MessageRelayHandler
	processesCmdMsgRelayHdl handler.MessageRelayHandler
}

func NewUpstreamHdl(cloudDeviceHdl handler.CloudDeviceHandler, localDeviceHdl handler.LocalDeviceHandler, deviceCmdMsgRelayHdl handler.MessageRelayHandler, processesCmdMsgRelayHdl handler.MessageRelayHandler, qos byte, timeout time.Duration) *UpstreamHandler {
	return &UpstreamHandler{
		wrapper: wrapper{
			qos:     qos,
			timeout: timeout,
		},
		cloudDeviceHdl:          cloudDeviceHdl,
		localDeviceHdl:          localDeviceHdl,
		deviceCmdMsgRelayHdl:    deviceCmdMsgRelayHdl,
		processesCmdMsgRelayHdl: processesCmdMsgRelayHdl,
	}
}

func (h *UpstreamHandler) HandleSubscriptions(client mqtt.Client) {
	devices := h.localDeviceHdl.GetDevices()
	for id, device := range devices {
		if device.State == model.Online {
			t := "command" + id + "/+"
			err := h.Subscribe(client, t, func(_ mqtt.Client, m mqtt.Message) {
				if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
					util.Logger.Errorf(relayMsgErr, m.Topic(), err)
				}
			})
			if err != nil {
				util.Logger.Errorf(subscribeErr, t, err)
			}
		}
	}
	if hubID := h.cloudDeviceHdl.GetHubID(); hubID != "" {
		t := "processes/" + hubID + "/cmd/#"
		err := h.Subscribe(client, t, func(_ mqtt.Client, m mqtt.Message) {
			if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(relayMsgErr, m.Topic(), err)
			}
		})
		if err != nil {
			util.Logger.Errorf(subscribeErr, t, err)
		}
	}
}

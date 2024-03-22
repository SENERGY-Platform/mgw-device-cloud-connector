package paho_mqtt

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

type CloudHandler struct {
	wrapper
	cloudDeviceHdl          handler.CloudDeviceHandler
	localDeviceHdl          handler.LocalDeviceHandler
	deviceCmdMsgRelayHdl    handler.MessageRelayHandler
	processesCmdMsgRelayHdl handler.MessageRelayHandler
}

func NewCloudHdl(client mqtt.Client, qos byte, timeout time.Duration, cloudDeviceHdl handler.CloudDeviceHandler, localDeviceHdl handler.LocalDeviceHandler, deviceCmdMsgRelayHdl, processesCmdMsgRelayHdl handler.MessageRelayHandler) *CloudHandler {
	return &CloudHandler{
		wrapper: wrapper{
			client:  client,
			qos:     qos,
			timeout: timeout,
		},
		cloudDeviceHdl:          cloudDeviceHdl,
		localDeviceHdl:          localDeviceHdl,
		deviceCmdMsgRelayHdl:    deviceCmdMsgRelayHdl,
		processesCmdMsgRelayHdl: processesCmdMsgRelayHdl,
	}
}

func (h *CloudHandler) HandleSubscriptions(_ mqtt.Client) {
	devices := h.localDeviceHdl.GetDevices()
	for id, device := range devices {
		if device.State == model.Online {
			t := "command" + id + "/+"
			err := h.Subscribe(t, func(_ mqtt.Client, m mqtt.Message) {
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
		err := h.Subscribe(t, func(_ mqtt.Client, m mqtt.Message) {
			if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(relayMsgErr, m.Topic(), err)
			}
		})
		if err != nil {
			util.Logger.Errorf(subscribeErr, t, err)
		}
	}
}

func (h *CloudHandler) HandleMissingDevices(missing []string) error {
	for _, id := range missing {
		t := "command" + id + "/+"
		if err := h.Unsubscribe(t); err != nil {
			util.Logger.Errorf(unsubscribeErr, t, err)
		}
	}
	return nil
}

func (h *CloudHandler) HandleDeviceStates(deviceStates map[string]string) (failed []string, err error) {
	for id, state := range deviceStates {
		t := "command" + id + "/+"
		switch state {
		case model.Online:
			err = h.Subscribe(t, func(_ mqtt.Client, m mqtt.Message) {
				if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
					util.Logger.Errorf(relayMsgErr, m.Topic(), err)
				}
			})
			if err != nil {
				util.Logger.Errorf(subscribeErr, t, err)
			}
		case model.Offline, "":
			if err = h.Unsubscribe(t); err != nil {
				util.Logger.Errorf(unsubscribeErr, t, err)
			}
		}
		if err != nil {
			failed = append(failed, id)
		}
	}
	return failed, nil
}

func (h *CloudHandler) HandleHubIDChange(oldID, newID string) error {
	t := "processes/" + newID + "/cmd/#"
	err := h.Subscribe(t, func(_ mqtt.Client, m mqtt.Message) {
		if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(relayMsgErr, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(subscribeErr, t, err)
	}
	t = "processes/" + oldID + "/cmd/#"
	if err = h.Unsubscribe(t); err != nil {
		util.Logger.Errorf(unsubscribeErr, t, err)
	}
	return nil
}

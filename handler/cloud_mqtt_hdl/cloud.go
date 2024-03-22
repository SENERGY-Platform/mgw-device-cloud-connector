package cloud_mqtt_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
)

type CloudMqttHandler struct {
	client                  handler.MqttClient
	cloudDeviceHdl          handler.CloudDeviceHandler
	localDeviceHdl          handler.LocalDeviceHandler
	deviceCmdMsgRelayHdl    handler.MessageRelayHandler
	processesCmdMsgRelayHdl handler.MessageRelayHandler
}

func NewCloudMqttHdl(cloudDeviceHdl handler.CloudDeviceHandler, localDeviceHdl handler.LocalDeviceHandler, deviceCmdMsgRelayHdl, processesCmdMsgRelayHdl handler.MessageRelayHandler) *CloudMqttHandler {
	return &CloudMqttHandler{
		cloudDeviceHdl:          cloudDeviceHdl,
		localDeviceHdl:          localDeviceHdl,
		deviceCmdMsgRelayHdl:    deviceCmdMsgRelayHdl,
		processesCmdMsgRelayHdl: processesCmdMsgRelayHdl,
	}
}

func (h *CloudMqttHandler) SetMqttClient(c handler.MqttClient) {
	h.client = c
}

func (h *CloudMqttHandler) HandleSubscriptions() {
	devices := h.localDeviceHdl.GetDevices()
	for id, device := range devices {
		if device.State == model.Online {
			t := "command" + id + "/+"
			err := h.client.Subscribe(t, func(m handler.Message) {
				if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
					util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
				}
			})
			if err != nil {
				util.Logger.Errorf(model.SubscribeErrString, t, err)
			}
		}
	}
	if hubID := h.cloudDeviceHdl.GetHubID(); hubID != "" {
		t := "processes/" + hubID + "/cmd/#"
		err := h.client.Subscribe(t, func(m handler.Message) {
			if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
				util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
			}
		})
		if err != nil {
			util.Logger.Errorf(model.SubscribeErrString, t, err)
		}
	}
}

func (h *CloudMqttHandler) HandleMissingDevices(missing []string) error {
	for _, id := range missing {
		t := "command" + id + "/+"
		if err := h.client.Unsubscribe(t); err != nil {
			util.Logger.Errorf(model.UnsubscribeErrString, t, err)
		}
	}
	return nil
}

func (h *CloudMqttHandler) HandleDeviceStates(deviceStates map[string]string) (failed []string, err error) {
	for id, state := range deviceStates {
		t := "command" + id + "/+"
		switch state {
		case model.Online:
			err = h.client.Subscribe(t, func(m handler.Message) {
				if err := h.deviceCmdMsgRelayHdl.Put(m); err != nil {
					util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
				}
			})
			if err != nil {
				util.Logger.Errorf(model.SubscribeErrString, t, err)
			}
		case model.Offline, "":
			if err = h.client.Unsubscribe(t); err != nil {
				util.Logger.Errorf(model.UnsubscribeErrString, t, err)
			}
		}
		if err != nil {
			failed = append(failed, id)
		}
	}
	return failed, nil
}

func (h *CloudMqttHandler) HandleHubIDChange(oldID, newID string) error {
	t := "processes/" + newID + "/cmd/#"
	err := h.client.Subscribe(t, func(m handler.Message) {
		if err := h.processesCmdMsgRelayHdl.Put(m); err != nil {
			util.Logger.Errorf(model.RelayMsgErrString, m.Topic(), err)
		}
	})
	if err != nil {
		util.Logger.Errorf(model.SubscribeErrString, t, err)
	}
	t = "processes/" + oldID + "/cmd/#"
	if err = h.client.Unsubscribe(t); err != nil {
		util.Logger.Errorf(model.UnsubscribeErrString, t, err)
	}
	return nil
}

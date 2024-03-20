package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

func HandleLocalDeviceEvent(m handler.Message) (string, []byte, error) {
	var dID, sID string
	if !parseTopic(topic.LocalDeviceEventSub, m.Topic(), &dID, &sID) {
		return "", nil, newParseErr(m.Topic())
	}
	b, err := json.Marshal(model.DeviceEventMessage{Data: string(m.Payload())})
	if err != nil {
		return "", nil, err
	}
	return topic.CloudDeviceEventPub + "/" + dID + "/" + sID, b, nil
}

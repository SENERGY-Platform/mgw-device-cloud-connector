package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

var LocalDeviceIDPrefix string

func HandleUpstreamDeviceEvent(m handler.Message) (string, []byte, error) {
	var dID, sID string
	if !parseTopic(topic.LocalDeviceEventSub, m.Topic(), &dID, &sID) {
		return "", nil, newParseErr(m.Topic())
	}
	b, err := json.Marshal(CloudDeviceEventMsg{Data: string(m.Payload())})
	if err != nil {
		return "", nil, err
	}
	return topic.CloudDeviceEventPub + "/" + LocalDeviceIDPrefix + dID + "/" + sID, b, nil
}

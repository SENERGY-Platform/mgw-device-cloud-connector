package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"time"
)

func HandleUpstreamDeviceEventAgeLimit(m handler.Message) (string, []byte, error) {
	if time.Since(m.Timestamp()) <= DeviceEventMaxAge {
		var dID, sID string
		if !parseTopic(topic.LocalDeviceEventSub, m.Topic(), &dID, &sID) {
			return "", nil, newParseErr(m.Topic())
		}
		b, err := cloudDeviceEventMsgPayload(m)
		if err != nil {
			return "", nil, err
		}
		return topic.Handler.CloudDeviceEventPub(LocalDeviceIDPrefix+dID, sID), b, nil
	}
	util.Logger.Warningf("%s ignored device event (%s)", logPrefix, m.Topic())
	return "", nil, model.NoMsgErr
}

func HandleUpstreamDeviceEvent(m handler.Message) (string, []byte, error) {
	var dID, sID string
	if !parseTopic(topic.LocalDeviceEventSub, m.Topic(), &dID, &sID) {
		return "", nil, newParseErr(m.Topic())
	}
	b, err := cloudDeviceEventMsgPayload(m)
	if err != nil {
		return "", nil, err
	}
	return topic.Handler.CloudDeviceEventPub(LocalDeviceIDPrefix+dID, sID), b, nil
}

func cloudDeviceEventMsgPayload(m handler.Message) ([]byte, error) {
	b1, err := json.Marshal(CSEMetadata{Timestamp: m.Timestamp().Format(time.RFC3339Nano)})
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(CloudDeviceEventMsg{
		Metadata: string(b1),
		Data:     string(m.Payload()),
	})
	if err != nil {
		return nil, err
	}
	return b2, nil
}

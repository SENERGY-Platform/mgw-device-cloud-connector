package message_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"strings"
)

func HandlerUpstreamDeviceConnectorErr(m handler.Message) (string, []byte, error) {
	if m.Topic() != topic.LocalDeviceConnectorErrSub {
		return "", nil, newParseErr(m.Topic())
	}
	return topic.CloudDeviceConnectorErrPub, m.Payload(), nil
}

func HandlerUpstreamDeviceErr(m handler.Message) (string, []byte, error) {
	var dID string
	if !parseTopic(topic.LocalDeviceErrSub, m.Topic(), &dID) {
		return "", nil, newParseErr(m.Topic())
	}
	return topic.CloudDeviceErrPub + "/" + dID, m.Payload(), nil
}

func HandlerUpstreamDeviceCmdErr(m handler.Message) (string, []byte, error) {
	var cID string
	if !parseTopic(topic.LocalDeviceCmdErrSub, m.Topic(), &cID) {
		return "", nil, newParseErr(m.Topic())
	}
	if strings.Contains(cID, DeviceCommandIDPrefix) {
		return topic.CloudDeviceCmdErrPub + "/" + strings.ReplaceAll(cID, DeviceCommandIDPrefix, ""), m.Payload(), nil
	}
	return "", nil, model.NoMsgErr
}

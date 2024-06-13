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
	return "error", m.Payload(), nil
}

func HandlerUpstreamDeviceErr(m handler.Message) (string, []byte, error) {
	var dID string
	if !parseTopic(topic.LocalDeviceErrSub, m.Topic(), &dID) {
		return "", nil, newParseErr(m.Topic())
	}
	return "error/device/" + UserID + "/" + LocalDeviceIDPrefix + dID, m.Payload(), nil //  error/device/{user_id}/{local_device_id}
}

func HandlerUpstreamDeviceCmdErr(m handler.Message) (string, []byte, error) {
	var cID string
	if !parseTopic(topic.LocalDeviceCmdErrSub, m.Topic(), &cID) {
		return "", nil, newParseErr(m.Topic())
	}
	if strings.Contains(cID, DeviceCommandIDPrefix) {
		return "error/command/" + UserID + "/" + strings.ReplaceAll(cID, DeviceCommandIDPrefix, ""), m.Payload(), nil // error/command/{user_id}/{correlation_id}
	}
	return "", nil, model.NoMsgErr
}

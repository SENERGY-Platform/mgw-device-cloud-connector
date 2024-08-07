package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"math"
	"strings"
	"time"
)

func HandleDownstreamDeviceCmd(m handler.Message) (string, []byte, error) {
	var dID, sID string
	if !parseTopic(topic.Handler.CloudDeviceCmdSub(), m.Topic(), &dID, &sID) {
		return "", nil, newParseErr(m.Topic())
	}
	var cmd CloudDeviceCmdMsg
	if err := json.Unmarshal(m.Payload(), &cmd); err != nil {
		return "", nil, err
	}
	sec, mSec := math.Modf(cmd.Timestamp)
	if time.Since(time.Unix(int64(sec), int64(mSec*10000000))) <= DeviceCommandMaxAge {
		b, err := json.Marshal(LocalDeviceCmdMsg{
			LocalDeviceCmdBase: LocalDeviceCmdBase{
				CommandID: DeviceCommandIDPrefix + cmd.CorrelationID,
				Data:      cmd.Payload.Data,
			},
			CompletionStrategy: cmd.CompletionStrategy,
		})
		if err != nil {
			return "", nil, err
		}
		return topic.Handler.LocalDeviceCmdPub(strings.ReplaceAll(dID, LocalDeviceIDPrefix, ""), sID), b, nil
	}
	util.Logger.Warningf("%s ignored device command (%s)", logPrefix, m.Topic())
	return "", nil, model.NoMsgErr
}

func HandleUpstreamDeviceCmdResponse(m handler.Message) (string, []byte, error) {
	var dID, sID string
	if !parseTopic(topic.LocalDeviceCmdResponseSub, m.Topic(), &dID, &sID) {
		return "", nil, newParseErr(m.Topic())
	}
	var cmdRes LocalDeviceCmdResponseMsg
	if err := json.Unmarshal(m.Payload(), &cmdRes); err != nil {
		return "", nil, err
	}
	if strings.HasPrefix(cmdRes.CommandID, DeviceCommandIDPrefix) {
		b, err := json.Marshal(CloudDeviceCmdResponseMsg{
			CorrelationID: strings.ReplaceAll(cmdRes.CommandID, DeviceCommandIDPrefix, ""),
			Payload:       CloudStandardEnvelope{Data: cmdRes.Data},
		})
		if err != nil {
			return "", nil, err
		}
		return topic.Handler.CloudDeviceCmdResponsePub(LocalDeviceIDPrefix+dID, sID), b, nil
	}
	return "", nil, model.NoMsgErr
}

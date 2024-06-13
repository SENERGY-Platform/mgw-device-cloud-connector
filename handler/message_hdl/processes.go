package message_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

func HandleDownstreamProcessesCmd(m handler.Message) (string, []byte, error) {
	var hID, subTopic string
	if !parseTopic(topic.CloudProcessesCmdSub, m.Topic(), &hID, &subTopic) {
		return "", nil, newParseErr(m.Topic())
	}
	return topic.LocalProcessesCmdPub + "/" + subTopic, m.Payload(), nil
}

func HandleUpstreamProcessesState(m handler.Message) (string, []byte, error) {
	var subTopic string
	if !parseTopic(topic.LocalProcessesStateSub, m.Topic(), &subTopic) {
		return "", nil, newParseErr(m.Topic())
	}
	return topic.CloudProcessesStatePub + "/" + NetworkID + "/state/" + subTopic, m.Payload(), nil
}

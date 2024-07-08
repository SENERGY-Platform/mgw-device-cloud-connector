package message_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

func HandleDownstreamProcessesCmd(m handler.Message) (string, []byte, error) {
	var subTopic string
	if !parseTopic(topic.Handler.CloudProcessesCmdSub(), m.Topic(), &subTopic) { // processes/{hub_id}/cmd/#
		return "", nil, newParseErr(m.Topic())
	}
	return topic.Handler.LocalProcessesCmdPub(subTopic), m.Payload(), nil
}

func HandleUpstreamProcessesState(m handler.Message) (string, []byte, error) {
	var subTopic string
	if !parseTopic(topic.LocalProcessesStateSub, m.Topic(), &subTopic) {
		return "", nil, newParseErr(m.Topic())
	}
	return topic.Handler.CloudProcessesStatePub(subTopic), m.Payload(), nil // processes/{hub_id}/state/{sub_topic}
}

package message_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

func HandleDownstreamProcessesCmd(m handler.Message) (string, []byte, error) {
	var hID, subTopic string
	if !parseTopic("processes/"+UserID+"/+/cmd/#", m.Topic(), &hID, &subTopic) { // processes/{user_id}/{hub_id}/cmd/#
		return "", nil, newParseErr(m.Topic())
	}
	return "processes/cmd/" + subTopic, m.Payload(), nil
}

func HandleUpstreamProcessesState(m handler.Message) (string, []byte, error) {
	var subTopic string
	if !parseTopic(topic.LocalProcessesStateSub, m.Topic(), &subTopic) {
		return "", nil, newParseErr(m.Topic())
	}
	return "processes/" + NetworkID + "/state/" + subTopic, m.Payload(), nil // processes/{hub_id}/state/{sub_topic}
}

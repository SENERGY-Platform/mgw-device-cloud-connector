package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"time"
)

func HandleUpstreamDeviceEvent(m handler.Message) (string, []byte, error) {
	if time.Since(m.Timestamp()) <= DeviceEventMaxAge {
		var dID, sID string
		if !parseTopic(topic.LocalDeviceEventSub, m.Topic(), &dID, &sID) {
			return "", nil, newParseErr(m.Topic())
		}
		b, err := json.Marshal(CloudDeviceEventMsg{Data: string(m.Payload())})
		if err != nil {
			return "", nil, err
		}
		return "event/" + UserID + "/" + LocalDeviceIDPrefix + dID + "/" + sID, b, nil // event/{user_id}/{local_device_id}/{service_id}
	}
	util.Logger.Warningf("%s ignored device event (%s)", logPrefix, m.Topic())
	return "", nil, model.NoMsgErr
}

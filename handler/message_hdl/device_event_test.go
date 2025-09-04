package message_hdl

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"reflect"
	"testing"
	"time"
)

func TestHandleUpstreamDeviceEvent(t *testing.T) {
	t.Cleanup(clearVars)
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	topic.InitTopicHandler("usrID", "netID")
	DeviceEventMaxAge = time.Second * 5
	LocalDeviceIDPrefix = "123"
	a := CloudDeviceEventMsg{Data: "test"}
	rt, rb, err := HandleUpstreamDeviceEventAgeLimit(&mockMessage{
		topic:     "event/a/b",
		payload:   []byte("test"),
		timestamp: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.CloudDeviceEventPub(LocalDeviceIDPrefix+"a", "b") {
		t.Error("expected", topic.Handler.CloudDeviceEventPub("a", "b"), "got", rt)
	}
	var b CloudDeviceEventMsg
	err = json.Unmarshal(rb, &b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("expected", a, "got", b)
	}
	t.Run("no message", func(t *testing.T) {
		_, _, err = HandleUpstreamDeviceEventAgeLimit(&mockMessage{
			topic:   "event/a/b",
			payload: []byte("test"),
		})
		if err != model.NoMsgErr {
			t.Error("expected no message error")
		}
	})
	t.Run("error", func(t *testing.T) {
		_, _, err = HandleUpstreamDeviceEventAgeLimit(&mockMessage{
			topic:     "test",
			timestamp: time.Now(),
		})
		if err == nil {
			t.Error("expected error")
		}
		if err == model.NoMsgErr {
			t.Error("wrong error type")
		}
	})
}

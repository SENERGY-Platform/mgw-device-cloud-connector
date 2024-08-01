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

func TestHandleDownstreamDeviceCmd(t *testing.T) {
	t.Cleanup(clearVars)
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	topic.InitTopicHandler("usrID", "netID")
	DeviceCommandMaxAge = time.Second * 5
	LocalDeviceIDPrefix = "123"
	DeviceCommandIDPrefix = "123"
	p, err := json.Marshal(CloudDeviceCmdMsg{
		CorrelationID: "corrID",
		Timestamp:     float64(time.Now().Unix()),
		Payload: CloudStandardEnvelope{
			Data: "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	a := LocalDeviceCmdMsg{
		LocalDeviceCmdBase: LocalDeviceCmdBase{
			CommandID: DeviceCommandIDPrefix + "corrID",
			Data:      "test",
		},
	}
	rt, rb, err := HandleDownstreamDeviceCmd(&mockMessage{
		topic:     "command/usrID/123a/b",
		payload:   p,
		timestamp: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.LocalDeviceCmdPub("a", "b") {
		t.Error("expected", topic.Handler.LocalDeviceCmdPub("a", "b"), "got", rt)
	}
	var b LocalDeviceCmdMsg
	err = json.Unmarshal(rb, &b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("expected", a, "got", b)
	}
	t.Run("no message", func(t *testing.T) {
		p, err = json.Marshal(CloudDeviceCmdMsg{
			CorrelationID: "corrID",
			Payload: CloudStandardEnvelope{
				Data: "test",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = HandleDownstreamDeviceCmd(&mockMessage{
			topic:   "command/usrID/123a/b",
			payload: p,
		})
		if err != model.NoMsgErr {
			t.Error("expected no message error")
		}
	})
	t.Run("error", func(t *testing.T) {
		_, _, err = HandleDownstreamDeviceCmd(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
		if err == model.NoMsgErr {
			t.Error("wrong error type")
		}
	})
}

func TestHandleUpstreamDeviceCmdResponse(t *testing.T) {
	t.Cleanup(clearVars)
	topic.InitTopicHandler("usrID", "netID")
	LocalDeviceIDPrefix = "123"
	DeviceCommandIDPrefix = "123"
	p, err := json.Marshal(LocalDeviceCmdResponseMsg{
		CommandID: DeviceCommandIDPrefix + "cmdID",
		Data:      "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	a := CloudDeviceCmdResponseMsg{
		CorrelationID: "cmdID",
		Payload:       CloudStandardEnvelope{Data: "test"},
	}
	rt, rb, err := HandleUpstreamDeviceCmdResponse(&mockMessage{
		topic:     "response/a/b",
		payload:   p,
		timestamp: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.CloudDeviceCmdResponsePub(LocalDeviceIDPrefix+"a", "b") {
		t.Error("expected", topic.Handler.CloudDeviceCmdResponsePub(LocalDeviceIDPrefix+"a", "b"), "got", rt)
	}
	var b CloudDeviceCmdResponseMsg
	err = json.Unmarshal(rb, &b)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("expected", a, "got", b)
	}
	t.Run("no message", func(t *testing.T) {
		p, err = json.Marshal(LocalDeviceCmdResponseMsg{
			CommandID: "cmdID",
			Data:      "test",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = HandleUpstreamDeviceCmdResponse(&mockMessage{
			topic:   "response/a/b",
			payload: p,
		})
		if err != model.NoMsgErr {
			t.Error("expected no message error")
		}
	})
	t.Run("error", func(t *testing.T) {
		_, _, err = HandleUpstreamDeviceCmdResponse(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
		if err == model.NoMsgErr {
			t.Error("wrong error type")
		}
	})
}

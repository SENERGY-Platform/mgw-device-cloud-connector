package message_hdl

import (
	"bytes"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"testing"
)

func TestHandlerUpstreamDeviceConnectorErr(t *testing.T) {
	topic.InitTopicHandler("usrID", "netID")
	rt, rb, err := HandlerUpstreamDeviceConnectorErr(&mockMessage{
		topic:   "error/client",
		payload: []byte("test"),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != "error" {
		t.Error("expected error, got", rt)
	}
	if !bytes.Equal(rb, []byte("test")) {
		t.Error("expected test, got", string(rb))
	}
	t.Run("error", func(t *testing.T) {
		_, _, err = HandlerUpstreamDeviceConnectorErr(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestHandlerUpstreamDeviceErr(t *testing.T) {
	topic.InitTopicHandler("usrID", "netID")
	rt, rb, err := HandlerUpstreamDeviceErr(&mockMessage{
		topic:   "error/device/test",
		payload: []byte("test"),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.CloudDeviceErrPub(LocalDeviceIDPrefix+"test") {
		t.Error("expected", topic.Handler.CloudDeviceErrPub(LocalDeviceIDPrefix+"test"), "got", rt)
	}
	if !bytes.Equal(rb, []byte("test")) {
		t.Error("expected test, got", string(rb))
	}
	t.Run("error", func(t *testing.T) {
		_, _, err = HandlerUpstreamDeviceErr(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestHandlerUpstreamDeviceCmdErr(t *testing.T) {
	topic.InitTopicHandler("usrID", "netID")
	DeviceCommandIDPrefix = "123"
	rt, rb, err := HandlerUpstreamDeviceCmdErr(&mockMessage{
		topic:   "error/command/123test",
		payload: []byte("test"),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.CloudDeviceCmdErrPub("test") {
		t.Error("expected", topic.Handler.CloudDeviceCmdErrPub("test"), "got", rt)
	}
	if !bytes.Equal(rb, []byte("test")) {
		t.Error("expected test, got", string(rb))
	}
	t.Run("no message", func(t *testing.T) {
		_, _, err = HandlerUpstreamDeviceCmdErr(&mockMessage{
			topic:   "error/command/test",
			payload: []byte("test"),
		})
		if err != model.NoMsgErr {
			t.Error("expected no message error")
		}
	})
	t.Run("error", func(t *testing.T) {
		_, _, err = HandlerUpstreamDeviceCmdErr(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}

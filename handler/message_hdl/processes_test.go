package message_hdl

import (
	"bytes"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"testing"
)

func TestHandleDownstreamProcessesCmd(t *testing.T) {
	t.Cleanup(clearVars)
	topic.InitTopicHandler("usrID", "netID")
	rt, rb, err := HandleDownstreamProcessesCmd(&mockMessage{
		topic:   "processes/netID/cmd/a/b",
		payload: []byte("test"),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.LocalProcessesCmdPub("a/b") {
		t.Error("expected", topic.Handler.LocalProcessesCmdPub("a/b"), "got", rt)
	}
	if !bytes.Equal(rb, []byte("test")) {
		t.Error("expected test, got", string(rb))
	}
	t.Run("error", func(t *testing.T) {
		_, _, err = HandleDownstreamProcessesCmd(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestHandleUpstreamProcessesState(t *testing.T) {
	t.Cleanup(clearVars)
	topic.InitTopicHandler("usrID", "netID")
	rt, rb, err := HandleUpstreamProcessesState(&mockMessage{
		topic:   "processes/state/a/b",
		payload: []byte("test"),
	})
	if err != nil {
		t.Error(err)
	}
	if rt != topic.Handler.CloudProcessesStatePub("a/b") {
		t.Error("expected", topic.Handler.CloudProcessesStatePub("a/b"), "got", rt)
	}
	if !bytes.Equal(rb, []byte("test")) {
		t.Error("expected test, got", string(rb))
	}
	t.Run("error", func(t *testing.T) {
		_, _, err = HandleUpstreamProcessesState(&mockMessage{
			topic: "test",
		})
		if err == nil {
			t.Error("expected error")
		}
	})
}

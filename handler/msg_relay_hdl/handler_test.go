package msg_relay_hdl

import (
	"bytes"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"reflect"
	"testing"
	"time"
)

type mockMessage struct {
	topic     string
	payload   []byte
	timestamp time.Time
}

func (m *mockMessage) Topic() string {
	return m.topic
}

func (m *mockMessage) Payload() []byte {
	return m.payload
}

func (m *mockMessage) Timestamp() time.Time {
	return m.timestamp
}

func TestHandler(t *testing.T) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	msg := &mockMessage{
		topic:     "test",
		payload:   []byte("test"),
		timestamp: time.Now(),
	}
	testMsgHdl := func(m handler.Message) (topic string, data []byte, err error) {
		if !reflect.DeepEqual(m, msg) {
			t.Error("expected", msg, "got", m)
		}
		return "topic", []byte("payload"), nil
	}
	testSendFunc := func(topic string, data []byte) error {
		if topic != "topic" {
			t.Error("expected topic got", topic)
		}
		if !bytes.Equal(data, []byte("payload")) {
			t.Error("expected payload got", string(data))
		}
		return nil
	}
	h := New(1, testMsgHdl, testSendFunc)
	err := h.Put(msg)
	if err != nil {
		t.Error(err)
	}
	if len(h.messages) != 1 {
		t.Error("message not in channel")
	}
	h.Start()
	time.Sleep(1 * time.Second)
	if len(h.messages) > 0 {
		t.Error("message not consumed")
	}
	h.Stop()
	t.Run("message handler error", func(t *testing.T) {
		testMsgHdl = func(m handler.Message) (topic string, data []byte, err error) {
			return "", nil, errors.New("test error")
		}
		testSendFunc = func(topic string, data []byte) error {
			t.Error("illegal call")
			return nil
		}
		h = New(1, testMsgHdl, testSendFunc)
		err = h.Put(msg)
		if err != nil {
			t.Error(err)
		}
		if len(h.messages) != 1 {
			t.Error("message not in channel")
		}
		h.Start()
		time.Sleep(1 * time.Second)
		if len(h.messages) > 0 {
			t.Error("message not consumed")
		}
		h.Stop()
	})
	t.Run("no message", func(t *testing.T) {
		testMsgHdl = func(m handler.Message) (topic string, data []byte, err error) {
			return "", nil, model.NoMsgErr
		}
		testSendFunc = func(topic string, data []byte) error {
			t.Error("illegal call")
			return nil
		}
		h = New(1, testMsgHdl, testSendFunc)
		err = h.Put(msg)
		if err != nil {
			t.Error(err)
		}
		if len(h.messages) != 1 {
			t.Error("message not in channel")
		}
		h.Start()
		time.Sleep(1 * time.Second)
		if len(h.messages) > 0 {
			t.Error("message not consumed")
		}
		h.Stop()
	})
	t.Run("buffer full", func(t *testing.T) {
		testMsgHdl = func(m handler.Message) (topic string, data []byte, err error) {
			return "", nil, nil
		}
		testSendFunc = func(topic string, data []byte) error {
			return nil
		}
		h = New(1, testMsgHdl, testSendFunc)
		err = h.Put(msg)
		if err != nil {
			t.Error(err)
		}
		err = h.Put(msg)
		if err == nil {
			t.Error(err)
		}
		if len(h.messages) != 1 {
			t.Error("message not in channel")
		}
		h.Start()
		time.Sleep(1 * time.Second)
		if len(h.messages) > 0 {
			t.Error("message not consumed")
		}
		h.Stop()
	})
}

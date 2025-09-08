package persistent_msg_relay_hdl

import (
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"os"
	"path"
	"slices"
	"testing"
	"time"
)

func TestHandler_read(t *testing.T) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	timeNow := time.Now()
	messages := []StorageMessage{
		{
			ID:           "a",
			DayHour:      1,
			Number:       1,
			MsgTopic:     "a/b",
			MsgPayload:   []byte("test"),
			MsgTimestamp: timeNow,
		},
		{
			ID:           "b",
			DayHour:      1,
			Number:       2,
			MsgTopic:     "a/b",
			MsgPayload:   []byte("test"),
			MsgTimestamp: timeNow.Add(time.Second),
		},
	}
	t.Run("success", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{
			Messages: slices.Clone(messages),
		}
		var i int
		h := New("", 0, stgHdlMock, func(m handler.Message) (topic string, data []byte, err error) {
			return m.Topic(), m.Payload(), nil
		}, func(topic string, data []byte) error {
			i++
			return nil
		}, 100)
		err := h.read(t.Context())
		if err != nil {
			t.Error(err)
		}
		if i != 2 {
			t.Error("invalid read message count")
		}
		if len(stgHdlMock.Messages) != 0 {
			t.Error("invalid messages length")
		}
	})
	t.Run("send error", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{
			Messages: slices.Clone(messages),
		}
		var i int
		h := New("", 0, stgHdlMock, func(m handler.Message) (topic string, data []byte, err error) {
			return m.Topic(), m.Payload(), nil
		}, func(topic string, data []byte) error {
			i++
			return errors.New("test")
		}, 100)
		err := h.read(t.Context())
		if err != nil {
			t.Error(err)
		}
		if i != 1 {
			t.Error("invalid read message count")
		}
		if len(stgHdlMock.Messages) != 2 {
			t.Error("invalid messages length")
		}
	})
	t.Run("delete error", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{
			Messages:          slices.Clone(messages),
			DeleteMessagesErr: errors.New("test"),
		}
		tmpDir := t.TempDir()
		var i int
		h := New(tmpDir, 0, stgHdlMock, func(m handler.Message) (topic string, data []byte, err error) {
			return m.Topic(), m.Payload(), nil
		}, func(topic string, data []byte) error {
			i++
			return nil
		}, 100)
		err := h.read(t.Context())
		if err == nil {
			t.Error("expected error")
		}
		if i != 2 {
			t.Error("invalid read message count")
		}
		_, err = os.Stat(path.Join(tmpDir, cleanupFile))
		if err != nil {
			t.Error(err)
		}
	})
}

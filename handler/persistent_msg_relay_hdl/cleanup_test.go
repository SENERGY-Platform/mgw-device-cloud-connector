package persistent_msg_relay_hdl

import (
	"errors"
	"os"
	"path"
	"testing"
)

func TestHandler_cleanup(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		storageHdlMock := &storageHandlerMock{
			Messages: []StorageMessage{
				{
					ID: "a",
				},
				{
					ID: "b",
				},
				{
					ID: "c",
				},
			},
		}
		h := New(tmpDir, 0, storageHdlMock, nil, nil, 0)
		err := h.createCleanupFile([]string{"a", "b"})
		if err != nil {
			t.Error(err)
		}
		_, err = os.Stat(path.Join(tmpDir, cleanupFile))
		if err != nil {
			t.Error(err)
		}
		err = h.cleanup(t.Context())
		if err != nil {
			t.Error(err)
		}
		_, err = os.Stat(path.Join(tmpDir, cleanupFile))
		if !errors.Is(err, os.ErrNotExist) {
			t.Error("expected error")
		}
		if len(storageHdlMock.Messages) != 1 {
			t.Fatal("invalid messages length")
		}
		if storageHdlMock.Messages[0].ID != "c" {
			t.Error("wrong message ID")
		}
	})
	t.Run("no file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storageHdlMock := &storageHandlerMock{
			Messages: []StorageMessage{
				{
					ID: "a",
				},
				{
					ID: "b",
				},
				{
					ID: "c",
				},
			},
		}
		h := New(tmpDir, 0, storageHdlMock, nil, nil, 0)
		err := h.cleanup(t.Context())
		if err != nil {
			t.Error(err)
		}
		if storageHdlMock.DeleteMessagesC > 0 {
			t.Error("illegal call")
		}
		if len(storageHdlMock.Messages) != 3 {
			t.Error("invalid messages length")
		}
	})
}

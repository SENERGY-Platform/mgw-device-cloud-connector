package persistent_msg_relay_hdl

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
)

func TestHandler_Init(t *testing.T) {
	t.Run("with messages", func(t *testing.T) {
		storageHdlMock := &storageHandlerMock{
			Messages: []StorageMessage{
				{
					ID:      "a",
					DayHour: 1,
					Number:  1,
				},
			},
		}
		h := New("", 0, storageHdlMock, nil, nil, 0)
		err := h.Init(t.Context())
		if err != nil {
			t.Error(err)
		}
		if h.msgDayHour != 1 {
			t.Errorf("expected 1, got %d", h.msgDayHour)
		}
		if h.msgNumber != 1 {
			t.Errorf("expected 1, got %d", h.msgNumber)
		}
	})
	t.Run("no messages", func(t *testing.T) {
		h := New("", 0, &storageHandlerMock{}, nil, nil, 0)
		err := h.Init(t.Context())
		if err != nil {
			t.Error(err)
		}
		if h.msgDayHour != 0 {
			t.Errorf("expected 0, got %d", h.msgDayHour)
		}
		if h.msgNumber != 0 {
			t.Errorf("expected 0, got %d", h.msgNumber)
		}
	})
}

type storageHandlerMock struct {
	Messages          []StorageMessage
	Err               error
	DeleteMessagesErr error
	IsFull            bool
	MaxMsg            int
	DeleteMessagesC   int
	LimitVal          int
	Mu                sync.RWMutex
}

func (m *storageHandlerMock) CreateMessages(_ context.Context, msg []StorageMessage) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if m.IsFull || (m.MaxMsg > 0 && len(m.Messages) == m.MaxMsg) {
		return NewFullErr(errors.New("mock"))
	}
	m.Messages = append(m.Messages, msg...)
	return nil
}

func (m *storageHandlerMock) ReadMessages(_ context.Context, limit int) ([]StorageMessage, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	if m.Err != nil {
		return nil, m.Err
	}
	m.LimitVal = limit
	if len(m.Messages) == 0 {
		return nil, NoResultsErr
	}
	if len(m.Messages) < limit {
		limit = len(m.Messages)
	}
	var msg []StorageMessage
	for i := 0; i < limit; i++ {
		msg = append(msg, m.Messages[i])
	}
	return msg, nil
}

func (m *storageHandlerMock) DeleteMessages(_ context.Context, ids []string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.DeleteMessagesC++
	if m.Err != nil {
		return m.Err
	}
	if m.DeleteMessagesErr != nil {
		return m.DeleteMessagesErr
	}
	if len(m.Messages) == 0 {
		return nil
	}
	var messages []StorageMessage
	for _, msg := range m.Messages {
		if slices.Contains(ids, msg.ID) {
			continue
		}
		messages = append(messages, msg)
	}
	m.Messages = messages
	return nil
}

func (m *storageHandlerMock) DeleteFirstNMessages(_ context.Context, limit int) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if len(m.Messages) == 0 {
		return nil
	}
	if len(m.Messages) < limit {
		limit = len(m.Messages)
	}
	m.Messages = m.Messages[limit:]
	return nil
}

func (m *storageHandlerMock) LastPosition(_ context.Context) (dayHour int64, msgNum int64, err error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	if m.Err != nil {
		return 0, 0, m.Err
	}
	if len(m.Messages) == 0 {
		return 0, 0, NoResultsErr
	}
	msg := m.Messages[len(m.Messages)-1]
	return msg.DayHour, msg.Number, nil
}

func (m *storageHandlerMock) NoEntries(_ context.Context) (bool, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	if m.Err != nil {
		return false, m.Err
	}
	return len(m.Messages) == 0, nil
}

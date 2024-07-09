package message_hdl

import "time"

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

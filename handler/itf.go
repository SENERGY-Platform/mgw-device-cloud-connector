package handler

type MessageHandler func(m Message) (topic string, data []byte, err error)

type Message interface {
	Topic() string
	Payload() []byte
}

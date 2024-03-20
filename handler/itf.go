package handler

type Message interface {
	Topic() string
	Payload() []byte
}

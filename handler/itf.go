package handler

import "github.com/SENERGY-Platform/mgw-device-cloud-connector/model"

type MessageHandler func(m Message) (topic string, data []byte, err error)

type Message interface {
	Topic() string
	Payload() []byte
}

type MessageRelayHandler interface {
	Put(m Message) error
}

type CloudDeviceHandler interface {
	GetHubID() string
}

type LocalDeviceHandler interface {
	GetDevices() map[string]model.Device
}

type MqttClient interface {
	Subscribe(topic string, messageHandler func(m Message)) error
	Unsubscribe(topic string) error
	Publish(topic string, retained bool, payload any) error
}

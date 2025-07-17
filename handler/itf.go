package handler

import (
	"context"
	cm_models_service "github.com/SENERGY-Platform/mgw-cloud-proxy/cert-manager/lib/models/service"
	"time"
)

type MessageHandler func(m Message) (topic string, data []byte, err error)

type Message interface {
	Topic() string
	Payload() []byte
	Timestamp() time.Time
}

type MessageRelayHandler interface {
	Put(m Message) error
}

type MqttClient interface {
	Subscribe(topic string, qos byte, messageHandler func(m Message)) error
	Unsubscribe(topic string) error
	Publish(topic string, qos byte, retained bool, payload any) error
}

type CertManagerClient interface {
	NetworkInfo(ctx context.Context, token string) (cm_models_service.NetworkInfo, error)
}

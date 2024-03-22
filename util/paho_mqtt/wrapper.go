package paho_mqtt

import (
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Wrapper struct {
	client  mqtt.Client
	timeout time.Duration
}

func NewWrapper(client mqtt.Client, timeout time.Duration) *Wrapper {
	return &Wrapper{
		client:  client,
		timeout: timeout,
	}
}

func (w *Wrapper) Subscribe(topic string, qos byte, msgHandler func(m handler.Message)) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Subscribe(topic, qos, func(_ mqtt.Client, message mqtt.Message) {
		msgHandler(message)
	})
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *Wrapper) Unsubscribe(topic string) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Unsubscribe(topic)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *Wrapper) Publish(topic string, qos byte, retained bool, payload any) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Publish(topic, qos, retained, payload)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *Wrapper) Connect() mqtt.Token {
	return w.client.Connect()
}

func (w *Wrapper) Disconnect(quiesce uint) {
	w.client.Disconnect(quiesce)
}

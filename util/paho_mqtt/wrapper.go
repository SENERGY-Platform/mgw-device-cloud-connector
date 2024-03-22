package paho_mqtt

import (
	"errors"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

const (
	relayMsgErr    = "relaying message failed: topic=%s err=%s"
	subscribeErr   = "subscribing to '%s' failed: %s"
	unsubscribeErr = "unsubscribing from '%s' failed: %s"
)

type Wrapper struct {
	client  mqtt.Client
	qos     byte
	timeout time.Duration
}

func NewWrapper(client mqtt.Client, qos byte, timeout time.Duration) *Wrapper {
	return &Wrapper{
		client:  client,
		qos:     qos,
		timeout: timeout,
	}
}

func (w *Wrapper) Subscribe(topic string, callback mqtt.MessageHandler) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Subscribe(topic, w.qos, callback)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *Wrapper) Unsubscribe(topics ...string) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Unsubscribe(topics...)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *Wrapper) Publish(topic string, retained bool, payload any) error {
	if !w.client.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := w.client.Publish(topic, w.qos, retained, payload)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

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

type wrapper struct {
	qos     byte
	timeout time.Duration
}

func (w *wrapper) Subscribe(c mqtt.Client, topic string, callback mqtt.MessageHandler) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Subscribe(topic, w.qos, callback)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *wrapper) Unsubscribe(c mqtt.Client, topics ...string) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Unsubscribe(topics...)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func (w *wrapper) Publish(c mqtt.Client, topic string, retained bool, payload any) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Publish(topic, w.qos, retained, payload)
	if !t.WaitTimeout(w.timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

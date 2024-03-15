package paho_mqtt

import (
	"errors"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

func Subscribe(c mqtt.Client, timeout time.Duration, topic string, qos byte, callback mqtt.MessageHandler) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Subscribe(topic, qos, callback)
	if !t.WaitTimeout(timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func Unsubscribe(c mqtt.Client, timeout time.Duration, topics ...string) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Unsubscribe(topics...)
	if !t.WaitTimeout(timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

func Publish(c mqtt.Client, timeout time.Duration, topic string, qos byte, retained bool, payload any) error {
	if !c.IsConnectionOpen() {
		return errors.New("not connected")
	}
	t := c.Publish(topic, qos, retained, payload)
	if !t.WaitTimeout(timeout) {
		return errors.New("timeout")
	}
	return t.Error()
}

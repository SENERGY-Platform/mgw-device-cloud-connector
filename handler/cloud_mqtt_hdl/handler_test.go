package cloud_mqtt_hdl

import (
	"errors"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
	"testing"
)

func TestHandler_HandleOnDisconnect(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	h := New(0)
	h.SetMqttClient(nil)
	h.subscriptions = map[string]struct{}{
		"a": {},
		"b": {},
	}
	h.HandleOnDisconnect()
	if len(h.subscriptions) != 0 {
		t.Error("subscriptions should be empty")
	}
}

func TestHandler_HandleSubscriptions(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	topic.InitTopicHandler("userID", "netID")
	mc := &mockMqttClient{
		Topics: map[string]struct{}{
			topic.Handler.CloudProcessesCmdSub():        {},
			topic.Handler.CloudDeviceServiceCmdSub("a"): {},
			topic.Handler.CloudDeviceServiceCmdSub("b"): {},
			topic.Handler.CloudDeviceServiceCmdSub("c"): {},
			topic.Handler.CloudDeviceServiceCmdSub("d"): {},
			topic.Handler.CloudDeviceServiceCmdSub("x"): {},
		},
		QOS: 1,
		T:   t,
	}
	h := Handler{
		client: mc,
		subscriptions: map[string]struct{}{
			topic.Handler.CloudDeviceServiceCmdSub("a"): {},
			topic.Handler.CloudDeviceServiceCmdSub("b"): {},
			topic.Handler.CloudDeviceServiceCmdSub("x"): {},
		},
		qos: 1,
	}
	h.HandleSubscriptions(nil, map[string]model.Device{
		"a": {State: model.Online},
		"c": {State: model.Online},
		"x": {State: model.Offline},
	}, nil, nil, nil)
	if len(h.subscriptions) != 3 {
		t.Error("invalid length")
	}
	if _, ok := h.subscriptions[topic.Handler.CloudProcessesCmdSub()]; !ok {
		t.Errorf("%s not in map", topic.Handler.CloudProcessesCmdSub())
	}
	if _, ok := h.subscriptions[topic.Handler.CloudDeviceServiceCmdSub("a")]; !ok {
		t.Errorf("%s not in map", topic.Handler.CloudDeviceServiceCmdSub("a"))
	}
	if _, ok := h.subscriptions[topic.Handler.CloudDeviceServiceCmdSub("c")]; !ok {
		t.Errorf("%s not in map", topic.Handler.CloudDeviceServiceCmdSub("c"))
	}
	if mc.SubscribeC != 2 {
		t.Error("missing call")
	}
	if mc.UnsubscribeC != 2 {
		t.Error("missing call")
	}
	t.Run("error", func(t *testing.T) {
		mc.UnsubErr = errors.New("test error")
		h.HandleSubscriptions(nil, map[string]model.Device{
			"a": {State: model.Online},
			"d": {State: model.Online},
		}, nil, nil, nil)
		if len(h.subscriptions) != 4 {
			t.Error("invalid length")
		}
		if _, ok := h.subscriptions[topic.Handler.CloudDeviceServiceCmdSub("d")]; !ok {
			t.Errorf("%s not in map", topic.Handler.CloudDeviceServiceCmdSub("d"))
		}
	})
	t.Run("not connected", func(t *testing.T) {
		mc.UnsubErr = model.NotConnectedErr
		h.HandleSubscriptions(nil, map[string]model.Device{
			"a": {State: model.Online},
			"d": {State: model.Online},
			"e": {State: model.Online},
		}, nil, nil, nil)
		if len(h.subscriptions) != 4 {
			t.Error("invalid length")
		}
		if _, ok := h.subscriptions[topic.Handler.CloudDeviceServiceCmdSub("e")]; ok {
			t.Errorf("%s in map", topic.Handler.CloudDeviceServiceCmdSub("e"))
		}
	})
}

func TestHandler_subscribe(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("does not exist", func(t *testing.T) {
		mc := &mockMqttClient{
			Topics: map[string]struct{}{
				"test": {},
			},
			QOS: 1,
			T:   t,
		}
		h := Handler{
			client:        mc,
			subscriptions: make(map[string]struct{}),
			qos:           1,
		}
		err := h.subscribe("test", func(m handler.Message) {})
		if err != nil {
			t.Error()
		}
		if _, ok := h.subscriptions["test"]; !ok {
			t.Error("topic not in map")
		}
		if mc.SubscribeC != 1 {
			t.Error("missing call")
		}
	})
	t.Run("exists", func(t *testing.T) {
		mc := &mockMqttClient{
			Topics: map[string]struct{}{
				"test": {},
			},
			QOS: 1,
			T:   t,
		}
		h := Handler{
			client: mc,
			subscriptions: map[string]struct{}{
				"test": {},
			},
			qos: 1,
		}
		err := h.subscribe("test", func(m handler.Message) {})
		if err != nil {
			t.Error()
		}
		if _, ok := h.subscriptions["test"]; !ok {
			t.Error("topic not in map")
		}
		if mc.SubscribeC > 0 {
			t.Error("illegal call")
		}
	})
	t.Run("error", func(t *testing.T) {
		mc := &mockMqttClient{
			T:      t,
			SubErr: errors.New("test error"),
		}
		h := Handler{
			client:        mc,
			subscriptions: make(map[string]struct{}),
		}
		err := h.subscribe("test", func(m handler.Message) {})
		if err == nil {
			t.Error("expected error")
		}
		if _, ok := h.subscriptions["test"]; ok {
			t.Error("topic should not be in map")
		}
	})
}

func TestHandler_unsubscribe(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("does not exist", func(t *testing.T) {
		mc := &mockMqttClient{
			Topics: map[string]struct{}{
				"test": {},
			},
			T: t,
		}
		h := Handler{
			client:        mc,
			subscriptions: make(map[string]struct{}),
		}
		err := h.unsubscribe("test")
		if err != nil {
			t.Error()
		}
		if mc.UnsubscribeC > 0 {
			t.Error("illegal call")
		}
	})
	t.Run("exists", func(t *testing.T) {
		mc := &mockMqttClient{
			Topics: map[string]struct{}{
				"test": {},
			},
			QOS: 1,
			T:   t,
		}
		h := Handler{
			client: mc,
			subscriptions: map[string]struct{}{
				"test": {},
			},
			qos: 1,
		}
		err := h.unsubscribe("test")
		if err != nil {
			t.Error()
		}
		if _, ok := h.subscriptions["test"]; ok {
			t.Error("topic should not be in map")
		}
		if mc.UnsubscribeC != 1 {
			t.Error("missing call")
		}
	})
	t.Run("error", func(t *testing.T) {
		mc := &mockMqttClient{
			T:        t,
			UnsubErr: errors.New("test error"),
		}
		h := Handler{
			client: mc,
			subscriptions: map[string]struct{}{
				"test": {},
			},
		}
		err := h.unsubscribe("test")
		if err == nil {
			t.Error("expected error")
		}
		if _, ok := h.subscriptions["test"]; !ok {
			t.Error("topic should be in map")
		}
	})
}

func TestHandler_diffSubs(t *testing.T) {
	h := Handler{
		subscriptions: map[string]struct{}{
			"a": {},
			"b": {},
		},
	}
	missingTopics, newTopics := h.diffSubs(map[string]struct{}{"a": {}, "c": {}})
	if missingTopics[0] != "b" {
		t.Error("missing topic not in slice")
	}
	if newTopics[0] != "c" {
		t.Error("new topic not in slice")
	}
}

type mockMqttClient struct {
	SubErr       error
	UnsubErr     error
	Topics       map[string]struct{}
	QOS          byte
	SubscribeC   int
	UnsubscribeC int
	T            *testing.T
}

func (m *mockMqttClient) Subscribe(topic string, qos byte, messageHandler func(m handler.Message)) error {
	m.SubscribeC++
	if m.SubErr != nil {
		return m.SubErr
	}
	if _, ok := m.Topics[topic]; !ok {
		m.T.Error("invalid topic, ", topic)
	}
	if qos != m.QOS {
		m.T.Error("expected qos ", m.QOS, ", got ", qos)
	}
	return nil
}

func (m *mockMqttClient) Unsubscribe(topic string) error {
	m.UnsubscribeC++
	if m.UnsubErr != nil {
		return m.UnsubErr
	}
	if _, ok := m.Topics[topic]; !ok {
		m.T.Error("invalid topic, ", topic)
	}
	return nil
}

func (m *mockMqttClient) Publish(_ string, _ byte, _ bool, _ any) error {
	panic("not implemented")
}

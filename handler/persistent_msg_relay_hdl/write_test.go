package persistent_msg_relay_hdl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
)

func TestHandler_writer(t *testing.T) {
	t.Run("large", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{}
		h := New("", 100, stgHdlMock, nil, nil, 0)
		ctx, cf := context.WithCancel(t.Context())
		go h.writer(ctx)
		for i := 0; i < 100; i++ {
			err := h.Put(util.Message{
				TopicVal: fmt.Sprintf("%d", i),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		cf()
		<-h.writerDone
		if len(stgHdlMock.Messages) != 100 {
			t.Error("invalid messages length")
		}
		for i := 0; i < 100; i++ {
			if stgHdlMock.Messages[i].MsgTopic != fmt.Sprintf("%d", i) {
				t.Fatalf("wrong message order")
			}
		}
	})
	t.Run("small", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{}
		h := New("", 100, stgHdlMock, nil, nil, 0)
		ctx, cf := context.WithCancel(t.Context())
		go h.writer(ctx)
		for i := 0; i < 5; i++ {
			err := h.Put(util.Message{
				TopicVal: fmt.Sprintf("%d", i),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		cf()
		<-h.writerDone
		if len(stgHdlMock.Messages) != 5 {
			t.Error("invalid messages length")
		}
		for i := 0; i < 5; i++ {
			if stgHdlMock.Messages[i].MsgTopic != fmt.Sprintf("%d", i) {
				t.Fatalf("wrong message order")
			}
		}
	})
	t.Run("successive", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{}
		h := New("", 100, stgHdlMock, nil, nil, 0)
		ctx, cf := context.WithCancel(t.Context())
		go h.writer(ctx)
		for i := 0; i < 5; i++ {
			err := h.Put(util.Message{
				TopicVal: fmt.Sprintf("%d", i),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		for i := 5; i < 10; i++ {
			err := h.Put(util.Message{
				TopicVal: fmt.Sprintf("%d", i),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(time.Second)
		cf()
		<-h.writerDone
		if len(stgHdlMock.Messages) != 10 {
			t.Error("invalid messages length")
		}
		for i := 0; i < 10; i++ {
			if stgHdlMock.Messages[i].MsgTopic != fmt.Sprintf("%d", i) {
				t.Fatalf("wrong message order")
			}
		}
	})
}

func TestHandler_write(t *testing.T) {
	timeNow := time.Now()
	messages := []handler.Message{
		util.Message{
			TopicVal:   "a/b",
			PayloadVal: "test1",
			TimeVal:    timeNow,
		},
		util.Message{
			TopicVal:   "c/d",
			PayloadVal: "test2",
			TimeVal:    timeNow.Add(time.Second),
		},
	}
	t.Run("success", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{}
		h := New("", 0, stgHdlMock, nil, nil, 0)
		err := h.write(messages)
		if err != nil {
			t.Error(err)
		}
		if len(stgHdlMock.Messages) != 2 {
			t.Fatal("invalid messages length")
		}
		if stgHdlMock.Messages[0].DayHour == stgHdlMock.Messages[1].DayHour {
			if !(stgHdlMock.Messages[0].Number == 1 && stgHdlMock.Messages[1].Number == 2) {
				t.Error("invalid messages numbers")
			}
		} else {
			if !(stgHdlMock.Messages[0].Number == 1 && stgHdlMock.Messages[1].Number == 1) {
				t.Error("invalid messages numbers")
			}
		}
		for i, message := range stgHdlMock.Messages {
			if message.ID != fmt.Sprintf("%d_%d", message.DayHour, message.Number) {
				t.Errorf("expected %s, got %s", fmt.Sprintf("%d_%d", message.DayHour, message.Number), message.ID)
			}
			if message.MsgTopic != messages[i].Topic() {
				t.Errorf("expected %s, got %s", messages[i].Topic(), message.MsgTopic)
			}
			if !bytes.Equal(message.MsgPayload, messages[i].Payload()) {
				t.Errorf("expected %s, got %s", messages[i].Payload(), message.MsgPayload)
			}
			if message.MsgTimestamp != messages[i].Timestamp() {
				t.Errorf("expected %s, got %s", messages[i].Timestamp(), message.MsgTimestamp)
			}
		}
	})
	t.Run("exceeds size constraint", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{IsFull: true}
		h := New("", 0, stgHdlMock, nil, nil, 0)
		err := h.write(messages)
		if err == nil {
			t.Error("expected error")
		}
		if !errors.Is(err, ExceedsSizeConstraintErr) {
			t.Errorf("expected %T, got %T", ExceedsSizeConstraintErr, err)
		}
	})
	t.Run("storage full", func(t *testing.T) {
		stgHdlMock := &storageHandlerMock{
			Messages: []StorageMessage{
				{
					ID:      "a",
					DayHour: 1,
					Number:  1,
				},
				{
					ID:      "b",
					DayHour: 1,
					Number:  2,
				},
			},
			MaxMsg: 2,
		}
		h := New("", 0, stgHdlMock, nil, nil, 0)
		err := h.write(messages)
		if err != nil {
			t.Error(err)
		}
		if len(stgHdlMock.Messages) != 2 {
			t.Fatal("invalid messages length")
		}
		for i, message := range stgHdlMock.Messages {
			if message.ID != fmt.Sprintf("%d_%d", message.DayHour, message.Number) {
				t.Errorf("expected %s, got %s", fmt.Sprintf("%d_%d", message.DayHour, message.Number), message.ID)
			}
			if message.MsgTopic != messages[i].Topic() {
				t.Errorf("expected %s, got %s", messages[i].Topic(), message.MsgTopic)
			}
			if !bytes.Equal(message.MsgPayload, messages[i].Payload()) {
				t.Errorf("expected %s, got %s", messages[i].Payload(), message.MsgPayload)
			}
			if message.MsgTimestamp != messages[i].Timestamp() {
				t.Errorf("expected %s, got %s", messages[i].Timestamp(), message.MsgTimestamp)
			}
		}
	})
}

func TestHandler_advanceDayHourAndNumber(t *testing.T) {
	h := New("", 0, nil, nil, nil, 0)
	h.advanceDayHourAndNumber()
	if h.msgDayHour == 0 {
		t.Error("zero value")
	}
	if h.msgNumber != 1 {
		t.Errorf("expected 1, got %d", h.msgNumber)
	}
}

func Test_getDayHourUnix(t *testing.T) {
	tNow := time.Now().UTC()
	a := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), tNow.Hour(), 0, 0, 0, time.UTC).Unix()
	b := getDayHourUnix(tNow)
	if a != b {
		t.Errorf("expected %d, got %d", a, b)
	}
}

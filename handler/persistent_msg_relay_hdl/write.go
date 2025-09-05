package persistent_msg_relay_hdl

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"time"
)

func (h *Handler) writer(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	loop := true
	var messages []handler.Message
	var err error
	for loop {
		select {
		case <-ctx.Done():
			loop = false
			break
		case <-ticker.C:
			err = h.write(messages)
			if err != nil {
				util.Logger.Errorf("%s write messages: %s", logPrefix, err)
			}
			messages = nil
		case msg := <-h.messages:
			messages = append(messages, msg)
			if len(messages) >= 100 {
				err = h.write(messages)
				if err != nil {
					util.Logger.Errorf("%s write messages: %s", logPrefix, err)
				}
				messages = nil
			}
		}
	}
	err = h.write(messages)
	if err != nil {
		util.Logger.Errorf("%s write messages: %s", logPrefix, err)
	}
	h.writerDone <- struct{}{}
}

func (h *Handler) write(messages []handler.Message) error {
	messagesLen := len(messages)
	if messagesLen == 0 {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	storageMessages := make([]StorageMessage, 0, messagesLen)
	for _, message := range messages {
		h.advanceDayHourAndNumber()
		storageMessages = append(storageMessages, newStorageMessage(h.msgDayHour, h.msgNumber, message))
	}
	for {
		err := h.storageHdl.CreateMessages(h.writerCtx, storageMessages)
		if err != nil {
			if !errors.Is(err, &FullErr{}) {
				return err
			}
			ok, err := h.storageHdl.NoEntries(h.writerCtx)
			if err != nil {
				return err
			}
			if ok {
				return ExceedsSizeConstraintErr
			}
			err = h.storageHdl.DeleteFirstNMessages(h.writerCtx, len(messages))
			if err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func (h *Handler) advanceDayHourAndNumber() {
	dayHour := getDayHourUnix()
	if h.msgDayHour != dayHour {
		h.msgDayHour = dayHour
		h.msgNumber = 0
	}
	h.msgNumber++
}

func newStorageMessage(dayHour, number int64, m handler.Message) StorageMessage {
	return StorageMessage{
		ID:           fmt.Sprintf("%d_%d", dayHour, number),
		DayHour:      dayHour,
		Number:       number,
		MsgTopic:     m.Topic(),
		MsgPayload:   m.Payload(),
		MsgTimestamp: m.Timestamp(),
	}
}

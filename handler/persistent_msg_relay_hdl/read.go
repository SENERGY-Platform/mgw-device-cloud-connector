package persistent_msg_relay_hdl

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"time"
)

func (h *Handler) reader(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()
	loop := true
	var err error
	for loop {
		select {
		case <-ticker.C:
			err = h.read(ctx)
			if err != nil {
				loop = false
				break
			}
		case <-ctx.Done():
			loop = false
			break
		}
	}
	h.readerDone <- struct{}{}
	if err != nil {
		h.errorStateMu.Lock()
		defer h.errorStateMu.Unlock()
		h.errorState = true
	}
}

func (h *Handler) read(ctx context.Context) error {
	messages, err := h.storageHdl.ReadMessages(ctx, h.limit)
	if err != nil && !errors.Is(err, NoResultsErr) {
		util.Logger.Errorf("%s read messages: %s", logPrefix, err)
		return err
	}
	messagesLen := len(messages)
	if messagesLen == 0 {
		return nil
	}
	sentMsgIDs := h.sendMessages(ctx, messages, messagesLen)
	if len(sentMsgIDs) > 0 {
		dErr := h.storageHdl.DeleteMessages(context.Background(), sentMsgIDs)
		if dErr != nil {
			util.Logger.Errorf("%s delete messages: %s", logPrefix, dErr)
			cErr := h.createCleanupFile(sentMsgIDs)
			if cErr != nil {
				util.Logger.Errorf("%s create cleanup file: %s", logPrefix, cErr)
				return errors.Join(dErr, cErr)
			}
			return dErr
		}
	}
	return nil
}

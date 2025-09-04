package persistent_msg_relay_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"sync"
)

func (h *Handler) sendMessages(ctx context.Context, messages []StorageMessage, messagesLen int) []string {
	msgMap := make(map[string][]StorageMessage)
	for _, message := range messages {
		msgMap[message.MsgTopic] = append(msgMap[message.MsgTopic], message)
	}
	var sentMsgIDs []string
	if len(msgMap) > 1 && messagesLen > 20 {
		syncSl := &syncSlice{}
		wg := &sync.WaitGroup{}
		for _, msgSl := range msgMap {
			wg.Add(1)
			go h.sendRoutine(ctx, wg, msgSl, syncSl)
		}
		wg.Wait()
		sentMsgIDs = syncSl.Values()
	} else {
		var err error
		sentMsgIDs, err = h.send(ctx, messages)
		if err != nil {
			util.Logger.Errorf("%s send messages: %s", logPrefix, err)
		}
	}
	return sentMsgIDs
}

func (h *Handler) sendRoutine(ctx context.Context, wg *sync.WaitGroup, messages []StorageMessage, syncSl *syncSlice) {
	defer wg.Done()
	msgIDs, err := h.send(ctx, messages)
	if err != nil {
		util.Logger.Errorf("%s send messages: %s", logPrefix, err)
	}
	syncSl.Append(msgIDs...)
}

func (h *Handler) send(ctx context.Context, messages []StorageMessage) ([]string, error) {
	var sentMsgIDs []string
	for _, message := range messages {
		if ctx.Err() != nil {
			return sentMsgIDs, ctx.Err()
		}
		topic, data, err := h.handleFunc(message)
		if err != nil {
			if err != model.NoMsgErr {
				util.Logger.Errorf("%s handle message: %s", logPrefix, err)
			}
			continue
		}
		if err = h.sendFunc(topic, data); err != nil {
			return sentMsgIDs, err
		}
		sentMsgIDs = append(sentMsgIDs, message.ID)
	}
	return sentMsgIDs, nil
}

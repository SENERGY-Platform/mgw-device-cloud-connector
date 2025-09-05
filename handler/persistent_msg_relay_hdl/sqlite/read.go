package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/persistent_msg_relay_hdl"
	"time"
)

const readMessagesStmt = "SELECT id, day_hour, msg_num, topic, payload, msg_time FROM messages ORDER BY day_hour ASC, msg_num ASC LIMIT %d;"

func (h *Handler) ReadMessages(ctx context.Context, limit int) ([]persistent_msg_relay_hdl.StorageMessage, error) {
	h.rwMu.RLock()
	defer h.rwMu.RUnlock()
	rows, err := h.db.QueryContext(ctx, fmt.Sprintf(readMessagesStmt, limit))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, persistent_msg_relay_hdl.NoResultsErr
		}
		return nil, err
	}
	defer rows.Close()
	var messages []persistent_msg_relay_hdl.StorageMessage
	for rows.Next() {
		var msg persistent_msg_relay_hdl.StorageMessage
		var receivedTime int64
		err = rows.Scan(&msg.ID, &msg.DayHour, &msg.Number, &msg.MsgTopic, &msg.MsgPayload, &receivedTime)
		if err != nil {
			return nil, err
		}
		msg.MsgTimestamp = time.Unix(0, receivedTime)
		messages = append(messages, msg)
	}
	return messages, nil
}

const lastPositionStmt = "SELECT day_hour, msg_num FROM messages ORDER BY day_hour DESC, msg_num DESC LIMIT 1;"

func (h *Handler) LastPosition(ctx context.Context) (dayHour int64, msgNum int64, err error) {
	h.rwMu.RLock()
	defer h.rwMu.RUnlock()
	row := h.db.QueryRowContext(ctx, lastPositionStmt)
	if err = row.Scan(&dayHour, &msgNum); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = persistent_msg_relay_hdl.NoResultsErr
		}
		return
	}
	return
}

const countRowsStmt = "SELECT count(*) FROM messages;"

func (h *Handler) NoEntries(ctx context.Context) (bool, error) {
	h.rwMu.RLock()
	defer h.rwMu.RUnlock()
	row := h.db.QueryRowContext(ctx, countRowsStmt)
	var c int64
	if err := row.Scan(&c); err != nil {
		return false, err
	}
	return c == 0, nil
}

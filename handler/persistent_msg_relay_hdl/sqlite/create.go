package sqlite

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler/persistent_msg_relay_hdl"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const insertMessageStmt = "INSERT INTO messages (id, day_hour, msg_num, topic, payload, msg_time) VALUES (?, ?, ?, ?, ?, ?);"

func (h *Handler) CreateMessages(ctx context.Context, messages []persistent_msg_relay_hdl.StorageMessage) error {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	if len(messages) == 1 {
		msg := messages[0]
		_, err := h.db.ExecContext(ctx, insertMessageStmt, msg.ID, msg.DayHour, msg.Number, msg.MsgTopic, msg.MsgPayload, msg.MsgTimestamp.UnixNano())
		if err != nil {
			return handleCreateMessageErr(err)
		}
		return nil
	}
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, msg := range messages {
		_, err = tx.ExecContext(ctx, insertMessageStmt, msg.ID, msg.DayHour, msg.Number, msg.MsgTopic, msg.MsgPayload, msg.MsgTimestamp.UnixNano())
		if err != nil {
			return handleCreateMessageErr(err)
		}
	}
	return tx.Commit()
}

func handleCreateMessageErr(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_FULL {
		return persistent_msg_relay_hdl.NewFullErr(err)
	}
	return err
}

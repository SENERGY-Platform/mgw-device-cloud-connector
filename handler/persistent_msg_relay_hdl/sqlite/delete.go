package sqlite

import (
	"context"
	"fmt"
)

const (
	deleteMessageStmt  = "DELETE FROM messages WHERE id = ?"
	deleteMessagesStmt = "DELETE FROM messages WHERE id IN (%s);"
)

func (h *Handler) DeleteMessages(ctx context.Context, ids []string) error {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	lenIDs := len(ids)
	if lenIDs == 1 {
		_, err := h.db.ExecContext(ctx, deleteMessageStmt, ids[0])
		if err != nil {
			return err
		}
		return nil
	}
	anySl := make([]any, lenIDs)
	for i, id := range ids {
		anySl[i] = id
	}
	_, err := h.db.ExecContext(ctx, fmt.Sprintf(deleteMessagesStmt, genQuestionMarks(lenIDs)), anySl...)
	if err != nil {
		return err
	}
	return nil
}

const deleteFirstNMessagesStmt = "DELETE FROM messages WHERE id IN (SELECT id FROM messages ORDER BY day_hour ASC, msg_num ASC LIMIT %d);"

func (h *Handler) DeleteFirstNMessages(ctx context.Context, limit int) error {
	h.rwMu.Lock()
	defer h.rwMu.Unlock()
	_, err := h.db.ExecContext(ctx, fmt.Sprintf(deleteFirstNMessagesStmt, limit))
	if err != nil {
		return err
	}
	return nil
}

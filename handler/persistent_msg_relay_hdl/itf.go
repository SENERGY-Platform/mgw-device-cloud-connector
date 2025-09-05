package persistent_msg_relay_hdl

import (
	"context"
)

type storageHandler interface {
	CreateMessages(ctx context.Context, msg []StorageMessage) error
	ReadMessages(ctx context.Context, limit int) ([]StorageMessage, error)
	DeleteMessages(ctx context.Context, ids []string) error
	DeleteFirstNMessages(ctx context.Context, limit int) error
	LastPosition(ctx context.Context) (dayHour int64, msgNum int64, err error)
	NoEntries(ctx context.Context) (bool, error)
}

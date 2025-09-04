package persistent_msg_relay_hdl

import (
	"errors"
	"time"
)

type StorageMessage struct {
	ID           string
	DayHour      int64
	Number       int64
	MsgTopic     string
	MsgPayload   []byte
	MsgTimestamp time.Time
}

func (s StorageMessage) Topic() string {
	return s.MsgTopic
}

func (s StorageMessage) Payload() []byte {
	return s.MsgPayload
}

func (s StorageMessage) Timestamp() time.Time {
	return s.MsgTimestamp
}

type FullErr struct {
	err error
}

func NewFullErr(err error) *FullErr {
	return &FullErr{err: err}
}

func (e *FullErr) Error() string {
	return e.err.Error()
}

func (e *FullErr) Unwrap() error {
	return e.err
}

func (e *FullErr) Is(target error) bool {
	_, ok := target.(*FullErr)
	return ok
}

var ExceedsSizeConstraintErr = errors.New("exceeds size constraint")

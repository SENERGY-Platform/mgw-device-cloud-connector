package model

import "errors"

var NoMsgErr = errors.New("")

var NotConnectedErr = errors.New("not connected")

var OperationTimeoutErr = errors.New("waiting for operation timed out, might still succeed in background")

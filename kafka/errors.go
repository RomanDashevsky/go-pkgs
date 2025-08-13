package kafka

import "errors"

// Common errors
var (
	ErrTimeout        = errors.New("kafka operation timeout")
	ErrConnectionLost = errors.New("kafka connection lost")
	ErrBadHandler     = errors.New("kafka unknown handler")
	ErrInternalServer = errors.New("kafka internal server error")
	ErrInvalidTopic   = errors.New("kafka invalid topic")
	ErrInvalidMessage = errors.New("kafka invalid message")
)

// Status constants for message processing
const (
	Success = "success"
	Failed  = "failed"
)

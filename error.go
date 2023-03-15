package buffer

import "errors"

var (
	// ErrClosed indicates the buffer is closed and can no longer be used.
	ErrClosed = errors.New("buffer is closed")
)

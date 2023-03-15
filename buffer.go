package buffer

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
)

// Buffer is lock-free
// It would start a background goroutine continuously consume data from channel and write to container.
// When container is "full", it would execute flush on container synchronously or asynchronously based on `isFlushAsync`.
//
//	@author kevineluo
//	@update 2023-03-15 09:40:02
type Buffer[T any] struct {
	Config

	container Container[T]

	// used to free context
	context context.Context
	cancel  context.CancelFunc
	// free lock for async putting data in container
	dataChan chan T
	// used to disable insertion into buffer
	isClosed *atomic.Bool
}

// Config Buffer Config
//
//	@author kevineluo
//	@update 2023-03-15 09:31:54
type Config struct {
	chanBufSize   int           // lock free channel size(when data in channel reach chanBufSize, it will block in Buffer.Put)
	flushInterval time.Duration // automate flush data every [flushInterval] duration

	logger *logr.Logger // logger implement logr.LogSinker
}

// Check check config and set default value
//
//	@receiver config *Config
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:11:28
func (config *Config) Check() (err error) {
	if config.chanBufSize == 0 {
		config.chanBufSize = 100
	}
	if config.flushInterval == 0 {
		config.flushInterval = 15 * time.Second
	}
	if config.logger == nil {
		logger := logr.Discard()
		config.logger = &logger
	}
	return
}

// NewBuffer return buffer in type `T`
func NewBuffer[T any](container Container[T], config Config) (buffer *Buffer[T], errChan <-chan error, err error) {
	err = config.Check()
	if err != nil {
		return
	}
	isClosed := new(atomic.Bool)
	isClosed.Store(false)
	ctx, cancel := context.WithCancel(context.Background())
	buffer = &Buffer[T]{
		Config:    config,
		container: container,
		context:   ctx,
		cancel:    cancel,
		dataChan:  make(chan T, config.chanBufSize),
		isClosed:  isClosed,
	}
	// error channel with size 1 to avoid block
	originErrChan := make(chan error, 1)
	go func(errChan chan<- error, atMost time.Duration) {
		buffer.logger.Info("buffer start writer goroutine")
		ticker := time.NewTicker(atMost)
		for {
			select {
			case data := <-buffer.dataChan:
				if err := putAndCheck(buffer, data); err != nil {
					errChan <- err
					buffer.container.reset()
				}
			case <-ticker.C:
				if err := buffer.container.execute(); err != nil {
					errChan <- err
					buffer.container.reset()
				}
			case <-buffer.context.Done():
				buffer.logger.Info("buffer are closing, shut down writer goroutine")
				return
			}
		}
	}(originErrChan, config.flushInterval)

	errChan = originErrChan

	return
}

// Put element into buffer synchronously
func (buffer *Buffer[T]) Put(data T) error {
	if buffer.isClosed.Load() {
		return errors.New("buffer is closed")
	}
	buffer.dataChan <- data
	return nil
}

// Close would gracefully shut down the buffer.
// 1. wait channel to be empty and close channel
// 2. flush data in container synchronously
func (buffer *Buffer[T]) Close() error {
	buffer.isClosed.Store(true)
	// wait until buffer is empty
	for len(buffer.dataChan) != 0 {
	}
	// shutdown goroutine
	buffer.cancel()
	close(buffer.dataChan)
	return buffer.container.execute()
}

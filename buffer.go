// Package buffer represents a buffer that asynchronously flushes its contents.
// It is useful for applications that need to aggregate data before writing it to an external storage.
// A buffer can be flushed manually, or automatically when it becomes full or after an interval has elapsed, whichever comes first.
//
//	@update 2023-03-15 10:40:31
package buffer

import (
	"context"
	"time"
)

type void struct{}

// Buffer is lock-free
// It would start a background goroutine continuously consume data from channel and write to container.
// When container is "full", it would execute flush on container synchronously or asynchronously based on `isFlushAsync`.
//
//	@author kevineluo
//	@update 2023-03-15 09:40:02
type Buffer[T any] struct {
	Config

	container Container[T] // hold data in buffer, implement Container interface

	context context.Context
	cancel  context.CancelFunc // used to send close buffer signal

	dataChan          chan T       // free lock for async putting data in container
	flushSignalChan   chan void    // channel for flush data signal
	cleanedSignalChan chan void    // channel for buffer cleaned up signal
	errChan           chan<- error // channel for sending error to buffer user
}

// NewBuffer return buffer in type `T`, and start handling data
//
//	@param container Container[T]
//	@param config Config
//	@return buffer *Buffer[T]
//	@return errChan <-chan error
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:38:29
func NewBuffer[T any](container Container[T], config Config) (buffer *Buffer[T], errChan <-chan error, err error) {
	err = config.Check()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	// error channel with size 1 to avoid block
	originErrChan := make(chan error, 1)
	errChan = originErrChan
	buffer = &Buffer[T]{
		Config:            config,
		container:         container,
		context:           ctx,
		cancel:            cancel,
		dataChan:          make(chan T, config.chanBufSize),
		flushSignalChan:   make(chan void),
		cleanedSignalChan: make(chan void),
		errChan:           originErrChan,
	}

	// start
	go buffer.run()

	return
}

// Put put data into buffer async
//
//	@param buffer *Buffer[T]
//	@return Put
//	@author kevineluo
//	@update 2023-03-15 11:09:25
func (buffer *Buffer[T]) Put(data T) error {
	if buffer.closed() {
		return ErrClosed
	}
	buffer.dataChan <- data
	return nil
}

// Flush manually flush the buffer
//
//	@param buffer *Buffer[T]
//	@return Flush
//	@author kevineluo
//	@update 2023-03-15 11:01:42
func (buffer *Buffer[T]) Flush() error {
	if buffer.closed() {
		return ErrClosed
	}
	buffer.flushSignalChan <- void{}
	return nil
}

// Close would gracefully shut down the buffer.
//
//	@param buffer *Buffer[T]
//	@return Close
//	@author kevineluo
//	@update 2023-03-16 11:12:33
func (buffer *Buffer[T]) Close() error {
	if buffer.closed() {
		return ErrClosed
	}

	// call cancel func to prevent buffer.Put, buffer.Flush and buffer.Close
	buffer.cancel()

	// wait till buffer cleaned up(buffer.cleanedSignalChan be closed by buffer.run goroutine)
	<-buffer.cleanedSignalChan
	close(buffer.dataChan)
	close(buffer.flushSignalChan)
	return nil
}

// run start handling data
//
//	@param buffer *Buffer[T]
//	@return run
//	@author kevineluo
//	@update 2023-03-15 10:34:29
func (buffer *Buffer[T]) run() {
	buffer.logger.Info("start buffer writer goroutine")

	ticker := time.NewTicker(buffer.flushInterval)
	defer ticker.Stop()

	// send buffer cleaned up signal
	defer close(buffer.cleanedSignalChan)

	for {
		select {
		case data := <-buffer.dataChan:
			// receive one piece of data
			putAndCheck(buffer, data)
		case <-ticker.C:
			// automate flush buffer
			ticker.Stop()
			if err := buffer.container.flush(); err != nil {
				buffer.errChan <- err
				buffer.container.reset()
			}
			ticker.Reset(buffer.flushInterval)
		case <-buffer.context.Done():
			// receive buffer close signal, clean up buffer and return
			if err := buffer.container.flush(); err != nil {
				buffer.errChan <- err
				buffer.container.reset()
			}
			return
		case <-buffer.flushSignalChan:
			// manually flush buffer
			if err := buffer.container.flush(); err != nil {
				buffer.errChan <- err
				buffer.container.reset()
			}
		}
	}
}

// closed check if the Buffer is closed
//
//	@param buffer *Buffer[T]
//	@return closed
//	@author kevineluo
//	@update 2023-03-15 11:09:03
func (buffer *Buffer[T]) closed() bool {
	select {
	case <-buffer.context.Done():
		return true
	default:
		return false
	}
}

// Package buffer represents a Generic buffer that asynchronously flushes its contents.
// It is useful for applications that need to aggregate data before writing it to an external storage.
// A buffer can be flushed manually, or automatically when it becomes full or after an interval has elapsed, whichever comes first.
//
//	@update 2023-03-15 10:40:31
package buffer

import (
	"context"
	"time"

	"github.com/Kevinello/go-buffer/container"
)

type (
	void        struct{}
	flushSignal struct {
		async bool        // determine the flush is async or not
		done  chan<- void // for receiving flush done signal
	}
)

// Buffer is a lock-free buffer
// It would start a background goroutine continuously consume data from channel and write to container
// When container is "full", it would asynchronously flush data on container by default
//
//	@author kevineluo
//	@update 2023-03-15 09:40:02
type Buffer[T any] struct {
	Config

	container container.Container[T] // hold data in buffer, implement Container interface

	context context.Context
	cancel  context.CancelFunc // used to send close buffer signal

	autoFlushTicker *time.Ticker      // ticker for automate flush data
	dataChan        chan T            // free lock for async putting data in container
	flushSignalChan chan *flushSignal // channel for flush data signal
	errChan         chan<- error      // channel for sending error to buffer user
}

// NewBuffer creates a buffer in type `T`, and start handling data
// return the buffer and a error channel for user to handle error from putting data and flushing data
//
//	@param ctx context.Context
//	@param container container.Container[T]
//	@param config Config
//	@return buffer *Buffer[T]
//	@return errChan <-chan error
//	@return err error
//	@author kevineluo
//	@update 2023-03-30 01:58:14
func NewBuffer[T any](ctx context.Context, container container.Container[T], config Config) (buffer *Buffer[T], errChan <-chan error, err error) {
	if err = config.Validate(); err != nil {
		return
	}
	subCtx, cancel := context.WithCancel(ctx)
	buffer = &Buffer[T]{
		Config:          config,
		container:       container,
		context:         subCtx,
		cancel:          cancel,
		dataChan:        make(chan T, config.ChanBufSize),
		flushSignalChan: make(chan *flushSignal),
		errChan:         make(chan error, 1), // error channel with size 1 to avoid block
	}

	// wait for context cancellation
	go buffer.cleanup()

	// active buffer
	go buffer.run()

	return
}

// Put put data into buffer asynchronously
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
//	@receiver buffer *Buffer
//	@param async bool
//	@return error
//	@author kevineluo
//	@update 2023-03-27 02:00:56
func (buffer *Buffer[T]) Flush(async bool) error {
	if buffer.closed() {
		return ErrClosed
	}

	if !async {
		// synchronously flush
		done := make(chan void)
		buffer.flushSignalChan <- &flushSignal{
			async: async,
			done:  done,
		}
		// block til flush done
		<-done
	} else {
		// asynchronously flush
		buffer.flushSignalChan <- &flushSignal{async: async}
	}
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

	// call cancel func to prevent buffer.Put, buffer.Flush and buffer.Close, and start cleanup
	buffer.cancel()

	return nil
}

// run start handling data
//
//	@param buffer *Buffer[T]
//	@return run
//	@author kevineluo
//	@update 2023-03-15 10:34:29
func (buffer *Buffer[T]) run() {
	defer close(buffer.errChan)
	buffer.Logger.Info("buffer start handling data", "ID", buffer.ID)

	buffer.autoFlushTicker = time.NewTicker(buffer.FlushInterval)
	if buffer.DisableAutoFlush {
		buffer.autoFlushTicker.Stop()
	} else {
		defer buffer.autoFlushTicker.Stop()
	}

	// main signal monitoring loop for safe buffer life cycle
	for {
		select {
		case <-buffer.context.Done():
			// receive buffer close signal, stop running
			buffer.Logger.Info("[Buffer.run] receive buffer close signal, stop running", "time", time.Now().Format(time.DateTime))
			return
		case data := <-buffer.dataChan:
			// receive one piece of data
			buffer.putAndCheck(data)
		case <-buffer.autoFlushTicker.C:
			// automate flush buffer(will temporarily stop the timer)
			buffer.Logger.Info("[Buffer.run] tick for automate flush data reach, will call container.Flush")
			buffer.autoFlushTicker.Stop()
			if buffer.SyncAutoFlush {
				if err := buffer.container.Flush(); err != nil {
					buffer.Logger.Error(err, "[Buffer.run] error when call Container.Flush")
					buffer.errChan <- err
					buffer.container.Reset()
				}
			} else {
				go func() {
					if err := buffer.container.Flush(); err != nil {
						buffer.Logger.Error(err, "[Buffer.run] error when call Container.Flush")
						buffer.errChan <- err
						buffer.container.Reset()
					}
				}()
			}
			buffer.autoFlushTicker.Reset(buffer.FlushInterval)
		case flushSignal := <-buffer.flushSignalChan:
			// manually flush buffer
			if err := buffer.container.Flush(); err != nil {
				buffer.Logger.Error(err, "[Buffer.run] error when call Container.Flush")
				buffer.errChan <- err
				buffer.container.Reset()
			}
			if !flushSignal.async {
				// send flush done signal for synchronously flush
				flushSignal.done <- void{}
			}
		}
	}
}

func (buffer *Buffer[T]) cleanup() {
	<-buffer.context.Done()
	// receive buffer close signal, clean up buffer and return
	// clean dataChan(there is no more data send into dataChan), and close all channels
	for {
		select {
		case data := <-buffer.dataChan:
			// receive one piece of data
			buffer.putAndCheck(data)
		default:
			// call last flush
			if err := buffer.container.Flush(); err != nil {
				buffer.Logger.Error(err, "[Buffer.cleanup] error when call Container.Flush")
			}
			close(buffer.dataChan)
			close(buffer.flushSignalChan)
			return
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

// putAndCheck put a piece of data into container and flush container when full
//
//	@param buffer *Buffer[T]
//	@param data T
//	@return error
//	@author kevineluo
//	@update 2023-03-15 09:46:37
func (buffer *Buffer[T]) putAndCheck(data T) {
	if err := buffer.container.Put(data); err != nil {
		buffer.Logger.Error(err, "[Buffer.putAndCheck] buffer cannot write message to container")
		buffer.errChan <- err
	}

	if buffer.container.IsFull() {
		buffer.Logger.Info("[Buffer.putAndCheck] buffer if full, will call container.Flush")
		buffer.autoFlushTicker.Stop()
		if buffer.SyncAutoFlush {
			if err := buffer.container.Flush(); err != nil {
				buffer.Logger.Error(err, "[Buffer.putAndCheck] error when call Container.Flush")
				buffer.errChan <- err
				buffer.container.Reset()
			}
		} else {
			go func() {
				if err := buffer.container.Flush(); err != nil {
					buffer.Logger.Error(err, "[Buffer.putAndCheck] error when call Container.Flush")
					buffer.errChan <- err
					buffer.container.Reset()
				}
			}()
		}
		buffer.autoFlushTicker.Reset(buffer.FlushInterval)
	}
}

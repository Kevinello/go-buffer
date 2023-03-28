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

	ticker            *time.Ticker      // ticker for automate flush data
	dataChan          chan T            // free lock for async putting data in container
	flushSignalChan   chan *flushSignal // channel for flush data signal
	cleanedSignalChan chan void         // channel for buffer cleaned up signal
	errChan           chan<- error      // channel for sending error to buffer user
}

// NewBuffer creates a buffer in type `T`, and start handling data
// return the buffer and a error channel for user to handle error from putting data and flushing data
//
//	@param container Container[T]
//	@param config Config
//	@return buffer *Buffer[T]
//	@return errChan <-chan error
//	@return err error
//	@author kevineluo
//	@update 2023-03-15 10:38:29
func NewBuffer[T any](container container.Container[T], config Config) (buffer *Buffer[T], errChan <-chan error, err error) {
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
		flushSignalChan:   make(chan *flushSignal),
		cleanedSignalChan: make(chan void),
		errChan:           originErrChan,
	}

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
	buffer.logger.Info("buffer start handling data", "ID", buffer.ID)

	buffer.ticker = time.NewTicker(buffer.flushInterval)
	if buffer.disableAutoFlush {
		buffer.ticker.Stop()
	} else {
		defer buffer.ticker.Stop()
	}

	// send buffer cleaned up signal
	defer close(buffer.cleanedSignalChan)

	// main signal monitoring loop for safe buffer life cycle
	for {
		select {
		case data := <-buffer.dataChan:
			// receive one piece of data
			putAndCheck(buffer, data)
		case <-buffer.ticker.C:
			// automate flush buffer(will temporarily stop the timer)
			buffer.logger.Info("tick for automate flush data reach, will call container.Flush")
			buffer.ticker.Stop()
			if buffer.syncAutoFlush {
				if err := buffer.container.Flush(); err != nil {
					buffer.logger.Error(err, "error when call Container.Flush")
					buffer.errChan <- err
					buffer.container.Reset()
				}
			} else {
				go func() {
					if err := buffer.container.Flush(); err != nil {
						buffer.logger.Error(err, "error when call Container.Flush")
						buffer.errChan <- err
						buffer.container.Reset()
					}
				}()
			}
			buffer.ticker.Reset(buffer.flushInterval)
		case <-buffer.context.Done():
			// receive buffer close signal, clean up buffer and return
			// firstly, clean dataChan(there is no more data send into dataChan)
			for {
				select {
				case data := <-buffer.dataChan:
					// receive one piece of data
					putAndCheck(buffer, data)
				default:
					// call last flush
					if err := buffer.container.Flush(); err != nil {
						buffer.logger.Error(err, "error when call Container.Flush")
						buffer.errChan <- err
						buffer.container.Reset()
					}
					return
				}
			}
		case flushSignal := <-buffer.flushSignalChan:
			// manually flush buffer
			if err := buffer.container.Flush(); err != nil {
				buffer.logger.Error(err, "error when call Container.Flush")
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
func putAndCheck[T any](buffer *Buffer[T], data T) {
	if err := buffer.container.Put(data); err != nil {
		buffer.logger.Error(err, "buffer cannot write message to container")
		buffer.errChan <- err
	}

	if buffer.container.IsFull() {
		buffer.logger.Info("buffer if full, will call container.Flush")
		buffer.ticker.Stop()
		if buffer.syncAutoFlush {
			if err := buffer.container.Flush(); err != nil {
				buffer.logger.Error(err, "error when call Container.Flush")
				buffer.errChan <- err
				buffer.container.Reset()
			}
		} else {
			go func() {
				if err := buffer.container.Flush(); err != nil {
					buffer.logger.Error(err, "error when call Container.Flush")
					buffer.errChan <- err
					buffer.container.Reset()
				}
			}()
		}
		buffer.ticker.Reset(buffer.flushInterval)
	}
}

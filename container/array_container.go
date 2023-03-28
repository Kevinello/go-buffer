// Package container implementation of some Container interfaces
//
//	@update 2023-03-26 03:57:31
package container

import (
	"fmt"
	"log"
)

var _ Container[int] = &ArrayContainer[int]{}

// ArrayContainer not thread safe
//
//	@author kevineluo
//	@update 2023-03-25 03:51:33
type ArrayContainer[T any] struct {
	flushAsync bool                  // enable async flush, default is false
	flushBatch func(array []T) error // custom flush buffer function
	array      []T                   // slice holding data
	flushSize  int                   // determine the flush size
}

// NewArrayContainer new an ArrayContainer
//
//	@param size int
//	@param flushAsync bool
//	@param flushBatch func(array []T) error
//	@return *ArrayContainer
//	@author kevineluo
//	@update 2023-03-26 03:46:01
func NewArrayContainer[T any](flushSize int, flushAsync bool, flushBatch func(array []T) error) *ArrayContainer[T] {
	return &ArrayContainer[T]{
		flushAsync: flushAsync,
		flushBatch: flushBatch,
		array:      make([]T, 0, flushSize),
		flushSize:  flushSize,
	}
}

// Put implement interface Container
//
//	@param container *ArrayContainer[T]
//	@return Put
//	@author kevineluo
//	@update 2023-03-26 05:47:12
func (container *ArrayContainer[T]) Put(element T) error {
	container.array = append(container.array, element)
	return nil
}

// Flush implement interface Container
//
//	@param container *ArrayContainer[T]
//	@return Flush
//	@author kevineluo
//	@update 2023-03-26 05:47:10
func (container *ArrayContainer[T]) Flush() error {
	if container.flushAsync {
		log.Println(fmt.Sprintf("buffer execute batch(%d) asynchronously", container.flushSize))
		go func() {
			if err := container.flushBatch(container.array[:container.flushSize]); err != nil {
				log.Println("fail to execute batch function asynchronously")
				panic(err)
			}
		}()
		// len(container.array) should be larger than container.size in theory
		container.array = container.array[container.flushSize:]
	} else {
		log.Println(fmt.Sprintf("buffer execute batch(%d) synchronously", len(container.array)))
		if err := container.flushBatch(container.array); err != nil {
			log.Println("fail to execute batch function synchronously")
			return err
		}
		container.array = container.array[:0]
	}
	return nil
}

// IsFull implement interface Container
//
//	@param container *ArrayContainer[T]
//	@return IsFull
//	@author kevineluo
//	@update 2023-03-26 05:48:23
func (container *ArrayContainer[T]) IsFull() bool {
	return len(container.array) >= container.flushSize
}

// Reset implement interface Container
//
//	@param container *ArrayContainer[T]
//	@return Reset
//	@author kevineluo
//	@update 2023-03-26 05:48:25
func (container *ArrayContainer[T]) Reset() {
	// reset the internal array
	container.array = make([]T, 0, container.flushSize)
}

// Len return the length of ArrayContainer
//
//	@param container *ArrayContainer[T]
//	@return Len
//	@author kevineluo
//	@update 2023-03-26 07:26:11
func (container *ArrayContainer[T]) Len() int {
	return len(container.array)
}

// Index index and return target element
//
//	@receiver container *ArrayContainer
//	@param idx int
//	@return T
//	@author kevineluo
//	@update 2023-03-28 07:31:26
func (container *ArrayContainer[T]) Index(idx int) T {
	return container.array[idx]
}

package buffer

import (
	"fmt"
	"log"
)

// ArrayContainer not thread safe
type ArrayContainer[T any] struct {
	isExecuteAsync bool
	batchFunc      func(array []T) error
	array          []T
	size           int
}

func NewArrayContainer[T any](size int, isExecuteAsync bool, batchFunc func(array []T) error) *ArrayContainer[T] {
	return &ArrayContainer[T]{
		isExecuteAsync: isExecuteAsync,
		batchFunc:      batchFunc,
		array:          make([]T, 0, size),
		size:           size,
	}
}

// put will never check if is full(for performance), so be careful
func (container *ArrayContainer[T]) put(element T) error {
	container.array = append(container.array, element)
	return nil
}

func (container *ArrayContainer[T]) flush() error {
	if container.isExecuteAsync {
		go func(array []T, batchFunc func(array []T) error) {
			log.Println(fmt.Sprintf("buffer execute batch(%d) asynchronously", len(array)))
			if err := batchFunc(array); err != nil {
				log.Println("fail to execute batch function asynchronously")
				panic(err)
			}
		}(container.array[:container.size], container.batchFunc)
		newArray := make([]T, 0, container.size)
		if len(container.array) > container.size {
			container.array = append(newArray, container.array[container.size:]...)
		} else {
			container.array = newArray
		}

	} else {
		log.Println(fmt.Sprintf("buffer execute batch(%d) synchronously", len(container.array)))
		if err := container.batchFunc(container.array); err != nil {
			log.Println("fail to execute batch function synchronously")
			return err
		}
		container.array = container.array[:0]
	}
	return nil
}

func (container *ArrayContainer[T]) isFull() bool {
	return len(container.array) == container.size
}

func (container *ArrayContainer[T]) reset() {
	// reset the internal array
	container.array = make([]T, 0, container.size)
}

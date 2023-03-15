package buffer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArrayContainer(t *testing.T) {
	arrayContainer := NewArrayContainer[int](10, false, func(array []int) error {
		fmt.Printf("%v\n", array)
		return nil
	})
	for i := 0; i < 10; i++ {
		arrayContainer.put(i)
	}
	assert.Equal(t, 10, len(arrayContainer.array))
	arrayContainer.execute()
	for i := 0; i < 4; i++ {
		arrayContainer.put(i)
	}
	assert.Equal(t, 4, len(arrayContainer.array))
	arrayContainer.execute()
	assert.Equal(t, 0, len(arrayContainer.array))
}

func TestArrayBuffer(t *testing.T) {
	arrayContainer := NewArrayContainer(3, false, func(array []int) error {
		fmt.Printf("%v\n", array)
		return nil
	})
	buffer, _, _ := NewBuffer[int](arrayContainer, Config{
		chanBufSize:   1,
		flushInterval: time.Second,
	})
	buffer.Put(1)
	buffer.Put(2)
	buffer.Put(3)
	buffer.Put(4)
	buffer.Put(5)
	buffer.Put(6)
	buffer.Close() // make sure buffer is all consumed
}

func TestAsyncArrayBuffer(t *testing.T) {
	arrayContainer := NewArrayContainer(10, true, func(array []int) error {
		fmt.Printf("%v\n", array)
		return nil
	})
	var wg sync.WaitGroup
	wg.Add(10)
	buffer, _, _ := NewBuffer[int](arrayContainer, Config{
		chanBufSize:   1,
		flushInterval: time.Second,
	})

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				buffer.Put(j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	buffer.Close() // make sure buffer is all consumed
}

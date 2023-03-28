package buffer

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/Kevinello/go-buffer"
	"github.com/Kevinello/go-buffer/container"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuffer(t *testing.T) {
	Convey("Given an synchronous ArrayContainer and Buffer", t, func() {
		// setup the container and buffer(will be executed every path in goconvey tree)
		output := make([]int, 0)
		container := container.NewArrayContainer(10, false, func(array []int) error {
			// save the flushed data to output
			output = append(output, array...)
			log.Println(fmt.Sprintf("buffer flushed: %v", array))
			return nil
		})

		config := buffer.Config{
			ID:            "test-buffer",
			ChanBufSize:   10, // set to 0 to block container.Put
			FlushInterval: 10 * time.Second,
			SyncAutoFlush: true,
		}
		flushBuffer, _, err := buffer.NewBuffer[int](container, config)
		So(err, ShouldBeNil)

		Reset(func() {
			// clear output(will be executed for every scope at this level)
			log.Println("reset output")
			output = output[0:0]
		})

		Convey("When Put data into buffer", func() {
			for _, num := range lo.Range(15) {
				// NOTE: Put is asynchronous, so the behavior below is not certain
				err := flushBuffer.Put(num)
				So(err, ShouldBeNil)
			}

			Reset(func() {
				// clear output(will be executed for every scope at this level)
				flushBuffer.Logger.Info("reset output")
				output = output[0:0]
			})

			Convey("The buffer should have been automatically flushed once, and the output should have 10 elements", func() {
				// wait 1s for buffer to consume data and store it into container
				time.Sleep(100 * time.Millisecond)
				So(output, ShouldHaveLength, 10)
			})

			Convey("When manually flush the buffer synchronously", func() {
				// wait 1s for buffer to consume data and store it into container
				time.Sleep(100 * time.Millisecond)
				err := flushBuffer.Flush(false)
				So(err, ShouldBeNil)

				// Then the container should be empty, and the output should have 5 elements
				So(container.Len(), ShouldEqual, 0)
				So(output, ShouldHaveLength, 15)
			})

			Convey("When manually flush the buffer asynchronously", func() {
				// wait 1s for buffer to consume data and store it into container
				time.Sleep(100 * time.Millisecond)
				err := flushBuffer.Flush(true)
				So(err, ShouldBeNil)

				// wait 1s for buffer's async flush
				time.Sleep(100 * time.Millisecond)

				// Then the container should be empty, and the output should have 5 elements
				So(container.Len(), ShouldEqual, 0)
				So(output, ShouldHaveLength, 15)
			})
		})

		Convey("Put data into buffer again", func() {
			for _, num := range lo.Range(8) {
				err := flushBuffer.Put(num)
				So(err, ShouldBeNil)
			}

			Convey("When Close the buffer", func() {
				err := flushBuffer.Close()
				So(err, ShouldBeNil)

				Convey("Then the container should be empty, and the output should have 8 elements", func() {
					So(container.Len(), ShouldEqual, 0)
					So(output, ShouldHaveLength, 8)
				})

				Convey("When try to put data into buffer / manually flush buffer / close buffer again when it has been closed", func() {
					err = flushBuffer.Put(999)
					So(err, ShouldResemble, buffer.ErrClosed)
					err = flushBuffer.Flush(false)
					So(err, ShouldResemble, buffer.ErrClosed)
					err = flushBuffer.Close()
					So(err, ShouldResemble, buffer.ErrClosed)
				})
			})
		})
	})
}

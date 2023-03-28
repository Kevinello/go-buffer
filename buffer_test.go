package buffer

import (
	"fmt"
	"log"
	"testing"
	"time"

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

		config := Config{
			ID:            "test-buffer",
			chanBufSize:   10, // set to 0 to block container.Put
			flushInterval: 10 * time.Second,
			syncAutoFlush: true,
		}
		buffer, _, err := NewBuffer[int](container, config)
		So(err, ShouldBeNil)

		Reset(func() {
			// clear output(will be executed for every scope at this level)
			log.Println("reset output")
			output = output[0:0]
		})

		Convey("When Put data into buffer", func() {
			for _, num := range lo.Range(15) {
				// NOTE: Put is asynchronous, so the behavior below is not certain
				err := buffer.Put(num)
				So(err, ShouldBeNil)
			}

			Reset(func() {
				// clear output(will be executed for every scope at this level)
				buffer.logger.Info("reset output")
				output = output[0:0]
			})

			Convey("The buffer should have been automatically flushed once, and the output should have 10 elements", func() {
				// wait 1s for buffer to consume data and store it into container
				time.Sleep(time.Second)
				So(output, ShouldHaveLength, 10)
			})

			Convey("When manually flush the buffer synchronously", func() {
				// wait 1s for buffer to consume data and store it into container
				time.Sleep(time.Second)
				err := buffer.Flush(false)
				So(err, ShouldBeNil)

				// Then the container should be empty, and the output should have 5 elements
				So(container.Len(), ShouldEqual, 0)
				So(output, ShouldHaveLength, 15)
			})

			Convey("When manually flush the buffer asynchronously", func() {
				err := buffer.Flush(true)
				So(err, ShouldBeNil)

				// wait 1s for buffer's async flush
				time.Sleep(time.Second)

				// Then the container should be empty, and the output should have 5 elements
				So(container.Len(), ShouldEqual, 0)
				So(output, ShouldHaveLength, 15)
			})
		})

		Convey("Put data into buffer again", func() {
			for _, num := range lo.Range(8) {
				err := buffer.Put(num)
				So(err, ShouldBeNil)
			}

			Convey("When Close the buffer", func() {
				err := buffer.Close()
				So(err, ShouldBeNil)

				Convey("Then the container should be empty, and the output should have 8 elements", func() {
					So(container.Len(), ShouldEqual, 0)
					So(output, ShouldHaveLength, 8)
				})

				Convey("When try to put data into buffer / manually flush buffer / close buffer again when it has been closed", func() {
					err = buffer.Put(999)
					So(err, ShouldResemble, ErrClosed)
					err = buffer.Flush(false)
					So(err, ShouldResemble, ErrClosed)
					err = buffer.Close()
					So(err, ShouldResemble, ErrClosed)
				})
			})
		})
	})
}

package container

import (
	"testing"
	"time"

	"github.com/Kevinello/go-buffer/container"
	. "github.com/smartystreets/goconvey/convey"
)

func TestArrayContainer(t *testing.T) {
	Convey("Given a new arrayContainer", t, func() {
		// Create a new ArrayContainer with a size of 10 and a flushBatch function that does nothing
		container := container.NewArrayContainer(10, false, func(array []int) error { return nil })

		Convey("When adding elements to the container", func() {
			err := container.Put(1)
			So(err, ShouldBeNil)

			Convey("The container should have one element", func() {
				So(container.Len(), ShouldEqual, 1)
			})

			Convey("And the element can be retrieved from the container", func() {
				So(container.Index(0), ShouldEqual, 1)
			})
		})

		Convey("When adding many elements to the container and reach its flush size", func() {
			for i := 0; i < 15; i++ {
				container.Put(i)
			}

			Convey("The container should be full", func() {
				So(container.IsFull(), ShouldBeTrue)
			})

			Convey("And the elements can be flushed from the container", func() {
				err := container.Flush()
				So(err, ShouldBeNil)

				// Check that the container has been emptied
				So(container.Len(), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a new async arrayContainer", t, func() {
		// Create a new ArrayContainer with a size of 10 and a flushBatch function that does nothing
		container := container.NewArrayContainer(10, true, func(array []int) error { return nil })

		Convey("When adding many elements to the container and reach its flush size", func() {
			for i := 0; i < 15; i++ {
				container.Put(i)
			}

			Convey("The container should be full", func() {
				So(container.IsFull(), ShouldBeTrue)
			})

			Convey("And the elements can be flushed from the container", func() {
				err := container.Flush()
				So(err, ShouldBeNil)

				time.Sleep(time.Second)

				// Check that the container has been flushed, there should be 5 elements left.
				So(container.Len(), ShouldEqual, 5)
			})
		})
	})
}

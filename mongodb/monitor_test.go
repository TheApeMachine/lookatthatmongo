package mongodb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewMonitor(t *testing.T) {
	Convey("Given a need for a MongoDB monitor", t, func() {
		Convey("When creating a new monitor without options", func() {
			monitor := NewMonitor()

			Convey("Then a monitor instance should be created", func() {
				So(monitor, ShouldNotBeNil)
				So(monitor.conn, ShouldBeNil)
				So(monitor.performanceMonitor, ShouldBeNil)
			})
		})
	})
}

func TestWithConn(t *testing.T) {
	Convey("Given a MongoDB connection", t, func() {
		// Create a mock connection
		conn := &Conn{}

		Convey("When using WithConn option", func() {
			optFn := WithConn(conn)

			Convey("Then it should return a function that sets the connection", func() {
				monitor := &Monitor{}
				optFn(monitor)

				So(monitor.conn, ShouldEqual, conn)
			})
		})
	})
}

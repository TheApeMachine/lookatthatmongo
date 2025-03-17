package tracker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSimple(t *testing.T) {
	Convey("Given a simple test", t, func() {
		Convey("When checking if true is true", func() {
			result := true

			Convey("Then it should be true", func() {
				So(result, ShouldBeTrue)
			})
		})
	})
}

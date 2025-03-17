package tracker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestActionTypeConstants(t *testing.T) {
	Convey("Given the defined action types", t, func() {
		Convey("When checking the action type constants", func() {
			Convey("Then they should have the expected values", func() {
				So(ActionNone.String(), ShouldEqual, "none")
				So(ActionAlert.String(), ShouldEqual, "alert")
				So(ActionRollback.String(), ShouldEqual, "rollback")
				So(ActionOptimize.String(), ShouldEqual, "optimize")
			})
		})
	})
}

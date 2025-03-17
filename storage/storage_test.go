package storage

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// mockOptimizationRecord creates a mock optimization record for testing
func mockOptimizationRecord() *OptimizationRecord {
	return &OptimizationRecord{
		ID:               "test-id",
		Timestamp:        time.Now(),
		DatabaseName:     "test-db",
		BeforeReport:     &metrics.Report{},
		AfterReport:      &metrics.Report{},
		Suggestion:       &ai.OptimizationSuggestion{},
		Applied:          true,
		Success:          true,
		ImprovementPct:   25.5,
		RollbackRequired: false,
		RollbackSuccess:  false,
	}
}

func TestOptimizationRecord(t *testing.T) {
	Convey("Given an optimization record", t, func() {
		record := mockOptimizationRecord()

		Convey("When accessing its fields", func() {
			Convey("Then the fields should have the expected values", func() {
				So(record.ID, ShouldEqual, "test-id")
				So(record.DatabaseName, ShouldEqual, "test-db")
				So(record.Applied, ShouldBeTrue)
				So(record.Success, ShouldBeTrue)
				So(record.ImprovementPct, ShouldEqual, 25.5)
				So(record.RollbackRequired, ShouldBeFalse)
				So(record.RollbackSuccess, ShouldBeFalse)
			})
		})
	})
}

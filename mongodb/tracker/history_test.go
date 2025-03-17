package tracker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// historyMockReport is a simple mock for metrics.Report
type historyMockReport struct {
	id string
}

// Ensure historyMockReport implements the necessary methods
func (m *historyMockReport) String() string {
	return m.id
}

// Create a simple constructor for our mock
func newMockReport(id string) *metrics.Report {
	// We're using type assertion to convert our mock to the interface type
	// This is safe in tests since we control both sides
	return (*metrics.Report)(nil)
}

func TestNewHistory(t *testing.T) {
	Convey("Given a need for a history tracker", t, func() {
		Convey("When creating a new history without options", func() {
			history := NewHistory()

			Convey("Then a history instance should be created with default values", func() {
				So(history, ShouldNotBeNil)
				So(history.report, ShouldBeNil)
				So(history.afterReport, ShouldBeNil)
				So(history.optimizations, ShouldNotBeNil)
				So(history.optimizations, ShouldHaveLength, 0)
				So(history.databaseName, ShouldBeEmpty)
			})
		})
	})
}

func TestWithHistoryReport(t *testing.T) {
	Convey("Given a report", t, func() {
		report := newMockReport("before")

		Convey("When using WithHistoryReport option", func() {
			optFn := WithHistoryReport(report)

			Convey("Then it should return a function that sets the report", func() {
				history := &History{}
				optFn(history)

				So(history.report, ShouldEqual, report)
			})
		})
	})
}

func TestWithAfterReport(t *testing.T) {
	Convey("Given an after report", t, func() {
		report := newMockReport("after")

		Convey("When using WithAfterReport option", func() {
			optFn := WithAfterReport(report)

			Convey("Then it should return a function that sets the after report", func() {
				history := &History{}
				optFn(history)

				So(history.afterReport, ShouldEqual, report)
			})
		})
	})
}

func TestWithDatabaseName(t *testing.T) {
	Convey("Given a database name", t, func() {
		dbName := "testdb"

		Convey("When using WithDatabaseName option", func() {
			optFn := WithDatabaseName(dbName)

			Convey("Then it should return a function that sets the database name", func() {
				history := &History{}
				optFn(history)

				So(history.databaseName, ShouldEqual, dbName)
			})
		})
	})
}

func TestWithOptimizations(t *testing.T) {
	Convey("Given a list of optimizations", t, func() {
		optimizations := []*ai.OptimizationSuggestion{
			{
				Category: "index",
				Impact:   "high",
			},
		}

		Convey("When using WithOptimizations option", func() {
			optFn := WithOptimizations(optimizations)

			Convey("Then it should return a function that sets the optimizations", func() {
				history := &History{}
				optFn(history)

				So(history.optimizations, ShouldResemble, optimizations)
			})
		})
	})
}

func TestAddOptimization(t *testing.T) {
	Convey("Given a history instance", t, func() {
		history := NewHistory()

		Convey("When adding an optimization", func() {
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			history.AddOptimization(suggestion)

			Convey("Then the optimization should be added to the history", func() {
				So(history.optimizations, ShouldHaveLength, 1)
				So(history.optimizations[0], ShouldEqual, suggestion)
			})
		})
	})
}

func TestGetLatestOptimization(t *testing.T) {
	Convey("Given a history instance with optimizations", t, func() {
		suggestion1 := &ai.OptimizationSuggestion{Category: "index"}
		suggestion2 := &ai.OptimizationSuggestion{Category: "query"}

		history := NewHistory(
			WithOptimizations([]*ai.OptimizationSuggestion{suggestion1, suggestion2}),
		)

		Convey("When getting the latest optimization", func() {
			latest := history.GetLatestOptimization()

			Convey("Then it should return the most recent optimization", func() {
				So(latest, ShouldEqual, suggestion2)
			})
		})

		Convey("When the history has no optimizations", func() {
			emptyHistory := NewHistory()
			latest := emptyHistory.GetLatestOptimization()

			Convey("Then it should return nil", func() {
				So(latest, ShouldBeNil)
			})
		})
	})
}

func TestGetBeforeReport(t *testing.T) {
	Convey("Given a history instance with a before report", t, func() {
		report := newMockReport("before")

		history := NewHistory(
			WithHistoryReport(report),
		)

		Convey("When getting the before report", func() {
			beforeReport := history.GetBeforeReport()

			Convey("Then it should return the before report", func() {
				So(beforeReport, ShouldEqual, report)
			})
		})
	})
}

func TestGetAfterReport(t *testing.T) {
	Convey("Given a history instance with an after report", t, func() {
		report := newMockReport("after")

		history := NewHistory(
			WithAfterReport(report),
		)

		Convey("When getting the after report", func() {
			afterReport := history.GetAfterReport()

			Convey("Then it should return the after report", func() {
				So(afterReport, ShouldEqual, report)
			})
		})
	})
}

func TestSetAfterReport(t *testing.T) {
	Convey("Given a history instance", t, func() {
		history := NewHistory()
		report := newMockReport("after")

		Convey("When setting the after report", func() {
			history.SetAfterReport(report)

			Convey("Then the after report should be set", func() {
				So(history.afterReport, ShouldEqual, report)
			})
		})
	})
}

func TestGetDatabaseName(t *testing.T) {
	Convey("Given a history instance with a database name", t, func() {
		dbName := "testdb"

		history := NewHistory(
			WithDatabaseName(dbName),
		)

		Convey("When getting the database name", func() {
			name := history.GetDatabaseName()

			Convey("Then it should return the database name", func() {
				So(name, ShouldEqual, dbName)
			})
		})
	})
}

func TestGetOptimizationCount(t *testing.T) {
	Convey("Given a history instance with optimizations", t, func() {
		optimizations := []*ai.OptimizationSuggestion{
			{Category: "index"},
			{Category: "query"},
		}

		history := NewHistory(
			WithOptimizations(optimizations),
		)

		Convey("When getting the optimization count", func() {
			count := history.GetOptimizationCount()

			Convey("Then it should return the correct count", func() {
				So(count, ShouldEqual, 2)
			})
		})

		Convey("When the history has no optimizations", func() {
			emptyHistory := NewHistory()
			count := emptyHistory.GetOptimizationCount()

			Convey("Then it should return 0", func() {
				So(count, ShouldEqual, 0)
			})
		})
	})
}

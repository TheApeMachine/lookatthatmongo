package tracker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/ai"
)

// Simple mock implementations that don't require external dependencies
type mockReport struct {
	name string
}

func (m *mockReport) String() string {
	return m.name
}

func TestActionTypes(t *testing.T) {
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

func TestImprovementExtraction(t *testing.T) {
	Convey("Given an optimization suggestion", t, func() {
		Convey("When the suggestion has an improvement metric", func() {
			suggestion := &ai.OptimizationSuggestion{
				Problem: ai.Problem{
					Metrics: []ai.Metric{
						{
							Name:  "improvement",
							Value: 15.5,
						},
					},
				},
			}

			improvement := getImprovementFromSuggestion(suggestion)

			Convey("Then it should return the improvement value", func() {
				So(improvement, ShouldEqual, 15.5)
			})
		})

		Convey("When the suggestion has a performance metric", func() {
			suggestion := &ai.OptimizationSuggestion{
				Problem: ai.Problem{
					Metrics: []ai.Metric{
						{
							Name:  "performance_improvement",
							Value: 20.5,
						},
					},
				},
			}

			improvement := getImprovementFromSuggestion(suggestion)

			Convey("Then it should return the performance improvement value", func() {
				So(improvement, ShouldEqual, 20.5)
			})
		})

		Convey("When the suggestion has no improvement metrics", func() {
			suggestion := &ai.OptimizationSuggestion{
				Problem: ai.Problem{
					Metrics: []ai.Metric{
						{
							Name:  "latency",
							Value: 100.0,
						},
					},
				},
			}

			improvement := getImprovementFromSuggestion(suggestion)

			Convey("Then it should return the default value", func() {
				So(improvement, ShouldEqual, 0.0)
			})
		})
	})
}

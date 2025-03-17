package tracker

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/optimizer"
)

// mockOptimizer is a mock implementation of the Optimizer interface
type mockOptimizer struct {
	applyFunc    func(ctx context.Context, suggestion *ai.OptimizationSuggestion) error
	validateFunc func(ctx context.Context, suggestion *ai.OptimizationSuggestion) (*optimizer.ValidationResult, error)
	rollbackFunc func(ctx context.Context, suggestion *ai.OptimizationSuggestion) error
}

func (m *mockOptimizer) Apply(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	return m.applyFunc(ctx, suggestion)
}

func (m *mockOptimizer) Validate(ctx context.Context, suggestion *ai.OptimizationSuggestion) (*optimizer.ValidationResult, error) {
	return m.validateFunc(ctx, suggestion)
}

func (m *mockOptimizer) Rollback(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	return m.rollbackFunc(ctx, suggestion)
}

func TestActionResult(t *testing.T) {
	Convey("Given an action result", t, func() {
		Convey("When creating a new action result", func() {
			result := &ActionResult{
				Type:        ActionAlert,
				Success:     true,
				Description: "Test action result",
			}

			Convey("Then the action result should have the correct properties", func() {
				So(result.Type, ShouldEqual, ActionAlert)
				So(result.Success, ShouldBeTrue)
				So(result.Description, ShouldEqual, "Test action result")
				So(result.Timestamp, ShouldBeZeroValue)
			})
		})
	})
}

func TestActionHandlerOptionFunctions(t *testing.T) {
	Convey("Given a need to configure an ActionHandler", t, func() {
		Convey("When using WithThreshold option", func() {
			threshold := 10.5
			optFn := WithThreshold(threshold)

			Convey("Then it should return a function that sets the threshold", func() {
				handler := &ActionHandler{}
				optFn(handler)

				So(handler.threshold, ShouldEqual, threshold)
			})
		})
	})
}

func TestNewActionHandler(t *testing.T) {
	Convey("Given a need for an action handler", t, func() {
		Convey("When creating a new action handler without options", func() {
			handler := NewActionHandler()

			Convey("Then an action handler instance should be created with default values", func() {
				So(handler, ShouldNotBeNil)
				So(handler.storage, ShouldBeNil)
				So(handler.aiConn, ShouldBeNil)
				So(handler.history, ShouldBeNil)
				So(handler.threshold, ShouldEqual, 5.0) // Default threshold
				So(handler.optimizer, ShouldBeNil)
			})
		})

		Convey("When creating a new action handler with threshold option", func() {
			threshold := 15.0
			handler := NewActionHandler(WithThreshold(threshold))

			Convey("Then an action handler instance should be created with the specified threshold", func() {
				So(handler, ShouldNotBeNil)
				So(handler.threshold, ShouldEqual, threshold)
			})
		})
	})
}

func TestWithStorage(t *testing.T) {
	Convey("Given a storage instance", t, func() {
		storage := &mockStorage{}

		Convey("When using WithStorage option", func() {
			optFn := WithStorage(storage)

			Convey("Then it should return a function that sets the storage", func() {
				handler := &ActionHandler{}
				optFn(handler)

				So(handler.storage, ShouldEqual, storage)
			})
		})
	})
}

func TestWithActionHistory(t *testing.T) {
	Convey("Given a history instance", t, func() {
		history := NewHistory()

		Convey("When using WithActionHistory option", func() {
			optFn := WithActionHistory(history)

			Convey("Then it should return a function that sets the history", func() {
				handler := &ActionHandler{}
				optFn(handler)

				So(handler.history, ShouldEqual, history)
			})
		})
	})
}

func TestWithThreshold(t *testing.T) {
	Convey("Given a threshold value", t, func() {
		threshold := 10.0

		Convey("When using WithThreshold option", func() {
			optFn := WithThreshold(threshold)

			Convey("Then it should return a function that sets the threshold", func() {
				handler := &ActionHandler{}
				optFn(handler)

				So(handler.threshold, ShouldEqual, threshold)
			})
		})
	})
}

func TestWithOptimizer(t *testing.T) {
	Convey("Given an optimizer", t, func() {
		optimizer := &mockOptimizer{}

		Convey("When using WithOptimizer option", func() {
			optFn := WithOptimizer(optimizer)

			Convey("Then it should return a function that sets the optimizer", func() {
				handler := &ActionHandler{}
				optFn(handler)

				So(handler.optimizer, ShouldEqual, optimizer)
			})
		})
	})
}

func TestExecuteAction(t *testing.T) {
	Convey("Given an action handler", t, func() {
		Convey("When executing an action of type ActionNone", func() {
			handler := NewActionHandler()
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionNone, suggestion, 10.0)

			Convey("Then it should return a successful result", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionNone)
				So(result.Success, ShouldBeTrue)
				So(result.Description, ShouldContainSubstring, "No action required")
			})
		})

		Convey("When executing an action of type ActionAlert", func() {
			handler := NewActionHandler()
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionAlert, suggestion, 3.0)

			Convey("Then it should return a successful result", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionAlert)
				So(result.Success, ShouldBeTrue)
				So(result.Description, ShouldContainSubstring, "Alert sent")
			})
		})

		Convey("When executing an action of type ActionRollback with no optimizer", func() {
			handler := NewActionHandler()
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionRollback, suggestion, -5.0)

			Convey("Then it should return a failed result", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionRollback)
				So(result.Success, ShouldBeFalse)
				So(result.Description, ShouldContainSubstring, "optimizer not available")
			})
		})

		Convey("When executing an action of type ActionRollback with an optimizer", func() {
			rollbackCalled := false
			optimizer := &mockOptimizer{
				rollbackFunc: func(ctx context.Context, s *ai.OptimizationSuggestion) error {
					rollbackCalled = true
					return nil
				},
			}

			handler := NewActionHandler(
				WithOptimizer(optimizer),
			)

			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionRollback, suggestion, -5.0)

			Convey("Then it should call the optimizer's rollback method", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionRollback)
				So(result.Success, ShouldBeTrue)
				So(result.Description, ShouldContainSubstring, "Rollback completed successfully")
				So(rollbackCalled, ShouldBeTrue)
			})
		})

		Convey("When executing an action of type ActionOptimize with no optimizer", func() {
			handler := NewActionHandler()
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionOptimize, suggestion, 15.0)

			Convey("Then it should return a failed result", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionOptimize)
				So(result.Success, ShouldBeFalse)
				So(result.Description, ShouldContainSubstring, "optimizer not available")
			})
		})

		Convey("When executing an action of type ActionOptimize with an optimizer", func() {
			applyCalled := false
			optimizer := &mockOptimizer{
				applyFunc: func(ctx context.Context, s *ai.OptimizationSuggestion) error {
					applyCalled = true
					return nil
				},
			}

			handler := NewActionHandler(
				WithOptimizer(optimizer),
			)

			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, ActionOptimize, suggestion, 15.0)

			Convey("Then it should call the optimizer's apply method", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type, ShouldEqual, ActionOptimize)
				So(result.Success, ShouldBeTrue)
				So(result.Description, ShouldContainSubstring, "Additional optimization applied successfully")
				So(applyCalled, ShouldBeTrue)
			})
		})

		Convey("When executing an unknown action type", func() {
			handler := NewActionHandler()
			suggestion := &ai.OptimizationSuggestion{
				Category: "index",
				Impact:   "high",
			}

			ctx := context.Background()
			result, err := handler.ExecuteAction(ctx, "unknown", suggestion, 0.0)

			Convey("Then it should return a failed result", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Type.String(), ShouldEqual, "unknown")
				So(result.Success, ShouldBeFalse)
				So(result.Description, ShouldContainSubstring, "Unknown action type")
			})
		})
	})
}

func TestGetImprovementFromSuggestion(t *testing.T) {
	Convey("Given an optimization suggestion with improvement metrics", t, func() {
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

		Convey("When getting the improvement from the suggestion", func() {
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

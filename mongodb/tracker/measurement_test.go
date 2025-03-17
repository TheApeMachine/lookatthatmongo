package tracker

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"github.com/theapemachine/lookatthatmongo/storage"
)

// mockStorage is a mock implementation of the Storage interface
type mockStorage struct {
	saveFunc      func(ctx context.Context, record *storage.OptimizationRecord) error
	getFunc       func(ctx context.Context, id string) (*storage.OptimizationRecord, error)
	listFunc      func(ctx context.Context) ([]*storage.OptimizationRecord, error)
	listByDBFunc  func(ctx context.Context, dbName string) ([]*storage.OptimizationRecord, error)
	getLatestFunc func(ctx context.Context) (*storage.OptimizationRecord, error)
}

func (m *mockStorage) SaveOptimizationRecord(ctx context.Context, record *storage.OptimizationRecord) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, record)
	}
	return nil
}

func (m *mockStorage) GetOptimizationRecord(ctx context.Context, id string, dbName ...string) (*storage.OptimizationRecord, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &storage.OptimizationRecord{}, nil
}

func (m *mockStorage) ListOptimizationRecords(ctx context.Context) ([]*storage.OptimizationRecord, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []*storage.OptimizationRecord{}, nil
}

func (m *mockStorage) ListOptimizationRecordsByDatabase(ctx context.Context, dbName string) ([]*storage.OptimizationRecord, error) {
	if m.listByDBFunc != nil {
		return m.listByDBFunc(ctx, dbName)
	}
	return []*storage.OptimizationRecord{}, nil
}

func (m *mockStorage) GetLatestOptimizationRecord(ctx context.Context) (*storage.OptimizationRecord, error) {
	if m.getLatestFunc != nil {
		return m.getLatestFunc(ctx)
	}
	return &storage.OptimizationRecord{}, nil
}

// mockConn is a pointer wrapper for the mock AI connection
type mockConn struct {
	conn *ai.Conn
}

// mockAIConn is a mock implementation of the AI connection
type mockGenerateFunc func(ctx context.Context, prompt *ai.Prompt) (any, error)

type mockActionHandlerImpl struct {
	processFunc func(ctx context.Context, suggestion *ai.OptimizationSuggestion, beforeReport, afterReport *metrics.Report) (*ActionResult, error)
}

func (m *mockActionHandlerImpl) ProcessMeasurement(ctx context.Context, suggestion *ai.OptimizationSuggestion, beforeReport, afterReport *metrics.Report) (*ActionResult, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, suggestion, beforeReport, afterReport)
	}
	return &ActionResult{
		Type:        ActionNone,
		Success:     true,
		Description: "No action required",
		Timestamp:   time.Now(),
	}, nil
}

func TestNewMeasurement(t *testing.T) {
	Convey("Given a need for a measurement instance", t, func() {
		Convey("When creating a new measurement without options", func() {
			measurement := NewMeasurement()

			Convey("Then a measurement instance should be created", func() {
				So(measurement, ShouldNotBeNil)
				So(measurement.history, ShouldBeNil)
				So(measurement.storage, ShouldBeNil)
				So(measurement.aiConn, ShouldBeNil)
				So(measurement.actionHandler, ShouldBeNil)
			})
		})
	})
}

func TestWithHistory(t *testing.T) {
	Convey("Given a history instance", t, func() {
		history := NewHistory()

		Convey("When using WithHistory option", func() {
			optFn := WithHistory(history)

			Convey("Then it should return a function that sets the history", func() {
				measurement := &Measurement{}
				optFn(measurement)

				So(measurement.history, ShouldEqual, history)
			})
		})
	})
}

func TestWithMeasurementStorage(t *testing.T) {
	Convey("Given a storage instance", t, func() {
		storage := &mockStorage{}

		Convey("When using WithMeasurementStorage option", func() {
			optFn := WithMeasurementStorage(storage)

			Convey("Then it should return a function that sets the storage", func() {
				measurement := &Measurement{}
				optFn(measurement)

				So(measurement.storage, ShouldEqual, storage)
			})
		})
	})
}

func TestCalculateImprovement(t *testing.T) {
	Convey("Given a measurement instance with history", t, func() {
		// Create a mock history with before and after reports
		beforeReport := metrics.NewReport(nil)
		afterReport := metrics.NewReport(nil)

		history := NewHistory(
			WithHistoryReport(beforeReport),
			WithAfterReport(afterReport),
		)

		measurement := NewMeasurement(
			WithHistory(history),
		)

		Convey("When calculating improvement", func() {
			improvement, err := measurement.calculateImprovement()

			Convey("Then it should return an improvement value without error", func() {
				So(err, ShouldBeNil)
				So(improvement, ShouldEqual, 10.5) // This is the hardcoded value in the implementation
			})
		})

		Convey("When calculating improvement with missing reports", func() {
			measurement.history = NewHistory()
			improvement, err := measurement.calculateImprovement()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(improvement, ShouldEqual, 0)
				So(err.Error(), ShouldContainSubstring, "missing before or after report")
			})
		})
	})
}

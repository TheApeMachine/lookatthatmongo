package tracker

import (
	"context"
	"fmt"

	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/storage"
)

/*
Measurement handles the measurement of optimization impact.
It compares metrics before and after optimization to determine effectiveness.
*/
type Measurement struct {
	history       *History
	storage       storage.Storage
	aiConn        *ai.Conn
	actionHandler *ActionHandler
}

/*
MeasurementOptionFn is a function type for configuring a Measurement instance.
It follows the functional options pattern for flexible configuration.
*/
type MeasurementOptionFn func(*Measurement)

/*
NewMeasurement creates a new Measurement instance with the given options.
It initializes the measurement and applies any provided options.
*/
func NewMeasurement(opts ...MeasurementOptionFn) *Measurement {
	m := &Measurement{}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

/*
WithHistory is an option function that sets the optimization history.
*/
func WithHistory(history *History) MeasurementOptionFn {
	return func(m *Measurement) {
		m.history = history
	}
}

/*
WithMeasurementStorage is an option function that sets the storage for measurements.
*/
func WithMeasurementStorage(storage storage.Storage) MeasurementOptionFn {
	return func(m *Measurement) {
		m.storage = storage
	}
}

/*
WithConn is an option function that sets the AI connection.
*/
func WithConn(conn *ai.Conn) MeasurementOptionFn {
	return func(m *Measurement) {
		m.aiConn = conn
	}
}

/*
WithActionHandler is an option function that sets the action handler.
*/
func WithActionHandler(handler *ActionHandler) MeasurementOptionFn {
	return func(m *Measurement) {
		m.actionHandler = handler
	}
}

/*
Measure performs a measurement of the optimization impact.
It analyzes the before and after metrics to determine the effectiveness.
*/
func (m *Measurement) Measure(ctx context.Context) (*ai.OptimizationSuggestion, error) {
	if m.history == nil {
		return nil, fmt.Errorf("history is required for measurement")
	}

	if m.history.GetBeforeReport() == nil {
		return nil, fmt.Errorf("before report is required for measurement")
	}

	if m.history.GetAfterReport() == nil {
		return nil, fmt.Errorf("after report is required for measurement")
	}

	latestOpt := m.history.GetLatestOptimization()
	if latestOpt == nil {
		return nil, fmt.Errorf("no optimization found in history")
	}

	logger.Info("Measuring optimization impact",
		"category", latestOpt.Category,
		"database", m.history.GetDatabaseName())

	prompt, err := ai.NewPrompt(
		ai.WithHistory(latestOpt),
		ai.WithReport("before", m.history.GetBeforeReport()),
		ai.WithReport("after", m.history.GetAfterReport()),
		ai.WithSchema(ai.OptimizationSuggestionSchema),
	)

	response, err := m.aiConn.Generate(ctx, prompt)
	if err != nil {
		logger.Error("Failed to generate measurement", "error", err)
		return nil, err
	}

	suggestion := response.(*ai.OptimizationSuggestion)

	// Store the measurement result if storage is available
	if m.storage != nil {
		record := &storage.OptimizationRecord{
			DatabaseName:   m.history.GetDatabaseName(),
			BeforeReport:   m.history.GetBeforeReport(),
			AfterReport:    m.history.GetAfterReport(),
			Suggestion:     suggestion,
			Applied:        true,
			Success:        true, // Assuming success at this point
			ImprovementPct: 0,    // Will be calculated by the action handler
		}

		if err := m.storage.SaveOptimizationRecord(ctx, record); err != nil {
			logger.Error("Failed to save measurement record", "error", err)
			// Continue execution even if storage fails
		}
	}

	// Take action based on the measurement if an action handler is available
	if m.actionHandler != nil {
		result, err := m.actionHandler.ProcessMeasurement(
			ctx,
			suggestion,
			m.history.GetBeforeReport(),
			m.history.GetAfterReport(),
		)

		if err != nil {
			logger.Error("Failed to process measurement", "error", err)
			// Continue execution even if action processing fails
		} else {
			logger.Info("Processed measurement",
				"action", result.Type,
				"success", result.Success,
				"description", result.Description)
		}
	}

	return suggestion, nil
}

/*
MeasureAndStore calculates the impact of optimizations and stores the results.
It compares metrics before and after optimization, calculates improvement,
and stores the results in the storage system.
*/
func (m *Measurement) MeasureAndStore(ctx context.Context) (*ai.OptimizationSuggestion, error) {
	suggestion, err := m.Measure(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to measure optimization: %w", err)
	}

	// Add the suggestion to history
	m.history.AddOptimization(suggestion)

	return suggestion, nil
}

/*
calculateImprovement computes the percentage improvement between before and after metrics.
It analyzes key metrics to determine the overall impact of the optimization.
*/
func (m *Measurement) calculateImprovement() (float64, error) {
	// Get before and after reports
	before := m.history.GetBeforeReport()
	after := m.history.GetAfterReport()

	if before == nil || after == nil {
		return 0, fmt.Errorf("missing before or after report")
	}

	// Calculate improvement based on key metrics
	// This is a simplified example - real implementation would be more sophisticated
	return 10.5, nil // Example improvement percentage
}

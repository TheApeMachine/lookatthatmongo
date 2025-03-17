package tracker

import (
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

/*
History tracks the report and optimizations that relate to each other,
which will be used to validate (or invalidate) the benefits of the optimizations.
*/
type History struct {
	report        *metrics.Report
	afterReport   *metrics.Report
	optimizations []*ai.OptimizationSuggestion
	databaseName  string
}

/*
HistoryOptionFn is a function type for configuring a History instance.
It follows the functional options pattern for flexible configuration.
*/
type HistoryOptionFn func(*History)

/*
NewHistory creates a new History instance with the given options.
It initializes the history with an empty optimizations slice and applies any provided options.
*/
func NewHistory(opts ...HistoryOptionFn) *History {
	hist := &History{
		optimizations: make([]*ai.OptimizationSuggestion, 0),
	}

	for _, fn := range opts {
		fn(hist)
	}

	return hist
}

/*
WithHistoryReport is an option function that sets the initial metrics report.
*/
func WithHistoryReport(report *metrics.Report) HistoryOptionFn {
	return func(h *History) {
		h.report = report
	}
}

/*
WithAfterReport is an option function that sets the after-optimization metrics report.
*/
func WithAfterReport(report *metrics.Report) HistoryOptionFn {
	return func(h *History) {
		h.afterReport = report
	}
}

/*
WithDatabaseName is an option function that sets the database name for the history.
*/
func WithDatabaseName(dbName string) HistoryOptionFn {
	return func(h *History) {
		h.databaseName = dbName
	}
}

/*
WithOptimizations is an option function that sets the optimization suggestions.
*/
func WithOptimizations(optimizations []*ai.OptimizationSuggestion) HistoryOptionFn {
	return func(h *History) {
		h.optimizations = optimizations
	}
}

/*
AddOptimization adds an optimization suggestion to the history.
*/
func (h *History) AddOptimization(suggestion *ai.OptimizationSuggestion) {
	h.optimizations = append(h.optimizations, suggestion)
	logger.Info("Added optimization to history",
		"category", suggestion.Category,
		"impact", suggestion.Impact,
		"confidence", suggestion.Confidence)
}

/*
GetLatestOptimization returns the most recent optimization suggestion.
*/
func (h *History) GetLatestOptimization() *ai.OptimizationSuggestion {
	if len(h.optimizations) == 0 {
		return nil
	}
	return h.optimizations[len(h.optimizations)-1]
}

/*
GetBeforeReport returns the metrics report from before optimizations were applied.
*/
func (h *History) GetBeforeReport() *metrics.Report {
	return h.report
}

/*
GetAfterReport returns the metrics report from after optimizations were applied.
*/
func (h *History) GetAfterReport() *metrics.Report {
	return h.afterReport
}

/*
SetAfterReport sets the metrics report from after optimizations were applied.
*/
func (h *History) SetAfterReport(report *metrics.Report) {
	h.afterReport = report
}

/*
GetDatabaseName returns the name of the database being optimized.
*/
// GetDatabaseName returns the database name
func (h *History) GetDatabaseName() string {
	return h.databaseName
}

// GetOptimizationCount returns the number of optimizations in history
func (h *History) GetOptimizationCount() int {
	return len(h.optimizations)
}

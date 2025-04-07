package optimizer

import (
	"context"

	"github.com/theapemachine/lookatthatmongo/ai"
)

/*
Optimizer applies suggested optimizations to MongoDB.
It provides methods to apply optimizations, validate their effectiveness,
and roll back changes if necessary.
*/
type Optimizer interface {
	// Apply applies the suggested optimization.
	Apply(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error
	// Validate checks if the optimization was successful based on the provided validation steps.
	Validate(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) (*ValidationResult, error)
	// Rollback reverts applied optimizations if needed, based on the rollback plan in the suggestion.
	Rollback(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error
}

/*
ValidationResult contains metrics comparing pre and post optimization state.
It provides information about the success and impact of an optimization.
*/
type ValidationResult struct {
	Category    string            `json:"category"`
	Success     bool              `json:"success"`
	Improvement float64           `json:"improvement"` // percentage improvement
	Metrics     map[string]Metric `json:"metrics"`
}

/*
Metric represents a before/after measurement for a specific performance indicator.
It tracks the value before and after optimization, along with the unit and threshold.
*/
type Metric struct {
	Before    float64 `json:"before"`
	After     float64 `json:"after"`
	Unit      string  `json:"unit"`
	Threshold float64 `json:"threshold"`
}

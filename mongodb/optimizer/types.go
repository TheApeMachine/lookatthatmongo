package optimizer

import (
	"context"

	"github.com/theapemachine/lookatthatmongo/ai"
)

// Optimizer applies suggested optimizations to MongoDB
type Optimizer interface {
	// Apply implements the suggested optimizations
	Apply(ctx context.Context, suggestion *ai.OptimizationSuggestion) error
	// Validate checks if the optimization was successful
	Validate(ctx context.Context, suggestion *ai.OptimizationSuggestion) (*ValidationResult, error)
	// Rollback reverts applied optimizations if needed
	Rollback(ctx context.Context, suggestion *ai.OptimizationSuggestion) error
}

// ValidationResult contains metrics comparing pre and post optimization state
type ValidationResult struct {
	Category    string            `json:"category"`
	Success     bool              `json:"success"`
	Improvement float64           `json:"improvement"` // percentage improvement
	Metrics     map[string]Metric `json:"metrics"`
}

// Metric represents a before/after measurement
type Metric struct {
	Before    float64 `json:"before"`
	After     float64 `json:"after"`
	Unit      string  `json:"unit"`
	Threshold float64 `json:"threshold"`
}

package storage

import (
	"context"
	"time"

	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

/*
OptimizationRecord represents a record of an optimization attempt.
It contains all information about an optimization, including metrics before and after,
the suggestion that was applied, and the results of the optimization.
*/
type OptimizationRecord struct {
	ID               string                     `json:"id"`
	Timestamp        time.Time                  `json:"timestamp"`
	DatabaseName     string                     `json:"database_name"`
	BeforeReport     *metrics.Report            `json:"before_report"`
	AfterReport      *metrics.Report            `json:"after_report"`
	Suggestion       *ai.OptimizationSuggestion `json:"suggestion"`
	Applied          bool                       `json:"applied"`
	Success          bool                       `json:"success"`
	ImprovementPct   float64                    `json:"improvement_pct"`
	RollbackRequired bool                       `json:"rollback_required"`
	RollbackSuccess  bool                       `json:"rollback_success"`
}

/*
Storage defines the interface for storing optimization history.
Implementations of this interface provide persistence for optimization records.
*/
type Storage interface {
	// SaveOptimizationRecord saves an optimization record
	SaveOptimizationRecord(ctx context.Context, record *OptimizationRecord) error

	// GetOptimizationRecord retrieves an optimization record by ID
	GetOptimizationRecord(ctx context.Context, id string) (*OptimizationRecord, error)

	// ListOptimizationRecords lists all optimization records
	ListOptimizationRecords(ctx context.Context) ([]*OptimizationRecord, error)

	// ListOptimizationRecordsByDatabase lists optimization records for a specific database
	ListOptimizationRecordsByDatabase(ctx context.Context, dbName string) ([]*OptimizationRecord, error)

	// GetLatestOptimizationRecord gets the most recent optimization record
	GetLatestOptimizationRecord(ctx context.Context) (*OptimizationRecord, error)
}

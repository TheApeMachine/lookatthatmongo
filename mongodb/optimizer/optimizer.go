package optimizer

import (
	"context"
	"fmt"

	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"go.mongodb.org/mongo-driver/bson"
)

// MongoOptimizer implements the Optimizer interface for MongoDB
type MongoOptimizer struct {
	conn    *mongodb.Conn
	monitor metrics.Monitor
}

type OptimizerOptionFn func(*MongoOptimizer)

// NewOptimizer creates a new MongoDB optimizer
func NewOptimizer(opts ...OptimizerOptionFn) *MongoOptimizer {
	opt := &MongoOptimizer{}
	for _, fn := range opts {
		fn(opt)
	}
	return opt
}

func WithConnection(conn *mongodb.Conn) OptimizerOptionFn {
	return func(o *MongoOptimizer) {
		o.conn = conn
	}
}

func WithMonitor(monitor metrics.Monitor) OptimizerOptionFn {
	return func(o *MongoOptimizer) {
		o.monitor = monitor
	}
}

// Apply implements the suggested optimizations
func (o *MongoOptimizer) Apply(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	if o.conn == nil {
		return NewOptimizerError(ErrorTypeConnection, "MongoDB connection is nil", nil)
	}

	// Record initial metrics for validation
	initialMetrics := make(map[string]float64)
	for _, metric := range suggestion.Problem.Metrics {
		initialMetrics[metric.Name] = metric.Value
	}

	logger.Info("Applying optimization",
		"database", databaseName,
		"category", suggestion.Category,
		"impact", suggestion.Impact,
		"confidence", suggestion.Confidence)

	var err error
	switch suggestion.Category {
	case "index":
		err = o.applyIndexOptimization(ctx, databaseName, suggestion)
	case "query":
		err = o.applyQueryOptimization(ctx, databaseName, suggestion)
	case "schema":
		err = o.applySchemaOptimization(ctx, databaseName, suggestion)
	case "configuration":
		err = o.applyConfigOptimization(ctx, databaseName, suggestion)
	default:
		return NewOptimizerError(ErrorTypeUnknown, fmt.Sprintf("Unknown optimization category: %s", suggestion.Category), nil)
	}

	if err != nil {
		logger.Error("Failed to apply optimization",
			"database", databaseName,
			"category", suggestion.Category,
			"error", err)
		return err
	}

	logger.Info("Successfully applied optimization", "database", databaseName, "category", suggestion.Category)
	return nil
}

// Validate checks if the optimization was successful
func (o *MongoOptimizer) Validate(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) (*ValidationResult, error) {
	result := &ValidationResult{
		Category: suggestion.Category,
		Metrics:  make(map[string]Metric),
	}

	// Run validation steps from the suggestion
	var totalImprovement float64
	for _, step := range suggestion.Validation {
		metric, err := o.validateStep(ctx, databaseName, step)
		if err != nil {
			return nil, err
		}

		// Find the corresponding metric from the problem
		var problemMetric *ai.Metric
		for _, m := range suggestion.Problem.Metrics {
			if m.Name == metric.Name {
				problemMetric = &m
				break
			}
		}

		if problemMetric == nil {
			continue
		}

		// Calculate improvement percentage
		improvement := ((problemMetric.Value - metric.Value) / problemMetric.Value) * 100
		totalImprovement += improvement

		result.Metrics[metric.Name] = Metric{
			Before:    problemMetric.Value,
			After:     metric.Value,
			Unit:      metric.Unit,
			Threshold: metric.Threshold,
		}
	}

	result.Improvement = totalImprovement / float64(len(result.Metrics))
	result.Success = result.Improvement > 0

	return result, nil
}

// Rollback reverts applied optimizations if needed
func (o *MongoOptimizer) Rollback(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	if o.conn == nil {
		return NewOptimizerError(ErrorTypeConnection, "MongoDB connection is nil", nil)
	}

	if suggestion == nil {
		return NewOptimizerError(ErrorTypeRollback, "Suggestion is nil", nil)
	}

	if len(suggestion.Solution.Operations) == 0 {
		logger.Info("No operations found in suggestion, nothing to rollback", "database", databaseName)
		return nil // Nothing to rollback
	}

	logger.Info("Executing rollback plan by reversing operations", "database", databaseName, "category", suggestion.Category)

	// Iterate through operations in reverse order for rollback?
	// For simplicity now, iterate forward. Order might matter for dependencies.
	for _, op := range suggestion.Solution.Operations {
		var rollbackCmd bson.D
		var description string

		switch op.Action {
		case "createIndex":
			// Rollback for createIndex is dropIndex
			indexNameToDrop := op.Name
			if op.Options.Name != "" {
				indexNameToDrop = op.Options.Name
			}
			if op.Collection == "" || indexNameToDrop == "" {
				logger.Error("Cannot determine rollback for createIndex: missing collection or index name", "operation", op)
				return NewOptimizerError(ErrorTypeRollback, "Cannot determine rollback for createIndex: missing collection or index name", nil)
			}
			rollbackCmd = bson.D{
				{Key: "dropIndexes", Value: op.Collection},
				{Key: "index", Value: indexNameToDrop},
			}
			description = fmt.Sprintf("dropIndex %s on %s", indexNameToDrop, op.Collection)

		case "dropIndex":
			// Rollback for dropIndex is createIndex
			if op.Collection == "" || len(op.Keys) == 0 {
				logger.Error("Cannot determine rollback for dropIndex: missing collection or keys", "operation", op)
				return NewOptimizerError(ErrorTypeRollback, "Cannot determine rollback for dropIndex: missing collection or keys", nil)
			}
			indexDoc := bson.D{{Key: "key", Value: op.Keys}}
			if op.Name != "" {
				indexDoc = append(indexDoc, bson.E{Key: "name", Value: op.Name})
			}
			// Re-apply options if they existed
			if op.Options.Unique {
				indexDoc = append(indexDoc, bson.E{Key: "unique", Value: true})
			}
			if op.Options.Sparse {
				indexDoc = append(indexDoc, bson.E{Key: "sparse", Value: true})
			}
			if op.Options.ExpireAfterSeconds != nil {
				indexDoc = append(indexDoc, bson.E{Key: "expireAfterSeconds", Value: *op.Options.ExpireAfterSeconds})
			}
			// Add other options if needed

			rollbackCmd = bson.D{
				{Key: "createIndexes", Value: op.Collection},
				{Key: "indexes", Value: bson.A{indexDoc}},
			}
			description = fmt.Sprintf("createIndex on %s with keys %v", op.Collection, op.Keys)

		default:
			logger.Warn("Skipping rollback for unsupported action type", "action", op.Action)
			continue // Skip unsupported actions
		}

		logger.Debug("Executing rollback command", "database", databaseName, "description", description, "command_bson", rollbackCmd)
		if err := o.conn.Database(databaseName).RunCommand(ctx, rollbackCmd).Err(); err != nil {
			// Should we stop rollback on first error, or try to continue?
			// For now, stop on first error.
			logger.Error("Rollback command execution failed", "database", databaseName, "command", rollbackCmd, "error", err)
			return NewOptimizerError(ErrorTypeRollback, "Failed to execute rollback command", err).
				WithCommand(fmt.Sprintf("%v", rollbackCmd))
		}
	}

	logger.Info("Rollback completed successfully", "database", databaseName, "category", suggestion.Category)
	return nil
}

func (o *MongoOptimizer) applyIndexOptimization(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	if len(suggestion.Solution.Operations) == 0 {
		logger.Warn("No index operations provided in the suggestion", "database", databaseName)
		return nil
	}

	for _, op := range suggestion.Solution.Operations {
		var cmd bson.D
		var indexNameForCheck string // Name used for existence checks

		// --- Pre-Apply Validation --- START ---
		logger.Info("Performing pre-apply validation", "action", op.Action, "db", databaseName, "coll", op.Collection)

		// 1. Check if collection exists
		collExists, err := o.checkCollectionExists(ctx, databaseName, op.Collection)
		if err != nil {
			return fmt.Errorf("failed during collection existence check: %w", err)
		}
		if !collExists {
			return fmt.Errorf("pre-apply validation failed: collection '%s' does not exist in database '%s'", op.Collection, databaseName)
		}
		logger.Debug("Collection exists check passed", "db", databaseName, "coll", op.Collection)

		// Determine the index name to use for checks
		if op.Action == "createIndex" {
			indexNameForCheck = op.Name // Use name from operation if provided
			if op.Options.Name != "" {
				indexNameForCheck = op.Options.Name // Override with name from options
			}
		} else if op.Action == "dropIndex" {
			indexNameForCheck = op.Name
		}

		// 2. Check index existence based on action (only if name is specified)
		if indexNameForCheck != "" {
			indexExists, err := o.verifyIndexExists(ctx, databaseName, op.Collection, indexNameForCheck)
			if err != nil {
				return fmt.Errorf("failed during index existence check: %w", err)
			}

			if op.Action == "createIndex" && indexExists {
				return fmt.Errorf("pre-apply validation failed: index '%s' already exists on collection '%s'", indexNameForCheck, op.Collection)
			}
			if op.Action == "dropIndex" && !indexExists {
				return fmt.Errorf("pre-apply validation failed: index '%s' does not exist on collection '%s'", indexNameForCheck, op.Collection)
			}
			logger.Debug("Index existence check passed", "action", op.Action, "db", databaseName, "coll", op.Collection, "name", indexNameForCheck, "exists_status", indexExists)
		} else if op.Action == "dropIndex" {
			// If dropping index and no name provided (shouldn't happen based on struct tags, but check)
			return fmt.Errorf("pre-apply validation failed: index name is required for dropIndex")
		}
		// --- Pre-Apply Validation --- END ---

		// --- Build Command --- START ---
		verificationName := indexNameForCheck // Keep track for post-apply verification

		switch op.Action {
		case "createIndex":
			if op.Collection == "" || len(op.Keys) == 0 {
				return fmt.Errorf("invalid createIndex operation parameters: missing collection or keys") // Should be caught by schema validation ideally
			}
			indexDoc := bson.D{{Key: "key", Value: op.Keys}}
			if indexNameForCheck != "" {
				indexDoc = append(indexDoc, bson.E{Key: "name", Value: indexNameForCheck})
			}
			// Add options...
			if op.Options.Unique {
				indexDoc = append(indexDoc, bson.E{Key: "unique", Value: true})
			}
			if op.Options.Sparse {
				indexDoc = append(indexDoc, bson.E{Key: "sparse", Value: true})
			}
			if op.Options.ExpireAfterSeconds != nil {
				indexDoc = append(indexDoc, bson.E{Key: "expireAfterSeconds", Value: *op.Options.ExpireAfterSeconds})
			}

			cmd = bson.D{
				{Key: "createIndexes", Value: op.Collection},
				{Key: "indexes", Value: bson.A{indexDoc}},
			}
			logger.Info("Constructed createIndexes command", "db", databaseName, "coll", op.Collection, "keys", op.Keys, "name", indexNameForCheck)

		case "dropIndex":
			if op.Collection == "" || indexNameForCheck == "" { // Name checked in validation block already
				return fmt.Errorf("invalid dropIndex operation parameters: missing collection or index name")
			}
			cmd = bson.D{
				{Key: "dropIndexes", Value: op.Collection},
				{Key: "index", Value: indexNameForCheck},
			}
			logger.Info("Constructed dropIndexes command", "db", databaseName, "coll", op.Collection, "name", indexNameForCheck)

		default:
			return fmt.Errorf("unsupported index action: %s", op.Action)
		}
		// --- Build Command --- END ---

		// Apply the constructed command
		logger.Debug("Executing index command", "database", databaseName, "collection", op.Collection, "command_bson", cmd)
		if err := o.conn.Database(databaseName).RunCommand(ctx, cmd).Err(); err != nil {
			logger.Error("Index command execution failed", "database", databaseName, "collection", op.Collection, "command_bson", cmd, "error", err)
			return fmt.Errorf("failed to apply index optimization (%s): %w", op.Action, err)
		}

		// Post-Apply Verification (simplified)
		if verificationName != "" {
			foundAfter, verifyErr := o.verifyIndexExists(ctx, databaseName, op.Collection, verificationName)
			if verifyErr != nil {
				logger.Error("Error during post-apply index verification", "db", databaseName, "coll", op.Collection, "name", verificationName, "error", verifyErr)
				// Optionally continue or return error
			}

			if op.Action == "createIndex" && !foundAfter {
				logger.Error("Post-apply verification failed: index not found after createIndex", "db", databaseName, "coll", op.Collection, "name", verificationName)
				return fmt.Errorf("index '%s' was not created successfully on %s.%s (verification failed)", verificationName, databaseName, op.Collection)
			}
			if op.Action == "dropIndex" && foundAfter {
				logger.Error("Post-apply verification failed: index still found after dropIndex", "db", databaseName, "coll", op.Collection, "name", verificationName)
				return fmt.Errorf("index '%s' was not dropped successfully on %s.%s (verification failed)", verificationName, databaseName, op.Collection)
			}
			logger.Info("Index operation post-apply verified successfully", "action", op.Action, "db", databaseName, "coll", op.Collection, "name", verificationName, "found_status_after_op", foundAfter)
		}
	}
	return nil
}

// checkCollectionExists checks if a collection exists in a database.
func (o *MongoOptimizer) checkCollectionExists(ctx context.Context, dbName, collName string) (bool, error) {
	filter := bson.M{"name": collName}
	names, err := o.conn.Database(dbName).ListCollectionNames(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to list collections for existence check: %w", err)
	}
	return len(names) > 0, nil
}

// verifyIndexExists checks if an index with a specific name exists.
func (o *MongoOptimizer) verifyIndexExists(ctx context.Context, dbName, collName, indexName string) (bool, error) {
	indexesView := o.conn.Database(dbName).Collection(collName).Indexes()
	cursor, err := indexesView.List(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list indexes for verification: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			// Log error but continue checking other indexes if possible
			logger.Error("Failed to decode index during verification", "error", err)
			continue
		}
		if name, ok := idx["name"].(string); ok && name == indexName {
			return true, nil // Found the index
		}
	}
	return false, cursor.Err() // Return false and any cursor error
}

func (o *MongoOptimizer) applyQueryOptimization(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	// TODO: Implement query optimization based on suggestion.Solution.Operations
	if len(suggestion.Solution.Operations) > 0 {
		logger.Warn("Query optimization via structured operations not yet implemented", "database", databaseName)
		// Example: Loop through ops, check op.Action, build and run find/update/aggregate commands
	}
	return fmt.Errorf("query optimization not implemented") // Return error until implemented
	// Old logic using suggestion.Solution.Commands removed
}

func (o *MongoOptimizer) applySchemaOptimization(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	// TODO: Implement schema optimization based on suggestion.Solution.Operations
	if len(suggestion.Solution.Operations) > 0 {
		logger.Warn("Schema optimization via structured operations not yet implemented", "database", databaseName)
		// Example: Loop through ops, check op.Action, build and run collMod commands
	}
	return fmt.Errorf("schema optimization not implemented") // Return error until implemented
	// Old logic using suggestion.Solution.Commands removed
}

func (o *MongoOptimizer) applyConfigOptimization(ctx context.Context, databaseName string, suggestion *ai.OptimizationSuggestion) error {
	// TODO: Implement config optimization based on suggestion.Solution.Operations
	if len(suggestion.Solution.Operations) > 0 {
		logger.Warn("Configuration optimization via structured operations not yet implemented", "database", databaseName)
		// Example: Loop through ops, check op.Action, build and run setParameter commands
	}
	return fmt.Errorf("config optimization not implemented") // Return error until implemented
	// Old logic using suggestion.Solution.Commands removed
}

// validateStep might need databaseName too, depending on implementation
func (o *MongoOptimizer) validateStep(ctx context.Context, databaseName string, step string) (*ai.Metric, error) {
	// TODO: Implement validation logic - this likely involves running commands
	// or queries against the specific databaseName and parsing results.
	// For now, return a dummy metric.
	logger.Warn("Validation step execution not fully implemented", "step", step, "database", databaseName)
	return &ai.Metric{Name: step, Value: 0.0}, nil // Placeholder
}

// Helper functions extractCollectionName and compareIndexes are no longer needed and can be removed.
/*
func extractCollectionName(cmd bson.D) (string, error) {
	// ... implementation ...
}
*/

/*
func compareIndexes(idx bson.M, createCmd bson.D) bool {
	// ... implementation ...
}
*/

package optimizer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
func (o *MongoOptimizer) Apply(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	if o.conn == nil {
		return NewOptimizerError(ErrorTypeConnection, "MongoDB connection is nil", nil)
	}

	// Record initial metrics for validation
	initialMetrics := make(map[string]float64)
	for _, metric := range suggestion.Problem.Metrics {
		initialMetrics[metric.Name] = metric.Value
	}

	logger.Info("Applying optimization",
		"category", suggestion.Category,
		"impact", suggestion.Impact,
		"confidence", suggestion.Confidence)

	var err error
	switch suggestion.Category {
	case "index":
		err = o.applyIndexOptimization(ctx, suggestion)
	case "query":
		err = o.applyQueryOptimization(ctx, suggestion)
	case "schema":
		err = o.applySchemaOptimization(ctx, suggestion)
	case "configuration":
		err = o.applyConfigOptimization(ctx, suggestion)
	default:
		return NewOptimizerError(ErrorTypeUnknown, fmt.Sprintf("Unknown optimization category: %s", suggestion.Category), nil)
	}

	if err != nil {
		logger.Error("Failed to apply optimization",
			"category", suggestion.Category,
			"error", err)
		return err
	}

	logger.Info("Successfully applied optimization", "category", suggestion.Category)
	return nil
}

// Validate checks if the optimization was successful
func (o *MongoOptimizer) Validate(ctx context.Context, suggestion *ai.OptimizationSuggestion) (*ValidationResult, error) {
	result := &ValidationResult{
		Category: suggestion.Category,
		Metrics:  make(map[string]Metric),
	}

	// Run validation steps from the suggestion
	var totalImprovement float64
	for _, step := range suggestion.Validation {
		metric, err := o.validateStep(ctx, step)
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
func (o *MongoOptimizer) Rollback(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	if o.conn == nil {
		return NewOptimizerError(ErrorTypeConnection, "MongoDB connection is nil", nil)
	}

	if suggestion == nil {
		return NewOptimizerError(ErrorTypeRollback, "Suggestion is nil", nil)
	}

	if suggestion.RollbackPlan == "" {
		return NewOptimizerError(ErrorTypeRollback, "No rollback plan provided", nil)
	}

	logger.Info("Executing rollback plan", "category", suggestion.Category)

	var rollbackCmd bson.D
	if err := bson.UnmarshalExtJSON([]byte(suggestion.RollbackPlan), true, &rollbackCmd); err != nil {
		return NewOptimizerError(ErrorTypeRollback, "Invalid rollback plan format", err)
	}

	if err := o.conn.Database("admin").RunCommand(ctx, rollbackCmd).Err(); err != nil {
		return NewOptimizerError(ErrorTypeRollback, "Failed to execute rollback command", err).
			WithCommand(fmt.Sprintf("%v", rollbackCmd))
	}

	logger.Info("Rollback completed successfully", "category", suggestion.Category)
	return nil
}

func (o *MongoOptimizer) applyIndexOptimization(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	for _, cmd := range suggestion.Solution.Commands {
		var indexCmd bson.D
		if err := bson.UnmarshalExtJSON([]byte(cmd), true, &indexCmd); err != nil {
			return fmt.Errorf("invalid index command: %w", err)
		}

		// Extract database and collection from the command
		dbName, collName, err := extractDbCollection(indexCmd)
		if err != nil {
			return err
		}

		// Apply the index command
		if err := o.conn.Database(dbName).RunCommand(ctx, indexCmd).Err(); err != nil {
			return fmt.Errorf("failed to apply index optimization: %w", err)
		}

		// Verify index was created
		indexes := o.conn.Database(dbName).Collection(collName).Indexes()
		cursor, err := indexes.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to verify index creation: %w", err)
		}
		defer cursor.Close(ctx)

		var found bool
		for cursor.Next(ctx) {
			var idx bson.M
			if err := cursor.Decode(&idx); err != nil {
				return fmt.Errorf("failed to decode index: %w", err)
			}
			if compareIndexes(idx, indexCmd) {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("index was not created successfully")
		}
	}
	return nil
}

func (o *MongoOptimizer) applyQueryOptimization(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	for _, cmd := range suggestion.Solution.Commands {
		var queryCmd struct {
			Collection string  `json:"collection"`
			Database   string  `json:"database"`
			Operation  string  `json:"operation"`
			Query      bson.D  `json:"query"`
			Update     bson.D  `json:"update,omitempty"`
			Options    *bson.D `json:"options,omitempty"`
			Hint       string  `json:"hint,omitempty"`
		}

		if err := json.Unmarshal([]byte(cmd), &queryCmd); err != nil {
			return fmt.Errorf("invalid query command: %w", err)
		}

		coll := o.conn.Database(queryCmd.Database).Collection(queryCmd.Collection)

		switch queryCmd.Operation {
		case "createView":
			if err := coll.Database().CreateView(ctx, queryCmd.Collection+"_view", queryCmd.Collection, mongo.Pipeline{queryCmd.Query}); err != nil {
				return fmt.Errorf("failed to create view: %w", err)
			}
		case "addHint":
			// Add hint to collection metadata
			if err := coll.Database().RunCommand(ctx, bson.D{
				{Key: "collMod", Value: queryCmd.Collection},
				{Key: "index", Value: queryCmd.Hint},
			}).Err(); err != nil {
				return fmt.Errorf("failed to add hint: %w", err)
			}
		default:
			return fmt.Errorf("unsupported query operation: %s", queryCmd.Operation)
		}
	}
	return nil
}

func (o *MongoOptimizer) applySchemaOptimization(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	for _, cmd := range suggestion.Solution.Commands {
		var schemaCmd struct {
			Collection string `json:"collection"`
			Database   string `json:"database"`
			Operation  string `json:"operation"`
			Pipeline   bson.A `json:"pipeline,omitempty"`
			Validator  bson.D `json:"validator,omitempty"`
		}

		if err := json.Unmarshal([]byte(cmd), &schemaCmd); err != nil {
			return fmt.Errorf("invalid schema command: %w", err)
		}

		db := o.conn.Database(schemaCmd.Database)

		switch schemaCmd.Operation {
		case "aggregate":
			// Run aggregation pipeline to transform data
			if _, err := db.Collection(schemaCmd.Collection).Aggregate(ctx, schemaCmd.Pipeline); err != nil {
				return fmt.Errorf("failed to run schema transformation: %w", err)
			}
		case "setValidator":
			// Update collection with schema validation
			if err := db.RunCommand(ctx, bson.D{
				{Key: "collMod", Value: schemaCmd.Collection},
				{Key: "validator", Value: schemaCmd.Validator},
			}).Err(); err != nil {
				return fmt.Errorf("failed to set validator: %w", err)
			}
		default:
			return fmt.Errorf("unsupported schema operation: %s", schemaCmd.Operation)
		}
	}
	return nil
}

func (o *MongoOptimizer) applyConfigOptimization(ctx context.Context, suggestion *ai.OptimizationSuggestion) error {
	for _, cmd := range suggestion.Solution.Commands {
		var configCmd bson.D
		if err := bson.UnmarshalExtJSON([]byte(cmd), true, &configCmd); err != nil {
			return fmt.Errorf("invalid config command: %w", err)
		}

		// Apply configuration change
		if err := o.conn.Database("admin").RunCommand(ctx, configCmd).Err(); err != nil {
			return fmt.Errorf("failed to apply config change: %w", err)
		}

		// Verify the change was applied
		result := o.conn.Database("admin").RunCommand(ctx, bson.D{{Key: "getParameter", Value: "*"}})
		var params bson.M
		if err := result.Decode(&params); err != nil {
			return fmt.Errorf("failed to verify config change: %w", err)
		}

		// Check if the parameter was set correctly
		for _, elem := range configCmd {
			if elem.Key == "setParameter" {
				paramValue, ok := params[elem.Key]
				if !ok || paramValue != elem.Value {
					return fmt.Errorf("config parameter %s was not set correctly", elem.Key)
				}
			}
		}
	}
	return nil
}

func (o *MongoOptimizer) validateStep(ctx context.Context, step string) (*ai.Metric, error) {
	// Parse the validation step
	var validation struct {
		MetricName string   `json:"metric_name"`
		Command    bson.D   `json:"command"`
		Extract    []string `json:"extract"`
	}

	if err := json.Unmarshal([]byte(step), &validation); err != nil {
		return nil, fmt.Errorf("invalid validation step: %w", err)
	}

	// Run the validation command
	result := o.conn.Database("admin").RunCommand(ctx, validation.Command)
	var response bson.M
	if err := result.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to run validation: %w", err)
	}

	// Extract the metric value using the path
	value, err := extractMetricValue(response, validation.Extract)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metric: %w", err)
	}

	return &ai.Metric{
		Name:  validation.MetricName,
		Value: value,
	}, nil
}

// Helper functions

func extractDbCollection(cmd bson.D) (string, string, error) {
	var dbName, collName string
	for _, elem := range cmd {
		switch elem.Key {
		case "createIndexes":
			collName = elem.Value.(string)
		case "$db":
			dbName = elem.Value.(string)
		}
	}
	if dbName == "" || collName == "" {
		return "", "", fmt.Errorf("could not extract database and collection names")
	}
	return dbName, collName, nil
}

func compareIndexes(existing bson.M, new bson.D) bool {
	// Convert new index to map for comparison
	newMap := make(map[string]interface{})
	for _, elem := range new {
		newMap[elem.Key] = elem.Value
	}

	// Compare relevant fields
	relevantFields := []string{"key", "unique", "sparse", "background"}
	for _, field := range relevantFields {
		if existing[field] != newMap[field] {
			return false
		}
	}
	return true
}

func extractMetricValue(data bson.M, path []string) (float64, error) {
	current := interface{}(data)
	for _, key := range path {
		switch v := current.(type) {
		case bson.M:
			var ok bool
			current, ok = v[key]
			if !ok {
				return 0, fmt.Errorf("key %s not found in path %s", key, strings.Join(path, "."))
			}
		default:
			return 0, fmt.Errorf("invalid path %s", strings.Join(path, "."))
		}
	}

	switch v := current.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("value at path %s is not a number", strings.Join(path, "."))
	}
}

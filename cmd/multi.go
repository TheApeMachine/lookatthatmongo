package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"github.com/theapemachine/lookatthatmongo/mongodb/optimizer"
	"github.com/theapemachine/lookatthatmongo/mongodb/tracker"
	"github.com/theapemachine/lookatthatmongo/storage"
)

var (
	databases    []string
	maxParallel  int
	databaseList string
	compareOnly  bool
)

// Result represents the result of optimizing a single database
type Result struct {
	DatabaseName  string
	Success       bool
	Error         error
	Improvement   float64
	Duration      time.Duration
	SuggestionIDs []string
}

/*
multiCmd represents the multi command that optimizes multiple databases simultaneously.
It can be used to optimize a list of databases in parallel and provide comparative analysis.
*/
var multiCmd = &cobra.Command{
	Use:   "multi",
	Short: "Optimize multiple MongoDB databases simultaneously",
	Long: `Run optimizations on multiple MongoDB databases simultaneously.
This command allows you to specify a list of databases to optimize in parallel
and provides a comparative analysis of the results.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Apply logging configuration
		cfg.ApplyLogging()

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Parse database list
		if databaseList == "" && len(databases) == 0 {
			return fmt.Errorf("at least one database must be specified using --databases or --db-list")
		}

		// Parse comma-separated database list if provided
		if databaseList != "" {
			databases = append(databases, parseDatabaseList(databaseList)...)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting multi-database optimization",
			"databases", fmt.Sprint(databases),
			"max_parallel", maxParallel,
			"compare_only", compareOnly)

		// Create storage for optimization history
		var store storage.Storage
		var err error

		// Initialize storage based on configuration
		if cfg.StorageType == "s3" {
			store, err = storage.NewS3Storage(cmd.Context(),
				storage.WithBucket(cfg.S3Bucket),
				storage.WithRegion(cfg.S3Region),
				storage.WithPrefix(cfg.S3Prefix),
			)
		} else {
			store, err = storage.NewFileStorage(cfg.StoragePath)
		}
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		// Process databases with concurrency control
		sem := make(chan struct{}, maxParallel)
		var wg sync.WaitGroup
		results := make([]Result, 0, len(databases))
		resultsChan := make(chan Result, len(databases))

		// Connect to MongoDB (shared connection)
		conn, err := mongodb.NewConn(cmd.Context(), cfg.MongoURI, "admin")
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		defer conn.Close(cmd.Context())

		// Start a goroutine for each database
		for _, dbName := range databases {
			wg.Add(1)
			sem <- struct{}{} // Acquire semaphore

			go func(ctx context.Context, dbName string) {
				defer wg.Done()
				defer func() { <-sem }() // Release semaphore

				start := time.Now()
				result := Result{
					DatabaseName: dbName,
					Success:      false,
				}

				// Process the database
				err := processDatabase(ctx, conn, store, dbName, compareOnly)
				duration := time.Since(start)

				if err != nil {
					result.Error = err
					logger.Error("Failed to process database",
						"database", dbName,
						"error", err,
						"duration", duration)
				} else {
					result.Success = true
					logger.Info("Successfully processed database",
						"database", dbName,
						"duration", duration)
				}

				result.Duration = duration
				resultsChan <- result
			}(cmd.Context(), dbName)
		}

		// Wait for all goroutines to complete and close the results channel
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Collect results
		for result := range resultsChan {
			results = append(results, result)
		}

		// Generate comparative report
		printComparativeReport(results)

		return nil
	},
}

// processDatabase processes a single database
func processDatabase(ctx context.Context, conn *mongodb.Conn, store storage.Storage, dbName string, compareOnly bool) error {
	logger.Info("Processing database", "database", dbName)

	// Set up monitoring
	monitor := mongodb.NewMonitor(mongodb.WithConn(conn))
	beforeReport := metrics.NewReport(monitor)

	// Collect metrics before optimization
	logger.Info("Collecting metrics", "database", dbName)
	err := beforeReport.Collect(ctx, dbName, func() ([]string, error) {
		return conn.Database(dbName).ListCollectionNames(ctx, struct{}{})
	})
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// If compare only, just store the report and exit
	if compareOnly {
		logger.Info("Compare-only mode, skipping optimization", "database", dbName)

		// Create a record with just the before report
		record := &storage.OptimizationRecord{
			ID:           generateRecordID(),
			Timestamp:    time.Now(),
			DatabaseName: dbName,
			BeforeReport: beforeReport,
			Applied:      false,
		}

		if err := store.SaveOptimizationRecord(ctx, record); err != nil {
			return fmt.Errorf("failed to save comparison record: %w", err)
		}

		return nil
	}

	// Create history tracker
	history := tracker.NewHistory(
		tracker.WithHistoryReport(beforeReport),
		tracker.WithDatabaseName(dbName),
	)

	// Generate optimization suggestions
	logger.Info("Generating optimization suggestions", "database", dbName)
	aiconn := ai.NewConn()
	prompt, err := ai.NewPrompt(ai.WithReport("before", beforeReport))
	if err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}
	suggestion, err := aiconn.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate optimization suggestions: %w", err)
	}

	// Add suggestion to history
	// Convert the generic map[string]interface{} to *ai.OptimizationSuggestion
	var typedSuggestion *ai.OptimizationSuggestion
	suggestionBytes, err := json.Marshal(suggestion)
	if err != nil {
		return fmt.Errorf("failed to marshal raw suggestion: %w", err)
	}
	if err := json.Unmarshal(suggestionBytes, &typedSuggestion); err != nil {
		return fmt.Errorf("failed to unmarshal suggestion into typed struct: %w", err)
	}
	history.AddOptimization(typedSuggestion)

	// Apply optimizations
	logger.Info("Applying optimizations",
		"database", dbName,
		"category", typedSuggestion.Category,
		"impact", typedSuggestion.Impact)

	opt := optimizer.NewOptimizer(
		optimizer.WithConnection(conn),
		optimizer.WithMonitor(monitor),
	)

	if err := opt.Apply(ctx, dbName, typedSuggestion); err != nil { // Pass dbName
		// Attempt rollback on failure if enabled
		if cfg.EnableRollback {
			logger.Error("Optimization failed, attempting rollback",
				"database", dbName,
				"error", err)
			if rbErr := opt.Rollback(ctx, dbName, typedSuggestion); rbErr != nil { // Pass dbName
				return fmt.Errorf("optimization failed and rollback failed: %v (rollback: %v)", err, rbErr)
			}
			return fmt.Errorf("optimization failed but rolled back successfully: %v", err)
		}
		return fmt.Errorf("optimization failed: %v", err)
	}

	// Collect metrics after optimization
	logger.Info("Collecting metrics after optimization", "database", dbName)
	afterReport := metrics.NewReport(monitor)
	err = afterReport.Collect(ctx, dbName, func() ([]string, error) {
		return conn.Database(dbName).ListCollectionNames(ctx, struct{}{})
	})
	if err != nil {
		return fmt.Errorf("failed to collect metrics after optimization: %w", err)
	}

	// Update history with after report
	history.SetAfterReport(afterReport)

	// Create action handler
	actionHandler := tracker.NewActionHandler(
		tracker.WithStorage(store),
		tracker.WithAIConn(aiconn),
		tracker.WithActionHistory(history),
		tracker.WithThreshold(cfg.ImprovementThreshold),
		tracker.WithOptimizer(opt),
	)

	// Create measurement
	measurement := tracker.NewMeasurement(
		tracker.WithConn(aiconn),
		tracker.WithHistory(history),
		tracker.WithMeasurementStorage(store),
		tracker.WithActionHandler(actionHandler),
	)

	// Measure and take action
	logger.Info("Measuring optimization impact", "database", dbName)
	_, err = measurement.MeasureAndStore(ctx)
	if err != nil {
		return fmt.Errorf("measurement failed: %w", err)
	}

	// Validate the changes
	logger.Info("Validating optimization", "database", dbName)
	result, err := opt.Validate(ctx, dbName, typedSuggestion) // Pass dbName
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	logger.Info("Optimization completed",
		"database", dbName,
		"improvement", result.Improvement)

	return nil
}

// parseDatabaseList splits a comma-separated list of databases
func parseDatabaseList(list string) []string {
	if list == "" {
		return []string{}
	}

	// Split by commas, handle spaces
	var result []string
	for _, db := range splitAndTrim(list, ',') {
		if db != "" {
			result = append(result, db)
		}
	}
	return result
}

// splitAndTrim splits a string by the given separator and trims spaces
func splitAndTrim(s string, sep rune) []string {
	var result []string
	var builder strings.Builder
	var inQuotes bool

	for _, r := range s {
		if r == '"' {
			inQuotes = !inQuotes
			continue
		}
		if r == sep && !inQuotes {
			result = append(result, strings.TrimSpace(builder.String()))
			builder.Reset()
			continue
		}
		builder.WriteRune(r)
	}

	if builder.Len() > 0 {
		result = append(result, strings.TrimSpace(builder.String()))
	}

	return result
}

// generateRecordID generates a unique ID for a record
func generateRecordID() string {
	return fmt.Sprintf("record-%d", time.Now().UnixNano())
}

// printComparativeReport prints a comparative report of database optimizations
func printComparativeReport(results []Result) {
	logger.Info("===== Multi-Database Optimization Report =====")

	var successes, failures int
	for _, result := range results {
		if result.Success {
			successes++
		} else {
			failures++
		}

		statusStr := "Success"
		if !result.Success {
			statusStr = fmt.Sprintf("Failed: %v", result.Error)
		}

		logger.Info(fmt.Sprintf("Database: %s", result.DatabaseName),
			"status", statusStr,
			"duration", result.Duration.Round(time.Millisecond),
			"improvement", result.Improvement)
	}

	logger.Info("Summary",
		"total", len(results),
		"successes", successes,
		"failures", failures)
}

func init() {
	rootCmd.AddCommand(multiCmd)

	// Add flags specific to the multi command
	multiCmd.Flags().StringSliceVar(&databases, "databases", []string{}, "List of databases to optimize")
	multiCmd.Flags().StringVar(&databaseList, "db-list", "", "Comma-separated list of databases to optimize")
	multiCmd.Flags().IntVar(&maxParallel, "parallel", 2, "Maximum number of databases to process in parallel")
	multiCmd.Flags().BoolVar(&compareOnly, "compare-only", false, "Only collect metrics and compare databases, don't optimize")
}

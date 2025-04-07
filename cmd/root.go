/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/config"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/mongodb"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"github.com/theapemachine/lookatthatmongo/mongodb/optimizer"
	"github.com/theapemachine/lookatthatmongo/mongodb/tracker"
	"github.com/theapemachine/lookatthatmongo/storage"
)

var (
	cfg = config.New()
)

/*
rootCmd represents the base command when called without any subcommands.
It handles the main functionality of the MongoDB optimization platform.
*/
var rootCmd = &cobra.Command{
	Use:   "lookatthatmongo",
	Short: "AI-powered MongoDB optimization platform",
	Long:  rootLong,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Apply logging configuration
		cfg.ApplyLogging()

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting MongoDB optimization", "database", cfg.DatabaseName)

		// Create storage for optimization history based on configuration
		var store storage.Storage
		var err error

		if cfg.StorageType == config.S3Storage {
			logger.Info("Using S3 storage", "bucket", cfg.S3Bucket, "region", cfg.S3Region)
			store, err = storage.NewS3Storage(cmd.Context(),
				storage.WithBucket(cfg.S3Bucket),
				storage.WithRegion(cfg.S3Region),
				storage.WithPrefix(cfg.S3Prefix),
			)
		} else {
			// Default to file storage
			logger.Info("Using file storage", "path", cfg.StoragePath)
			store, err = storage.NewFileStorage(cfg.StoragePath)
		}

		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		// Connect to MongoDB
		conn, err := mongodb.NewConn(cmd.Context(), cfg.MongoURI, cfg.DatabaseName)
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		defer conn.Close(cmd.Context())
		logger.Info("Connected to MongoDB", "database", cfg.DatabaseName)

		// Set up monitoring
		monitor := mongodb.NewMonitor(mongodb.WithConn(conn))
		beforeReport := metrics.NewReport(monitor)

		// Collect metrics before optimization
		logger.Info("Collecting metrics before optimization")
		err = beforeReport.Collect(cmd.Context(), cfg.DatabaseName, func() ([]string, error) {
			return conn.Database(cfg.DatabaseName).ListCollectionNames(cmd.Context(), struct{}{})
		})
		if err != nil {
			return fmt.Errorf("failed to collect metrics: %w", err)
		}

		// Create history tracker
		history := tracker.NewHistory(
			tracker.WithHistoryReport(beforeReport),
			tracker.WithDatabaseName(cfg.DatabaseName),
		)

		// Generate optimization suggestions
		logger.Info("Generating optimization suggestions")
		aiconn := ai.NewConn()
		prompt, err := ai.NewPrompt(
			ai.WithReport("before", beforeReport),
			ai.WithSchema(ai.OptimizationSuggestionSchema),
		)
		if err != nil {
			return fmt.Errorf("failed to create prompt: %w", err)
		}
		rawSuggestion, err := aiconn.Generate(cmd.Context(), prompt)
		if err != nil {
			return fmt.Errorf("failed to generate optimization suggestions: %w", err)
		}

		// Convert the generic map[string]interface{} to *ai.OptimizationSuggestion
		var typedSuggestion *ai.OptimizationSuggestion
		suggestionBytes, err := json.Marshal(rawSuggestion)
		if err != nil {
			return fmt.Errorf("failed to marshal raw suggestion: %w", err)
		}
		if err := json.Unmarshal(suggestionBytes, &typedSuggestion); err != nil {
			return fmt.Errorf("failed to unmarshal suggestion into typed struct: %w", err)
		}

		// Now use typedSuggestion for subsequent operations
		history.AddOptimization(typedSuggestion)

		// Apply optimizations
		logger.Info("Applying optimizations",
			"category", typedSuggestion.Category,
			"impact", typedSuggestion.Impact)
		opt := optimizer.NewOptimizer(
			optimizer.WithConnection(conn),
			optimizer.WithMonitor(monitor),
		)

		if err := opt.Apply(cmd.Context(), cfg.DatabaseName, typedSuggestion); err != nil {
			// Attempt rollback on failure if enabled
			if cfg.EnableRollback {
				logger.Error("Optimization failed, attempting rollback", "error", err)
				if rbErr := opt.Rollback(cmd.Context(), cfg.DatabaseName, typedSuggestion); rbErr != nil {
					return fmt.Errorf("optimization failed and rollback failed: %v (rollback: %v)", err, rbErr)
				}
				return fmt.Errorf("optimization failed but rolled back successfully: %v", err)
			}
			return fmt.Errorf("optimization failed: %v", err)
		}

		// Collect metrics after optimization
		logger.Info("Collecting metrics after optimization")
		afterReport := metrics.NewReport(monitor)
		err = afterReport.Collect(cmd.Context(), cfg.DatabaseName, func() ([]string, error) {
			return conn.Database(cfg.DatabaseName).ListCollectionNames(cmd.Context(), struct{}{})
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
		logger.Info("Measuring optimization impact")
		_, err = measurement.MeasureAndStore(cmd.Context())
		if err != nil {
			return fmt.Errorf("measurement failed: %w", err)
		}

		// Validate the changes
		logger.Info("Validating optimization")
		result, err := opt.Validate(cmd.Context(), cfg.DatabaseName, typedSuggestion)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		logger.Info("Optimization completed",
			"improvement", result.Improvement,
			"database", cfg.DatabaseName)

		return nil
	},
}

/*
Execute adds all child commands to the root command and sets flags appropriately.
This is called by main.main(). It only needs to happen once to the rootCmd.
*/
func Execute() error {
	return rootCmd.Execute()
}

/*
init initializes the command by setting up flags and configuration.
It is automatically called when the package is imported.
*/
func init() {
	// Define flags
	rootCmd.Flags().StringVar(&cfg.DatabaseName, "db", cfg.DatabaseName, "MongoDB database name to optimize")

	// Storage flags
	rootCmd.Flags().StringVar((*string)(&cfg.StorageType), "storage-type", string(cfg.StorageType), "Storage type (file or s3)")
	rootCmd.Flags().StringVar(&cfg.StoragePath, "storage-path", cfg.StoragePath, "Path to store optimization history (for file storage)")

	// S3 storage flags
	rootCmd.Flags().StringVar(&cfg.S3Bucket, "s3-bucket", cfg.S3Bucket, "S3 bucket name for optimization history")
	rootCmd.Flags().StringVar(&cfg.S3Region, "s3-region", cfg.S3Region, "AWS region for S3 (default: us-east-1)")
	rootCmd.Flags().StringVar(&cfg.S3Prefix, "s3-prefix", cfg.S3Prefix, "Prefix for S3 objects (default: optimization-records/)")
	rootCmd.Flags().IntVar(&cfg.S3RetentionDays, "s3-retention-days", cfg.S3RetentionDays, "Number of days to keep records in S3 before auto-deletion")
	rootCmd.Flags().StringVar(&cfg.S3CredentialsFile, "s3-credentials", cfg.S3CredentialsFile, "Path to AWS credentials file")

	// Logging flags
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	// Optimization flags
	rootCmd.Flags().Float64Var(&cfg.ImprovementThreshold, "threshold", cfg.ImprovementThreshold, "Improvement threshold percentage")
	rootCmd.Flags().BoolVar(&cfg.EnableRollback, "enable-rollback", cfg.EnableRollback, "Enable automatic rollback on failure")
	rootCmd.Flags().IntVar(&cfg.MaxOptimizations, "max-optimizations", cfg.MaxOptimizations, "Maximum number of optimizations to apply")

	// Set up log level from flag
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cfg.SetLogLevel(logLevel)
	}
}

var (
	// Flag variables
	logLevel string
)

/*
rootLong contains the detailed description of the application shown in help.
*/
var rootLong = `
Look At That Mon Go - AI-powered MongoDB optimization platform

This tool intelligently analyzes, optimizes, and monitors your MongoDB deployments.
It collects metrics, identifies performance bottlenecks, recommends and applies
optimizations, and validates their impact—all with minimal human intervention.

Example usage:
  MONGO_URI="mongodb://username:password@hostname:port" lookatthatmongo --db myDatabase
`

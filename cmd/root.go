/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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

		// Create storage for optimization history
		store, err := storage.NewFileStorage(cfg.StoragePath)
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
		prompt := ai.NewPrompt(ai.WithReport("before", beforeReport))
		suggestion, err := aiconn.Generate(cmd.Context(), prompt)
		if err != nil {
			return fmt.Errorf("failed to generate optimization suggestions: %w", err)
		}

		// Add suggestion to history
		history.AddOptimization(suggestion.(*ai.OptimizationSuggestion))

		// Apply optimizations
		logger.Info("Applying optimizations",
			"category", suggestion.(*ai.OptimizationSuggestion).Category,
			"impact", suggestion.(*ai.OptimizationSuggestion).Impact)
		opt := optimizer.NewOptimizer(
			optimizer.WithConnection(conn),
			optimizer.WithMonitor(monitor),
		)

		if err := opt.Apply(cmd.Context(), suggestion.(*ai.OptimizationSuggestion)); err != nil {
			// Attempt rollback on failure if enabled
			if cfg.EnableRollback {
				logger.Error("Optimization failed, attempting rollback", "error", err)
				if rbErr := opt.Rollback(cmd.Context(), suggestion.(*ai.OptimizationSuggestion)); rbErr != nil {
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
		result, err := opt.Validate(cmd.Context(), suggestion.(*ai.OptimizationSuggestion))
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
	rootCmd.Flags().StringVar(&cfg.StoragePath, "storage", cfg.StoragePath, "Path to store optimization history")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
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

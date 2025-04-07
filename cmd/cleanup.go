package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/theapemachine/lookatthatmongo/config"
	"github.com/theapemachine/lookatthatmongo/logger"
	"github.com/theapemachine/lookatthatmongo/storage"
)

var (
	retentionDays int
)

/*
cleanupCmd represents the cleanup command that deletes old optimization records.
This is mainly useful for S3 storage to control storage costs, but can also be used with file storage.
*/
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up old optimization records",
	Long: `Delete optimization records that are older than the specified retention period.
This command is particularly useful for S3 storage to control storage costs.`,
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
		logger.Info("Starting cleanup of old optimization records",
			"retention_days", retentionDays,
			"storage_type", cfg.StorageType)

		// Create storage client based on configuration
		var store storage.Storage
		var err error

		if cfg.StorageType == config.S3Storage {
			logger.Info("Using S3 storage", "bucket", cfg.S3Bucket, "region", cfg.S3Region)
			store, err = storage.NewS3Storage(cmd.Context(),
				storage.WithBucket(cfg.S3Bucket),
				storage.WithRegion(cfg.S3Region),
				storage.WithPrefix(cfg.S3Prefix),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize S3 storage: %w", err)
			}

			// S3 storage has a DeleteOldRecords method
			s3Store, ok := store.(*storage.S3Storage)
			if !ok {
				return fmt.Errorf("unexpected error: S3 storage type assertion failed")
			}

			retention := time.Duration(retentionDays) * 24 * time.Hour
			count, err := s3Store.DeleteOldRecords(cmd.Context(), retention)
			if err != nil {
				return fmt.Errorf("failed to delete old records: %w", err)
			}

			logger.Info("Cleanup completed successfully", "deleted_records", count)
			return nil
		} else {
			// File storage - currently no built-in cleanup mechanism
			logger.Info("File storage cleanup not implemented",
				"storage_path", cfg.StoragePath)
			logger.Info("To clean up file storage, manually delete files older than the retention period")
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)

	// Add flags specific to the cleanup command
	cleanupCmd.Flags().IntVar(&retentionDays, "retention-days", 90, "Delete records older than this many days")
}

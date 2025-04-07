package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/lookatthatmongo/logger"
)

// StorageType represents the type of storage to use
type StorageType string

const (
	// FileStorage represents file-based storage
	FileStorage StorageType = "file"
	// S3Storage represents S3-based storage
	S3Storage StorageType = "s3"
)

/*
Config holds the application configuration including MongoDB connection settings,
storage settings, logging settings, and optimization parameters.
*/
type Config struct {
	// MongoDB connection settings
	MongoURI     string
	DatabaseName string

	// Storage settings
	StorageType StorageType
	StoragePath string

	// S3 Storage settings
	S3Bucket          string
	S3Region          string
	S3Prefix          string
	S3RetentionDays   int    // Number of days to keep records before auto-deletion
	S3CredentialsFile string // Path to AWS credentials file

	// Logging settings
	LogLevel log.Level

	// Optimization settings
	ImprovementThreshold float64
	EnableRollback       bool
	MaxOptimizations     int
}

/*
New creates a new configuration with default values.
It initializes configuration from environment variables or falls back to defaults.
*/
func New() *Config {
	// Set up default storage path in user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	defaultStoragePath := filepath.Join(home, ".lookatthatmongo", "history")

	return &Config{
		MongoURI:             os.Getenv("MONGO_URI"),
		DatabaseName:         getEnvWithDefault("MONGO_DB", "FanAppDev2"),
		StorageType:          StorageType(getEnvWithDefault("STORAGE_TYPE", string(FileStorage))),
		StoragePath:          getEnvWithDefault("STORAGE_PATH", defaultStoragePath),
		S3Bucket:             getEnvWithDefault("S3_BUCKET", ""),
		S3Region:             getEnvWithDefault("S3_REGION", "us-east-1"),
		S3Prefix:             getEnvWithDefault("S3_PREFIX", "optimization-records/"),
		S3RetentionDays:      parseInt(getEnvWithDefault("S3_RETENTION_DAYS", "90")),
		S3CredentialsFile:    getEnvWithDefault("S3_CREDENTIALS_FILE", ""),
		LogLevel:             parseLogLevel(getEnvWithDefault("LOG_LEVEL", "debug")),
		ImprovementThreshold: parseFloat(getEnvWithDefault("IMPROVEMENT_THRESHOLD", "5.0")),
		EnableRollback:       parseBool(getEnvWithDefault("ENABLE_ROLLBACK", "true")),
		MaxOptimizations:     parseInt(getEnvWithDefault("MAX_OPTIMIZATIONS", "3")),
	}
}

/*
Validate checks if the configuration is valid.
It ensures required fields are set and have appropriate values.
*/
func (c *Config) Validate() error {
	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI environment variable is required")
	}

	if c.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate storage-specific settings
	if c.StorageType == S3Storage {
		if c.S3Bucket == "" {
			return fmt.Errorf("S3_BUCKET environment variable is required when STORAGE_TYPE=s3")
		}
	} else if c.StorageType == FileStorage {
		// For file storage, no additional validation needed
	} else {
		return fmt.Errorf("invalid storage type: %s (valid values: file, s3)", c.StorageType)
	}

	return nil
}

/*
SetDatabaseName sets the database name in the configuration.
*/
func (c *Config) SetDatabaseName(name string) {
	if name != "" {
		c.DatabaseName = name
	}
}

/*
SetStoragePath sets the storage path in the configuration.
*/
func (c *Config) SetStoragePath(path string) {
	if path != "" {
		c.StoragePath = path
	}
}

/*
SetLogLevel sets the log level in the configuration.
*/
func (c *Config) SetLogLevel(level string) {
	if level != "" {
		c.LogLevel = parseLogLevel(level)
	}
}

/*
SetImprovementThreshold sets the improvement threshold in the configuration.
*/
func (c *Config) SetImprovementThreshold(threshold float64) {
	if threshold > 0 {
		c.ImprovementThreshold = threshold
	}
}

/*
ApplyLogging configures the logger based on the current configuration.
*/
func (c *Config) ApplyLogging() {
	logger.SetLevel(c.LogLevel)
	logger.Info("Logging configured",
		"level", c.LogLevel.String(),
		"database", c.DatabaseName)
}

/*
getEnvWithDefault retrieves an environment variable or returns a default value if not set.
*/
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

/*
parseLogLevel converts a string log level to a log.Level value.
*/
func parseLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

/*
parseFloat converts a string to a float64 value.
*/
func parseFloat(value string) float64 {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 5.0 // Default value
	}
	return f
}

/*
parseInt converts a string to an int value.
*/
func parseInt(value string) int {
	i, err := strconv.Atoi(value)
	if err != nil {
		return 3 // Default value
	}
	return i
}

/*
parseBool converts a string to a boolean value.
*/
func parseBool(value string) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return true // Default value
	}
	return b
}

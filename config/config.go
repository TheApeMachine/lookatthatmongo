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

/*
Config holds the application configuration including MongoDB connection settings,
storage settings, logging settings, and optimization parameters.
*/
type Config struct {
	// MongoDB connection settings
	MongoURI     string
	DatabaseName string

	// Storage settings
	StoragePath string

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
		StoragePath:          getEnvWithDefault("STORAGE_PATH", defaultStoragePath),
		LogLevel:             parseLogLevel(getEnvWithDefault("LOG_LEVEL", "info")),
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

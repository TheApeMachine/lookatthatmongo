package config

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("When creating a new configuration", t, func() {
		// Save original environment variables to restore later
		originalURI := os.Getenv("MONGO_URI")
		originalDB := os.Getenv("MONGO_DB")
		originalStoragePath := os.Getenv("STORAGE_PATH")
		originalLogLevel := os.Getenv("LOG_LEVEL")
		originalThreshold := os.Getenv("IMPROVEMENT_THRESHOLD")
		originalRollback := os.Getenv("ENABLE_ROLLBACK")
		originalMaxOpt := os.Getenv("MAX_OPTIMIZATIONS")

		// Clean up environment after test
		defer func() {
			os.Setenv("MONGO_URI", originalURI)
			os.Setenv("MONGO_DB", originalDB)
			os.Setenv("STORAGE_PATH", originalStoragePath)
			os.Setenv("LOG_LEVEL", originalLogLevel)
			os.Setenv("IMPROVEMENT_THRESHOLD", originalThreshold)
			os.Setenv("ENABLE_ROLLBACK", originalRollback)
			os.Setenv("MAX_OPTIMIZATIONS", originalMaxOpt)
		}()

		Convey("With default values", func() {
			// Clear environment variables to test defaults
			os.Unsetenv("MONGO_URI")
			os.Unsetenv("MONGO_DB")
			os.Unsetenv("STORAGE_PATH")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("IMPROVEMENT_THRESHOLD")
			os.Unsetenv("ENABLE_ROLLBACK")
			os.Unsetenv("MAX_OPTIMIZATIONS")

			config := New()

			home, _ := os.UserHomeDir()
			defaultStoragePath := filepath.Join(home, ".lookatthatmongo", "history")

			So(config.MongoURI, ShouldEqual, "")
			So(config.DatabaseName, ShouldEqual, "FanAppDev2")
			So(config.StoragePath, ShouldEqual, defaultStoragePath)
			So(config.ImprovementThreshold, ShouldEqual, 5.0)
			So(config.EnableRollback, ShouldBeTrue)
			So(config.MaxOptimizations, ShouldEqual, 3)
		})

		Convey("With environment variables set", func() {
			// Set environment variables
			os.Setenv("MONGO_URI", "mongodb://localhost:27017")
			os.Setenv("MONGO_DB", "testdb")
			os.Setenv("STORAGE_PATH", "/tmp/test-storage")
			os.Setenv("LOG_LEVEL", "debug")
			os.Setenv("IMPROVEMENT_THRESHOLD", "0.2")
			os.Setenv("ENABLE_ROLLBACK", "false")
			os.Setenv("MAX_OPTIMIZATIONS", "5")

			config := New()

			So(config.MongoURI, ShouldEqual, "mongodb://localhost:27017")
			So(config.DatabaseName, ShouldEqual, "testdb")
			So(config.StoragePath, ShouldEqual, "/tmp/test-storage")
			So(config.ImprovementThreshold, ShouldEqual, 0.2)
			So(config.EnableRollback, ShouldBeFalse)
			So(config.MaxOptimizations, ShouldEqual, 5)
		})
	})
}

func TestValidate(t *testing.T) {
	Convey("When validating configuration", t, func() {
		Convey("With valid configuration", func() {
			config := &Config{
				MongoURI:     "mongodb://localhost:27017",
				DatabaseName: "testdb",
				StoragePath:  "/tmp/test-storage",
			}

			err := config.Validate()
			So(err, ShouldBeNil)
		})

		Convey("With missing MongoDB URI", func() {
			config := &Config{
				MongoURI:     "",
				DatabaseName: "testdb",
				StoragePath:  "/tmp/test-storage",
			}

			err := config.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "MONGO_URI")
		})

		Convey("With missing database name", func() {
			config := &Config{
				MongoURI:     "mongodb://localhost:27017",
				DatabaseName: "",
				StoragePath:  "/tmp/test-storage",
			}

			err := config.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "database name")
		})
	})
}

func TestSetDatabaseName(t *testing.T) {
	Convey("When setting database name", t, func() {
		config := New()

		Convey("It should update the database name", func() {
			config.SetDatabaseName("newdb")
			So(config.DatabaseName, ShouldEqual, "newdb")
		})
	})
}

func TestSetStoragePath(t *testing.T) {
	Convey("When setting storage path", t, func() {
		config := New()

		Convey("It should update the storage path", func() {
			config.SetStoragePath("/new/path")
			So(config.StoragePath, ShouldEqual, "/new/path")
		})
	})
}

func TestSetLogLevel(t *testing.T) {
	Convey("When setting log level", t, func() {
		config := New()

		Convey("It should update the log level", func() {
			config.SetLogLevel("debug")
			So(config.LogLevel.String(), ShouldEqual, "debug")

			config.SetLogLevel("info")
			So(config.LogLevel.String(), ShouldEqual, "info")

			config.SetLogLevel("warn")
			So(config.LogLevel.String(), ShouldEqual, "warn")

			config.SetLogLevel("error")
			So(config.LogLevel.String(), ShouldEqual, "error")

			// Invalid level should default to info
			config.SetLogLevel("invalid")
			So(config.LogLevel.String(), ShouldEqual, "info")
		})
	})
}

func TestSetImprovementThreshold(t *testing.T) {
	Convey("When setting improvement threshold", t, func() {
		config := New()

		Convey("It should update the improvement threshold", func() {
			config.SetImprovementThreshold(0.5)
			So(config.ImprovementThreshold, ShouldEqual, 0.5)
		})
	})
}

func TestApplyLogging(t *testing.T) {
	Convey("When applying logging configuration", t, func() {
		config := New()

		Convey("It should not panic", func() {
			So(func() { config.ApplyLogging() }, ShouldNotPanic)
		})
	})
}

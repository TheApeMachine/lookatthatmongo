package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewFileStorage(t *testing.T) {
	Convey("Given a need for file storage", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		Convey("When creating a new file storage", func() {
			storage, err := NewFileStorage(tempDir)

			Convey("Then it should be created successfully", func() {
				So(err, ShouldBeNil)
				So(storage, ShouldNotBeNil)
				So(storage.basePath, ShouldEqual, tempDir)
			})
		})

		Convey("When creating a file storage with a non-existent path", func() {
			nonExistentPath := filepath.Join(tempDir, "non_existent")
			storage, err := NewFileStorage(nonExistentPath)

			Convey("Then it should create the directory and succeed", func() {
				So(err, ShouldBeNil)
				So(storage, ShouldNotBeNil)

				// Check that the directory was created
				_, err := os.Stat(nonExistentPath)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestSaveOptimizationRecord(t *testing.T) {
	Convey("Given a file storage instance", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		So(err, ShouldBeNil)

		Convey("When saving an optimization record", func() {
			record := mockOptimizationRecord()
			ctx := context.Background()

			err := storage.SaveOptimizationRecord(ctx, record)

			Convey("Then it should save successfully", func() {
				So(err, ShouldBeNil)

				// Check that the file was created
				expectedPath := filepath.Join(tempDir, record.DatabaseName, record.ID+".json")
				_, err := os.Stat(expectedPath)
				So(err, ShouldBeNil)
			})
		})

		Convey("When saving a record with no ID", func() {
			record := mockOptimizationRecord()
			record.ID = ""
			ctx := context.Background()

			err := storage.SaveOptimizationRecord(ctx, record)

			Convey("Then it should generate an ID and save successfully", func() {
				So(err, ShouldBeNil)
				So(record.ID, ShouldNotBeEmpty)

				// Check that the file was created
				expectedPath := filepath.Join(tempDir, record.DatabaseName, record.ID+".json")
				_, err := os.Stat(expectedPath)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetOptimizationRecord(t *testing.T) {
	Convey("Given a file storage instance with a saved record", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		So(err, ShouldBeNil)

		// Save a record first
		record := mockOptimizationRecord()
		ctx := context.Background()
		err = storage.SaveOptimizationRecord(ctx, record)
		So(err, ShouldBeNil)

		Convey("When retrieving the record by ID", func() {
			retrievedRecord, err := storage.GetOptimizationRecord(ctx, record.ID)

			Convey("Then it should retrieve the record successfully", func() {
				So(err, ShouldBeNil)
				So(retrievedRecord, ShouldNotBeNil)
				So(retrievedRecord.ID, ShouldEqual, record.ID)
				So(retrievedRecord.DatabaseName, ShouldEqual, record.DatabaseName)
			})
		})

		Convey("When retrieving a non-existent record", func() {
			retrievedRecord, err := storage.GetOptimizationRecord(ctx, "non-existent-id")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(retrievedRecord, ShouldBeNil)
			})
		})
	})
}

func TestListOptimizationRecords(t *testing.T) {
	Convey("Given a file storage instance with multiple saved records", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		So(err, ShouldBeNil)

		// Save multiple records
		ctx := context.Background()
		record1 := mockOptimizationRecord()
		record1.ID = "test-id-1"
		record1.Timestamp = time.Now().Add(-1 * time.Hour)

		record2 := mockOptimizationRecord()
		record2.ID = "test-id-2"
		record2.Timestamp = time.Now()

		err = storage.SaveOptimizationRecord(ctx, record1)
		So(err, ShouldBeNil)

		err = storage.SaveOptimizationRecord(ctx, record2)
		So(err, ShouldBeNil)

		Convey("When listing all records", func() {
			records, err := storage.ListOptimizationRecords(ctx)

			Convey("Then it should return all records", func() {
				So(err, ShouldBeNil)
				So(records, ShouldNotBeNil)
				So(len(records), ShouldEqual, 2)

				// Records should be sorted by timestamp (newest first)
				So(records[0].ID, ShouldEqual, record2.ID)
				So(records[1].ID, ShouldEqual, record1.ID)
			})
		})
	})
}

func TestListOptimizationRecordsByDatabase(t *testing.T) {
	Convey("Given a file storage instance with records from different databases", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		So(err, ShouldBeNil)

		// Save records for different databases
		ctx := context.Background()
		record1 := mockOptimizationRecord()
		record1.ID = "test-id-1"
		record1.DatabaseName = "db1"

		record2 := mockOptimizationRecord()
		record2.ID = "test-id-2"
		record2.DatabaseName = "db2"

		err = storage.SaveOptimizationRecord(ctx, record1)
		So(err, ShouldBeNil)

		err = storage.SaveOptimizationRecord(ctx, record2)
		So(err, ShouldBeNil)

		Convey("When listing records for a specific database", func() {
			records, err := storage.ListOptimizationRecordsByDatabase(ctx, "db1")

			Convey("Then it should return only records for that database", func() {
				So(err, ShouldBeNil)
				So(records, ShouldNotBeNil)
				So(len(records), ShouldEqual, 1)
				So(records[0].ID, ShouldEqual, record1.ID)
				So(records[0].DatabaseName, ShouldEqual, "db1")
			})
		})
	})
}

func TestGetLatestOptimizationRecord(t *testing.T) {
	Convey("Given a file storage instance with multiple records", t, func() {
		// Create a temporary directory for testing
		tempDir, err := os.MkdirTemp("", "file_storage_test")
		So(err, ShouldBeNil)

		// Clean up after the test
		defer os.RemoveAll(tempDir)

		storage, err := NewFileStorage(tempDir)
		So(err, ShouldBeNil)

		// Save multiple records with different timestamps
		ctx := context.Background()
		record1 := mockOptimizationRecord()
		record1.ID = "test-id-1"
		record1.Timestamp = time.Now().Add(-1 * time.Hour)

		record2 := mockOptimizationRecord()
		record2.ID = "test-id-2"
		record2.Timestamp = time.Now()

		err = storage.SaveOptimizationRecord(ctx, record1)
		So(err, ShouldBeNil)

		err = storage.SaveOptimizationRecord(ctx, record2)
		So(err, ShouldBeNil)

		Convey("When getting the latest record", func() {
			latestRecord, err := storage.GetLatestOptimizationRecord(ctx)

			Convey("Then it should return the most recent record", func() {
				So(err, ShouldBeNil)
				So(latestRecord, ShouldNotBeNil)
				So(latestRecord.ID, ShouldEqual, record2.ID)
			})
		})
	})
}

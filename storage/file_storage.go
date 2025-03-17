package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/theapemachine/lookatthatmongo/logger"
)

/*
FileStorage implements the Storage interface using the local filesystem.
It stores optimization records as JSON files in a directory structure.
*/
type FileStorage struct {
	basePath string
	mu       sync.RWMutex
}

/*
NewFileStorage creates a new file storage instance.
It initializes the storage directory if it doesn't exist.
*/
func NewFileStorage(basePath string) (*FileStorage, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

/*
SaveOptimizationRecord saves an optimization record to a file.
It generates a unique ID and timestamp if not provided.
*/
func (fs *FileStorage) SaveOptimizationRecord(ctx context.Context, record *OptimizationRecord) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate a unique ID if not provided
	if record.ID == "" {
		record.ID = uuid.New().String()
	}

	// Ensure timestamp is set
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// Ensure the database directory exists
	dbDir := filepath.Join(fs.basePath, record.DatabaseName)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Marshal the record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal optimization record: %w", err)
	}

	// Save to a file named after the record ID
	filePath := filepath.Join(dbDir, record.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write optimization record: %w", err)
	}

	logger.Info("Saved optimization record",
		"id", record.ID,
		"database", record.DatabaseName,
		"improvement", record.ImprovementPct,
		"success", record.Success)

	return nil
}

/*
GetOptimizationRecord retrieves an optimization record by ID.
It reads the record from the file system.
*/
func (fs *FileStorage) GetOptimizationRecord(ctx context.Context, id string, dbName ...string) (*OptimizationRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// If database name is provided, use it; otherwise search in all databases
	if len(dbName) > 0 && dbName[0] != "" {
		// Look in the specified database directory
		dbDir := filepath.Join(fs.basePath, dbName[0])
		filePath := filepath.Join(dbDir, id+".json")
		return fs.readRecordFromFile(filePath)
	}

	// Search in all database directories
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			filePath := filepath.Join(fs.basePath, entry.Name(), id+".json")
			if record, err := fs.readRecordFromFile(filePath); err == nil {
				return record, nil
			}
		}
	}

	return nil, fmt.Errorf("optimization record not found: %s", id)
}

// Helper function to read a record from a file
func (fs *FileStorage) readRecordFromFile(filePath string) (*OptimizationRecord, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read record file: %w", err)
	}

	var record OptimizationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return &record, nil
}

/*
ListOptimizationRecords lists all optimization records.
It returns all records across all databases sorted by timestamp (newest first).
*/
func (fs *FileStorage) ListOptimizationRecords(ctx context.Context) ([]*OptimizationRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var records []*OptimizationRecord

	// Read all database directories
	dbDirs, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	// Iterate through each database directory
	for _, dbDir := range dbDirs {
		if !dbDir.IsDir() {
			continue
		}

		// Read all records in this database directory
		dbDirPath := filepath.Join(fs.basePath, dbDir.Name())
		files, err := os.ReadDir(dbDirPath)
		if err != nil {
			// Skip directories we can't read
			continue
		}

		// Process each file in the database directory
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			filePath := filepath.Join(dbDirPath, file.Name())
			record, err := fs.readRecordFromFile(filePath)
			if err != nil {
				// Skip files we can't read
				continue
			}

			records = append(records, record)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	if len(records) == 0 {
		// Return empty slice rather than nil to avoid test failures
		return []*OptimizationRecord{}, nil
	}

	return records, nil
}

/*
ListOptimizationRecordsByDatabase lists all optimization records for a specific database.
It filters the records by the specified database name.
*/
func (fs *FileStorage) ListOptimizationRecordsByDatabase(ctx context.Context, dbName string) ([]*OptimizationRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var records []*OptimizationRecord

	// Check if the database directory exists
	dbDirPath := filepath.Join(fs.basePath, dbName)
	files, err := os.ReadDir(dbDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty slice rather than nil to avoid test failures
			return []*OptimizationRecord{}, nil
		}
		return nil, fmt.Errorf("failed to read database directory: %w", err)
	}

	// Process each file in the database directory
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dbDirPath, file.Name())
		record, err := fs.readRecordFromFile(filePath)
		if err != nil {
			// Skip files we can't read
			continue
		}

		records = append(records, record)
	}

	// Sort by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	if len(records) == 0 {
		// Return empty slice rather than nil to avoid test failures
		return []*OptimizationRecord{}, nil
	}

	return records, nil
}

/*
GetLatestOptimizationRecord retrieves the most recent optimization record.
It returns the record with the latest timestamp.
*/
func (fs *FileStorage) GetLatestOptimizationRecord(ctx context.Context) (*OptimizationRecord, error) {
	records, err := fs.ListOptimizationRecords(ctx)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no optimization records found")
	}

	// Records are already sorted by timestamp (newest first)
	return records[0], nil
}

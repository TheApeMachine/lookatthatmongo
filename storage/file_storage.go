package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	// Set timestamp if not provided
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	// Create the file path
	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%s.json", record.ID))

	// Marshal the record to JSON
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write record file: %w", err)
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
It reads and deserializes the JSON file for the specified record.
*/
func (fs *FileStorage) GetOptimizationRecord(ctx context.Context, id string) (*OptimizationRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%s.json", id))

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
It scans the storage directory for record files and deserializes them.
*/
func (fs *FileStorage) ListOptimizationRecords(ctx context.Context) ([]*OptimizationRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	files, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var records []*OptimizationRecord
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(fs.basePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read record file", "file", file.Name(), "error", err)
			continue
		}

		var record OptimizationRecord
		if err := json.Unmarshal(data, &record); err != nil {
			logger.Warn("Failed to unmarshal record", "file", file.Name(), "error", err)
			continue
		}

		records = append(records, &record)
	}

	// Sort records by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

/*
ListOptimizationRecordsByDatabase lists optimization records for a specific database.
It filters the records by database name.
*/
func (fs *FileStorage) ListOptimizationRecordsByDatabase(ctx context.Context, dbName string) ([]*OptimizationRecord, error) {
	records, err := fs.ListOptimizationRecords(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*OptimizationRecord
	for _, record := range records {
		if record.DatabaseName == dbName {
			filtered = append(filtered, record)
		}
	}

	return filtered, nil
}

/*
GetLatestOptimizationRecord gets the most recent optimization record.
It sorts all records by timestamp and returns the most recent one.
*/
func (fs *FileStorage) GetLatestOptimizationRecord(ctx context.Context) (*OptimizationRecord, error) {
	records, err := fs.ListOptimizationRecords(ctx)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no optimization records found")
	}

	return records[0], nil
}

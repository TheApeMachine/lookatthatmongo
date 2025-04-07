package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/charmbracelet/log"
	"github.com/theapemachine/lookatthatmongo/logger"
)

const (
	// DefaultRecordsPrefix is the default prefix for optimization records in S3
	DefaultRecordsPrefix = "optimization-records/"
)

// S3StorageError defines custom errors for S3Storage
type S3StorageError struct {
	Message string
	Err     error
}

// Error returns the error message
func (e *S3StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// S3StorageOption defines options for the S3Storage
type S3StorageOption func(*S3Storage)

// WithBucket sets the S3 bucket name
func WithBucket(bucket string) S3StorageOption {
	return func(s *S3Storage) {
		s.bucket = bucket
	}
}

// WithPrefix sets the S3 object prefix
func WithPrefix(prefix string) S3StorageOption {
	return func(s *S3Storage) {
		s.prefix = prefix
	}
}

// WithRegion sets the AWS region
func WithRegion(region string) S3StorageOption {
	return func(s *S3Storage) {
		s.region = region
	}
}

// WithClient sets a custom S3 client
func WithClient(client *s3.Client) S3StorageOption {
	return func(s *S3Storage) {
		s.client = client
	}
}

// S3Storage implements the Storage interface using AWS S3
type S3Storage struct {
	bucket string
	prefix string
	region string
	client s3ClientAPI
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(ctx context.Context, opts ...S3StorageOption) (*S3Storage, error) {
	storage := &S3Storage{
		prefix: DefaultRecordsPrefix,
		region: "us-east-1", // Default region
	}

	// Apply options
	for _, opt := range opts {
		opt(storage)
	}

	// Bucket name is required
	if storage.bucket == "" {
		return nil, &S3StorageError{Message: "bucket name is required"}
	}

	// If client is not provided, create one
	if storage.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(storage.region))
		if err != nil {
			return nil, &S3StorageError{Message: "failed to load AWS config", Err: err}
		}
		storage.client = s3.NewFromConfig(cfg)
	}

	// Ensure the bucket exists
	_, err := storage.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(storage.bucket),
	})
	if err != nil {
		return nil, &S3StorageError{Message: "failed to access S3 bucket", Err: err}
	}

	logger.Info("S3 storage initialized", "bucket", storage.bucket, "prefix", storage.prefix)
	return storage, nil
}

// getObjectKey generates the S3 object key for a record
func (s *S3Storage) getObjectKey(record *OptimizationRecord) string {
	// Format: optimization-records/database-name/record-id.json
	return filepath.Join(s.prefix, record.DatabaseName, fmt.Sprintf("%s.json", record.ID))
}

// parseObjectKey extracts record ID and database name from an object key
func (s *S3Storage) parseObjectKey(key string) (id, dbName string) {
	// Remove the prefix
	key = strings.TrimPrefix(key, s.prefix)
	key = strings.TrimPrefix(key, "/")

	// Split by directory separator
	parts := strings.Split(key, "/")
	if len(parts) < 2 {
		return "", ""
	}

	// Last part is the filename (record-id.json)
	id = strings.TrimSuffix(parts[len(parts)-1], ".json")

	// Second-to-last part is the database name
	dbName = parts[len(parts)-2]

	return id, dbName
}

// SaveOptimizationRecord saves an optimization record to S3
func (s *S3Storage) SaveOptimizationRecord(ctx context.Context, record *OptimizationRecord) error {
	if record == nil {
		return &S3StorageError{Message: "record cannot be nil"}
	}

	// Generate the object key
	key := s.getObjectKey(record)

	// Convert record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return &S3StorageError{Message: "failed to marshal record to JSON", Err: err}
	}

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return &S3StorageError{Message: "failed to upload record to S3", Err: err}
	}

	logger.Debug("Saved optimization record to S3", "id", record.ID, "key", key)
	return nil
}

// GetOptimizationRecord retrieves an optimization record by ID
func (s *S3Storage) GetOptimizationRecord(ctx context.Context, id string, dbName ...string) (*OptimizationRecord, error) {
	if id == "" {
		return nil, &S3StorageError{Message: "record ID cannot be empty"}
	}

	var prefix string
	if len(dbName) > 0 && dbName[0] != "" {
		// If database name is provided, use it to narrow down the search
		prefix = filepath.Join(s.prefix, dbName[0])
	} else {
		// Otherwise, search in all databases
		prefix = s.prefix
	}

	// List objects in the bucket with the specified prefix
	resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, &S3StorageError{Message: "failed to list objects in S3", Err: err}
	}

	// Search for the record with matching ID
	for _, obj := range resp.Contents {
		objID, _ := s.parseObjectKey(*obj.Key)
		if objID == id {
			return s.getRecord(ctx, *obj.Key)
		}
	}

	return nil, &S3StorageError{Message: fmt.Sprintf("record with ID %s not found", id)}
}

// getRecord retrieves and parses a record from S3
func (s *S3Storage) getRecord(ctx context.Context, key string) (*OptimizationRecord, error) {
	// Get the object from S3
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, &S3StorageError{Message: fmt.Sprintf("record with key %s not found", key)}
		}
		return nil, &S3StorageError{Message: "failed to get object from S3", Err: err}
	}
	defer resp.Body.Close()

	// Read the object body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &S3StorageError{Message: "failed to read object body", Err: err}
	}

	// Parse the JSON data
	var record OptimizationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, &S3StorageError{Message: "failed to unmarshal record from JSON", Err: err}
	}

	return &record, nil
}

// ListOptimizationRecords lists all optimization records
func (s *S3Storage) ListOptimizationRecords(ctx context.Context) ([]*OptimizationRecord, error) {
	records, err := s.listRecordsByPrefix(ctx, s.prefix)
	if err != nil {
		return nil, err
	}

	// Sort records by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

// ListOptimizationRecordsByDatabase lists optimization records for a specific database
func (s *S3Storage) ListOptimizationRecordsByDatabase(ctx context.Context, dbName string) ([]*OptimizationRecord, error) {
	if dbName == "" {
		return nil, &S3StorageError{Message: "database name cannot be empty"}
	}

	prefix := filepath.Join(s.prefix, dbName)
	records, err := s.listRecordsByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Sort records by timestamp (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

// listRecordsByPrefix lists all records with a specific prefix
func (s *S3Storage) listRecordsByPrefix(ctx context.Context, prefix string) ([]*OptimizationRecord, error) {
	var records []*OptimizationRecord

	// Use pagination to handle large numbers of objects
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &S3StorageError{Message: "failed to list objects in S3", Err: err}
		}

		// Process each object in the page
		for _, obj := range page.Contents {
			record, err := s.getRecord(ctx, *obj.Key)
			if err != nil {
				log.Warn("Failed to get record", "key", *obj.Key, "error", err)
				continue
			}
			records = append(records, record)
		}
	}

	return records, nil
}

// GetLatestOptimizationRecord gets the most recent optimization record
func (s *S3Storage) GetLatestOptimizationRecord(ctx context.Context) (*OptimizationRecord, error) {
	// List all records
	records, err := s.ListOptimizationRecords(ctx)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, &S3StorageError{Message: "no optimization records found"}
	}

	// Records are already sorted by timestamp (newest first)
	return records[0], nil
}

// DeleteOldRecords deletes records older than the specified duration
func (s *S3Storage) DeleteOldRecords(ctx context.Context, age time.Duration) (int, error) {
	// List all records
	records, err := s.ListOptimizationRecords(ctx)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-age)
	var objectsToDelete []types.ObjectIdentifier
	var count int

	// Identify records to delete
	for _, record := range records {
		if record.Timestamp.Before(cutoff) {
			key := s.getObjectKey(record)
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: aws.String(key),
			})
			count++
		}
	}

	// If no objects to delete, return early
	if len(objectsToDelete) == 0 {
		return 0, nil
	}

	// Delete objects in batches (S3 allows up to 1000 objects per request)
	const batchSize = 1000
	for i := 0; i < len(objectsToDelete); i += batchSize {
		end := i + batchSize
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]
		_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return count - (len(objectsToDelete) - i), &S3StorageError{
				Message: "failed to delete objects from S3",
				Err:     err,
			}
		}
	}

	logger.Info("Deleted old optimization records", "count", count, "age", age.String())
	return count, nil
}

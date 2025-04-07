package storage

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theapemachine/lookatthatmongo/ai"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// mockS3Client is a mock implementation of the S3 client for testing
type mockS3Client struct {
	mock.Mock
}

// HeadBucket mocks the HeadBucket operation
func (m *mockS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.HeadBucketOutput), args.Error(1)
}

// PutObject mocks the PutObject operation
func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

// GetObject mocks the GetObject operation
func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

// ListObjectsV2 mocks the ListObjectsV2 operation
func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

// DeleteObjects mocks the DeleteObjects operation
func (m *mockS3Client) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.DeleteObjectsOutput), args.Error(1)
}

// newMockS3Output creates a mock output for GetObject with the provided JSON content
func newMockS3Output(content string) *s3.GetObjectOutput {
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(content)),
	}
}

// createTestRecord creates a test optimization record for use in tests
func createTestRecord(id string, dbName string) *OptimizationRecord {
	return &OptimizationRecord{
		ID:           id,
		Timestamp:    time.Now(),
		DatabaseName: dbName,
		BeforeReport: &metrics.Report{},
		AfterReport:  &metrics.Report{},
		Suggestion: &ai.OptimizationSuggestion{
			Category: "indexes",
			Impact:   "high",
		},
		Applied:          true,
		Success:          true,
		ImprovementPct:   25.0,
		RollbackRequired: false,
		RollbackSuccess:  false,
	}
}

// mockS3Option returns an S3StorageOption that sets a mock client
func mockS3Option(mockClient *mockS3Client) S3StorageOption {
	return func(s *S3Storage) {
		s.client = mockClient
	}
}

func TestNewS3Storage(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockS3Client)

	// Test case: successful initialization
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	}).Return(&s3.HeadBucketOutput{}, nil)

	storage, err := NewS3Storage(ctx, WithBucket("test-bucket"), mockS3Option(mockClient))
	assert.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "test-bucket", storage.bucket)
	assert.Equal(t, DefaultRecordsPrefix, storage.prefix)

	// Test case: missing bucket name
	storage, err = NewS3Storage(ctx, mockS3Option(mockClient))
	assert.Error(t, err)
	assert.Nil(t, storage)
	assert.Contains(t, err.Error(), "bucket name is required")

	// Test case: bucket does not exist
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("non-existent-bucket"),
	}).Return(nil, &types.NoSuchBucket{})

	storage, err = NewS3Storage(ctx, WithBucket("non-existent-bucket"), mockS3Option(mockClient))
	assert.Error(t, err)
	assert.Nil(t, storage)
	assert.Contains(t, err.Error(), "failed to access S3 bucket")

	mockClient.AssertExpectations(t)
}

func TestS3Storage_SaveOptimizationRecord(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockS3Client)

	// Mock HeadBucket for initialization
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	}).Return(&s3.HeadBucketOutput{}, nil)

	storage, err := NewS3Storage(ctx, WithBucket("test-bucket"), mockS3Option(mockClient))
	assert.NoError(t, err)

	// Test case: successful save
	record := createTestRecord("test-id", "test-db")
	expectedKey := "optimization-records/test-db/test-id.json"

	mockClient.On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		return *input.Bucket == "test-bucket" && *input.Key == expectedKey
	})).Return(&s3.PutObjectOutput{}, nil)

	err = storage.SaveOptimizationRecord(ctx, record)
	assert.NoError(t, err)

	// Test case: nil record
	err = storage.SaveOptimizationRecord(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record cannot be nil")

	// Test case: S3 error
	record2 := createTestRecord("error-id", "test-db")
	expectedKey2 := "optimization-records/test-db/error-id.json"

	mockClient.On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		return *input.Bucket == "test-bucket" && *input.Key == expectedKey2
	})).Return(nil, assert.AnError)

	err = storage.SaveOptimizationRecord(ctx, record2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload record to S3")

	mockClient.AssertExpectations(t)
}

func TestS3Storage_GetOptimizationRecord(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockS3Client)

	// Mock HeadBucket for initialization
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	}).Return(&s3.HeadBucketOutput{}, nil)

	storage, err := NewS3Storage(ctx, WithBucket("test-bucket"), mockS3Option(mockClient))
	assert.NoError(t, err)

	// Test case: empty ID
	_, err = storage.GetOptimizationRecord(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record ID cannot be empty")

	// Test case: record found with database name
	expectedKey := "optimization-records/test-db/test-id.json"

	// First mock the list operation to find the key with a database name
	mockClient.On("ListObjectsV2", ctx, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return *input.Bucket == "test-bucket" &&
			*input.Prefix == "optimization-records/test-db"
	})).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String(expectedKey)},
		},
	}, nil)

	// Then mock the get operation to retrieve the content
	jsonData := `{"id":"test-id","timestamp":"2023-01-01T12:00:00Z","database_name":"test-db","before_report":{},"after_report":{},"suggestion":{"category":"indexes","impact":"high"},"applied":true,"success":true,"improvement_pct":25,"rollback_required":false,"rollback_success":false}`

	mockClient.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String(expectedKey),
	}).Return(newMockS3Output(jsonData), nil)

	retrievedRecord, err := storage.GetOptimizationRecord(ctx, "test-id", "test-db")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedRecord)
	assert.Equal(t, "test-id", retrievedRecord.ID)

	// Test case: record not found when no database name is specified
	mockClient.On("ListObjectsV2", ctx, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return *input.Bucket == "test-bucket" &&
			*input.Prefix == "optimization-records"
	})).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("optimization-records/test-db/other-id.json")},
		},
	}, nil)

	_, err = storage.GetOptimizationRecord(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record with ID non-existent-id not found")

	mockClient.AssertExpectations(t)
}

func TestS3Storage_ListOptimizationRecords(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockS3Client)

	// Mock HeadBucket for initialization
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	}).Return(&s3.HeadBucketOutput{}, nil)

	storage, err := NewS3Storage(ctx, WithBucket("test-bucket"), mockS3Option(mockClient))
	assert.NoError(t, err)

	// Mock the paginator with a single page
	mockClient.On("ListObjectsV2", ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("test-bucket"),
		Prefix: aws.String("optimization-records"),
	}).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("optimization-records/test-db/id1.json")},
			{Key: aws.String("optimization-records/test-db/id2.json")},
		},
		IsTruncated: aws.Bool(false),
	}, nil)

	// Mock GetObject for each key
	jsonData1 := `{"id":"id1","timestamp":"2023-01-02T12:00:00Z","database_name":"test-db","suggestion":{"category":"indexes"}}`
	jsonData2 := `{"id":"id2","timestamp":"2023-01-01T12:00:00Z","database_name":"test-db","suggestion":{"category":"indexes"}}`

	mockClient.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("optimization-records/test-db/id1.json"),
	}).Return(newMockS3Output(jsonData1), nil)

	mockClient.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("optimization-records/test-db/id2.json"),
	}).Return(newMockS3Output(jsonData2), nil)

	records, err := storage.ListOptimizationRecords(ctx)
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	// Should be sorted by timestamp (newest first)
	assert.Equal(t, "id1", records[0].ID)
	assert.Equal(t, "id2", records[1].ID)

	mockClient.AssertExpectations(t)
}

func TestS3Storage_DeleteOldRecords(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockS3Client)

	// Mock HeadBucket for initialization
	mockClient.On("HeadBucket", ctx, &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	}).Return(&s3.HeadBucketOutput{}, nil)

	storage, err := NewS3Storage(ctx, WithBucket("test-bucket"), mockS3Option(mockClient))
	assert.NoError(t, err)

	// Set up records with different timestamps
	// First, mock the ListObjectsV2 call
	mockClient.On("ListObjectsV2", ctx, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return *input.Bucket == "test-bucket" &&
			*input.Prefix == "optimization-records"
	})).Return(&s3.ListObjectsV2Output{
		Contents: []types.Object{
			{Key: aws.String("optimization-records/test-db/old.json")},
			{Key: aws.String("optimization-records/test-db/new.json")},
		},
	}, nil)

	// Now mock the GetObject calls
	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now()

	// Convert to JSON with actual times
	oldJSON := `{"id":"old","timestamp":"` + oldTime.Format(time.RFC3339) + `","database_name":"test-db"}`
	newJSON := `{"id":"new","timestamp":"` + newTime.Format(time.RFC3339) + `","database_name":"test-db"}`

	mockClient.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("optimization-records/test-db/old.json"),
	}).Return(newMockS3Output(oldJSON), nil)

	mockClient.On("GetObject", ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("optimization-records/test-db/new.json"),
	}).Return(newMockS3Output(newJSON), nil)

	// Mock the DeleteObjects call for the old record
	mockClient.On("DeleteObjects", ctx, mock.MatchedBy(func(input *s3.DeleteObjectsInput) bool {
		if len(input.Delete.Objects) != 1 {
			return false
		}
		return *input.Bucket == "test-bucket" &&
			*input.Delete.Objects[0].Key == "optimization-records/test-db/old.json"
	})).Return(&s3.DeleteObjectsOutput{}, nil)

	// Test deletion of records older than 24 hours
	count, err := storage.DeleteOldRecords(ctx, 24*time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	mockClient.AssertExpectations(t)
}

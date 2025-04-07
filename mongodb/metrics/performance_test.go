package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCalculateRate(t *testing.T) {
	tests := []struct {
		name          string
		metric        string
		currentCount  int64
		prevCount     int64
		timeDiff      time.Duration
		expectedRate  float64
		setupPrevious bool
	}{
		{
			name:          "normal rate calculation",
			metric:        "test",
			currentCount:  100,
			prevCount:     50,
			timeDiff:      5 * time.Second,
			expectedRate:  10.0, // (100-50)/5 = 10 ops/sec
			setupPrevious: true,
		},
		{
			name:          "counter reset handling",
			metric:        "test",
			currentCount:  50,
			prevCount:     100,
			timeDiff:      5 * time.Second,
			expectedRate:  10.0, // 50/5 = 10 ops/sec (using current count)
			setupPrevious: true,
		},
		{
			name:          "first collection",
			metric:        "new_metric",
			currentCount:  100,
			timeDiff:      5 * time.Second,
			expectedRate:  0.0, // No previous data
			setupPrevious: false,
		},
		{
			name:          "zero duration",
			metric:        "test",
			currentCount:  100,
			prevCount:     50,
			timeDiff:      0,
			expectedRate:  0.0, // Avoid division by zero
			setupPrevious: true,
		},
		{
			name:          "negative duration",
			metric:        "test",
			currentCount:  100,
			prevCount:     50,
			timeDiff:      -1 * time.Second,
			expectedRate:  0.0, // Invalid time difference
			setupPrevious: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &PerformanceMonitor{
				prev: &previousCounts{
					counts: make(map[string]int64),
				},
			}
			now := time.Now()

			if tt.setupPrevious {
				monitor.prev.Lock()
				monitor.prev.counts[tt.metric] = tt.prevCount
				monitor.prev.timestamp = now.Add(-tt.timeDiff)
				monitor.prev.Unlock()
			}

			rate := monitor.calculateRate(tt.metric, tt.currentCount, now)
			assert.InDelta(t, tt.expectedRate, rate, 0.001, "rate calculation mismatch")
		})
	}
}

func TestRateCalculationHelpers(t *testing.T) {
	monitor := &PerformanceMonitor{
		prev: &previousCounts{
			counts: make(map[string]int64),
		},
	}

	// Test initial state
	now := time.Now()
	rate := monitor.calculateRate("test", 100, now)
	assert.Equal(t, float64(0), rate, "initial rate should be 0")

	// Test after storing initial value
	monitor.prev.Lock()
	monitor.prev.counts["test"] = 100
	monitor.prev.timestamp = now
	monitor.prev.Unlock()

	// Wait a bit and test rate calculation
	time.Sleep(time.Second)
	now = time.Now()
	rate = monitor.calculateRate("test", 150, now) // 50 operations in ~1 second
	assert.InDelta(t, float64(50), rate, 5.0, "rate should be approximately 50 ops/sec")

	// Test counter reset handling
	rate = monitor.calculateRate("test", 25, now) // Counter reset to lower value
	assert.InDelta(t, float64(25), rate, 5.0, "should use current count when counter resets")
}

func TestFormatQueryPattern(t *testing.T) {
	monitor := &PerformanceMonitor{
		prev: &previousCounts{
			counts: make(map[string]int64),
		},
	}

	objectID := primitive.NewObjectID()
	now := time.Now()

	tests := []struct {
		name     string
		query    any
		expected string
	}{
		{
			name:     "nil query",
			query:    nil,
			expected: "",
		},
		{
			name: "simple equality query",
			query: bson.D{
				{Key: "name", Value: "test"},
				{Key: "age", Value: 25},
			},
			expected: `{"name":"<?>","age":"<?>"}`,
		},
		{
			name: "nested document",
			query: bson.D{
				{Key: "user", Value: bson.D{
					{Key: "name", Value: "test"},
					{Key: "email", Value: "test@example.com"},
				}},
			},
			expected: `{"user":{"name":"<?>","email":"<?"}}`,
		},
		{
			name: "array values",
			query: bson.D{
				{Key: "tags", Value: []any{"tag1", "tag2", "tag3"}},
			},
			expected: `{"tags":[<?>,<?>,<?>]}`,
		},
		{
			name: "complex operators",
			query: bson.D{
				{Key: "age", Value: bson.D{
					{Key: "$gt", Value: 25},
					{Key: "$lt", Value: 50},
				}},
				{Key: "status", Value: bson.D{
					{Key: "$in", Value: []any{"active", "pending"}},
				}},
			},
			expected: `{"age":{"$gt":"<?>","$lt":"<?>"},"status":{"$in":[<?>,<?>]}}`,
		},
		{
			name: "special operator values preserved",
			query: bson.D{
				{Key: "field", Value: bson.D{
					{Key: "$exists", Value: true},
					{Key: "$type", Value: "string"},
				}},
			},
			expected: `{"field":{"$exists":true,"$type":string}}`,
		},
		{
			name: "ObjectID handling",
			query: bson.D{
				{Key: "_id", Value: objectID},
			},
			expected: `{"_id":"<ObjectId>"}`,
		},
		{
			name: "regex pattern",
			query: bson.D{
				{Key: "name", Value: primitive.Regex{Pattern: "^test", Options: "i"}},
			},
			expected: `{"name":"<regex:\"^test\">"}`,
		},
		{
			name: "timestamp handling",
			query: bson.D{
				{Key: "created", Value: now},
			},
			expected: `{"created":"<?>"}`,
		},
		{
			name: "mixed types",
			query: bson.D{
				{Key: "name", Value: "test"},
				{Key: "age", Value: 25},
				{Key: "active", Value: true},
				{Key: "score", Value: 98.6},
			},
			expected: `{"name":"<?>","age":"<?>","active":"<?>","score":"<?>"}`,
		},
		{
			name: "nested arrays and documents",
			query: bson.D{
				{Key: "items", Value: []any{
					bson.D{{Key: "id", Value: 1}, {Key: "qty", Value: 5}},
					bson.D{{Key: "id", Value: 2}, {Key: "qty", Value: 10}},
				}},
			},
			expected: `{"items":[{"id":"<?>","qty":"<?>"},{"id":"<?>","qty":"<?>"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.formatQueryPattern(tt.query)
			assert.Equal(t, tt.expected, result, "query pattern normalization mismatch")
		})
	}
}

func TestNormalizeEdgeCases(t *testing.T) {
	monitor := &PerformanceMonitor{
		prev: &previousCounts{
			counts: make(map[string]int64),
		},
	}

	tests := []struct {
		name     string
		query    any
		expected string
	}{
		{
			name:     "empty document",
			query:    bson.D{},
			expected: "{}",
		},
		{
			name:     "empty array",
			query:    []any{},
			expected: "[]",
		},
		{
			name:     "empty map",
			query:    bson.M{},
			expected: "{}",
		},
		{
			name:     "unknown type",
			query:    struct{ Name string }{"test"},
			expected: `<?:struct { Name string }>`,
		},
		{
			name: "deeply nested structure",
			query: bson.D{
				{Key: "level1", Value: bson.D{
					{Key: "level2", Value: bson.D{
						{Key: "level3", Value: bson.D{
							{Key: "value", Value: "deep"},
						}},
					}},
				}},
			},
			expected: `{"level1":{"level2":{"level3":{"value":"<?"}}}}`,
		},
		{
			name: "mixed nested arrays",
			query: bson.D{
				{Key: "mixed", Value: []any{
					1,
					[]any{"nested", "array"},
					bson.D{{Key: "doc", Value: "value"}},
				}},
			},
			expected: `{"mixed":[<?>,["<?>","<?>"],{"doc":"<?>"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.formatQueryPattern(tt.query)
			assert.Equal(t, tt.expected, result, "edge case normalization mismatch")
		})
	}
}

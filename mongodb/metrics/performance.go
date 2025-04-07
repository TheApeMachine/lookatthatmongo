package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OptionFn func(*PerformanceMonitor)

// previousCounts stores the previous counter values and their timestamps
type previousCounts struct {
	sync.RWMutex
	counts    map[string]int64
	timestamp time.Time
}

// PerformanceMonitor handles collection of performance-related metrics
type PerformanceMonitor struct {
	client *mongo.Client
	opts   []OptionFn
	prev   *previousCounts
}

// NewPerformanceMonitor creates a new performance monitor with the given options
func NewPerformanceMonitor(client *mongo.Client, opts ...OptionFn) *PerformanceMonitor {
	return &PerformanceMonitor{
		client: client,
		opts:   opts,
		prev: &previousCounts{
			counts:    make(map[string]int64),
			timestamp: time.Now(),
		},
	}
}

// CollectStats gathers all performance-related statistics
func (pm *PerformanceMonitor) CollectStats(
	ctx context.Context,
) (*PerformanceStats, error) {
	stats := &PerformanceStats{}

	for _, fn := range []func(ctx context.Context, stats *PerformanceStats) error{
		pm.collectLatencyStats,
		pm.collectThroughputStats,
		pm.collectResourceStats,
		pm.collectSlowOperations,
		pm.collectIndexUtilization,
	} {
		if err := fn(ctx, stats); err != nil {
			return nil, fmt.Errorf("failed to collect stats: %w", err)
		}
	}

	return stats, nil
}

func (pm *PerformanceMonitor) collectLatencyStats(ctx context.Context, stats *PerformanceStats) error {
	var result bson.M
	if err := pm.runCommand(ctx, "admin", bson.D{
		{Key: "serverStatus", Value: 1},
		{Key: "metrics", Value: 1},
	}, &result); err != nil {
		return err
	}

	metrics, ok := result["metrics"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid metrics format in server status")
	}

	opLatencies, ok := metrics["operationLatencies"].(bson.M)
	if !ok {
		return nil // Not an error, just no latency metrics available
	}

	stats.Latency = LatencyStats{
		ReadLatencyMicros:    pm.parseLatency(opLatencies, "reads"),
		WriteLatencyMicros:   pm.parseLatency(opLatencies, "writes"),
		CommandLatencyMicros: pm.parseLatency(opLatencies, "commands"),
	}

	return nil
}

func (pm *PerformanceMonitor) collectThroughputStats(ctx context.Context, stats *PerformanceStats) error {
	var result bson.M
	if err := pm.runCommand(ctx, "admin", bson.D{
		{Key: "serverStatus", Value: 1},
		{Key: "opcounters", Value: 1},
		{Key: "network", Value: 1},
	}, &result); err != nil {
		return err
	}

	network, ok := result["network"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid network metrics format")
	}

	opcounters, ok := result["opcounters"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid opcounters format")
	}

	now := time.Now()

	// Calculate rates for each operation type
	reads := getMetric[int64](pm, opcounters, "query")
	writes := getMetric[int64](pm, opcounters, "insert") +
		getMetric[int64](pm, opcounters, "update") +
		getMetric[int64](pm, opcounters, "delete")
	commands := getMetric[int64](pm, opcounters, "command")

	// Update stats with calculated rates
	stats.Throughput = ThroughputStats{
		ReadsPerSecond:    pm.calculateRate("reads", reads, now),
		WritesPerSecond:   pm.calculateRate("writes", writes, now),
		CommandsPerSecond: pm.calculateRate("commands", commands, now),
		NetworkInBytes:    getMetric[int64](pm, network, "bytesIn"),
		NetworkOutBytes:   getMetric[int64](pm, network, "bytesOut"),
	}

	// Store current values for next calculation
	pm.prev.Lock()
	pm.prev.counts["reads"] = reads
	pm.prev.counts["writes"] = writes
	pm.prev.counts["commands"] = commands
	pm.prev.timestamp = now
	pm.prev.Unlock()

	return nil
}

func (pm *PerformanceMonitor) collectResourceStats(ctx context.Context, stats *PerformanceStats) error {
	var result bson.M
	if err := pm.runCommand(ctx, "admin", bson.D{
		{Key: "serverStatus", Value: 1},
		{Key: "system", Value: 1},
		{Key: "mem", Value: 1},
		{Key: "connections", Value: 1},
	}, &result); err != nil {
		return err
	}

	system, ok := result["system"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid system metrics format")
	}

	mem, ok := result["mem"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid memory metrics format")
	}

	conn, ok := result["connections"].(bson.M)
	if !ok {
		return fmt.Errorf("invalid connection metrics format")
	}

	stats.ResourceUsage = ResourceUsageStats{
		CPUUsagePercent:    getMetric[float64](pm, system, "cpu", "user"),
		MemoryUsagePercent: float64(getMetric[int64](pm, mem, "resident")),
		ConnectionPoolStats: PoolStats{
			InUse:     getMetric[int64](pm, conn, "current"),
			Available: getMetric[int64](pm, conn, "available"),
			Created:   getMetric[int64](pm, conn, "totalCreated"),
		},
	}

	return nil
}

func (pm *PerformanceMonitor) collectSlowOperations(ctx context.Context, stats *PerformanceStats) error {
	var result bson.M
	if err := pm.runCommand(ctx, "admin", bson.D{
		{Key: "currentOp", Value: true},
		{Key: "microsecs_running", Value: bson.D{{Key: "$gt", Value: 100}}},
	}, &result); err != nil {
		return err
	}

	inprog, ok := result["inprog"].([]any)
	if !ok {
		return nil // No slow operations
	}

	for _, op := range inprog {
		opMap, ok := op.(bson.M)
		if !ok {
			continue
		}

		stats.SlowOperations = append(stats.SlowOperations, SlowOperation{
			OpID:         fmt.Sprintf("%v", opMap["opid"]),
			Type:         fmt.Sprintf("%v", opMap["op"]),
			Namespace:    fmt.Sprintf("%v", opMap["ns"]),
			Duration:     time.Duration(getMetric[int64](pm, opMap, "microsecs_running")) * time.Microsecond,
			QueryPattern: pm.formatQueryPattern(opMap["query"]),
			Plan:         getMetric[string](pm, opMap, "planSummary"),
			Timestamp:    time.Now(),
		})
	}

	return nil
}

func (pm *PerformanceMonitor) collectIndexUtilization(ctx context.Context, stats *PerformanceStats) error {
	dbs, err := pm.listDatabases(ctx)
	if err != nil {
		return err
	}

	for _, dbName := range dbs {
		if err := pm.collectDatabaseIndexStats(ctx, dbName, stats); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods

func (pm *PerformanceMonitor) runCommand(ctx context.Context, db string, cmd any, result any) error {
	return pm.client.Database(db).RunCommand(ctx, cmd).Decode(result)
}

func (pm *PerformanceMonitor) parseLatency(metrics bson.M, opType string) OperationLatency {
	latency, ok := metrics[opType].(bson.M)
	if !ok {
		return OperationLatency{}
	}

	return OperationLatency{
		P50:  getMetric[float64](pm, latency, "latency", "50"),
		P95:  getMetric[float64](pm, latency, "latency", "95"),
		P99:  getMetric[float64](pm, latency, "latency", "99"),
		Max:  getMetric[float64](pm, latency, "latency", "max"),
		Mean: getMetric[float64](pm, latency, "latency", "mean"),
	}
}

func (pm *PerformanceMonitor) calculateRate(metric string, currentCount int64, now time.Time) float64 {
	pm.prev.RLock()
	prevCount, exists := pm.prev.counts[metric]
	prevTime := pm.prev.timestamp
	pm.prev.RUnlock()

	if !exists {
		return 0 // First collection, can't calculate rate yet
	}

	// Calculate time difference in seconds
	duration := now.Sub(prevTime).Seconds()
	if duration <= 0 {
		return 0 // Avoid division by zero or negative rates
	}

	// Calculate rate (operations per second)
	countDiff := currentCount - prevCount
	if countDiff < 0 {
		// Counter reset (e.g., server restart), use current value
		countDiff = currentCount
	}

	return float64(countDiff) / duration
}

func getMetric[T any](perfMon *PerformanceMonitor, m bson.M, keys ...string) T {
	current := m
	for _, key := range keys[:len(keys)-1] {
		next, ok := current[key].(bson.M)
		if !ok {
			return *new(T)
		}
		current = next
	}

	if num, ok := current[keys[len(keys)-1]].(T); ok {
		return num
	}

	return *new(T)
}

// formatQueryPattern normalizes a MongoDB query pattern by replacing specific values
// with placeholders while preserving the query structure. This helps identify similar
// queries that differ only in their parameter values.
func (pm *PerformanceMonitor) formatQueryPattern(query any) string {
	if query == nil {
		return ""
	}

	switch q := query.(type) {
	case bson.D:
		return pm.normalizeDocument(q)
	case bson.M:
		// Convert bson.M to bson.D for consistent ordering
		d := make(bson.D, 0, len(q))
		for k, v := range q {
			d = append(d, bson.E{Key: k, Value: v})
		}
		return pm.normalizeDocument(d)
	case []any:
		return pm.normalizeArray(q)
	case string, float64, int64, int32, bool:
		return "<?>"
	default:
		return fmt.Sprintf("<?:%T>", q)
	}
}

// normalizeDocument normalizes a BSON document by replacing values with placeholders
func (pm *PerformanceMonitor) normalizeDocument(doc bson.D) string {
	if len(doc) == 0 {
		return "{}"
	}

	var parts []string
	for _, elem := range doc {
		normalized := pm.normalizeValue(elem.Value)

		// Special handling for operators
		if len(elem.Key) > 0 && elem.Key[0] == '$' {
			// Keep actual values for certain operators where the value type matters
			switch elem.Key {
			case "$type", "$size", "$exists", "$mod", "$options":
				normalized = fmt.Sprintf("%v", elem.Value)
			}
		}

		parts = append(parts, fmt.Sprintf("%q:%s", elem.Key, normalized))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// normalizeArray normalizes an array of values
func (pm *PerformanceMonitor) normalizeArray(arr []any) string {
	if len(arr) == 0 {
		return "[]"
	}

	var parts []string
	for _, v := range arr {
		parts = append(parts, pm.normalizeValue(v))
	}

	return "[" + strings.Join(parts, ",") + "]"
}

// normalizeValue normalizes a single value based on its type
func (pm *PerformanceMonitor) normalizeValue(val any) string {
	switch v := val.(type) {
	case nil:
		return "null"
	case bson.D:
		return pm.normalizeDocument(v)
	case bson.M:
		d := make(bson.D, 0, len(v))
		for k, mv := range v {
			d = append(d, bson.E{Key: k, Value: mv})
		}
		return pm.normalizeDocument(d)
	case []any:
		return pm.normalizeArray(v)
	case map[string]any:
		d := make(bson.D, 0, len(v))
		for k, mv := range v {
			d = append(d, bson.E{Key: k, Value: mv})
		}
		return pm.normalizeDocument(d)
	case string, float64, int64, int32, bool, time.Time:
		return "<?>"
	case primitive.ObjectID:
		return "<ObjectId>"
	case primitive.Regex:
		return fmt.Sprintf("<regex:%q>", v.Pattern)
	default:
		return fmt.Sprintf("<?:%T>", v)
	}
}

func (pm *PerformanceMonitor) listDatabases(ctx context.Context) ([]string, error) {
	dbs, err := pm.client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	// Filter out system databases
	var filtered []string
	for _, db := range dbs {
		if db != "admin" && db != "local" && db != "config" {
			filtered = append(filtered, db)
		}
	}
	return filtered, nil
}

func (pm *PerformanceMonitor) collectDatabaseIndexStats(ctx context.Context, dbName string, stats *PerformanceStats) error {
	collections, err := pm.client.Database(dbName).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, collName := range collections {
		if err := pm.collectCollectionIndexStats(ctx, dbName, collName, stats); err != nil {
			return err
		}
	}

	return nil
}

func (pm *PerformanceMonitor) collectCollectionIndexStats(ctx context.Context, dbName, collName string, stats *PerformanceStats) error {
	var result bson.M
	if err := pm.runCommand(ctx, dbName, bson.D{
		{Key: "aggregate", Value: collName},
		{Key: "pipeline", Value: bson.A{
			bson.D{{Key: "$indexStats", Value: bson.D{}}},
		}},
		{Key: "cursor", Value: bson.D{}},
	}, &result); err != nil {
		return err
	}

	cursor, ok := result["cursor"].(bson.M)
	if !ok {
		return nil
	}

	batch, ok := cursor["firstBatch"].([]any)
	if !ok {
		return nil
	}

	for _, stat := range batch {
		indexStat, ok := stat.(bson.M)
		if !ok {
			continue
		}

		stats.IndexUtilization = append(stats.IndexUtilization, IndexUtilizationStat{
			DatabaseName:   dbName,
			CollectionName: collName,
			IndexName:      getMetric[string](pm, indexStat, "name"),
			UsageCount:     getMetric[int64](pm, indexStat, "accesses", "ops"),
			LastUsed:       time.Now(),
		})
	}

	return nil
}

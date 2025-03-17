package metrics

import "time"

// ServerStats represents server-level statistics
type ServerStats struct {
	Host              string          `json:"host" bson:"host"`
	Version           string          `json:"version" bson:"version"`
	Uptime            float64         `json:"uptime" bson:"uptime"`
	LocalTime         time.Time       `json:"localTime" bson:"localTime"`
	Connections       ConnectionStats `json:"connections" bson:"connections"`
	Memory            MemoryStats     `json:"memory" bson:"memory"`
	OperationCounts   OpCountStats    `json:"opcounters" bson:"opcounters"`
	ReplicationStatus RepStats        `json:"replicationStatus" bson:"replicationStatus"`
}

// ConnectionStats tracks connection metrics
type ConnectionStats struct {
	Current      int64 `json:"current" bson:"current"`
	Available    int64 `json:"available" bson:"available"`
	TotalCreated int64 `json:"totalCreated" bson:"totalCreated"`
}

// MemoryStats tracks memory usage
type MemoryStats struct {
	Resident   int64 `json:"resident" bson:"resident"`     // Resident memory in MB
	Virtual    int64 `json:"virtual" bson:"virtual"`       // Virtual memory in MB
	PageFaults int64 `json:"pageFaults" bson:"pageFaults"` // Number of page faults
}

// OpCountStats tracks operation counts
type OpCountStats struct {
	Insert  int64 `json:"insert" bson:"insert"`
	Query   int64 `json:"query" bson:"query"`
	Update  int64 `json:"update" bson:"update"`
	Delete  int64 `json:"delete" bson:"delete"`
	GetMore int64 `json:"getmore" bson:"getmore"`
	Command int64 `json:"command" bson:"command"`
}

// RepStats tracks replication status
type RepStats struct {
	IsReplicaSet bool            `json:"isReplicaSet" bson:"isReplicaSet"`
	IsMaster     bool            `json:"isMaster" bson:"isMaster"`
	SetName      string          `json:"setName,omitempty" bson:"setName,omitempty"`
	Members      []ReplicaMember `json:"members,omitempty" bson:"members,omitempty"`
}

// ReplicaMember represents a member in a replica set
type ReplicaMember struct {
	Name          string    `json:"name" bson:"name"`
	Health        bool      `json:"health" bson:"health"`
	State         int       `json:"state" bson:"state"`
	StateStr      string    `json:"stateStr" bson:"stateStr"`
	LastHeartbeat time.Time `json:"lastHeartbeat,omitempty" bson:"lastHeartbeat,omitempty"`
}

// DatabaseStats represents database-level statistics
type DatabaseStats struct {
	Name        string  `json:"name" bson:"name"`
	Collections int     `json:"collections" bson:"collections"`
	Objects     int64   `json:"objects" bson:"objects"`
	DataSize    float64 `json:"dataSize" bson:"dataSize"`       // in bytes
	StorageSize float64 `json:"storageSize" bson:"storageSize"` // in bytes
	IndexSize   float64 `json:"indexSize" bson:"indexSize"`     // in bytes
	IndexCount  int     `json:"indexCount" bson:"indexCount"`
}

// CollectionStats represents collection-level statistics
type CollectionStats struct {
	Name         string             `json:"name" bson:"name"`
	Size         float64            `json:"size" bson:"size"` // in bytes
	Count        int64              `json:"count" bson:"count"`
	AvgObjSize   float64            `json:"avgObjSize" bson:"avgObjSize"`   // in bytes
	StorageSize  float64            `json:"storageSize" bson:"storageSize"` // in bytes
	Capped       bool               `json:"capped" bson:"capped"`
	MaxSize      float64            `json:"maxSize,omitempty" bson:"maxSize,omitempty"`
	IndexSizes   map[string]float64 `json:"indexSizes" bson:"indexSizes"`
	IndexDetails []IndexStats       `json:"indexDetails" bson:"indexDetails"`
}

// IndexStats represents index statistics
type IndexStats struct {
	Name       string  `json:"name" bson:"name"`
	Size       float64 `json:"size" bson:"size"` // in bytes
	KeyPattern string  `json:"keyPattern" bson:"keyPattern"`
	Unique     bool    `json:"unique" bson:"unique"`
	Sparse     bool    `json:"sparse" bson:"sparse"`
	UseCount   int64   `json:"useCount" bson:"useCount"`
}

// PerformanceStats captures detailed performance metrics
type PerformanceStats struct {
	Latency          LatencyStats           `json:"latency" bson:"latency"`
	Throughput       ThroughputStats        `json:"throughput" bson:"throughput"`
	ResourceUsage    ResourceUsageStats     `json:"resourceUsage" bson:"resourceUsage"`
	SlowOperations   []SlowOperation        `json:"slowOperations" bson:"slowOperations"`
	IndexUtilization []IndexUtilizationStat `json:"indexUtilization" bson:"indexUtilization"`
}

// LatencyStats tracks operation latency metrics
type LatencyStats struct {
	ReadLatencyMicros    OperationLatency `json:"readLatency" bson:"readLatency"`
	WriteLatencyMicros   OperationLatency `json:"writeLatency" bson:"writeLatency"`
	CommandLatencyMicros OperationLatency `json:"commandLatency" bson:"commandLatency"`
}

// OperationLatency contains percentile latency measurements
type OperationLatency struct {
	P50  float64 `json:"p50" bson:"p50"` // 50th percentile (median)
	P95  float64 `json:"p95" bson:"p95"` // 95th percentile
	P99  float64 `json:"p99" bson:"p99"` // 99th percentile
	Max  float64 `json:"max" bson:"max"`
	Mean float64 `json:"mean" bson:"mean"`
}

// ThroughputStats tracks operation throughput
type ThroughputStats struct {
	ReadsPerSecond    float64 `json:"readsPerSecond" bson:"readsPerSecond"`
	WritesPerSecond   float64 `json:"writesPerSecond" bson:"writesPerSecond"`
	CommandsPerSecond float64 `json:"commandsPerSecond" bson:"commandsPerSecond"`
	NetworkInBytes    int64   `json:"networkInBytes" bson:"networkInBytes"`
	NetworkOutBytes   int64   `json:"networkOutBytes" bson:"networkOutBytes"`
}

// ResourceUsageStats tracks system resource utilization
type ResourceUsageStats struct {
	CPUUsagePercent     float64   `json:"cpuUsagePercent" bson:"cpuUsagePercent"`
	MemoryUsagePercent  float64   `json:"memoryUsagePercent" bson:"memoryUsagePercent"`
	DiskIOPSRead        int64     `json:"diskIopsRead" bson:"diskIopsRead"`
	DiskIOPSWrite       int64     `json:"diskIopsWrite" bson:"diskIopsWrite"`
	DiskLatencyRead     float64   `json:"diskLatencyRead" bson:"diskLatencyRead"`   // in milliseconds
	DiskLatencyWrite    float64   `json:"diskLatencyWrite" bson:"diskLatencyWrite"` // in milliseconds
	ConnectionPoolStats PoolStats `json:"connectionPoolStats" bson:"connectionPoolStats"`
}

// PoolStats tracks connection pool metrics
type PoolStats struct {
	InUse     int64 `json:"inUse" bson:"inUse"`
	Available int64 `json:"available" bson:"available"`
	Created   int64 `json:"created" bson:"created"`
	Returned  int64 `json:"returned" bson:"returned"`
	Cleared   int64 `json:"cleared" bson:"cleared"`
	TimedOut  int64 `json:"timedOut" bson:"timedOut"`
}

// SlowOperation represents a slow query or operation
type SlowOperation struct {
	OpID         string        `json:"opId" bson:"opId"`
	Type         string        `json:"type" bson:"type"` // query, update, command, etc.
	Namespace    string        `json:"namespace" bson:"namespace"`
	Duration     time.Duration `json:"duration" bson:"duration"`
	QueryPattern string        `json:"queryPattern" bson:"queryPattern"` // Normalized query pattern
	Plan         string        `json:"plan" bson:"plan"`                 // Query plan summary
	Timestamp    time.Time     `json:"timestamp" bson:"timestamp"`
}

// IndexUtilizationStat tracks index usage statistics
type IndexUtilizationStat struct {
	DatabaseName   string    `json:"databaseName" bson:"databaseName"`
	CollectionName string    `json:"collectionName" bson:"collectionName"`
	IndexName      string    `json:"indexName" bson:"indexName"`
	UsageCount     int64     `json:"usageCount" bson:"usageCount"`
	SizeBytes      int64     `json:"sizeBytes" bson:"sizeBytes"`
	LastUsed       time.Time `json:"lastUsed" bson:"lastUsed"`
	IsSparse       bool      `json:"isSparse" bson:"isSparse"`
	IsUnique       bool      `json:"isUnique" bson:"isUnique"`
	IsMultiKey     bool      `json:"isMultiKey" bson:"isMultiKey"`
}

// QueryPatternStats tracks query pattern performance
type QueryPatternStats struct {
	Pattern             string        `json:"pattern" bson:"pattern"`
	ExecutionCount      int64         `json:"executionCount" bson:"executionCount"`
	AverageLatency      time.Duration `json:"averageLatency" bson:"averageLatency"`
	IndexesUsed         []string      `json:"indexesUsed" bson:"indexesUsed"`
	CollectionScans     int64         `json:"collectionScans" bson:"collectionScans"`
	InMemorySort        int64         `json:"inMemorySort" bson:"inMemorySort"`
	AverageDocsScanned  int64         `json:"averageDocsScanned" bson:"averageDocsScanned"`
	AverageKeysExamined int64         `json:"averageKeysExamined" bson:"averageKeysExamined"`
	AverageDocsReturned int64         `json:"averageDocsReturned" bson:"averageDocsReturned"`
	LastExecuted        time.Time     `json:"lastExecuted" bson:"lastExecuted"`
}

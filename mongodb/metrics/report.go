package metrics

import (
	"encoding/json"
	"time"
)

// Report represents a comprehensive performance report
type Report struct {
	Timestamp     time.Time                     `json:"timestamp"`
	ServerStats   *ServerStats                  `json:"serverStats"`
	DatabaseStats map[string]*DatabaseStats     `json:"databaseStats"`
	Collections   map[string][]*CollectionStats `json:"collections"`
	Indexes       map[string][]*IndexStats      `json:"indexes"`
	monitor       Monitor
}

// Monitor interface defines methods for collecting MongoDB metrics
type Monitor interface {
	GetServerStats(ctx any) (*ServerStats, error)
	GetDatabaseStats(ctx any, dbName string) (*DatabaseStats, error)
	GetCollectionStats(ctx any, dbName, collName string) (*CollectionStats, error)
	GetIndexStats(ctx any, dbName, collName string) ([]IndexStats, error)
}

// NewReport creates a new report instance
func NewReport(monitor Monitor) *Report {
	return &Report{
		Timestamp:     time.Now(),
		DatabaseStats: make(map[string]*DatabaseStats),
		Collections:   make(map[string][]*CollectionStats),
		Indexes:       make(map[string][]*IndexStats),
		monitor:       monitor,
	}
}

// Collect gathers all metrics for the specified database
func (r *Report) Collect(ctx any, dbName string, listCollections func() ([]string, error)) error {
	var err error

	// Get server stats
	if r.ServerStats, err = r.monitor.GetServerStats(ctx); err != nil {
		return err
	}

	// Get database stats
	dbStats, err := r.monitor.GetDatabaseStats(ctx, dbName)
	if err != nil {
		return err
	}
	r.DatabaseStats[dbName] = dbStats

	// Get collections
	collections, err := listCollections()
	if err != nil {
		return err
	}

	// Collect stats for each collection
	for _, collName := range collections {
		if err := r.collectCollectionMetrics(ctx, dbName, collName); err != nil {
			return err
		}
	}

	return nil
}

// collectCollectionMetrics gathers metrics for a single collection
func (r *Report) collectCollectionMetrics(ctx any, dbName, collName string) error {
	// Get collection stats
	collStats, err := r.monitor.GetCollectionStats(ctx, dbName, collName)
	if err != nil {
		return err
	}
	r.Collections[collName] = []*CollectionStats{collStats}

	// Get index stats
	indexStats, err := r.monitor.GetIndexStats(ctx, dbName, collName)
	if err != nil {
		return err
	}

	// Convert to pointers
	indexPtrs := make([]*IndexStats, len(indexStats))
	for i := range indexStats {
		indexPtrs[i] = &indexStats[i]
	}
	r.Indexes[collName] = indexPtrs

	return nil
}

func (r *Report) String() string {
	json, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(json)
}

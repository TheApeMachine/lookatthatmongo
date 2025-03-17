package mongodb

import (
	"context"
	"fmt"

	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
	"go.mongodb.org/mongo-driver/bson"
)

/*
Monitor provides methods to collect various MongoDB performance metrics.
It interfaces with different MongoDB statistics endpoints to gather comprehensive
performance data about the server, databases, collections, and indexes.
*/
type Monitor struct {
	conn               *Conn
	performanceMonitor *metrics.PerformanceMonitor
}

/*
MonitorOptionFn is a function type for configuring a Monitor instance.
It follows the functional options pattern for flexible configuration.
*/
type MonitorOptionFn func(*Monitor)

/*
NewMonitor creates a new Monitor instance.
It accepts optional configuration functions to customize the monitor.
*/
func NewMonitor(opts ...MonitorOptionFn) *Monitor {
	monitor := &Monitor{}

	for _, opt := range opts {
		opt(monitor)
	}

	if monitor.conn != nil {
		monitor.performanceMonitor = metrics.NewPerformanceMonitor(monitor.conn.Client)
	}

	return monitor
}

/*
WithConn is an option function that sets the MongoDB connection for the monitor.
*/
func WithConn(conn *Conn) MonitorOptionFn {
	return func(m *Monitor) {
		m.conn = conn
	}
}

/*
GetServerStats retrieves server-wide statistics from MongoDB.
It implements the metrics.Monitor interface.
*/
func (monitor *Monitor) GetServerStats(ctx any) (*metrics.ServerStats, error) {
	cmd := bson.D{{Key: "serverStatus", Value: 1}}
	var result = &metrics.ServerStats{}

	if err := monitor.conn.Database("admin").RunCommand(ctx.(context.Context), cmd).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to get server stats: %w", err)
	}

	return result, nil
}

/*
GetDatabaseStats retrieves statistics for a specific database.
*/
func (monitor *Monitor) GetDatabaseStats(ctx any, dbName string) (*metrics.DatabaseStats, error) {
	var result = &metrics.DatabaseStats{}

	if err := monitor.conn.Database(dbName).RunCommand(
		ctx.(context.Context),
		bson.D{{Key: "dbStats", Value: 1}},
	).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return result, nil
}

/*
GetCollectionStats retrieves statistics for a specific collection.
*/
func (monitor *Monitor) GetCollectionStats(ctx any, dbName, collName string) (*metrics.CollectionStats, error) {
	var result = &metrics.CollectionStats{}

	if err := monitor.conn.Database(dbName).RunCommand(
		ctx.(context.Context),
		bson.D{{Key: "collStats", Value: collName}},
	).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	return result, nil
}

/*
GetIndexStats retrieves statistics for all indexes in a collection.
*/
func (monitor *Monitor) GetIndexStats(ctx any, dbName, collName string) ([]metrics.IndexStats, error) {
	cursor, err := monitor.conn.Database(dbName).Collection(collName).Indexes().List(ctx.(context.Context))
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}

	var indexes []metrics.IndexStats
	for cursor.Next(ctx.(context.Context)) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			return nil, fmt.Errorf("failed to decode index: %w", err)
		}

		indexStats := metrics.IndexStats{
			Name:       idx["name"].(string),
			KeyPattern: fmt.Sprintf("%v", idx["key"]),
		}

		if unique, ok := idx["unique"]; ok {
			indexStats.Unique = unique.(bool)
		}

		if sparse, ok := idx["sparse"]; ok {
			indexStats.Sparse = sparse.(bool)
		}

		indexes = append(indexes, indexStats)
	}

	return indexes, nil
}

/*
GetPerformanceStats retrieves performance-related statistics.
*/
func (monitor *Monitor) GetPerformanceStats(ctx context.Context) (*metrics.PerformanceStats, error) {
	if monitor.performanceMonitor == nil {
		return nil, fmt.Errorf("performance monitor not initialized")
	}
	return monitor.performanceMonitor.CollectStats(ctx)
}

/*
GetReplicationStatus retrieves replication status information.
*/
func (monitor *Monitor) GetReplicationStatus(ctx context.Context) (*metrics.RepStats, error) {
	var result = &metrics.RepStats{}

	if err := monitor.conn.Database("admin").RunCommand(
		ctx,
		bson.D{{Key: "replSetGetStatus", Value: 1}},
	).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to get replication status: %w", err)
	}

	return result, nil
}

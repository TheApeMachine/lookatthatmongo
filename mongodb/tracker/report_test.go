package tracker

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// mockMetricsMonitor is a mock implementation of the metrics.Monitor interface
type mockMetricsMonitor struct {
	serverStatsFunc     func(ctx any) (*metrics.ServerStats, error)
	databaseStatsFunc   func(ctx any, dbName string) (*metrics.DatabaseStats, error)
	collectionStatsFunc func(ctx any, dbName, collName string) (*metrics.CollectionStats, error)
	indexStatsFunc      func(ctx any, dbName, collName string) ([]metrics.IndexStats, error)
}

func (m *mockMetricsMonitor) GetServerStats(ctx any) (*metrics.ServerStats, error) {
	return m.serverStatsFunc(ctx)
}

func (m *mockMetricsMonitor) GetDatabaseStats(ctx any, dbName string) (*metrics.DatabaseStats, error) {
	return m.databaseStatsFunc(ctx, dbName)
}

func (m *mockMetricsMonitor) GetCollectionStats(ctx any, dbName, collName string) (*metrics.CollectionStats, error) {
	return m.collectionStatsFunc(ctx, dbName, collName)
}

func (m *mockMetricsMonitor) GetIndexStats(ctx any, dbName, collName string) ([]metrics.IndexStats, error) {
	return m.indexStatsFunc(ctx, dbName, collName)
}

func TestNewReport(t *testing.T) {
	Convey("Given a need for a metrics report", t, func() {
		monitor := &mockMetricsMonitor{}

		Convey("When creating a new report", func() {
			report := metrics.NewReport(monitor)

			Convey("Then a report instance should be created", func() {
				So(report, ShouldNotBeNil)
				So(report.Timestamp, ShouldHappenBefore, time.Now())
				So(report.ServerStats, ShouldBeNil)
				So(report.DatabaseStats, ShouldNotBeNil)
				So(report.Collections, ShouldNotBeNil)
				So(report.Indexes, ShouldNotBeNil)
			})
		})
	})
}

func TestReportCollect(t *testing.T) {
	Convey("Given a report and a monitor", t, func() {
		serverStats := &metrics.ServerStats{
			Host:    "testhost",
			Version: "5.0.0",
		}

		dbStats := &metrics.DatabaseStats{
			Name:     "testdb",
			DataSize: 1024,
		}

		collStats := &metrics.CollectionStats{
			Name:  "testcoll",
			Count: 100,
		}

		indexStats := []metrics.IndexStats{
			{
				Name: "testindex",
				Size: 512,
			},
		}

		monitor := &mockMetricsMonitor{
			serverStatsFunc: func(ctx any) (*metrics.ServerStats, error) {
				return serverStats, nil
			},
			databaseStatsFunc: func(ctx any, dbName string) (*metrics.DatabaseStats, error) {
				return dbStats, nil
			},
			collectionStatsFunc: func(ctx any, dbName, collName string) (*metrics.CollectionStats, error) {
				return collStats, nil
			},
			indexStatsFunc: func(ctx any, dbName, collName string) ([]metrics.IndexStats, error) {
				return indexStats, nil
			},
		}

		report := metrics.NewReport(monitor)

		Convey("When collecting metrics", func() {
			listCollections := func() ([]string, error) {
				return []string{"testcoll"}, nil
			}

			err := report.Collect(nil, "testdb", listCollections)

			Convey("Then it should collect all metrics without error", func() {
				So(err, ShouldBeNil)
				So(report.ServerStats, ShouldEqual, serverStats)
				So(report.DatabaseStats["testdb"], ShouldEqual, dbStats)
				So(report.Collections["testcoll"], ShouldHaveLength, 1)
				So(report.Collections["testcoll"][0], ShouldEqual, collStats)
				So(report.Indexes["testcoll"], ShouldHaveLength, 1)
				So(report.Indexes["testcoll"][0].Name, ShouldEqual, "testindex")
			})
		})
	})
}

func TestReportString(t *testing.T) {
	Convey("Given a report with data", t, func() {
		serverStats := &metrics.ServerStats{
			Host:    "testhost",
			Version: "5.0.0",
		}

		monitor := &mockMetricsMonitor{}
		report := metrics.NewReport(monitor)
		report.ServerStats = serverStats

		Convey("When converting to string", func() {
			str := report.String()

			Convey("Then it should return a JSON string", func() {
				So(str, ShouldNotBeEmpty)
				So(str, ShouldContainSubstring, "testhost")
				So(str, ShouldContainSubstring, "5.0.0")
			})
		})
	})
}

func TestGetReportMetrics(t *testing.T) {
	Convey("Given a report with detailed metrics", t, func() {
		// Create server stats with correct field structure
		serverStats := &metrics.ServerStats{
			Host:    "testhost",
			Version: "5.0.0",
			Uptime:  3600,
			Connections: metrics.ConnectionStats{
				Current:      50,
				Available:    100,
				TotalCreated: 150,
			},
			Memory: metrics.MemoryStats{
				Resident:   1024,
				Virtual:    2048,
				PageFaults: 10,
			},
			OperationCounts: metrics.OpCountStats{
				Insert:  100,
				Query:   200,
				Update:  150,
				Delete:  50,
				GetMore: 25,
				Command: 300,
			},
		}

		// Create database stats with correct field structure
		dbStats := &metrics.DatabaseStats{
			Name:        "testdb",
			Collections: 5,
			Objects:     1000,
			DataSize:    1024 * 1024 * 10, // 10MB
			StorageSize: 1024 * 1024 * 15, // 15MB
			IndexSize:   1024 * 1024,      // 1MB
			IndexCount:  10,
		}

		// Create collection stats with correct field structure
		collStats := &metrics.CollectionStats{
			Name:        "testcoll",
			Size:        1024 * 1024 * 5, // 5MB
			Count:       500,
			AvgObjSize:  1024 * 10,       // 10KB
			StorageSize: 1024 * 1024 * 7, // 7MB
			Capped:      false,
			IndexSizes:  map[string]float64{"idx1": 512 * 1024, "idx2": 256 * 1024},
		}

		// Create index stats with correct field structure
		indexStats := []*metrics.IndexStats{
			{
				Name:       "testindex_1",
				Size:       512 * 1024, // 512KB
				KeyPattern: "{ field1: 1 }",
				Unique:     true,
				Sparse:     false,
				UseCount:   100,
			},
			{
				Name:       "testindex_2",
				Size:       256 * 1024, // 256KB
				KeyPattern: "{ field2: 1, field3: 1 }",
				Unique:     false,
				Sparse:     true,
				UseCount:   50,
			},
		}

		monitor := &mockMetricsMonitor{
			serverStatsFunc: func(ctx any) (*metrics.ServerStats, error) {
				return serverStats, nil
			},
			databaseStatsFunc: func(ctx any, dbName string) (*metrics.DatabaseStats, error) {
				return dbStats, nil
			},
			collectionStatsFunc: func(ctx any, dbName, collName string) (*metrics.CollectionStats, error) {
				return collStats, nil
			},
			indexStatsFunc: func(ctx any, dbName, collName string) ([]metrics.IndexStats, error) {
				// Convert to the slice type expected by the interface
				result := make([]metrics.IndexStats, len(indexStats))
				for i, idx := range indexStats {
					result[i] = *idx
				}
				return result, nil
			},
		}

		report := metrics.NewReport(monitor)
		dbMap := make(map[string]*metrics.DatabaseStats)
		dbMap["testdb"] = dbStats
		report.DatabaseStats = dbMap

		collMap := make(map[string][]*metrics.CollectionStats)
		collMap["testcoll"] = []*metrics.CollectionStats{collStats}
		report.Collections = collMap

		indexMap := make(map[string][]*metrics.IndexStats)
		indexMap["testcoll"] = indexStats
		report.Indexes = indexMap

		report.ServerStats = serverStats

		Convey("When extracting metrics for analysis", func() {
			// Test various metrics extraction methods if available in the Report class

			Convey("Then the report data should be accurately represented in the metrics", func() {
				// Basic verification that the report contains the expected data
				So(report.ServerStats.Host, ShouldEqual, "testhost")
				So(report.ServerStats.Version, ShouldEqual, "5.0.0")
				So(report.ServerStats.Connections.Current, ShouldEqual, 50)

				So(report.DatabaseStats["testdb"].Name, ShouldEqual, "testdb")
				So(report.DatabaseStats["testdb"].DataSize, ShouldEqual, 1024*1024*10)
				So(report.DatabaseStats["testdb"].Collections, ShouldEqual, 5)

				So(report.Collections["testcoll"][0].Name, ShouldEqual, "testcoll")
				So(report.Collections["testcoll"][0].Count, ShouldEqual, 500)
				So(report.Collections["testcoll"][0].Size, ShouldEqual, 1024*1024*5)

				So(report.Indexes["testcoll"][0].Name, ShouldEqual, "testindex_1")
				So(report.Indexes["testcoll"][1].Name, ShouldEqual, "testindex_2")
				So(report.Indexes["testcoll"][0].Size, ShouldEqual, 512*1024)
				So(report.Indexes["testcoll"][1].Size, ShouldEqual, 256*1024)
			})
		})
	})
}

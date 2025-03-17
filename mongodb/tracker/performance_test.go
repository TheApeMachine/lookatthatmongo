package tracker

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/lookatthatmongo/mongodb/metrics"
)

// PerformanceReport represents a performance report for testing
type PerformanceReport struct {
	Timestamp        time.Time
	ServerStats      *metrics.ServerStats
	DatabaseStats    *metrics.DatabaseStats
	CollectionStats  []*metrics.CollectionStats
	IndexStats       []*metrics.IndexStats
	ReplicationStats *metrics.RepStats
}

// NewPerformanceReport creates a new performance report for testing
func NewPerformanceReport(opts ...PerformanceReportOption) *PerformanceReport {
	report := &PerformanceReport{
		Timestamp: time.Now(),
	}

	for _, opt := range opts {
		opt(report)
	}

	return report
}

// PerformanceReportOption is a function that configures a PerformanceReport
type PerformanceReportOption func(*PerformanceReport)

// WithServerStats sets the server stats
func WithServerStats(stats *metrics.ServerStats) PerformanceReportOption {
	return func(r *PerformanceReport) {
		r.ServerStats = stats
	}
}

// WithDatabaseStats sets the database stats
func WithDatabaseStats(stats *metrics.DatabaseStats) PerformanceReportOption {
	return func(r *PerformanceReport) {
		r.DatabaseStats = stats
	}
}

// WithCollectionStats sets the collection stats
func WithCollectionStats(stats []*metrics.CollectionStats) PerformanceReportOption {
	return func(r *PerformanceReport) {
		r.CollectionStats = stats
	}
}

// WithIndexStats sets the index stats
func WithIndexStats(stats []*metrics.IndexStats) PerformanceReportOption {
	return func(r *PerformanceReport) {
		r.IndexStats = stats
	}
}

// WithReplicationStatus sets the replication status
func WithReplicationStatus(stats *metrics.RepStats) PerformanceReportOption {
	return func(r *PerformanceReport) {
		r.ReplicationStats = stats
	}
}

// GetServerStats returns the server stats
func (r *PerformanceReport) GetServerStats() *metrics.ServerStats {
	return r.ServerStats
}

// GetDatabaseStats returns the database stats
func (r *PerformanceReport) GetDatabaseStats() *metrics.DatabaseStats {
	return r.DatabaseStats
}

// GetCollectionStats returns the collection stats
func (r *PerformanceReport) GetCollectionStats() []*metrics.CollectionStats {
	return r.CollectionStats
}

// GetIndexStats returns the index stats
func (r *PerformanceReport) GetIndexStats() []*metrics.IndexStats {
	return r.IndexStats
}

// GetReplicationStatus returns the replication status
func (r *PerformanceReport) GetReplicationStatus() *metrics.RepStats {
	return r.ReplicationStats
}

// GetTimestamp returns the timestamp
func (r *PerformanceReport) GetTimestamp() time.Time {
	return r.Timestamp
}

func TestNewPerformanceReport(t *testing.T) {
	Convey("Given a need for a performance report", t, func() {
		Convey("When creating a new performance report without options", func() {
			report := NewPerformanceReport()

			Convey("Then a performance report instance should be created with default values", func() {
				So(report, ShouldNotBeNil)
				So(report.Timestamp, ShouldHappenBefore, time.Now())
				So(report.ServerStats, ShouldBeNil)
				So(report.DatabaseStats, ShouldBeNil)
				So(report.CollectionStats, ShouldBeNil)
				So(report.IndexStats, ShouldBeNil)
				So(report.ReplicationStats, ShouldBeNil)
			})
		})

		Convey("When creating a new performance report with options", func() {
			serverStats := &metrics.ServerStats{
				Host:    "testhost",
				Version: "5.0.0",
			}

			dbStats := &metrics.DatabaseStats{
				Name:     "testdb",
				DataSize: 1024,
			}

			collStats := []*metrics.CollectionStats{
				{
					Name:  "testcoll",
					Count: 100,
				},
			}

			indexStats := []*metrics.IndexStats{
				{
					Name: "testindex",
					Size: 512,
				},
			}

			replStatus := &metrics.RepStats{
				IsReplicaSet: true,
				Members: []metrics.ReplicaMember{
					{
						Name:     "node1",
						StateStr: "PRIMARY",
					},
				},
			}

			report := NewPerformanceReport(
				WithServerStats(serverStats),
				WithDatabaseStats(dbStats),
				WithCollectionStats(collStats),
				WithIndexStats(indexStats),
				WithReplicationStatus(replStatus),
			)

			Convey("Then a performance report instance should be created with the provided values", func() {
				So(report, ShouldNotBeNil)
				So(report.Timestamp, ShouldHappenBefore, time.Now())
				So(report.ServerStats, ShouldEqual, serverStats)
				So(report.DatabaseStats, ShouldEqual, dbStats)
				So(report.CollectionStats, ShouldResemble, collStats)
				So(report.IndexStats, ShouldResemble, indexStats)
				So(report.ReplicationStats, ShouldEqual, replStatus)
			})
		})
	})
}

func TestWithServerStats(t *testing.T) {
	Convey("Given server stats", t, func() {
		serverStats := &metrics.ServerStats{
			Host:    "testhost",
			Version: "5.0.0",
		}

		Convey("When using WithServerStats option", func() {
			optFn := WithServerStats(serverStats)

			Convey("Then it should return a function that sets the server stats", func() {
				report := &PerformanceReport{}
				optFn(report)

				So(report.ServerStats, ShouldEqual, serverStats)
			})
		})
	})
}

func TestWithDatabaseStats(t *testing.T) {
	Convey("Given database stats", t, func() {
		dbStats := &metrics.DatabaseStats{
			Name:     "testdb",
			DataSize: 1024,
		}

		Convey("When using WithDatabaseStats option", func() {
			optFn := WithDatabaseStats(dbStats)

			Convey("Then it should return a function that sets the database stats", func() {
				report := &PerformanceReport{}
				optFn(report)

				So(report.DatabaseStats, ShouldEqual, dbStats)
			})
		})
	})
}

func TestWithCollectionStats(t *testing.T) {
	Convey("Given collection stats", t, func() {
		collStats := []*metrics.CollectionStats{
			{
				Name:  "testcoll",
				Count: 100,
			},
		}

		Convey("When using WithCollectionStats option", func() {
			optFn := WithCollectionStats(collStats)

			Convey("Then it should return a function that sets the collection stats", func() {
				report := &PerformanceReport{}
				optFn(report)

				So(report.CollectionStats, ShouldResemble, collStats)
			})
		})
	})
}

func TestWithIndexStats(t *testing.T) {
	Convey("Given index stats", t, func() {
		indexStats := []*metrics.IndexStats{
			{
				Name: "testindex",
				Size: 512,
			},
		}

		Convey("When using WithIndexStats option", func() {
			optFn := WithIndexStats(indexStats)

			Convey("Then it should return a function that sets the index stats", func() {
				report := &PerformanceReport{}
				optFn(report)

				So(report.IndexStats, ShouldResemble, indexStats)
			})
		})
	})
}

func TestWithReplicationStatus(t *testing.T) {
	Convey("Given replication status", t, func() {
		replStatus := &metrics.RepStats{
			IsReplicaSet: true,
			Members: []metrics.ReplicaMember{
				{
					Name:     "node1",
					StateStr: "PRIMARY",
				},
			},
		}

		Convey("When using WithReplicationStatus option", func() {
			optFn := WithReplicationStatus(replStatus)

			Convey("Then it should return a function that sets the replication status", func() {
				report := &PerformanceReport{}
				optFn(report)

				So(report.ReplicationStats, ShouldEqual, replStatus)
			})
		})
	})
}

func TestGetServerStats(t *testing.T) {
	Convey("Given a performance report with server stats", t, func() {
		serverStats := &metrics.ServerStats{
			Host:    "testhost",
			Version: "5.0.0",
		}

		report := NewPerformanceReport(
			WithServerStats(serverStats),
		)

		Convey("When getting the server stats", func() {
			stats := report.GetServerStats()

			Convey("Then it should return the server stats", func() {
				So(stats, ShouldEqual, serverStats)
			})
		})
	})
}

func TestGetDatabaseStats(t *testing.T) {
	Convey("Given a performance report with database stats", t, func() {
		dbStats := &metrics.DatabaseStats{
			Name:     "testdb",
			DataSize: 1024,
		}

		report := NewPerformanceReport(
			WithDatabaseStats(dbStats),
		)

		Convey("When getting the database stats", func() {
			stats := report.GetDatabaseStats()

			Convey("Then it should return the database stats", func() {
				So(stats, ShouldEqual, dbStats)
			})
		})
	})
}

func TestGetCollectionStats(t *testing.T) {
	Convey("Given a performance report with collection stats", t, func() {
		collStats := []*metrics.CollectionStats{
			{
				Name:  "testcoll",
				Count: 100,
			},
		}

		report := NewPerformanceReport(
			WithCollectionStats(collStats),
		)

		Convey("When getting the collection stats", func() {
			stats := report.GetCollectionStats()

			Convey("Then it should return the collection stats", func() {
				So(stats, ShouldResemble, collStats)
			})
		})
	})
}

func TestGetIndexStats(t *testing.T) {
	Convey("Given a performance report with index stats", t, func() {
		indexStats := []*metrics.IndexStats{
			{
				Name: "testindex",
				Size: 512,
			},
		}

		report := NewPerformanceReport(
			WithIndexStats(indexStats),
		)

		Convey("When getting the index stats", func() {
			stats := report.GetIndexStats()

			Convey("Then it should return the index stats", func() {
				So(stats, ShouldResemble, indexStats)
			})
		})
	})
}

func TestGetReplicationStatus(t *testing.T) {
	Convey("Given a performance report with replication status", t, func() {
		replStatus := &metrics.RepStats{
			IsReplicaSet: true,
			Members: []metrics.ReplicaMember{
				{
					Name:     "node1",
					StateStr: "PRIMARY",
				},
			},
		}

		report := NewPerformanceReport(
			WithReplicationStatus(replStatus),
		)

		Convey("When getting the replication status", func() {
			status := report.GetReplicationStatus()

			Convey("Then it should return the replication status", func() {
				So(status, ShouldEqual, replStatus)
			})
		})
	})
}

func TestGetTimestamp(t *testing.T) {
	Convey("Given a performance report", t, func() {
		report := NewPerformanceReport()

		Convey("When getting the timestamp", func() {
			timestamp := report.GetTimestamp()

			Convey("Then it should return the timestamp", func() {
				So(timestamp, ShouldHappenBefore, time.Now())
				So(timestamp, ShouldHappenAfter, time.Now().Add(-time.Minute))
			})
		})
	})
}

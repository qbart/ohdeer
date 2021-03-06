package deer

import (
	"context"
	"time"
)

// Store interface.
type Store interface {
	// Migrate runs right after creating a connection to store.
	// Store implementation is responsible for creating/migrating schema.
	Migrate(ctx context.Context) error

	// Close shuts down the connection to store.
	Close(ctx context.Context)

	// Save inserts service check result to store.
	Save(ctx context.Context, result *CheckResult)

	// Read loads all metrics from store.
	Read(ctx context.Context, filter *ReadFilter) ([]*Metric, error)

	// Truncate removes all metrics from store.
	Truncate(ctx context.Context) error
}

// ReadFilter contains params to configure metrics read.
type ReadFilter struct {
	Since          time.Time
	TimeBucket     uint
	TimeBucketUnit string
	Interval       uint
	IntervalUnit   string
	ActiveServices map[string][]string
}

// Metric represents metric for given time bucket.
type Metric struct {
	MonitorID    string    `json:"monitor_id"`
	ServiceID    string    `json:"service_id"`
	Bucket       time.Time `json:"bucket"`
	Health       float64   `json:"health"`
	Details      Details   `json:"details"`
	PassedChecks uint64    `json:"passed_checks"`
	FailedChecks uint64    `json:"failed_checks"`
}

// Until calculates when interval should stop.
func (f *ReadFilter) Until() time.Time {
	return f.Since.Add(f.IntervalToDuration())
}

// IntervalToDuration converts user defined interval to time.Duration.
func (f *ReadFilter) IntervalToDuration() time.Duration {
	dur := time.Duration(f.Interval)

	switch f.IntervalUnit {
	case "hour":
		return dur * time.Hour
	case "day":
		return dur * 24 * time.Hour
	}

	return time.Duration(0)
}

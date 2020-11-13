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
	Read(ctx context.Context) ([]*Metric, error)
}

// Metric represents metric for given time bucket.
type Metric struct {
	MonitorID string    `json:"monitor_id"`
	ServiceID string    `json:"service_id"`
	Bucket    time.Time `json:"bucket"`
	Health    float64   `json:"health"`
	Details   Details   `json:"details"`
}

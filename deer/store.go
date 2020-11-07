package deer

import "context"

// Store interface.
type Store interface {
	// Init runs right after creating a connection to store.
	// Store implementation is responsible for creating/migrating schema.
	Init(ctx context.Context) error

	// Close shuts down the connection to store.
	Close(ctx context.Context)

	// Save inserts service check result to store.
	Save(ctx context.Context, result *CheckResult)
}

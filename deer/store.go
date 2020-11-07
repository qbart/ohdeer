package deer

// Store interface.
type Store interface {
	// Init runs right after creating a connection to store.
	// Store implementation is responsible for creating/migrating schema.
	Init()

	// Close shuts down the connection to store.
	Close()

	// Save inserts service check result to store.
	Save(result *CheckResult)
}

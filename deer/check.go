package deer

import "time"

// CheckResult is a struct.
type CheckResult struct {
	MonitorID string
	ServiceID string
	At        time.Time
	Success   bool
}

// Check is a interface for monitoring checks.
type Check interface {
	Validatable

	RunFn(s *Store) func()
	Check(resp *Response) bool
}

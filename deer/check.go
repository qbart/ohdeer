package deer

import (
	"net/http"
)

// CheckResult is a struct.
type CheckResult struct {
	MonitorID string
	ServiceID string
	Success   bool
}

// Check is a interface for monitoring checks.
type Check interface {
	Validatable

	RunFn(s *Store) func()
	Check(resp http.Response, err error) bool
}

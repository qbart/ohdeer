package deer

import "time"

// CheckResult is a struct.
type CheckResult struct {
	MonitorID string
	ServiceID string
	At        time.Time
	Success   bool
	Trace     *Trace
}

// Check is a interface for monitoring checks.
type Check interface {
	Validatable

	RunFn(s *Store) func()
	Check(resp *Response) bool
}

// Details for checks.
type Details struct {
	Trace *Trace `json:"trace"`
}

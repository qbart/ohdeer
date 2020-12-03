package deer

import "time"

// CheckResult is a struct.
type CheckResult struct {
	MonitorID  string
	ServiceID  string
	At         time.Time
	Success    bool
	Trace      *Trace
	Error      error
	StatusCode int
}

// Check is a interface for monitoring checks.
type Check interface {
	Validatable

	RunFn(s *Store) func()
	Check(resp *Response) bool
}

// Details for checks.
type Details struct {
	Trace    *Trace           `json:"trace"`
	Error    *ErrorDetails    `json:"error,omitempty"`
	Response *ResponseDetails `json:"response,omitempty"`
}

// ErrorDetails contains response error.
type ErrorDetails struct {
	Message string `json:"message"`
}

// ResponseDetails contains response error.
type ResponseDetails struct {
	StatusCode int `json:"status_code"`
}

package deer

import (
	"context"
	"fmt"
	"time"
)

// HTTPCheck defines http type check.
type HTTPCheck struct {
	ref

	// body
	IntervalSec  uint64   `hcl:"interval"`
	TimeoutSec   uint64   `hcl:"timeout"`
	Addr         string   `hcl:"addr"`
	Expectations []Expect `hcl:"expect,block"`
}

// Validate ensures correct values are set for http check.
func (h *HTTPCheck) Validate() error {
	switch {
	case h.TimeoutSec <= 0:
		return fmt.Errorf("Timeout must be > 0")

	case h.IntervalSec <= 0:
		return fmt.Errorf("Interval must be > 0")

	case len(h.Addr) == 0:
		return fmt.Errorf("Addr cannot be empty")

	case len(h.Expectations) == 0:
		return fmt.Errorf("At least one expectation fot http check is required")
	}

	for _, expect := range h.Expectations {
		if expect.Subject != "status" {
			return fmt.Errorf("Invalid expectation subject")
		}
	}

	return nil
}

// RunFn returns task function to run check.
func (h *HTTPCheck) RunFn(s Store) func() {
	store := s

	return func() {
		now := time.Now()
		req := Request{}
		resp := req.Get(h.Addr, time.Duration(h.TimeoutSec)*time.Second)

		success := h.Check(resp)
		result := CheckResult{
			MonitorID: h.ref.Monitor.ID,
			ServiceID: h.ref.Service.ID,
			At:        now,
			Success:   success,
			Trace:     &resp.Trace,
		}
		store.Save(context.Background(), &result)
	}
}

// Check verifies if check is valid or not.
func (h *HTTPCheck) Check(resp *Response) bool {
	if resp.Err != nil {
		return false
	}

	success := true

	for _, expect := range h.Expectations {
		switch expect.Subject {
		case "status":
			status := resp.Resp.StatusCode
			found := false
			for _, s := range expect.Inclusion {
				if s == status {
					found = true
				}
			}

			success = success && found
		}
	}
	return success
}

package deer

import "fmt"

// HttpCheck defines http type check.
type HttpCheck struct {
	// body
	IntervalSec  uint64   `hcl:"interval"`
	TimeoutSec   uint64   `hcl:"timeout"`
	Addr         string   `hcl:"addr"`
	Expectations []Expect `hcl:"expect,block"`
}

// Validate ensures correct values are set for http check.
func (h *HttpCheck) Validate() error {
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

	return nil
}

// RunFn returns task function to run check.
func (h *HttpCheck) RunFn() func() {
	return func() {
		fmt.Println("GET", h.Addr, "timeout", h.TimeoutSec)
	}
}

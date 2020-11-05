package deer

import (
	"fmt"
)

type Validatable interface {
	Validate() error
}

// Monitor keeps the configuration of a single monitor.
// It aggregates one or more services.
type Monitor struct {
	// label
	ID string `hcl:"id,label"`

	// body
	Name     string    `hcl:"name"`
	Services []Service `hcl:"service,block"`
}

// Service defines monitor checks.
type Service struct {
	// label
	ID string `hcl:"id,label"`

	// body
	Name       string      `hcl:"name"`
	HttpChecks []HttpCheck `hcl:"http,block"`
}

// HttpCheck defines http type check.
type HttpCheck struct {
	// body
	IntervalSec  int      `hcl:"interval"`
	TimeoutSec   int      `hcl:"timeout"`
	Addr         string   `hcl:"addr"`
	Expectations []Expect `hcl:"expect,block"`
}

func (h *HttpCheck) Validate() error {
	switch {
	case h.TimeoutSec <= 0:
		return fmt.Errorf("Timeout must be > 0")

	case h.IntervalSec <= 0:
		return fmt.Errorf("Interval must be > 0")

	case len(h.Addr) == 0:
		return fmt.Errorf("Addr cannot be empty")

	case len(h.Expectations) < 0:
		return fmt.Errorf("At least one expectation fot http check is required")
	}

	return nil
}

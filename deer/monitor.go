package deer

type Validatable interface {
	Validate() error
}

// Monitor keeps the configuration of a single monitor.
// It aggregates one or more services.
type Monitor struct {
	// label
	ID string `hcl:"id,label"`

	// body
	Name     string     `hcl:"name"`
	Services []*Service `hcl:"service,block"`
}

// Service defines monitor checks.
type Service struct {
	// label
	ID string `hcl:"id,label"`

	// body
	Name       string       `hcl:"name"`
	HttpChecks []*HttpCheck `hcl:"http,block"`
}

type ref struct {
	Monitor *Monitor
	Service *Service
}

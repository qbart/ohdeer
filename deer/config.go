package deer

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Config keeps monitor configuration.
type Config struct {
	Monitors []Monitor `hcl:"monitor,block"`
}

// ParseConfig load and parses config from given path.
func ParseConfig(path string, src []byte) (*Config, error) {
	var cfg Config
	err := hclsimple.Decode(path, src, nil, &cfg)

	if err == nil {
		for _, m := range cfg.Monitors {
			if len(m.ID) == 0 {
				return nil, fmt.Errorf("Monitor cannot have empty ID")
			}
			if len(m.Name) == 0 {
				return nil, fmt.Errorf("Monitor cannot have empty name")
			}

			for _, s := range m.Services {
				if len(s.ID) == 0 {
					return nil, fmt.Errorf("Service in monitor %s cannot have empty ID", m.ID)
				}
				if len(s.Name) == 0 {
					return nil, fmt.Errorf("Service in monitor %s cannot have empty name", m.ID)
				}

				for _, h := range s.HttpChecks {
					if err := h.Validate(); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return &cfg, err
}

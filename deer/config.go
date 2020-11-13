package deer

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Config keeps monitor configuration.
type Config struct {
	Tls      *Tls       `hcl:"tls,block"`
	Monitors []*Monitor `hcl:"monitor,block"`
}

type Tls struct {
	Domain   string `hcl:"domain"`
	CacheDir string `hcl:"cache_dir"`
}

// LoadConfig loads and parses config from given path.
func LoadConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseConfig(path, b)
}

// ParseConfig parses config from given data.
func ParseConfig(path string, src []byte) (*Config, error) {
	var cfg Config
	err := hclsimple.Decode(path, src, nil, &cfg)

	if err == nil {
		if cfg.Tls == nil {
			cfg.Tls = &Tls{}
		}
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
					h.ref = ref{Monitor: m, Service: s}

					if err := h.Validate(); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return &cfg, err
}

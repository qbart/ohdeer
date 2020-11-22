package deer

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Config keeps monitor configuration.
type Config struct {
	TLS      *TLS       `hcl:"tls,block"`
	Monitors []*Monitor `hcl:"monitor,block"`
}

// TLS configuration.
type TLS struct {
	Domain   string `hcl:"domain"`
	CacheDir string `hcl:"cache_dir"`
	CertFile string `hcl:"cert_file"`
	KeyFile  string `hcl:"key_file"`
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
		if cfg.TLS == nil {
			cfg.TLS = &TLS{}
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

				for _, h := range s.HTTPChecks {
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

// ActiveServices returns list of active services per monitor.
func (c *Config) ActiveServices() map[string][]string {
	r := make(map[string][]string, 0)
	for _, m := range c.Monitors {
		r[m.ID] = make([]string, 0)
		for _, s := range m.Services {
			r[m.ID] = append(r[m.ID], s.ID)
		}
	}
	return r
}

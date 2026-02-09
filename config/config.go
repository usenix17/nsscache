package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	LDAP   LDAPConfig   `yaml:"ldap"`
	Cache  CacheConfig  `yaml:"cache"`
	Server ServerConfig `yaml:"server"`
}

// LDAPConfig contains LDAP connection settings
type LDAPConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	UseTLS       bool   `yaml:"use_tls"`
	SkipVerify   bool   `yaml:"skip_verify"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	BaseDN       string `yaml:"base_dn"`
	UserFilter   string `yaml:"user_filter"`
	GroupFilter  string `yaml:"group_filter"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	TTL int `yaml:"ttl"` // seconds
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Listen string `yaml:"listen"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		// Defaults
		LDAP: LDAPConfig{
			Port:         636,
			UseTLS:       true,
			UserFilter:   "(objectClass=posixAccount)",
			GroupFilter:  "(objectClass=posixGroup)",
		},
		Cache: CacheConfig{
			TTL: 300,
		},
		Server: ServerConfig{
			Listen: ":8080",
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Override bind password from environment if set
	if envPass := os.Getenv("LDAP_BIND_PASSWORD"); envPass != "" {
		cfg.LDAP.BindPassword = envPass
	}

	return cfg, nil
}

package config

// Config represents the application configuration format (config.yaml).
type Config struct {
	Version     string             `yaml:"version"`
	Contexts    map[string]*Context `yaml:"contexts"`
	Preferences map[string]string  `yaml:"preferences"`
}

// Context represents a generic, provider-agnostic environment context configuration.
type Context struct {
	Provider string         `yaml:"provider"`
	Raw      map[string]any `yaml:",inline"`
}

package config

// Config represents the application configuration format (config.yaml).
type Config struct {
	Version     string                `yaml:"version"`
	Current     string                `yaml:"current"`
	Workspaces  map[string]*Workspace `yaml:"workspaces"`
	Preferences map[string]string     `yaml:"preferences"`
}

// Workspace represents a generic, provider-agnostic environment configuration.
type Workspace struct {
	Provider string         `yaml:"provider"`
	Raw      map[string]any `yaml:",inline"`
}

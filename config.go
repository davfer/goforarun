package goforarun

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BaseAppConfig is the base configuration for the service, it needs a name, a log level, a log format,
// and an observability mode.
type BaseAppConfig struct {
	// ServiceName is the identifier of the service
	ServiceName string `yaml:"service_name"`
	// LoggingConfig is the configuration of the logs
	LoggingConfig LoggingConfig `yaml:"logs"`
	// BuildInfo is the information of the build. Useful to identify running process for observability.
	BuildInfo *BuildInfo
}

// ServiceVersionedName returns the service name with the version if it exists, useful to group equal running services
func (c *BaseAppConfig) ServiceVersionedName() string {
	if c.BuildInfo != nil {
		return c.ServiceName + "-" + c.BuildInfo.Version
	}

	return c.ServiceName
}

// BuildInfo is the information of the build, it contains the version, the commit and the date.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewConfig creates a new configuration from a file path
func NewConfig[K any](configPath string) (K, error) {
	var config K

	file, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

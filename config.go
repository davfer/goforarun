package app

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

// BaseAppConfig is the base configuration for the service, it needs a name, a log level, a log format,
// and an observability mode.
type BaseAppConfig struct {
	// ServiceName is the identifier of the service
	ServiceName string `yaml:"service_name"`
	// LoggingLevel is the level of the logger
	LoggingLevel string `yaml:"log_level"`
	// LoggingFormat is the format of the logger (logrus)
	LoggingFormat string `yaml:"log_format"`
	// ObservabilityMode is the mode of the observability, currently only traces (otpl, file, stdout, disabled)
	ObservabilityMode string `yaml:"observability_mode"`
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

func NewConfig[K any](configPath string) (K, error) {
	var config K

	file, err := os.Open(configPath)
	if err != nil {
		return config, errors.Wrap(err, "failed to open config file")
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		return config, errors.Wrap(err, "failed to decode config file")
	}

	return config, nil
}

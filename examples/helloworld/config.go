package main

import (
	"github.com/davfer/goforarun/config"
)

type ExampleConfig struct {
	FrameworkConfig *config.BaseAppConfig `yaml:"framework"`
}

func (c *ExampleConfig) Framework() *config.BaseAppConfig {
	return c.FrameworkConfig
}

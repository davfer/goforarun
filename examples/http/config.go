package main

import (
	"github.com/davfer/goforarun/config"
)

type HttpServiceConfig struct {
	FrameworkConfig *config.BaseAppConfig `yaml:"framework"`
}

func (c *HttpServiceConfig) Framework() *config.BaseAppConfig {
	return c.FrameworkConfig
}

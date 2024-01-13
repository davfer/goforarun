package main

import app "github.com/davfer/goforarun"

type HttpServiceConfig struct {
	FrameworkConfig *app.BaseAppConfig `yaml:"framework"`
}

func (c *HttpServiceConfig) Framework() *app.BaseAppConfig {
	return c.FrameworkConfig
}

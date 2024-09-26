package main

import (
	app "github.com/davfer/goforarun"
)

type ExampleConfig struct {
	FrameworkConfig *app.BaseAppConfig `yaml:"framework"`
}

func (c *ExampleConfig) Framework() *app.BaseAppConfig {
	return c.FrameworkConfig
}

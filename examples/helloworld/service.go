package main

import (
	"context"
	"fmt"
	app "github.com/davfer/goforarun"
)

type ExampleService struct {
	cfg *ExampleConfig
}

func (e *ExampleService) Init(cfg *ExampleConfig) ([]app.RunnableServer, error) {
	e.cfg = cfg

	return []app.RunnableServer{}, nil
}

func (e *ExampleService) Run(ctx context.Context) error {
	fmt.Printf("Hello, %s!\n", e.cfg.Framework().ServiceName)

	return nil
}

func (e *ExampleService) Shutdown(ctx context.Context) error {
	return nil
}

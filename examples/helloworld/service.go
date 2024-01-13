package main

import (
	"context"
	"fmt"
	app "github.com/davfer/goforarun"
)

type ExampleService[Config app.Config] struct {
	cfg Config
}

func (e *ExampleService[Config]) Init(cfg Config) ([]app.RunnableServer, error) {
	e.cfg = cfg

	return []app.RunnableServer{}, nil
}

func (e *ExampleService[Config]) Run(ctx context.Context) error {
	fmt.Printf("Hello, %s!\n", e.cfg.Framework().ServiceName)

	return nil
}

func (e *ExampleService[Config]) Shutdown(ctx context.Context) error {
	return nil
}

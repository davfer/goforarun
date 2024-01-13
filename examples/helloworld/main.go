package main

import (
	"context"
	app "github.com/davfer/goforarun"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if svc, err := app.NewService[*ExampleService[*ExampleConfig], *ExampleConfig](
		&ExampleService[*ExampleConfig]{},
		&app.BuildInfo{Version: version, Commit: commit, Date: date},
	); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}

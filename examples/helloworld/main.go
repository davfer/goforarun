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
	build := &app.BuildInfo{Version: version, Commit: commit, Date: date}
	if svc, err := app.NewService[*ExampleService, *ExampleConfig](&ExampleService{}, build); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}

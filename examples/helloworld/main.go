package main

import (
	"context"
	"github.com/davfer/goforarun"
	"github.com/davfer/goforarun/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	build := &config.BuildInfo{Version: version, Commit: commit, Date: date}
	if svc, err := goforarun.NewService[*ExampleService, *ExampleConfig](&ExampleService{}, build); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}

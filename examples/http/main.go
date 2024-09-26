package main

import (
	"context"
	app "github.com/davfer/goforarun"
	"github.com/davfer/goforarun/config"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Setenv("DEBUG", "true")

	build := &config.BuildInfo{Version: version, Commit: commit, Date: date}
	if svc, err := app.NewService[*HttpService, *HttpServiceConfig](&HttpService{}, build); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}

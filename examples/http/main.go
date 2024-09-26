package main

import (
	"context"
	"os"

	app "github.com/davfer/goforarun"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Setenv("DEBUG", "true")

	build := &app.BuildInfo{Version: version, Commit: commit, Date: date}
	if svc, err := app.NewService[*HttpService, *HttpServiceConfig](&HttpService{}, build); err != nil {
		panic(err)
	} else {
		svc.Run(context.Background())
	}
}

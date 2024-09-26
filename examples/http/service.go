package main

import (
	"context"
	"log/slog"
	"net/http"

	app "github.com/davfer/goforarun"
	gofarhttp "github.com/davfer/goforarun/http"
	"github.com/davfer/goforarun/logger"
)

type HttpService struct {
	cfg    *HttpServiceConfig
	logger *slog.Logger
}

func (e *HttpService) Init(cfg *HttpServiceConfig) ([]app.RunnableServer, error) {
	e.cfg = cfg
	e.logger = logger.Get("http-server")

	server := gofarhttp.NewHttpBaseServer(&app.InfoServer{
		Net:  "tcp",
		Host: "",
		Port: "8090",
		Name: "server",
	}, func(w http.ResponseWriter, r *http.Request) {
		e.logger.Info("request received", slog.String("method", r.Method))
		w.Write([]byte("Hello, world!\n"))
	})

	return []app.RunnableServer{server}, nil
}

func (e *HttpService) Run(ctx context.Context) error {
	e.logger.Debug("running http service") // this will not be shown in stdout

	// Init consumers, etc.

	return nil
}

func (e *HttpService) Shutdown(ctx context.Context) error {
	return nil
}

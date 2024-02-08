package main

import (
	"context"
	"github.com/davfer/goforarun"
	gofarhttp "github.com/davfer/goforarun/http"
	"github.com/davfer/goforarun/observability"
	"net/http"
)

type HttpService struct {
	observability.ObservableStruct
	cfg *HttpServiceConfig
}

func (e *HttpService) Init(cfg *HttpServiceConfig) ([]goforarun.RunnableServer, error) {
	e.InitObservableStruct("my-server-service")
	e.cfg = cfg

	server := gofarhttp.NewHttpBaseServer(&goforarun.InfoServer{
		Net:  "tcp",
		Host: "",
		Port: "8081",
		Name: "server",
	}, func(w http.ResponseWriter, r *http.Request) {
		e.Logger.WithField("method", r.Method).Info("Request received")
		w.Write([]byte("Hello, world!\n"))
	})

	return []goforarun.RunnableServer{server}, nil
}

func (e *HttpService) Run(ctx context.Context) error {
	e.Logger.Debug("Running http service")

	// Init consumers, etc.

	return nil
}

func (e *HttpService) Shutdown(ctx context.Context) error {
	return nil
}

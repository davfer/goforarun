package main

import (
	"context"
	app "github.com/davfer/goforarun"
	"net/http"
)

type HttpService struct {
	app.ObservableStruct
	cfg *HttpServiceConfig
}

func (e *HttpService) Init(cfg *HttpServiceConfig) ([]app.RunnableServer, error) {
	e.InitObservableStruct("my-server-service")
	e.cfg = cfg

	server := app.NewHttpBaseServer(&app.InfoServer{
		Net:  "tcp",
		Host: "",
		Port: "8081",
		Name: "server",
	}, func(w http.ResponseWriter, r *http.Request) {
		e.Logger.WithField("method", r.Method).Info("Request received")
		w.Write([]byte("Hello, world!\n"))
	})

	return []app.RunnableServer{server}, nil
}

func (e *HttpService) Run(ctx context.Context) error {
	e.Logger.Debug("Running http service")

	// Init consumers, etc.

	return nil
}

func (e *HttpService) Shutdown(ctx context.Context) error {
	return nil
}

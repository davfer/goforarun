package http

import (
	"context"
	"github.com/davfer/goforarun"
	"github.com/davfer/goforarun/logger"
	"log/slog"
	"net/http"
)

type BaseServer struct {
	info       *goforarun.InfoServer
	httpServer *http.Server
	logger     *slog.Logger
	handler    http.HandlerFunc
}

func NewHttpBaseServer(info *goforarun.InfoServer, handler http.HandlerFunc) goforarun.RunnableServer {
	return &BaseServer{
		info:       info,
		logger:     logger.Get("http-server", slog.String("name", info.Name)),
		httpServer: nil,
		handler:    handler,
	}
}

func (cs *BaseServer) Run(ctx context.Context) error {
	cs.logger.With("connection", cs.info).Info("listening server")
	cs.httpServer = &http.Server{
		Addr:    cs.info.Host + ":" + cs.info.Port,
		Handler: cs.handler,
	}

	return cs.httpServer.ListenAndServe()
}

func (cs *BaseServer) Shutdown(ctx context.Context) error {
	return cs.httpServer.Shutdown(ctx)
}

func (cs *BaseServer) Info() *goforarun.InfoServer {
	return cs.info
}

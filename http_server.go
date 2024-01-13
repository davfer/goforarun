package app

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
)

type HttpBaseServer struct {
	info       *InfoServer
	httpServer *http.Server
	logger     *logrus.Entry
	handler    http.HandlerFunc
}

func NewHttpBaseServer(info *InfoServer, handler http.HandlerFunc) RunnableServer {
	return &HttpBaseServer{
		info:       info,
		logger:     NewLogger("http-server").WithField("name", info.Name),
		httpServer: nil,
		handler:    handler,
	}
}

func (cs *HttpBaseServer) Run(ctx context.Context) error {
	cs.logger.WithField("connection", cs.info).Info("listening server")
	cs.httpServer = &http.Server{
		Addr:    cs.info.Host + ":" + cs.info.Port,
		Handler: cs.handler,
	}

	return cs.httpServer.ListenAndServe()
}

func (cs *HttpBaseServer) Shutdown(ctx context.Context) error {
	return cs.httpServer.Shutdown(ctx)
}

func (cs *HttpBaseServer) Info() *InfoServer {
	return cs.info
}

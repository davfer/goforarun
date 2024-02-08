package grpc

import (
	"context"
	"github.com/davfer/goforarun"
	"github.com/davfer/goforarun/observability"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type BaseServer struct {
	info       *goforarun.InfoServer
	grpcServer *grpc.Server
	logger     *logrus.Entry
	registrars []ServiceRegisterFunc
}

type ServiceRegisterFunc func(s *grpc.Server)

func NewGrpcBaseServer(info *goforarun.InfoServer, registrars []ServiceRegisterFunc) goforarun.RunnableServer {
	return &BaseServer{
		info:       info,
		registrars: registrars,
		logger:     observability.NewLogger("grpc-server").WithField("name", info.Name),
	}
}

func (cs *BaseServer) Run(ctx context.Context) error {
	cs.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_logrus.UnaryServerInterceptor(cs.logger.WithField("type", "interceptor")),
				grpc_validator.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor(),
				//otelgrpc.UnaryServerInterceptor(),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_logrus.StreamServerInterceptor(cs.logger.WithField("type", "interceptor")),
				grpc_validator.StreamServerInterceptor(),
				grpc_recovery.StreamServerInterceptor(),
				//otelgrpc.StreamServerInterceptor(),
			),
		),
	)

	// server part
	cs.logger.WithField("connection", cs.info).Info("listening server")
	listen, err := net.Listen(cs.info.Net, cs.info.Host+":"+cs.info.Port)
	if err != nil {
		return errors.Wrap(err, "error listening server")
	}

	// register part
	for _, reg := range cs.registrars {
		cs.logger.Debug("registering grpc service")
		reg(cs.grpcServer)
	}

	// start server
	cs.logger.Info("starting grpc server")
	return cs.grpcServer.Serve(listen)
}

func (cs *BaseServer) Shutdown(ctx context.Context) error {
	cs.logger.Debug("shutting down grpc server")
	cs.grpcServer.GracefulStop()
	cs.logger.Info("grpc server stopped")

	return nil
}

func (cs *BaseServer) Info() *goforarun.InfoServer {
	return cs.info
}

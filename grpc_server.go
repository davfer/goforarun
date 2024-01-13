package app

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"net"
)

type GrpcBaseServer struct {
	info       *InfoServer
	grpcServer *grpc.Server
	logger     *logrus.Entry
	registrars []ServiceRegisterFunc
}

type ServiceRegisterFunc func(s *grpc.Server)

func NewGrpcBaseServer(info *InfoServer, registrars []ServiceRegisterFunc) RunnableServer {
	return &GrpcBaseServer{
		info:       info,
		registrars: registrars,
		logger:     NewLogger("grpc-server").WithField("name", info.Name),
	}
}

func (cs *GrpcBaseServer) Run(ctx context.Context) error {
	cs.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_logrus.UnaryServerInterceptor(cs.logger.WithField("type", "interceptor")),
				grpc_validator.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor(),
				otelgrpc.UnaryServerInterceptor(),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_logrus.StreamServerInterceptor(cs.logger.WithField("type", "interceptor")),
				grpc_validator.StreamServerInterceptor(),
				grpc_recovery.StreamServerInterceptor(),
				otelgrpc.StreamServerInterceptor(),
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

func (cs *GrpcBaseServer) Shutdown(ctx context.Context) error {
	cs.logger.Debug("shutting down grpc server")
	cs.grpcServer.GracefulStop()
	cs.logger.Info("grpc server stopped")

	return nil
}

func (cs *GrpcBaseServer) Info() *InfoServer {
	return cs.info
}

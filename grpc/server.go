package grpc

import (
	"context"
	"github.com/davfer/goforarun"
	"github.com/davfer/goforarun/logger"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcvalidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type BaseServer struct {
	info       *goforarun.InfoServer
	grpcServer *grpc.Server
	logger     *slog.Logger
	registrars []ServiceRegisterFunc
}

type ServiceRegisterFunc func(s *grpc.Server)

func NewGrpcBaseServer(info *goforarun.InfoServer, registrars []ServiceRegisterFunc) goforarun.RunnableServer {
	return &BaseServer{
		info:       info,
		registrars: registrars,
		logger:     logger.Get("grpc-server", slog.String("name", info.Name)),
	}
}

func (cs *BaseServer) Run(ctx context.Context) error {
	cs.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				//grpclogrus.UnaryServerInterceptor(cs.logger.With("type", "interceptor")),
				grpcvalidator.UnaryServerInterceptor(),
				grpcrecovery.UnaryServerInterceptor(),
				//otelgrpc.UnaryServerInterceptor(),
			),
		),
		grpc.StreamInterceptor(
			grpcmiddleware.ChainStreamServer(
				//grpclogrus.StreamServerInterceptor(cs.logger.WithField("type", "interceptor")),
				grpcvalidator.StreamServerInterceptor(),
				grpcrecovery.StreamServerInterceptor(),
				//otelgrpc.StreamServerInterceptor(),
			),
		),
	)

	// server part
	cs.logger.With("connection", cs.info).Info("listening server")
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

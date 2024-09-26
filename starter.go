package goforarun

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/davfer/goforarun/config"
	"github.com/davfer/goforarun/logger"
	"github.com/davfer/goforarun/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

const AppLoggerName = "gofar"

var ErrGracefulShutdown = errors.New("graceful shutdown")

type Config interface {
	Framework() *config.BaseAppConfig
}

type App[V any] interface {
	Init(cfg V) ([]RunnableServer, error)
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type Service[K App[V], V Config] struct {
	BaseService[V]
	app     K
	servers []RunnableServer
	logger  *slog.Logger
}

type BaseService[V any] struct {
	Cfg V
}

// NewService creates a new service with the given app and config.
// This is the main and only entry point for the GoForARun framework.
func NewService[K App[V], V Config](app K, buildInfo *config.BuildInfo) (*Service[K, V], error) {
	// config
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.NewConfig[V](configFile)
	if err != nil {
		return nil, fmt.Errorf("could not start without config: %w", err)
	}

	cfg.Framework().BuildInfo = buildInfo

	// observability
	var opts []observability.Customizer
	if cfg.Framework().ServiceName != "" {
		opts = append(opts, observability.WithServiceName(cfg.Framework().ServiceName))
	}
	if buildInfo.Version != "" {
		opts = append(opts, observability.WithServiceVersion(buildInfo.Version))
	}
	if cfg.Framework().LoggingConfig.Level != "" {
		var l slog.Leveler
		l, err = observability.ParseLevel(cfg.Framework().LoggingConfig.Level)
		if err != nil {
			return nil, err
		}
		opts = append(opts, observability.WithLoggerLevel(l))
	}
	if len(cfg.Framework().LoggingConfig.FilteredChannels) > 0 {
		opts = append(opts, observability.WithLoggerChannels(cfg.Framework().LoggingConfig.FilteredChannels))
	}
	if val, ok := os.LookupEnv("DEBUG"); ok && val == "true" {
		opts = append(opts, observability.WithLoggerStdout(true))
	}

	err = observability.StartObservability(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("could not start observability: %w", err)
	}
	l := logger.Get(AppLoggerName)

	/////////////////////
	// INIT USER APP
	l.With("build", buildInfo).Debug("initializing app")
	servers, err := app.Init(cfg)
	if err != nil {
		l.Error("could not initialize app", logger.AttrErr(err))
		return nil, fmt.Errorf("could not initialize app: %w", err)
	}
	/////////////////////

	return &Service[K, V]{
		BaseService[V]{Cfg: cfg},
		app,
		servers,
		l,
	}, nil
}

// Run starts the service and blocks until it receives a SIGINT or the app crashes.
// If the app Run method returns an error, the service will log it and exit.
func (s *Service[K, V]) Run(ctx context.Context) {
	tracedCtx, span := otel.Tracer(AppLoggerName).Start(ctx, "run")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error)
	for i := range s.servers {
		go func(server RunnableServer) {
			s.logger.With("server", server.Info().Name).Debug("starting unmanaged server")
			err := server.Run(tracedCtx)
			if err != nil {
				errCh <- err
			}
		}(s.servers[i])
	}
	// TODO: add condition to wait for all servers to be listening

	go func() {
		s.logger.Debug("starting app")
		err := s.app.Run(tracedCtx)
		s.logger.Debug("app ran successfully", logger.AttrErr(err))
		if err != nil || len(s.servers) == 0 {
			errCh <- err
		}
	}()

	for {
		select {
		case <-sigCh:
			s.logger.Info("starting soft shutdown")

			ctxShutdown, cancel := context.WithTimeout(tracedCtx, 10*time.Second)

			for _, server := range s.servers {
				s.logger.With("server", server.Info().Name).Debug("shutting down unmanaged server")
				err := server.Shutdown(ctxShutdown)
				if err != nil {
					s.logger.With("server", server.Info().Name).Error("error while shutting down unmanaged server", logger.AttrErr(err))
				}
			}

			s.logger.Debug("shutting down app")
			err := s.app.Shutdown(ctxShutdown)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				s.logger.Error("error while shutting down business app", logger.AttrErr(err))
			}
			span.End()

			s.logger.Debug("shutting down observability")
			err = observability.StopObservability(ctxShutdown)
			if err != nil {
				s.logger.Error("error while closing observability", logger.AttrErr(err))
			}

			cancel()
			s.logger.Debug("exiting")

			os.Exit(125)
		case err := <-errCh:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()

			if errors.Is(err, ErrGracefulShutdown) || err == nil {
				s.logger.Info("graceful shutdown")

				os.Exit(0)

				return
			}

			s.logger.Error("service crashed", logger.AttrErr(err))

			err = observability.StopObservability(context.Background())
			if err != nil {
				s.logger.Error("error while closing observability", logger.AttrErr(err))
			}

			os.Exit(1)
		}
	}
}

package app

import (
	"context"
	"flag"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/codes"
	"os"
	"os/signal"
	"time"
)

const (
	AppLoggerName = "gfar"
)

var ErrGracefulShutdown = errors.New("graceful shutdown")

type Config interface {
	Framework() *BaseAppConfig
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
}

type BaseService[V any] struct {
	Cfg V
}

// NewService creates a new service with the given app and config.
// This is the main and only entry point for the GoForARun framework.
func NewService[K App[V], V Config](app K, buildInfo *BuildInfo) (*Service[K, V], error) {
	// config
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := NewConfig[V](configFile)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start without config")
	}

	cfg.Framework().BuildInfo = buildInfo

	// observability
	err = SetObservabilityConfig(cfg.Framework())
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set observability config")
	}
	l := NewLogger(AppLoggerName)

	/////////////////////
	// INIT USER APP
	l.WithField("build", buildInfo).Debug("initializing app")
	servers, err := app.Init(cfg)
	if err != nil {
		l.WithError(err).Fatal("couldn't initialize app")
		return nil, errors.Wrap(err, "couldn't initialize app")
	}
	/////////////////////

	return &Service[K, V]{
		BaseService[V]{Cfg: cfg},
		app,
		servers,
	}, nil
}

// Run starts the service and blocks until it receives a SIGINT or the app crashes.
// If the app Run method returns an error, the service will log it and exit.
func (s *Service[K, V]) Run(ctx context.Context) {
	l := NewLogger(AppLoggerName)
	tracedCtx, span := NewTracer(AppLoggerName).Start(ctx, "run")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	errCh := make(chan error)
	for i := range s.servers {
		go func(server RunnableServer) {
			l.WithField("server", server.Info().Name).Debug("starting unmanaged server")
			err := server.Run(tracedCtx)
			if err != nil {
				errCh <- err
			}
		}(s.servers[i])
	}
	// TODO: add condition to wait for all servers to be listening

	go func() {
		l.Debug("starting app")
		err := s.app.Run(tracedCtx)
		l.WithError(err).Debug("finishing app")
		if err != nil || len(s.servers) == 0 {
			errCh <- err
		}
	}()

	for {
		select {
		case <-sigCh:
			l.Info("starting soft shutdown")

			ctxShutdown, cancel := context.WithTimeout(tracedCtx, 10*time.Second)

			for _, server := range s.servers {
				l.WithField("server", server.Info().Name).Debug("shutting down unmanaged server")
				err := server.Shutdown(ctxShutdown)
				if err != nil {
					l.WithField("server", server.Info().Name).WithError(err).Error("error while shutting down unmanaged server")
				}
			}

			l.Debug("shutting down app")
			err := s.app.Shutdown(ctxShutdown)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				l.WithError(err).Error("error while shutting down business app")
			}
			span.End()

			l.Debug("shutting down observability")
			err = CloseObservability(ctxShutdown)
			if err != nil {
				l.WithError(err).Error("error while closing observability")
			}

			cancel()
			l.Debug("exiting")

			os.Exit(125)
		case err := <-errCh:
			if errors.Is(err, ErrGracefulShutdown) || err == nil {
				l.Info("graceful shutdown")
				span.End()

				os.Exit(0)

				return
			}

			if _, ok := err.(error); ok {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()

			if err != nil {
				l.WithError(err).Error("service crashed")

				err = CloseObservability(context.Background())
				if err != nil {
					l.WithError(err).Error("error while closing observability")
				}

				l.Fatal(err)
			}
		}
	}
}

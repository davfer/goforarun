package observability

import (
	"fmt"
	"github.com/davfer/goforarun/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"io"
	"os"
)

var logger *Logger

type Logger struct {
	baseLogger       *logrus.Logger
	baseFields       logrus.Fields
	filteredChannels map[string]*logrus.Level
}

func InitLogger(cfg config.LoggingConfig, fields logrus.Fields) error {
	logger = &Logger{
		baseLogger:       logrus.New(),
		filteredChannels: make(map[string]*logrus.Level),
		baseFields:       fields,
	}

	filteredChannels, err := logger.parseFilteredChannels(cfg.FilteredChannels)
	if err != nil {
		return errors.Wrap(err, "couldn't set logger filtered channels")
	}
	logger.filteredChannels = filteredChannels

	level, err := logger.parseLevel(cfg.Level)
	if err != nil {
		return errors.Wrap(err, "couldn't set logger level")
	}
	logger.baseLogger.SetLevel(level)

	formatter, err := logger.parseFormatter(cfg.Format)
	if err != nil {
		return errors.Wrap(err, "couldn't set logger formatter")
	}
	logger.baseLogger.SetFormatter(formatter)

	out, err := logger.parseOutput(cfg.Output)
	if err != nil {
		return errors.Wrap(err, "couldn't set logger output")
	}
	logger.baseLogger.SetOutput(out)

	return nil
}

func (l *Logger) GetEntry(channel string) *logrus.Entry {
	if level, ok := l.filteredChannels[channel]; ok {
		if level == nil {
			return l.GetNilEntry()
		}

		return l.GetScopedEntry(channel, level)
	}

	return l.baseLogger.WithFields(l.baseFields).WithField("channel", channel)
}

func (l *Logger) GetNilEntry() *logrus.Entry {
	nilLogger := *logrus.New()
	nilLogger.SetOutput(io.Discard)

	return nilLogger.WithFields(l.baseFields)
}

func (l *Logger) GetScopedEntry(channel string, level *logrus.Level) *logrus.Entry {
	scopedLogger := *logrus.New()
	scopedLogger.SetOutput(l.baseLogger.Out)
	scopedLogger.SetFormatter(l.baseLogger.Formatter)
	scopedLogger.SetLevel(*level)
	return scopedLogger.WithFields(l.baseFields).WithField("channel", channel)
}

func (l *Logger) parseLevel(levelName string) (logrus.Level, error) {
	switch levelName {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warn":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	case "panic":
		return logrus.PanicLevel, nil
	default:
		return logrus.InfoLevel, errors.New(fmt.Sprintf("invalid log level: %s", levelName))
	}
}

func (l *Logger) parseFormatter(formatterName string) (logrus.Formatter, error) {
	switch formatterName {
	case "text":
		return new(logrus.TextFormatter), nil
	case "json":
		return new(logrus.JSONFormatter), nil
	default:
		return nil, errors.New("invalid log formatter")
	}
}

func (l *Logger) parseOutput(outputName string) (io.Writer, error) {
	switch outputName {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		return os.Create("logs.txt")
	default:
		return nil, errors.New("invalid log output")
	}
}

func (l *Logger) parseFilteredChannels(filteredChannels map[string]string) (map[string]*logrus.Level, error) {
	channels := make(map[string]*logrus.Level, len(filteredChannels))
	for channel, levelName := range filteredChannels {
		level, err := l.parseLevel(levelName)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't parse filtered channel level")
		}
		channels[channel] = &level
	}

	return channels, nil
}

func NewLogger(channel string) *logrus.Entry {
	return logger.GetEntry(channel)
}

func NilLogger() *logrus.Entry {
	return logger.GetNilEntry()
}

func getLogrusHook() logrus.Hook {
	return otellogrus.NewHook(
		otellogrus.WithLevels(
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		),
	)
}

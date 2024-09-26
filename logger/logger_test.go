package logger_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thejerf/slogassert"
	"log/slog"
	"strconv"
	"testing"

	"github.com/davfer/goforarun/logger"
)

func TestAttrErr(t *testing.T) {
	err := fmt.Errorf("test error")
	assert.Equal(t, slog.Any("error", err), logger.AttrErr(err))
}

type testCase struct {
	channel string
	level   slog.Leveler
	called  bool
}

func (tt *testCase) name() string {
	return fmt.Sprintf("ch:%s l:%s called:%s", tt.channel, tt.level, strconv.FormatBool(tt.called))
}

func TestChanneledHandler(t *testing.T) {
	handler := slogassert.New(t, slog.LevelDebug, nil)

	channeled := logger.NewChanneledHandler(handler, map[string]slog.Leveler{
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
	})
	slog.SetDefault(slog.New(channeled))

	testCases := []testCase{
		{"error", slog.LevelError, true},
		{"error", slog.LevelWarn, false},
		{"error", slog.LevelInfo, false},
		{"error", slog.LevelDebug, false},

		{"warn", slog.LevelError, true},
		{"warn", slog.LevelWarn, true},
		{"warn", slog.LevelInfo, false},
		{"warn", slog.LevelDebug, false},

		{"info", slog.LevelError, true},
		{"info", slog.LevelWarn, true},
		{"info", slog.LevelInfo, true},
		{"info", slog.LevelDebug, false},

		{"debug", slog.LevelError, true},
		{"debug", slog.LevelWarn, true},
		{"debug", slog.LevelInfo, true},
		{"debug", slog.LevelDebug, true},

		{"unregistered", slog.LevelError, true},
		{"unregistered", slog.LevelWarn, true},
		{"unregistered", slog.LevelInfo, true},
		{"unregistered", slog.LevelDebug, true},
	}

	for _, tt := range testCases {
		switch tt.level {
		case slog.LevelDebug:
			logger.Get(tt.channel).Debug(tt.name())
		case slog.LevelInfo:
			logger.Get(tt.channel).Info(tt.name())
		case slog.LevelWarn:
			logger.Get(tt.channel).Warn(tt.name())
		case slog.LevelError:
			logger.Get(tt.channel).Error(tt.name())
		}
	}

	var msgs []string
	for _, v := range handler.Unasserted() {
		msgs = append(msgs, v.Message)
	}

	for _, tt := range testCases {
		t.Run(tt.name(), func(t *testing.T) {
			if tt.called {
				assert.Contains(t, msgs, tt.name())
			} else {
				assert.NotContains(t, msgs, tt.name())
			}
		})
	}
}

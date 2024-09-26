package logger

import (
	"context"
	"log/slog"
)

func Get(channel string, attrs ...any) *slog.Logger {
	return slog.Default().With(slog.String("channel", channel)).With(attrs...)
}

func AttrErr(err error) slog.Attr {
	return slog.Any("error", err)
}

var _ slog.Handler = (*ChanneledHandler)(nil)

type ChanneledHandler struct {
	parent   *ChanneledHandler
	wrap     slog.Handler
	channel  string
	channels map[string]slog.Leveler
}

func NewChanneledHandler(h slog.Handler, channels map[string]slog.Leveler) slog.Handler {
	return &ChanneledHandler{
		wrap:     h,
		channels: channels,
	}
}

func (c *ChanneledHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if l, ok := c.channels[c.channel]; ok {
		if level < l.Level() {
			return false
		}
	}
	return c.wrap.Enabled(ctx, level)
}

func (c *ChanneledHandler) Handle(ctx context.Context, record slog.Record) error {
	return c.wrap.Handle(ctx, record)
}

func (c *ChanneledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var channel string
	for _, attr := range attrs {
		if attr.Key == "channel" {
			channel = attr.Value.String()
		}
	}
	return &ChanneledHandler{
		parent:   c,
		wrap:     c.wrap.WithAttrs(attrs),
		channel:  channel,
		channels: c.channels,
	}
}

func (c *ChanneledHandler) WithGroup(name string) slog.Handler {
	return &ChanneledHandler{
		parent:   c,
		wrap:     c.wrap.WithGroup(name),
		channel:  c.channel,
		channels: c.channels,
	}
}

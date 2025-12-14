// Package logger provides a global logger for tmpltype using slog.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// EnvLogLevel is the environment variable name for log level.
const EnvLogLevel = "TMPLTYPE_LOG_LEVEL"

var (
	globalLogger *slog.Logger
	globalLevel  = new(slog.LevelVar) // デフォルトは Info
)

func init() {
	initLogger()
}

// initLogger initializes the global logger from environment variables.
func initLogger() {
	// 環境変数からログレベルを設定
	if v := os.Getenv(EnvLogLevel); v != "" {
		switch strings.ToLower(v) {
		case "debug":
			globalLevel.Set(slog.LevelDebug)
		case "info":
			globalLevel.Set(slog.LevelInfo)
		}
	}

	// カスタムハンドラーで [scan:category] フォーマットを実現
	handler := &customHandler{
		out:   os.Stdout,
		level: globalLevel,
	}
	globalLogger = slog.New(handler)
}

// Debug logs a debug message with the given category and attributes.
// Output format: [scan:category] key=value key=value
func Debug(category string, args ...any) {
	attrs := []slog.Attr{slog.String("_category", category)}
	attrs = append(attrs, attrsFromArgs(args...)...)
	globalLogger.LogAttrs(context.Background(), slog.LevelDebug, "", attrs...)
}

// Info logs an info message.
func Info(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// attrsFromArgs converts variadic key-value pairs to slog.Attr slice.
func attrsFromArgs(args ...any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, _ := args[i].(string)
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	return attrs
}

// customHandler implements slog.Handler with custom formatting.
type customHandler struct {
	out   io.Writer
	level slog.Leveler
	attrs []slog.Attr
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	// カテゴリを取得
	var category string
	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "_category" {
			category = a.Value.String()
		} else {
			attrs = append(attrs, a)
		}
		return true
	})

	// フォーマット: [scan:category] key=value key=value
	var buf strings.Builder
	buf.WriteString("[scan:")
	buf.WriteString(category)
	buf.WriteString("]")

	if len(attrs) > 0 {
		buf.WriteString(" ")
		for i, attr := range attrs {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(attr.Key)
			buf.WriteString("=")
			buf.WriteString(fmt.Sprint(attr.Value.Any()))
		}
	}

	buf.WriteString("\n")
	_, err := h.out.Write([]byte(buf.String()))
	return err
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customHandler{
		out:   h.out,
		level: h.level,
		attrs: append(h.attrs, attrs...),
	}
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	// グループはサポートしない
	return h
}

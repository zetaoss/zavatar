// internal/zlog/zlog.go
package zlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

type ctxKeyLogger struct{}

func Ctx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

type MyHandler struct {
	out   *log.Logger
	level slog.Leveler
	attrs []string
}

func NewMyHandler(w io.Writer, level slog.Leveler) *MyHandler {
	return &MyHandler{
		out:   log.New(w, "", log.LstdFlags),
		level: level,
	}
}

func (h *MyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *MyHandler) Handle(_ context.Context, r slog.Record) error {
	sb := strings.Builder{}
	sb.WriteString(r.Level.String())
	sb.WriteString(" ")
	sb.WriteString(r.Message)

	if len(h.attrs) > 0 {
		sb.WriteString(" ")
		sb.WriteString(strings.Join(h.attrs, " "))
	}

	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value.Any()))
		return true
	})

	h.out.Println(sb.String())
	return nil
}

func (h *MyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]string, len(h.attrs))
	copy(newAttrs, h.attrs)

	for _, a := range attrs {
		newAttrs = append(newAttrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
	}

	return &MyHandler{
		out:   h.out,
		level: h.level,
		attrs: newAttrs,
	}
}

func (h *MyHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init() {
	envLevel := os.Getenv("LOG_LEVEL")
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)

	if envLevel != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(envLevel)); err == nil {
			logLevel.Set(level)
		} else {
			log.Printf("WARN: Invalid LOG_LEVEL value '%s', defaulting to Info\n", envLevel)
		}
	}

	handler := NewMyHandler(os.Stdout, logLevel)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Debug("Logger initialized. Debug logging is enabled.")
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := middleware.GetReqID(r.Context())

		reqLogger := slog.Default().With(
			"rid", rid,
			"method", r.Method,
			"path", r.URL.Path,
		)

		ctx := context.WithValue(r.Context(), ctxKeyLogger{}, reqLogger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

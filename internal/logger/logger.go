// Package logger предоставляет инициализацию структурированного логера и HTTP-middleware для логирования запросов.
package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Log — глобальный экземпляр логера. По умолчанию Nop, инициализируется через Initialize.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует глобальный логер с указанным уровнем логирования (debug, info, warn, error).
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

type (
	responseData struct {
		status int
		size   int
	}

	responseWriter struct {
		http.ResponseWriter
		data *responseData
	}
)

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.data.status = status
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.data.size += size
	return size, err
}

// WithLogger — middleware, логирующий входящие HTTP-запросы (метод, URI, статус, длительность, размер ответа).
func WithLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		rd := &responseData{
			status: 0,
			size:   0,
		}

		rw := responseWriter{
			ResponseWriter: w,
			data:           rd,
		}

		h.ServeHTTP(&rw, r)

		duration := time.Since(start)

		Log.Info("incoming HTTP request",
			zap.String("method", method),
			zap.String("uri", uri),
			zap.Int("status", rd.status),
			zap.Duration("duration", duration),
			zap.Int("size", rd.size),
		)
	})
}

// Err создаёт zap-поле для логирования ошибки.
func Err(err error) zap.Field {
	return zap.Error(err)
}

// Field создаёт zap-поле с произвольным именем и значением.
func Field(name string, field any) zap.Field {
	return zap.Any(name, field)
}

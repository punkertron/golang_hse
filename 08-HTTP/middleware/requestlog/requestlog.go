//go:build !solution

package requestlog

import (
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	defer l.Sync()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var m httpsnoop.Metrics

			id, _ := uuid.NewV4()

			l.Info("request started",
				zap.String("request_id", id.String()),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)

			defer func() {
				var msg string

				rec := recover()
				if rec != nil {
					msg = "request panicked"
				} else {
					msg = "request finished"
				}

				l.Info(msg,
					zap.String("request_id", id.String()),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("duration", m.Duration.String()),
					zap.Int("status_code", m.Code),
				)

				if rec != nil {
					panic(rec)
				}
			}()

			m = httpsnoop.CaptureMetrics(next, w, r)
			// next.ServeHTTP(w, r)
		})
	}
}

package appobs

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// LogMiddleware logs each request as a JSON line, including the trace id when a
// span is active (so logs correlate with traces).
func LogMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		attrs := []any{
			"method", r.Method,
			"route", routeOf(r),
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		}
		if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
			attrs = append(attrs, "trace_id", sc.TraceID().String())
		}
		logger.InfoContext(r.Context(), "http request", attrs...)
	})
}

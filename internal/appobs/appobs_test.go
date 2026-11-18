package appobs

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestNewServerTimeouts(t *testing.T) {
	srv := NewServer(":8080", http.NewServeMux())
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q", srv.Addr)
	}
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Errorf("timeouts not set: %+v", srv)
	}
}

func TestProblem(t *testing.T) {
	rr := httptest.NewRecorder()
	Problem(rr, http.StatusBadRequest, "invalid_body", "bad json")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("content-type = %q", ct)
	}
	var p ProblemDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Status != 400 || p.Code != "invalid_body" || p.Detail != "bad json" {
		t.Errorf("problem = %+v", p)
	}
}

func TestLogMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := LogMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/orders", nil))

	out := buf.String()
	if !strings.Contains(out, `"msg":"http request"`) || !strings.Contains(out, `"status":201`) {
		t.Errorf("log line = %q", out)
	}
}

func TestLogMiddlewareTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	req := httptest.NewRequest(http.MethodGet, "/orders", nil).WithContext(ctx)

	LogMiddleware(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), `"trace_id":"01000000000000000000000000000000"`) {
		t.Errorf("log line missing trace_id: %q", buf.String())
	}
}

func TestInitTracingShutdown(t *testing.T) {
	shutdown, err := InitTracing(context.Background(), "test-svc")
	if err != nil {
		t.Fatalf("InitTracing: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

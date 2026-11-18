// Package appobs holds the observability plumbing shared by the demo services:
// structured logging, RED metrics, OpenTelemetry tracing and an HTTP server with
// production timeouts.
package appobs

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// NewLogger returns a JSON slog logger writing to stdout (collected by Alloy).
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// NewServer returns an http.Server with sane production timeouts.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

// ProblemDetail is an RFC 9457 problem+json body.
type ProblemDetail struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// Problem writes an RFC 9457 application/problem+json response.
func Problem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ProblemDetail{
		Title:  http.StatusText(status),
		Status: status,
		Code:   code,
		Detail: detail,
	})
}

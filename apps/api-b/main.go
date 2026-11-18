// Command api-b is the payments service of the demo workload. It is called by
// api-a to demonstrate distributed tracing, and exposes RED metrics and slog
// logs for the observability stack.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/fabianoflorentino/srectl/internal/appobs"
)

func main() {
	logger := appobs.NewLogger()
	if err := run(logger); err != nil {
		logger.Error("api-b exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := appobs.InitTracing(ctx, "api-b")
	if err != nil {
		return err
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	metrics := appobs.NewMetrics("api_b")
	handler := otelhttp.NewHandler(
		appobs.LogMiddleware(logger, metrics.Middleware(routes(metrics))), "api-b")
	srv := appobs.NewServer(":"+port(), handler)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api-b listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

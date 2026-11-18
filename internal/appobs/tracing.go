package appobs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracing configures the global TracerProvider to export OTLP/HTTP traces
// to OTEL_EXPORTER_OTLP_ENDPOINT (default tempo.monitoring.svc:4318). It returns
// a shutdown function to flush spans on exit.
func InitTracing(ctx context.Context, service string) (func(context.Context) error, error) {
	opts, err := exporterOptions(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		return nil, err
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes("",
		attribute.String("service.name", service)))
	if err != nil {
		return nil, fmt.Errorf("building resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return func(ctx context.Context) error {
		c, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(c)
	}, nil
}

// exporterOptions turns an endpoint (host:port or URL) into OTLP/HTTP options.
func exporterOptions(endpoint string) ([]otlptracehttp.Option, error) {
	if endpoint == "" {
		endpoint = "tempo.monitoring.svc:4318"
	}
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(u.Host)}
		if u.Scheme != "https" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return opts, nil
	}
	return []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	}, nil
}

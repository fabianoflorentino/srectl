package appobs

import (
	"context"
	"testing"
)

func TestExporterOptions(t *testing.T) {
	for _, endpoint := range []string{"", "tempo:4318", "http://tempo:4318", "https://tempo:4318"} {
		opts, err := exporterOptions(endpoint)
		if err != nil {
			t.Errorf("exporterOptions(%q): %v", endpoint, err)
		}
		if len(opts) == 0 {
			t.Errorf("exporterOptions(%q) returned no options", endpoint)
		}
	}
}

func TestInitTracingBadService(t *testing.T) {
	// A non-empty service name always builds a resource; this mainly exercises
	// the happy path and the propagator setup.
	shutdown, err := InitTracing(context.Background(), "api-a")
	if err != nil {
		t.Fatalf("InitTracing: %v", err)
	}
	_ = shutdown(context.Background())
}

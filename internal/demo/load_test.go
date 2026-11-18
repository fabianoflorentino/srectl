package demo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadRequiresClient(t *testing.T) {
	if _, err := Load(context.Background(), LoadConfig{URL: "http://x", Duration: time.Millisecond}); err == nil {
		t.Error("want error when client is nil")
	}
}

func TestLoadGeneratesTraffic(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	res, err := Load(context.Background(), LoadConfig{
		URL:      srv.URL,
		Duration: 300 * time.Millisecond,
		RPS:      50,
		Client:   srv.Client(),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if res.Requests == 0 {
		t.Error("no requests generated")
	}
	if hits.Load() == 0 {
		t.Error("server saw no requests")
	}
}

func TestLoadClampsConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Zero RPS/duration must not panic (clamped to 1 rps / 1s), and an absurd
	// RPS is clamped to 1000.
	if _, err := Load(context.Background(), LoadConfig{
		URL: srv.URL, RPS: 0, Duration: 10 * time.Millisecond, Client: srv.Client(),
	}); err != nil {
		t.Fatalf("Load(zero): %v", err)
	}
	if _, err := Load(context.Background(), LoadConfig{
		URL: srv.URL, RPS: 100000, Duration: 10 * time.Millisecond, Client: srv.Client(),
	}); err != nil {
		t.Fatalf("Load(huge): %v", err)
	}
}

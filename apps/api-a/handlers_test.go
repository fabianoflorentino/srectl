package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fabianoflorentino/srectl/internal/appobs"
)

func TestHealth(t *testing.T) {
	for _, path := range []string{"/health", "/ready"} {
		rr := httptest.NewRecorder()
		health(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK || rr.Body.String() != "ok" {
			t.Errorf("%s = %d %q", path, rr.Code, rr.Body.String())
		}
	}
}

func TestOrdersSuccess(t *testing.T) {
	b := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payResult{Status: "approved", TransactionID: "txn-o1"})
	}))
	defer b.Close()

	rr := httptest.NewRecorder()
	orders(b.Client(), b.URL)(rr, httptest.NewRequest(http.MethodGet, "/orders", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var resp orderResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "confirmed" || resp.TransactionID != "txn-o1" || resp.OrderID == "" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestOrdersUpstreamFailures(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"upstream 500", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }},
		{"upstream invalid body", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := httptest.NewServer(tc.handler)
			defer b.Close()

			rr := httptest.NewRecorder()
			orders(b.Client(), b.URL)(rr, httptest.NewRequest(http.MethodGet, "/orders", nil))
			if rr.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want 502", rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("content-type = %q", ct)
			}
		})
	}
}

func TestOrdersUpstreamDown(t *testing.T) {
	b := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := b.URL
	b.Close()

	rr := httptest.NewRecorder()
	orders(b.Client(), url)(rr, httptest.NewRequest(http.MethodGet, "/orders", nil))
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rr.Code)
	}
}

func TestSlowHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	slowHandler(5*time.Millisecond)(rr, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if rr.Code != http.StatusOK || rr.Body.String() != "slow ok" {
		t.Errorf("slow = %d %q", rr.Code, rr.Body.String())
	}
}

func TestSlowHandlerCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil).WithContext(ctx)
	slowHandler(time.Second)(httptest.NewRecorder(), req) // must return immediately
}

func TestFail(t *testing.T) {
	rr := httptest.NewRecorder()
	fail(rr, httptest.NewRequest(http.MethodGet, "/error", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "forced_error") {
		t.Errorf("body = %q", rr.Body.String())
	}
}

func TestRoutesExposeMetrics(t *testing.T) {
	m := appobs.NewMetrics("api_a")
	mux := routes(m, http.DefaultClient, "http://unused")
	handler := m.Middleware(mux)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(rr.Body.String(), "api_a_http_requests_total") {
		t.Errorf("metrics missing counter")
	}
}

func TestPortAndAPIbURL(t *testing.T) {
	t.Setenv("PORT", "9001")
	t.Setenv("API_B_URL", "http://x")
	if port() != "9001" || apiBURL() != "http://x" {
		t.Errorf("port=%q apiBURL=%q", port(), apiBURL())
	}
	t.Setenv("PORT", "")
	t.Setenv("API_B_URL", "")
	if port() != "8080" || apiBURL() != "http://api-b.demo.svc.cluster.local" {
		t.Errorf("defaults: port=%q apiBURL=%q", port(), apiBURL())
	}
}

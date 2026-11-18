package appobs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsMiddlewareRecords(t *testing.T) {
	m := NewMetrics("api_a")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := m.Middleware(mux)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/orders", nil))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))

	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rr.Body.String()

	for _, want := range []string{
		"api_a_http_requests_total",
		`route="GET /orders"`,
		`status="201"`,
		`route="unmatched"`,
		"api_a_http_request_duration_seconds_bucket",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics missing %q", want)
		}
	}
}

func TestRouteOf(t *testing.T) {
	if got := routeOf(&http.Request{Pattern: "GET /orders"}); got != "GET /orders" {
		t.Errorf("routeOf = %q", got)
	}
	if got := routeOf(&http.Request{}); got != "unmatched" {
		t.Errorf("routeOf empty = %q", got)
	}
}

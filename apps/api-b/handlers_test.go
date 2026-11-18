package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fabianoflorentino/srectl/internal/appobs"
)

func TestHealth(t *testing.T) {
	for _, path := range []string{"/health", "/ready"} {
		rr := httptest.NewRecorder()
		health(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Errorf("%s status = %d", path, rr.Code)
		}
		if rr.Body.String() != "ok" {
			t.Errorf("%s body = %q", path, rr.Body.String())
		}
	}
}

func TestPay(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"valid", `{"order_id":"o1","amount":42}`, http.StatusOK},
		{"invalid json", `{`, http.StatusBadRequest},
		{"missing order id", `{"amount":1}`, http.StatusBadRequest},
		{"unknown field", `{"order_id":"o1","extra":true}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			pay(rr, httptest.NewRequest(http.MethodPost, "/pay", strings.NewReader(tc.body)))
			if rr.Code != tc.status {
				t.Fatalf("status = %d, want %d", rr.Code, tc.status)
			}
			if tc.status != http.StatusOK {
				if ct := rr.Header().Get("Content-Type"); ct != "application/problem+json" {
					t.Errorf("content-type = %q", ct)
				}
				return
			}
			var resp payResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Status != "approved" || resp.TransactionID != "txn-o1" {
				t.Errorf("resp = %+v", resp)
			}
		})
	}
}

func TestRoutesExposeMetrics(t *testing.T) {
	m := appobs.NewMetrics("api_b")
	mux := routes(m)
	handler := m.Middleware(mux)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("metrics status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "api_b_http_requests_total") {
		t.Errorf("metrics output missing counter")
	}
}

func TestPort(t *testing.T) {
	t.Setenv("PORT", "9000")
	if got := port(); got != "9000" {
		t.Errorf("port = %q, want 9000", got)
	}
	t.Setenv("PORT", "")
	if got := port(); got != "8080" {
		t.Errorf("default port = %q, want 8080", got)
	}
}

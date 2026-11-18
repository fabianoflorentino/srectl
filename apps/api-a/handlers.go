package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/fabianoflorentino/srectl/internal/appobs"
)

var orderSeq atomic.Uint64

type payRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

type payResult struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

type orderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

// health is used for both /health and /ready.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// orders checks out an order by calling api-b, propagating the trace context.
func orders(client *http.Client, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := "o-" + strconv.FormatUint(orderSeq.Add(1), 10)
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		payload, err := json.Marshal(payRequest{OrderID: orderID, Amount: 42.0})
		if err != nil {
			appobs.Problem(w, http.StatusInternalServerError, "encode", "failed to encode request")
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/pay", bytes.NewReader(payload))
		if err != nil {
			appobs.Problem(w, http.StatusInternalServerError, "request_build", "failed to build upstream request")
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			appobs.Problem(w, http.StatusBadGateway, "upstream_unavailable", "api-b is unavailable")
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			appobs.Problem(w, http.StatusBadGateway, "upstream_error", "api-b returned an error")
			return
		}
		var result payResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			appobs.Problem(w, http.StatusBadGateway, "upstream_error", "invalid api-b response")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(orderResponse{
			OrderID:       orderID,
			Status:        "confirmed",
			TransactionID: result.TransactionID,
		})
	}
}

// slowHandler sleeps for d (honouring request cancellation) to exercise latency.
func slowHandler(d time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow ok"))
	}
}

// fail always returns a 500 problem+json to drive error-rate signals.
func fail(w http.ResponseWriter, _ *http.Request) {
	appobs.Problem(w, http.StatusInternalServerError, "forced_error", "this endpoint always fails")
}

// routes builds the api-a mux (handler for /metrics included).
func routes(m *appobs.Metrics, client *http.Client, baseURL string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", health)
	mux.HandleFunc("GET /metrics", m.Handler().ServeHTTP)
	mux.HandleFunc("GET /orders", orders(client, baseURL))
	mux.HandleFunc("GET /slow", slowHandler(250*time.Millisecond))
	mux.HandleFunc("GET /error", fail)
	return mux
}

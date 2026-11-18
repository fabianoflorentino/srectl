package main

import (
	"encoding/json"
	"net/http"

	"github.com/fabianoflorentino/srectl/internal/appobs"
)

type payRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

type payResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

// health is used for both /health and /ready.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// pay charges an order and returns a transaction id.
func pay(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req payRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || req.OrderID == "" {
		appobs.Problem(w, http.StatusBadRequest, "invalid_body",
			"order_id is required and the body must be valid JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payResponse{
		Status:        "approved",
		TransactionID: "txn-" + req.OrderID,
	})
}

// routes builds the api-b mux (handler for /metrics included).
func routes(m *appobs.Metrics) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", health)
	mux.HandleFunc("GET /metrics", m.Handler().ServeHTTP)
	mux.HandleFunc("POST /pay", pay)
	return mux
}

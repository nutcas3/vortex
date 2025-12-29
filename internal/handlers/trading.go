package handlers

import (
	"encoding/json"
	"net/http"
)

type TradingHandler struct {
	// Add dependencies here (trading engine, order book, etc.)
}

func NewTradingHandler() *TradingHandler {
	return &TradingHandler{}
}

// Orders handles order management (GET/POST)
func (h *TradingHandler) Orders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get orders
		json.NewEncoder(w).Encode(map[string]string{"message": "GET orders - TODO: Implement"})
	case http.MethodPost:
		// Create order
		json.NewEncoder(w).Encode(map[string]string{"message": "POST order - TODO: Implement"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Order handles single order operations (GET/PUT/DELETE)
func (h *TradingHandler) Order(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get single order
		json.NewEncoder(w).Encode(map[string]string{"message": "GET order - TODO: Implement"})
	case http.MethodPut:
		// Update order
		json.NewEncoder(w).Encode(map[string]string{"message": "PUT order - TODO: Implement"})
	case http.MethodDelete:
		// Cancel order
		json.NewEncoder(w).Encode(map[string]string{"message": "DELETE order - TODO: Implement"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Fills handles trade fill history
func (h *TradingHandler) Fills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET fills - TODO: Implement"})
}

package handlers

import (
	"encoding/json"
	"net/http"
)

type MarketDataHandler struct {
	// Add dependencies here (order book, trade history, etc.)
}

func NewMarketDataHandler() *MarketDataHandler {
	return &MarketDataHandler{}
}

// Ticker24hr handles 24-hour ticker statistics
func (h *MarketDataHandler) Ticker24hr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET 24hr ticker - TODO: Implement"})
}

// OrderBook handles order book depth
func (h *MarketDataHandler) OrderBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET order book depth - TODO: Implement"})
}

// RecentTrades handles recent trade history
func (h *MarketDataHandler) RecentTrades(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET recent trades - TODO: Implement"})
}

// Klines handles candlestick/kline data
func (h *MarketDataHandler) Klines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "GET klines - TODO: Implement"})
}

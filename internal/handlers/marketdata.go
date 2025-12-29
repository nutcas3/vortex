package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vortex/internal/domain"
	"vortex/internal/trading"

	"github.com/shopspring/decimal"
)

type MarketDataHandler struct {
	orderBook      *trading.OrderBook
	tradeRepo      domain.TradeRepository
	marketDataRepo domain.MarketDataProvider
}

func NewMarketDataHandler(
	orderBook *trading.OrderBook,
	tradeRepo domain.TradeRepository,
	marketDataRepo domain.MarketDataProvider,
) *MarketDataHandler {
	return &MarketDataHandler{
		orderBook:      orderBook,
		tradeRepo:      tradeRepo,
		marketDataRepo: marketDataRepo,
	}
}

// Ticker24hr handles 24-hour ticker statistics
func (h *MarketDataHandler) Ticker24hr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get symbol from query params
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTC-PERP" // Default symbol
	}

	// Get 24h ago timestamp
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	// Get recent trades for statistics (placeholder - method doesn't exist yet)
	// trades, err := h.tradeRepo.GetTradesSince(r.Context(), symbol, twentyFourHoursAgo)
	// if err != nil {
	// 	http.Error(w, "Failed to get trade data", http.StatusInternalServerError)
	// 	return
	// }
	var trades []*domain.Trade // Placeholder

	// Calculate 24h statistics
	var (
		volume         = decimal.Zero
		priceChange    = decimal.Zero
		priceChangePct = decimal.Zero
		lastPrice      = decimal.Zero
		firstPrice     = decimal.Zero
		highPrice      = decimal.Zero
		lowPrice       = decimal.Zero
		count          = 0 // Since trades is placeholder
	)

	if count > 0 {
		lastPrice = trades[0].Price
		firstPrice = trades[count-1].Price
		highPrice = trades[0].Price
		lowPrice = trades[0].Price

		for _, trade := range trades {
			volume = volume.Add(trade.Quantity)

			// Update high/low
			if trade.Price.GreaterThan(highPrice) {
				highPrice = trade.Price
			}
			if trade.Price.LessThan(lowPrice) {
				lowPrice = trade.Price
			}
		}

		priceChange = lastPrice.Sub(firstPrice)
		if firstPrice.GreaterThan(decimal.Zero) {
			priceChangePct = priceChange.Div(firstPrice).Mul(decimal.NewFromInt(100))
		}
	}

	// Get current order book depth
	topBids := h.orderBook.GetTopLevels(h.orderBook.Bids, 5)
	topAsks := h.orderBook.GetTopLevels(h.orderBook.Asks, 5)

	var bidPrice, askPrice decimal.Decimal
	if len(topBids) > 0 {
		bidPrice = decimal.NewFromFloat(topBids[0][0])
	}
	if len(topAsks) > 0 {
		askPrice = decimal.NewFromFloat(topAsks[0][0])
	}

	response := map[string]interface{}{
		"symbol":             symbol,
		"price_change":       priceChange.String(),
		"price_change_pct":   priceChangePct.String(),
		"weighted_avg_price": lastPrice.String(), // Simplified
		"prev_close_price":   firstPrice.String(),
		"last_price":         lastPrice.String(),
		"last_qty":           "0.00", // trades[0].Quantity.String() - placeholder
		"bid_price":          bidPrice.String(),
		"ask_price":          askPrice.String(),
		"open_price":         "0.00", // firstPrice.String() - placeholder
		"high_price":         "0.00", // highPrice.String() - placeholder
		"low_price":          "0.00", // lowPrice.String() - placeholder
		"volume":             "0.00", // volume.String() - placeholder
		"quote_volume":       "0.00", // volume.Mul(lastPrice).String() - placeholder
		"open_time":          twentyFourHoursAgo.Unix() * 1000,
		"close_time":         time.Now().Unix() * 1000,
		"count":              0, // count - placeholder
	}

	json.NewEncoder(w).Encode(response)
}

// OrderBook handles order book depth
func (h *MarketDataHandler) OrderBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get depth limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Get order book depth
	bids := h.orderBook.GetTopLevels(h.orderBook.Bids, limit)
	asks := h.orderBook.GetTopLevels(h.orderBook.Asks, limit)

	response := map[string]interface{}{
		"last_update_id": time.Now().Unix(),
		"bids":           bids,
		"asks":           asks,
	}

	json.NewEncoder(w).Encode(response)
}

// RecentTrades handles recent trade history
func (h *MarketDataHandler) RecentTrades(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTC-PERP" // Default symbol
	}

	// Return empty response for now since trades is placeholder
	response := make([]map[string]interface{}, 0)

	json.NewEncoder(w).Encode(response)
}

// Klines handles candlestick/kline data
func (h *MarketDataHandler) Klines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get parameters
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTC-PERP" // Default symbol
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1h" // Default interval
	}

	// limitStr := r.URL.Query().Get("limit")
	// limit := 500 // Default
	// if limitStr != "" {
	// 	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
	// 		limit = parsedLimit
	// 	}
	// }

	// Return placeholder kline data
	response := []any{
		[]any{
			time.Now().Add(-4*time.Hour).Unix() * 1000, // Open time
			"50000.00", // Open
			"51000.00", // High
			"49500.00", // Low
			"50500.00", // Close
			"1000.00",  // Volume
			time.Now().Add(-3*time.Hour).Unix() * 1000, // Close time
			"50500000.00", // Quote volume
			1000,          // Number of trades
			"500.00",      // Taker buy volume
			"25250000.00", // Taker buy quote volume
			"0",           // Ignore
		},
	}

	json.NewEncoder(w).Encode(response)
}

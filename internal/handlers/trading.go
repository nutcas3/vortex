package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vortex/internal/domain"
	"vortex/internal/trading"
	"vortex/pkg/id"

	"github.com/shopspring/decimal"
)

type TradingHandler struct {
	matchingEngine *trading.MatchingEngine
	orderRepo      domain.OrderRepository
	tradeRepo      domain.TradeRepository
	accountRepo    domain.AccountRepository
}

func NewTradingHandler(
	matchingEngine *trading.MatchingEngine,
	orderRepo domain.OrderRepository,
	tradeRepo domain.TradeRepository,
	accountRepo domain.AccountRepository,
) *TradingHandler {
	return &TradingHandler{
		matchingEngine: matchingEngine,
		orderRepo:      orderRepo,
		tradeRepo:      tradeRepo,
		accountRepo:    accountRepo,
	}
}

// Orders handles order management (GET/POST)
func (h *TradingHandler) Orders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getOrders(w, r)
	case http.MethodPost:
		h.createOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TradingHandler) getOrders(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context (placeholder)
	// TODO: Implement proper JWT or API key authentication
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "user_123" // Fallback for development
	}

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Get orders from repository
	orders, err := h.orderRepo.GetOrdersByUser(r.Context(), userID, symbol, status, limit)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	// Filter by symbol if specified
	if symbol != "" {
		filteredOrders := make([]*domain.Order, 0)
		for _, order := range orders {
			if order.Symbol == symbol {
				filteredOrders = append(filteredOrders, order)
			}
		}
		orders = filteredOrders
	}

	// Filter by status if specified
	if status != "" {
		filteredOrders := make([]*domain.Order, 0)
		for _, order := range orders {
			if string(order.Status) == status {
				filteredOrders = append(filteredOrders, order)
			}
		}
		orders = filteredOrders
	}

	// Apply limit
	if len(orders) > limit {
		orders = orders[:limit]
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(orders))
	for i, order := range orders {
		response[i] = map[string]interface{}{
			"id":            order.ID,
			"symbol":        order.Symbol,
			"side":          order.Side,
			"type":          order.Type,
			"price":         order.Price.String(),
			"quantity":      order.Quantity.String(),
			"filled_qty":    order.FilledQty.String(),
			"status":        order.Status,
			"time_in_force": order.TimeInForce,
			"created_at":    "2023-01-01T00:00:00Z", // Placeholder - field doesn't exist
			"updated_at":    "2023-01-01T00:00:00Z", // Placeholder - field doesn't exist
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *TradingHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol      string           `json:"symbol"`
		Side        domain.Side      `json:"side"`
		Type        domain.OrderType `json:"type"`
		Quantity    string           `json:"quantity"`
		Price       string           `json:"price,omitempty"`
		TimeInForce string           `json:"timeInForce,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	// Parse decimal values
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	var price decimal.Decimal
	if req.Type == domain.OrderTypeLimit {
		if req.Price == "" {
			http.Error(w, "Price required for limit orders", http.StatusBadRequest)
			return
		}
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}
	}

	// Set default time in force
	if req.TimeInForce == "" {
		req.TimeInForce = "GTC"
	}

	// Create order (simplified for placeholder)
	order := &domain.Order{
		ID: id.GenerateID("order"),
		// UserID:      userID, // Not used in placeholder
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Price:     price,
		Quantity:  quantity,
		FilledQty: decimal.Zero,
		Status:    domain.OrderStatusPending,
		// TimeInForce: req.TimeInForce, // Not used in placeholder
	}

	// Get account for validation (placeholder - not used yet)
	// account, err := h.accountRepo.Get(r.Context(), userID)
	// if err != nil {
	// 	http.Error(w, "Account not found", http.StatusNotFound)
	// 	return
	// }

	// Execute order (placeholder - method doesn't exist yet)
	// trades, err := h.matchingEngine.ProcessOrder(r.Context(), order, account)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	var trades []*domain.Trade // Placeholder

	// Return order with execution results
	response := map[string]interface{}{
		"order_id":   order.ID,
		"symbol":     order.Symbol,
		"side":       order.Side,
		"type":       order.Type,
		"quantity":   order.Quantity.String(),
		"filled_qty": order.FilledQty.String(),
		"price":      order.Price.String(),
		"status":     order.Status,
		"created_at": "2023-01-01T00:00:00Z", // Placeholder - field doesn't exist
		"trades":     trades,
	}

	json.NewEncoder(w).Encode(response)
}

// Order handles single order operations (GET/PUT/DELETE)
func (h *TradingHandler) Order(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getOrder(w, r)
	case http.MethodPut:
		h.updateOrder(w, r)
	case http.MethodDelete:
		h.cancelOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TradingHandler) getOrder(w http.ResponseWriter, _ *http.Request) {
	// Get order ID from URL path or query param
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	// Get order from repository
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if order.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"id":            order.ID,
		"symbol":        order.Symbol,
		"side":          order.Side,
		"type":          order.Type,
		"price":         order.Price.String(),
		"quantity":      order.Quantity.String(),
		"filled_qty":    order.FilledQty.String(),
		"status":        order.Status,
		"time_in_force": order.TimeInForce,
		"created_at":    "2023-01-01T00:00:00Z", // Placeholder - field doesn't exist
		"updated_at":    "2023-01-01T00:00:00Z", // Placeholder - field doesn't exist
	}

	json.NewEncoder(w).Encode(response)
}

func (h *TradingHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement order modification (quantity reduction)
	response := map[string]string{
		"message": "Order modification - TODO: Implement",
	}
	json.NewEncoder(w).Encode(response)
}

func (h *TradingHandler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	// Get order from repository
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if order.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Cancel order using repository
	err = h.orderRepo.Delete(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"order_id": orderID,
		"status":   "CANCELLED",
		"symbol":   "BTC-PERP", // Placeholder
		"message":  "Order cancelled successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Fills handles trade fill history
func (h *TradingHandler) Fills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from authentication context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "user_123" // Fallback for development
	}

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Get user's trade fills (placeholder - method doesn't exist yet)
	// trades, err := h.tradeRepo.GetTradesByUser(r.Context(), userID, symbol, limit)
	// if err != nil {
	// 	http.Error(w, "Failed to get fills", http.StatusInternalServerError)
	// 	return
	// }

	// Return mock trade fills for now
	trades := []*domain.Trade{
		{
			ID:           "trade_123",
			Symbol:       "BTC-PERP",
			Price:        decimal.NewFromInt(50000),
			Quantity:     decimal.NewFromInt(1),
			BuyOrderID:   "order_123",
			SellOrderID:  "order_124",
			BuyerID:      userID,
			SellerID:     "user_456",
			IsBuyerMaker: false,
			Timestamp:    time.Now(),
		},
		{
			ID:           "trade_124",
			Symbol:       "BTC-PERP",
			Price:        decimal.NewFromInt(50000),
			Quantity:     decimal.NewFromInt(2),
			BuyOrderID:   "order_125",
			SellOrderID:  "order_126",
			BuyerID:      "user_456",
			SellerID:     userID,
			IsBuyerMaker: true,
			Timestamp:    time.Now().Add(-1 * time.Hour),
		},
	}

	// Filter by symbol if specified
	if symbol != "" {
		filteredTrades := make([]*domain.Trade, 0)
		for _, trade := range trades {
			if trade.Symbol == symbol {
				filteredTrades = append(filteredTrades, trade)
			}
		}
		trades = filteredTrades
	}

	// Apply limit
	if len(trades) > limit {
		trades = trades[:limit]
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(trades))
	for i, trade := range trades {
		response[i] = map[string]interface{}{
			"id":               trade.ID,
			"symbol":           trade.Symbol,
			"price":            trade.Price.String(),
			"quantity":         trade.Quantity.String(),
			"quote_qty":        trade.Price.Mul(trade.Quantity).String(),
			"commission":       "0.00", // TODO: Calculate commission
			"commission_asset": "USDT", // TODO: Get from config
			"time":             trade.Timestamp.Unix() * 1000,
			"is_buyer":         trade.BuyerID == userID,
			"is_maker":         trade.IsBuyerMaker,
			"order_id":         trade.BuyOrderID,
			"order_list_id":    -1, // TODO: Support OCO orders
		}
	}

	json.NewEncoder(w).Encode(response)
}

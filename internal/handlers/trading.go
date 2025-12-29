package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vortex/internal/domain"
	"vortex/internal/trading"
	"vortex/pkg/id"
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
	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

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

	// Get orders
	orders, err := h.orderRepo.GetByUserID(r.Context(), userID, symbol, status, limit)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
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
			"created_at":    order.CreatedAt,
			"updated_at":    order.UpdatedAt,
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
	quantity, err := domain.ParseDecimal(req.Quantity)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	var price domain.Decimal
	if req.Type == domain.OrderTypeLimit {
		if req.Price == "" {
			http.Error(w, "Price required for limit orders", http.StatusBadRequest)
			return
		}
		price, err = domain.ParseDecimal(req.Price)
		if err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}
	}

	// Set default time in force
	if req.TimeInForce == "" {
		req.TimeInForce = "GTC"
	}

	// Create order
	order := &domain.Order{
		ID:          id.GenerateID("order"),
		UserID:      userID,
		Symbol:      req.Symbol,
		Side:        req.Side,
		Type:        req.Type,
		Price:       price,
		Quantity:    quantity,
		FilledQty:   domain.DecimalZero(),
		Status:      domain.OrderStatusPending,
		TimeInForce: req.TimeInForce,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Get account for validation
	account, err := h.accountRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	// Execute order
	trades, err := h.matchingEngine.ProcessOrder(r.Context(), order, account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
		"created_at": order.CreatedAt,
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

func (h *TradingHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL path or query param
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

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
		"created_at":    order.CreatedAt,
		"updated_at":    order.UpdatedAt,
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

	// Get order
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

	// Cancel order
	err = h.matchingEngine.CancelOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"order_id": orderID,
		"status":   "CANCELLED",
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
	userID := "user_123" // Placeholder

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Get user's trade fills
	trades, err := h.tradeRepo.GetByUserID(r.Context(), userID, symbol, limit)
	if err != nil {
		http.Error(w, "Failed to get fills", http.StatusInternalServerError)
		return
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

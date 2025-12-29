package handlers

import (
	"encoding/json"
	"net/http"

	"vortex/internal/domain"

	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	accountRepo  domain.AccountRepository
	positionRepo domain.PositionRepository
}

func NewAccountHandler(accountRepo domain.AccountRepository, positionRepo domain.PositionRepository) *AccountHandler {
	return &AccountHandler{
		accountRepo:  accountRepo,
		positionRepo: positionRepo,
	}
}

// Account handles account information
func (h *AccountHandler) Account(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	account, err := h.accountRepo.Get(r.Context(), userID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	// Return account information (excluding sensitive data)
	response := map[string]interface{}{
		"user_id":          account.UserID,
		"balance":          account.Balance.String(),
		"available_margin": account.AvailableMargin.String(),
		"total_equity":     account.TotalEquity.String(),
		"locked_balance":   account.LockedBalance.String(),
		// "created_at":       account.CreatedAt, // Field doesn't exist in domain model
		// "updated_at":       account.UpdatedAt, // Field doesn't exist in domain model
	}

	json.NewEncoder(w).Encode(response)
}

// Positions handles position management
func (h *AccountHandler) Positions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getPositions(w, r)
	case http.MethodPost:
		h.createPosition(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AccountHandler) getPositions(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	positions, err := h.positionRepo.FindByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get positions", http.StatusInternalServerError)
		return
	}

	// Convert positions to response format
	response := make([]map[string]interface{}, len(positions))
	for i, pos := range positions {
		response[i] = map[string]interface{}{
			"id":                pos.ID,
			"symbol":            pos.Symbol,
			"side":              pos.Side,
			"size":              pos.Size.String(),
			"entry_price":       pos.EntryPrice.String(),
			"mark_price":        pos.MarkPrice.String(),
			"liquidation_price": pos.LiquidationPrice.String(),
			"margin":            pos.Margin.String(),
			"unrealized_pnl":    pos.UnrealizedPnL.String(),
			"realized_pnl":      pos.RealizedPnL.String(),
			"status":            pos.Status,
			"leverage":          pos.Leverage.String(),
			"created_at":        pos.CreatedAt,
			"updated_at":        pos.UpdatedAt,
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (h *AccountHandler) createPosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol   string      `json:"symbol"`
		Side     domain.Side `json:"side"`
		Size     string      `json:"size"`
		Price    string      `json:"price"`
		Leverage string      `json:"leverage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from authentication context
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "user_123" // Fallback for development
	}

	// Parse decimal values
	size, err := decimal.NewFromString(req.Size)
	if err != nil {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		return
	}

	var price decimal.Decimal
	if req.Price != "" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			http.Error(w, "Invalid price", http.StatusBadRequest)
			return
		}
	}

	var leverage decimal.Decimal
	if req.Leverage != "" {
		leverage, err = decimal.NewFromString(req.Leverage)
		if err != nil {
			http.Error(w, "Invalid leverage", http.StatusBadRequest)
			return
		}
	} else {
		leverage = decimal.NewFromInt(10) // Default 10x leverage
	}

	// Calculate margin requirement
	margin := size.Mul(price).Div(leverage)

	// Create position
	position := &domain.Position{
		ID:               "pos_" + userID + "_" + req.Symbol,
		UserID:           userID,
		Symbol:           req.Symbol,
		Side:             req.Side,
		Size:             size,
		EntryPrice:       price,
		MarkPrice:        price,        // Initially same as entry price
		LiquidationPrice: decimal.Zero, // Will be calculated
		Margin:           margin,
		UnrealizedPnL:    decimal.Zero,
		RealizedPnL:      decimal.Zero,
		Status:           domain.PositionStatusOpen,
		Leverage:         leverage,
	}

	// Save position to repository
	err = h.positionRepo.Save(r.Context(), position)
	if err != nil {
		http.Error(w, "Failed to create position", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Position created successfully",
		"position": map[string]interface{}{
			"id":                position.ID,
			"symbol":            position.Symbol,
			"side":              position.Side,
			"size":              position.Size.String(),
			"entry_price":       position.EntryPrice.String(),
			"mark_price":        position.MarkPrice.String(),
			"liquidation_price": position.LiquidationPrice.String(),
			"margin":            position.Margin.String(),
			"unrealized_pnl":    position.UnrealizedPnL.String(),
			"realized_pnl":      position.RealizedPnL.String(),
			"status":            position.Status,
			"leverage":          position.Leverage.String(),
		},
	}

	json.NewEncoder(w).Encode(response)
}

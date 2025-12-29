package handlers

import (
	"encoding/json"
	"net/http"

	"vortex/internal/domain"
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

	account, err := h.accountRepo.GetByUserID(r.Context(), userID)
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
		"created_at":       account.CreatedAt,
		"updated_at":       account.UpdatedAt,
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

	positions, err := h.positionRepo.GetByUserID(r.Context(), userID)
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

	// TODO: Get user ID from authentication context
	userID := "user_123" // Placeholder

	// TODO: Parse decimal values and create position
	// This would typically be done through order execution rather than direct position creation

	response := map[string]string{
		"message": "Position creation through orders - TODO: Implement order execution first",
	}

	json.NewEncoder(w).Encode(response)
}

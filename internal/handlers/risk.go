package handlers

import (
	"encoding/json"
	"net/http"

	"vortex/internal/domain"
	"vortex/internal/risk"

	"github.com/shopspring/decimal"
)

type RiskHandler struct {
	riskEngine    domain.RiskEngine
	positionRepo  domain.PositionRepository
	accountRepo   domain.AccountRepository
	markPriceCalc *risk.MarkPriceCalculator
}

func NewRiskHandler(
	riskEngine domain.RiskEngine,
	positionRepo domain.PositionRepository,
	accountRepo domain.AccountRepository,
	markPriceCalc *risk.MarkPriceCalculator,
) *RiskHandler {
	return &RiskHandler{
		riskEngine:    riskEngine,
		positionRepo:  positionRepo,
		accountRepo:   accountRepo,
		markPriceCalc: markPriceCalc,
	}
}

// Margin handles margin calculations
func (h *RiskHandler) Margin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	leverageStr := r.URL.Query().Get("leverage")
	side := r.URL.Query().Get("side")
	quantityStr := r.URL.Query().Get("quantity")
	priceStr := r.URL.Query().Get("price")

	if symbol == "" || side == "" || quantityStr == "" || priceStr == "" {
		http.Error(w, "Missing required parameters: symbol, side, quantity, price", http.StatusBadRequest)
		return
	}

	// Parse parameters
	var orderSide domain.Side
	if side == "BUY" {
		orderSide = domain.SideBuy
	} else if side == "SELL" {
		orderSide = domain.SideSell
	} else {
		http.Error(w, "Invalid side. Must be BUY or SELL", http.StatusBadRequest)
		return
	}

	quantity, err := decimal.NewFromString(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	var leverage decimal.Decimal
	if leverageStr != "" {
		leverage, err = decimal.NewFromString(leverageStr)
		if err != nil {
			http.Error(w, "Invalid leverage", http.StatusBadRequest)
			return
		}
	} else {
		leverage = decimal.NewFromInt(20) // Default 20x leverage
	}

	// Create sample position for calculation
	position := &domain.Position{
		Symbol:     symbol,
		Side:       orderSide,
		Size:       quantity,
		EntryPrice: price,
		Leverage:   leverage,
		Margin:     decimal.Zero, // Will be calculated
	}

	// Calculate required margin
	requiredMargin := h.riskEngine.CalculateMargin(position)

	// Calculate liquidation price
	liquidationPrice := h.riskEngine.CalculateLiquidationPrice(position)

	// Calculate initial margin requirement
	initialMargin := quantity.Mul(price).Div(leverage)

	// Calculate maintenance margin (typically 50% of initial)
	maintenanceMargin := initialMargin.Mul(decimal.NewFromFloat(0.5))

	response := map[string]interface{}{
		"symbol":             symbol,
		"side":               side,
		"quantity":           quantity.String(),
		"price":              price.String(),
		"leverage":           leverage.String(),
		"initial_margin":     initialMargin.String(),
		"maintenance_margin": maintenanceMargin.String(),
		"required_margin":    requiredMargin.String(),
		"liquidation_price":  liquidationPrice.String(),
		"margin_ratio":       requiredMargin.Div(initialMargin.Mul(decimal.NewFromInt(100))).String() + "%",
	}

	json.NewEncoder(w).Encode(response)
}

// Liquidation handles liquidation price calculations
func (h *RiskHandler) Liquidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	symbol := r.URL.Query().Get("symbol")
	userID := r.URL.Query().Get("userId")

	if symbol == "" {
		http.Error(w, "Symbol parameter required", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context if not provided
	if userID == "" {
		userID = "user_123" // Placeholder
	}

	// Get user's position for the symbol
	positions, err := h.positionRepo.FindByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get positions", http.StatusInternalServerError)
		return
	}

	// Find position for the specified symbol
	var targetPosition *domain.Position
	for _, pos := range positions {
		if pos.Symbol == symbol && pos.Status == domain.PositionStatusOpen {
			targetPosition = pos
			break
		}
	}

	if targetPosition == nil {
		http.Error(w, "No open position found for symbol", http.StatusNotFound)
		return
	}

	// Get current mark price
	markPrice, err := h.markPriceCalc.CalculateMarkPrice(r.Context(), symbol, targetPosition.MarkPrice)
	if err != nil {
		http.Error(w, "Failed to calculate mark price", http.StatusInternalServerError)
		return
	}

	// Check if position is at risk of liquidation
	isAtRisk := h.riskEngine.CheckLiquidation(targetPosition, markPrice)

	// Calculate distance to liquidation
	var distanceToLiquidation decimal.Decimal
	if targetPosition.Side == domain.SideBuy {
		// Long: Distance = Mark Price - Liquidation Price
		distanceToLiquidation = markPrice.Sub(targetPosition.LiquidationPrice)
	} else {
		// Short: Distance = Liquidation Price - Mark Price
		distanceToLiquidation = targetPosition.LiquidationPrice.Sub(markPrice)
	}

	// Calculate percentage distance
	var pctDistance decimal.Decimal
	if targetPosition.Side == domain.SideBuy {
		pctDistance = distanceToLiquidation.Div(markPrice).Mul(decimal.NewFromInt(100))
	} else {
		pctDistance = distanceToLiquidation.Div(markPrice).Mul(decimal.NewFromInt(100))
	}

	response := map[string]interface{}{
		"symbol":              symbol,
		"user_id":             userID,
		"side":                targetPosition.Side,
		"size":                targetPosition.Size.String(),
		"entry_price":         targetPosition.EntryPrice.String(),
		"mark_price":          markPrice.String(),
		"liquidation_price":   targetPosition.LiquidationPrice.String(),
		"margin":              targetPosition.Margin.String(),
		"unrealized_pnl":      targetPosition.UnrealizedPnL.String(),
		"is_at_risk":          isAtRisk,
		"distance_to_liq":     distanceToLiquidation.String(),
		"pct_distance_to_liq": pctDistance.String() + "%",
		"leverage":            targetPosition.Leverage.String(),
		"margin_ratio":        targetPosition.Margin.Div(targetPosition.Size.Mul(markPrice)).Mul(decimal.NewFromInt(100)).String() + "%",
	}

	json.NewEncoder(w).Encode(response)
}

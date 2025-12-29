package risk

import (
	"fmt"
	"time"

	"vortex/internal/domain"
	"vortex/pkg/utils"
	"github.com/shopspring/decimal"
)

type Config struct {
	Symbol                string
	MaintenanceMarginRate utils.Decimal // e.g., 0.005 = 0.5%
	InitialMarginRate     utils.Decimal // e.g., 0.01 = 1% (100x leverage)
	MaxLeverage           utils.Decimal
	FundingInterval       time.Duration // e.g., 8 hours
	FundingRateCap        utils.Decimal // e.g., 0.0075 = 0.75%
	ImpactMarginNotional  utils.Decimal // For mark price calculation
}

type Engine struct {
	config             Config
	markPriceCalc      *MarkPriceCalculator
	accountRepo        domain.AccountRepository
	positionRepo       domain.PositionRepository
	marketDataProvider domain.MarketDataProvider
}

func NewEngine(
	config Config,
	markPriceCalc *MarkPriceCalculator,
	accountRepo domain.AccountRepository,
	positionRepo domain.PositionRepository,
	mdp domain.MarketDataProvider,
) *Engine {
	return &Engine{
		config:             config,
		markPriceCalc:      markPriceCalc,
		accountRepo:        accountRepo,
		positionRepo:       positionRepo,
		marketDataProvider: mdp,
	}
}

// ValidateOrder checks if user has sufficient margin to place order
func (re *Engine) ValidateOrder(order *domain.Order, account *domain.Account) error {
	account.Mu.RLock()
	defer account.Mu.RUnlock()

	// 1. Check if order would reduce or increase position
	existingPos, hasPosition := account.Positions[order.Symbol]

	var requiredMargin utils.Decimal
	if hasPosition {
		// Calculate margin impact
		requiredMargin = re.calculateOrderMarginImpact(order, existingPos)
	} else {
		// New position: calculate initial margin
		notionalValue := order.Quantity * order.Price
		requiredMargin = notionalValue / re.config.MaxLeverage
	}

	// 2. Check available margin
	if requiredMargin > account.AvailableMargin {
		return fmt.Errorf("insufficient margin: need %.2f, have %.2f",
			requiredMargin, account.AvailableMargin)
	}

	// 3. Validate leverage doesn't exceed max
	if hasPosition {
		newSize := existingPos.Size
		if order.Side == domain.SideBuy {
			newSize += order.Quantity
		} else {
			newSize -= order.Quantity
		}

		effectiveLeverage := (newSize * order.Price) / account.TotalEquity
		if effectiveLeverage > re.config.MaxLeverage {
			return fmt.Errorf("leverage %.2fx exceeds maximum %.2fx",
				effectiveLeverage, re.config.MaxLeverage)
		}
	}

	return nil
}

// calculateOrderMarginImpact determines margin change from an order
func (re *Engine) calculateOrderMarginImpact(order *domain.Order, position *domain.Position) utils.Decimal {
	// If order reduces position, it frees margin
	isReducing := (position.Side == domain.SideBuy && order.Side == domain.SideSell) ||
		(position.Side == domain.SideSell && order.Side == domain.SideBuy)

	if isReducing {
		reduceQty := utils.Min(order.Quantity, utils.Abs(position.Size))
		freedMargin := (reduceQty / utils.Abs(position.Size)) * position.Margin
		return -freedMargin // Negative because it frees margin
	}

	// Order increases position
	notionalValue := order.Quantity * order.Price
	return notionalValue / position.Leverage
}

// CalculateMargin computes required margin for a position
func (re *Engine) CalculateMargin(position *domain.Position) utils.Decimal {
	notionalValue := utils.Abs(position.Size) * position.EntryPrice
	return notionalValue / position.Leverage
}

// CalculateLiquidationPrice computes the price at which position is liquidated
func (re *Engine) CalculateLiquidationPrice(position *domain.Position) utils.Decimal {
	notionalValue := utils.Abs(position.Size) * position.EntryPrice
	maintenanceMargin := notionalValue * re.config.MaintenanceMarginRate

	// Liquidation buffer = InitialMargin - MaintenanceMargin
	liquidationBuffer := position.Margin - maintenanceMargin

	var liqPrice utils.Decimal
	if position.Side == domain.SideBuy {
		// Long: LiqPrice = Entry - (Buffer / Size)
		liqPrice = position.EntryPrice - (liquidationBuffer / utils.Abs(position.Size))
	} else {
		// Short: LiqPrice = Entry + (Buffer / Size)
		liqPrice = position.EntryPrice + (liquidationBuffer / utils.Abs(position.Size))
	}

	// Ensure liquidation price is positive
	if liqPrice < 0 {
		liqPrice = 0.01
	}

	return liqPrice
}

// CheckLiquidation determines if a position should be liquidated
func (re *Engine) CheckLiquidation(position *domain.Position, markPrice utils.Decimal) bool {
	if position.Status != domain.PositionStatusOpen {
		return false
	}

	// Check if mark price crossed liquidation threshold
	if position.Side == domain.SideBuy && markPrice <= position.LiquidationPrice {
		return true
	}
	if position.Side == domain.SideSell && markPrice >= position.LiquidationPrice {
		return true
	}

	return false
}

// UpdatePositionMargin recalculates position metrics based on mark price
func (re *Engine) UpdatePositionMargin(position *domain.Position, markPrice utils.Decimal) {
	position.MarkPrice = markPrice

	// Calculate unrealized PnL
	if position.Side == domain.SideBuy {
		position.UnrealizedPnL = (markPrice - position.EntryPrice) * position.Size
	} else {
		position.UnrealizedPnL = (position.EntryPrice - markPrice) * utils.Abs(position.Size)
	}

	// Update liquidation price (may change with funding payments)
	position.LiquidationPrice = re.CalculateLiquidationPrice(position)
	position.UpdatedAt = time.Now()
}

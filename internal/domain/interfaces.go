package domain

import (
	"github.com/shopspring/decimal"
)

// RiskEngine defines the interface for risk management
type RiskEngine interface {
	ValidateOrder(order *Order, account *Account) error
	CalculateMargin(position *Position) decimal.Decimal
	CalculateLiquidationPrice(position *Position) decimal.Decimal
	CheckLiquidation(position *Position, markPrice decimal.Decimal) bool
	UpdatePositionMargin(position *Position, markPrice decimal.Decimal)
}

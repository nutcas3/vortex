package domain

import (
	"context"

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

// OrderRepository defines the interface for order management
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, orderID string) (*Order, error)
	GetOrdersByUser(ctx context.Context, userID, symbol, status string, limit int) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, orderID string) error
}

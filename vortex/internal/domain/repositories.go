package domain

import (
	"context"
	"github.com/shopspring/decimal"
)

type TradeRepository interface {
	Save(ctx context.Context, trade *Trade) error
	FindByUser(ctx context.Context, userID string) ([]*Trade, error)
	FindBySymbol(ctx context.Context, symbol string, limit int) ([]*Trade, error)
}

// AccountRepository manages user accounts
type AccountRepository interface {
	Get(ctx context.Context, userID string) (*Account, error)
	Update(ctx context.Context, account *Account) error
}

// PositionRepository manages trading positions
type PositionRepository interface {
	Get(ctx context.Context, userID, symbol string) (*Position, error)
	Save(ctx context.Context, position *Position) error
	FindByUser(ctx context.Context, userID string) ([]*Position, error)
}

// MarketDataProvider supplies real-time market data
type MarketDataProvider interface {
	GetMarkPrice(ctx context.Context, symbol string) (utils.Decimal, error)
	GetIndexPrice(ctx context.Context, symbol string) (utils.Decimal, error)
	SubscribePriceUpdates(symbol string) (<-chan utils.Decimal, error)
}

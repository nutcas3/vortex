package domain

import (
	"sync"
	"time"

	"vortex/pkg/utils"
	"github.com/shopspring/decimal"
)

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusOpen      OrderStatus = "OPEN"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

type PositionStatus string

const (
	PositionStatusOpen       PositionStatus = "OPEN"
	PositionStatusLiquidated PositionStatus = "LIQUIDATED"
	PositionStatusClosed     PositionStatus = "CLOSED"
)

type Order struct {
	ID          string
	UserID      string
	Symbol      string // e.g., "BTC-PERP"
	Side        Side
	Type        OrderType
	Price       utils.Decimal // For limit orders; 0 for market orders
	Quantity    utils.Decimal
	FilledQty   utils.Decimal
	Status      OrderStatus
	Timestamp   time.Time
	TimeInForce string // "GTC", "IOC", "FOK"
	ReduceOnly  bool   // Only reduce position, don't increase
	PostOnly    bool   // Only add liquidity, don't take
}

type Position struct {
	ID               string
	UserID           string
	Symbol           string
	Side             Side
	Size             utils.Decimal // Positive for long, negative for short
	EntryPrice       utils.Decimal
	MarkPrice        utils.Decimal
	LiquidationPrice utils.Decimal
	Leverage         utils.Decimal
	Margin           utils.Decimal // Initial margin allocated
	UnrealizedPnL    utils.Decimal
	RealizedPnL      utils.Decimal
	Status           PositionStatus
	LastFundingTime  time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Mu               sync.RWMutex
}

type Trade struct {
	ID           string
	Symbol       string
	Price        utils.Decimal
	Quantity     utils.Decimal
	BuyOrderID   string
	SellOrderID  string
	BuyerID      string
	SellerID     string
	IsBuyerMaker bool
	Timestamp    time.Time
}

type Account struct {
	UserID          string
	Balance         utils.Decimal // Collateral in USD
	LockedBalance   utils.Decimal // Margin locked in positions
	UnrealizedPnL   utils.Decimal
	TotalEquity     utils.Decimal // Balance + UnrealizedPnL
	AvailableMargin utils.Decimal
	UsedMargin      utils.Decimal
	Positions       map[string]*Position // keyed by symbol
	Mu              sync.RWMutex
}

type MarketData struct {
	Symbol      string
	LastPrice   utils.Decimal
	MarkPrice   utils.Decimal // Fair price for margin calculations
	IndexPrice  utils.Decimal // Spot price from index
	FundingRate utils.Decimal
	Volume24h   utils.Decimal
	UpdatedAt   time.Time
}

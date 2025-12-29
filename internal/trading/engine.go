package trading

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"vortex/internal/domain"
	"vortex/pkg/id"
	"vortex/pkg/math"
)

// MatchingEngine executes orders against the order book
type MatchingEngine struct {
	orderBooks map[string]*OrderBook
	riskEngine domain.RiskEngine
	tradeRepo  domain.TradeRepository
	mu         sync.RWMutex
}

// NewMatchingEngine creates a matching engine
func NewMatchingEngine(riskEngine domain.RiskEngine, tradeRepo domain.TradeRepository) *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
		riskEngine: riskEngine,
		tradeRepo:  tradeRepo,
	}
}

// GetOrCreateOrderBook retrieves or initializes an order book
func (me *MatchingEngine) GetOrCreateOrderBook(symbol string) *OrderBook {
	me.mu.Lock()
	defer me.mu.Unlock()

	book, exists := me.orderBooks[symbol]
	if !exists {
		book = NewOrderBook(symbol)
		me.orderBooks[symbol] = book
	}
	return book
}

// ProcessOrder is an alias for SubmitOrder for consistency with handler naming
func (me *MatchingEngine) ProcessOrder(ctx context.Context, order *domain.Order, account *domain.Account) ([]*domain.Trade, error) {
	return me.SubmitOrder(ctx, order, account)
}

// SubmitOrder processes a new order
func (me *MatchingEngine) SubmitOrder(ctx context.Context, order *domain.Order, account *domain.Account) ([]*domain.Trade, error) {
	// 1. Risk check before matching
	if err := me.riskEngine.ValidateOrder(order, account); err != nil {
		order.Status = domain.OrderStatusRejected
		return nil, fmt.Errorf("risk check failed: %w", err)
	}

	// 2. Get order book
	book := me.GetOrCreateOrderBook(order.Symbol)

	// 3. Execute matching
	trades := make([]*domain.Trade, 0)

	if order.Type == domain.OrderTypeMarket {
		trades = me.matchMarketOrder(order, book, account)
	} else {
		trades = me.matchLimitOrder(order, book, account)
	}

	// 4. Save trades
	for _, trade := range trades {
		if err := me.tradeRepo.Save(ctx, trade); err != nil {
			return trades, err
		}
	}

	return trades, nil
}

// matchLimitOrder matches a limit order against the book
func (me *MatchingEngine) matchLimitOrder(order *domain.Order, book *OrderBook, _ *domain.Account) []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	book.mu.Lock()
	defer book.mu.Unlock()

	// Select opposite side
	oppositeSide := book.Asks
	if order.Side == domain.SideSell {
		oppositeSide = book.Bids
	}

	// Match against existing orders
	for oppositeSide.Len() > 0 && order.FilledQty.LessThan(order.Quantity) {
		topLevel := oppositeSide.levels[0]

		// Check if price crosses
		canMatch := false
		if order.Side == domain.SideBuy && order.Price.GreaterThanOrEqual(topLevel.Price) {
			canMatch = true
		} else if order.Side == domain.SideSell && order.Price.LessThanOrEqual(topLevel.Price) {
			canMatch = true
		}

		if !canMatch {
			break
		}

		// Match with orders at this price level (FIFO)
		for len(topLevel.Orders) > 0 && order.FilledQty.LessThan(order.Quantity) {
			makerOrder := topLevel.Orders[0]
			matchQty := math.Min(order.Quantity.Sub(order.FilledQty), makerOrder.Quantity.Sub(makerOrder.FilledQty))

			// Create trade
			trade := &domain.Trade{
				ID:           id.GenerateID("trade"),
				Symbol:       order.Symbol,
				Price:        makerOrder.Price, // Taker gets maker's price
				Quantity:     matchQty,
				Timestamp:    time.Now(),
				IsBuyerMaker: makerOrder.Side == domain.SideBuy,
			}

			if order.Side == domain.SideBuy {
				trade.BuyOrderID = order.ID
				trade.SellOrderID = makerOrder.ID
				trade.BuyerID = order.UserID
				trade.SellerID = makerOrder.UserID
			} else {
				trade.BuyOrderID = makerOrder.ID
				trade.SellOrderID = order.ID
				trade.BuyerID = makerOrder.UserID
				trade.SellerID = order.UserID
			}

			trades = append(trades, trade)

			// Update filled quantities
			order.FilledQty = order.FilledQty.Add(matchQty)
			makerOrder.FilledQty = makerOrder.FilledQty.Add(matchQty)
			topLevel.Volume = topLevel.Volume.Sub(matchQty)

			// Remove fully filled orders
			if makerOrder.FilledQty.GreaterThanOrEqual(makerOrder.Quantity) {
				makerOrder.Status = domain.OrderStatusFilled
				topLevel.Orders = topLevel.Orders[1:]
				delete(book.orderMap, makerOrder.ID)
			}
		}

		// Remove empty price level
		if len(topLevel.Orders) == 0 {
			heap.Pop(oppositeSide)
		}
	}

	// Update order status
	if order.FilledQty.GreaterThanOrEqual(order.Quantity) {
		order.Status = domain.OrderStatusFilled
	} else if order.FilledQty.GreaterThan(math.Zero) {
		order.Status = domain.OrderStatusOpen
	}

	// Add remaining quantity to book if not fully filled and not IOC
	if order.Status == domain.OrderStatusOpen && order.TimeInForce != "IOC" {
		book.AddOrder(order)
	}

	return trades
}

// matchMarketOrder matches a market order (always takes liquidity)
func (me *MatchingEngine) matchMarketOrder(order *domain.Order, book *OrderBook, _ *domain.Account) []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	book.mu.Lock()
	defer book.mu.Unlock()

	oppositeSide := book.Asks
	if order.Side == domain.SideSell {
		oppositeSide = book.Bids
	}

	for oppositeSide.Len() > 0 && order.FilledQty.LessThan(order.Quantity) {
		topLevel := oppositeSide.levels[0]

		for len(topLevel.Orders) > 0 && order.FilledQty.LessThan(order.Quantity) {
			makerOrder := topLevel.Orders[0]
			matchQty := math.Min(order.Quantity.Sub(order.FilledQty), makerOrder.Quantity.Sub(makerOrder.FilledQty))

			trade := &domain.Trade{
				ID:           id.GenerateID("trade"),
				Symbol:       order.Symbol,
				Price:        makerOrder.Price,
				Quantity:     matchQty,
				Timestamp:    time.Now(),
				IsBuyerMaker: makerOrder.Side == domain.SideBuy,
			}

			if order.Side == domain.SideBuy {
				trade.BuyOrderID = order.ID
				trade.SellOrderID = makerOrder.ID
				trade.BuyerID = order.UserID
				trade.SellerID = makerOrder.UserID
			} else {
				trade.BuyOrderID = makerOrder.ID
				trade.SellOrderID = order.ID
				trade.BuyerID = makerOrder.UserID
				trade.SellerID = order.UserID
			}

			trades = append(trades, trade)

			order.FilledQty = order.FilledQty.Add(matchQty)
			makerOrder.FilledQty = makerOrder.FilledQty.Add(matchQty)
			topLevel.Volume = topLevel.Volume.Sub(matchQty)

			if makerOrder.FilledQty.GreaterThanOrEqual(makerOrder.Quantity) {
				makerOrder.Status = domain.OrderStatusFilled
				topLevel.Orders = topLevel.Orders[1:]
				delete(book.orderMap, makerOrder.ID)
			}
		}

		if len(topLevel.Orders) == 0 {
			heap.Pop(oppositeSide)
		}
	}

	if order.FilledQty.GreaterThanOrEqual(order.Quantity) {
		order.Status = domain.OrderStatusFilled
	} else {
		order.Status = domain.OrderStatusRejected // Market order couldn't fill completely
	}

	return trades
}

// GetOrderBooks returns all order books
func (me *MatchingEngine) OrderBooks() map[string]*OrderBook {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return me.orderBooks
}

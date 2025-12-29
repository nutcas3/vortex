package trading

import (
	"container/heap"
	"errors"
	"sync"

	"vortex/internal/domain"
	"vortex/pkg/math"

	"github.com/shopspring/decimal"
)

// PriceLevel represents all orders at a specific price
type PriceLevel struct {
	Price  decimal.Decimal
	Orders []*domain.Order
	Volume decimal.Decimal
	index  int // heap index
}

// OrderBookSide implements a min-heap (asks) or max-heap (bids)
type OrderBookSide struct {
	levels   []*PriceLevel
	isAsk    bool
	priceMap map[decimal.Decimal]*PriceLevel
}

func (obs *OrderBookSide) Len() int { return len(obs.levels) }

func (obs *OrderBookSide) Less(i, j int) bool {
	if obs.isAsk {
		return obs.levels[i].Price.LessThan(obs.levels[j].Price) // min-heap for asks
	}
	return obs.levels[i].Price.GreaterThan(obs.levels[j].Price) // max-heap for bids
}

func (obs *OrderBookSide) Swap(i, j int) {
	obs.levels[i], obs.levels[j] = obs.levels[j], obs.levels[i]
	obs.levels[i].index = i
	obs.levels[j].index = j
}

func (obs *OrderBookSide) Push(x interface{}) {
	n := len(obs.levels)
	level := x.(*PriceLevel)
	level.index = n
	obs.levels = append(obs.levels, level)
	obs.priceMap[level.Price] = level
}

func (obs *OrderBookSide) Pop() interface{} {
	old := obs.levels
	n := len(old)
	level := old[n-1]
	old[n-1] = nil
	level.index = -1
	obs.levels = old[0 : n-1]
	delete(obs.priceMap, level.Price)
	return level
}

// OrderBook maintains buy and sell orders for a symbol
type OrderBook struct {
	Symbol    string
	Bids      *OrderBookSide // Buy orders (max-heap)
	Asks      *OrderBookSide // Sell orders (min-heap)
	orderMap  map[string]*domain.Order
	mu        sync.RWMutex
	lastTrade *domain.Trade
}

// NewOrderBook creates an initialized order book
func NewOrderBook(symbol string) *OrderBook {
	bids := &OrderBookSide{
		levels:   make([]*PriceLevel, 0),
		isAsk:    false,
		priceMap: make(map[decimal.Decimal]*PriceLevel),
	}
	asks := &OrderBookSide{
		levels:   make([]*PriceLevel, 0),
		isAsk:    true,
		priceMap: make(map[decimal.Decimal]*PriceLevel),
	}
	heap.Init(bids)
	heap.Init(asks)

	return &OrderBook{
		Symbol:   symbol,
		Bids:     bids,
		Asks:     asks,
		orderMap: make(map[string]*domain.Order),
	}
}

// AddOrder inserts an order into the book
func (ob *OrderBook) AddOrder(order *domain.Order) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Status != domain.OrderStatusOpen {
		return errors.New("only open orders can be added")
	}

	ob.orderMap[order.ID] = order

	side := ob.Bids
	if order.Side == domain.SideSell {
		side = ob.Asks
	}

	// Find or create price level
	level, exists := side.priceMap[order.Price]
	if !exists {
		level = &PriceLevel{
			Price:  order.Price,
			Orders: make([]*domain.Order, 0),
			Volume: math.Zero,
		}
		heap.Push(side, level)
	}

	level.Orders = append(level.Orders, order)
	level.Volume = level.Volume.Add(order.Quantity.Sub(order.FilledQty))

	return nil
}

// RemoveOrder deletes an order from the book
func (ob *OrderBook) RemoveOrder(orderID string) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.orderMap[orderID]
	if !exists {
		return errors.New("order not found")
	}

	side := ob.Bids
	if order.Side == domain.SideSell {
		side = ob.Asks
	}

	level := side.priceMap[order.Price]
	if level == nil {
		return errors.New("price level not found")
	}

	// Remove order from level
	for i, o := range level.Orders {
		if o.ID == orderID {
			level.Orders = append(level.Orders[:i], level.Orders[i+1:]...)
			level.Volume = level.Volume.Sub(order.Quantity.Sub(order.FilledQty))
			break
		}
	}

	// Remove empty price level
	if len(level.Orders) == 0 {
		heap.Remove(side, level.index)
	}

	delete(ob.orderMap, orderID)
	return nil
}

// GetBestBid returns the highest buy price
func (ob *OrderBook) GetBestBid() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.Bids.Len() == 0 {
		return math.Zero, false
	}
	return ob.Bids.levels[0].Price, true
}

// GetBestAsk returns the lowest sell price
func (ob *OrderBook) GetBestAsk() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.Asks.Len() == 0 {
		return math.Zero, false
	}
	return ob.Asks.levels[0].Price, true
}

// GetSpread returns the bid-ask spread
func (ob *OrderBook) GetSpread() decimal.Decimal {
	bid, bidExists := ob.GetBestBid()
	ask, askExists := ob.GetBestAsk()

	if !bidExists || !askExists {
		return math.Zero
	}
	return ask.Sub(bid)
}

// GetTopLevels returns the top N price levels
func (ob *OrderBook) GetTopLevels(side *OrderBookSide, depth int) [][]float64 {
	result := make([][]float64, 0, depth)
	for i := 0; i < side.Len() && i < depth; i++ {
		level := side.levels[i]
		result = append(result, []float64{
			level.Price.InexactFloat64(),
			level.Volume.InexactFloat64(),
		})
	}
	return result
}

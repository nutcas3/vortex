package risk

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"vortex/internal/domain"
	"vortex/internal/trading"
	"vortex/pkg/utils"

	"github.com/shopspring/decimal"
)

// LiquidationEngine monitors and executes liquidations
type LiquidationEngine struct {
	riskEngine      *Engine
	positionRepo    domain.PositionRepository
	accountRepo     domain.AccountRepository
	matchingEngine  *trading.MatchingEngine
	markPriceCalc   *MarkPriceCalculator
	liquidationChan chan *domain.Position
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewLiquidationEngine creates a new liquidation engine
func NewLiquidationEngine(
	riskEngine *Engine,
	positionRepo domain.PositionRepository,
	accountRepo domain.AccountRepository,
	matchingEngine *trading.MatchingEngine,
	markPriceCalc *MarkPriceCalculator,
) *LiquidationEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &LiquidationEngine{
		riskEngine:      riskEngine,
		positionRepo:    positionRepo,
		accountRepo:     accountRepo,
		matchingEngine:  matchingEngine,
		markPriceCalc:   markPriceCalc,
		liquidationChan: make(chan *domain.Position, 1000),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins monitoring positions for liquidation
func (le *LiquidationEngine) Start() {
	le.wg.Add(2)

	// Worker 1: Monitor all positions
	go le.monitorPositions()

	// Worker 2: Execute liquidations
	go le.executeLiquidations()

	log.Println("Liquidation Engine started")
}

// Stop gracefully shuts down the liquidation engine
func (le *LiquidationEngine) Stop() {
	le.cancel()
	close(le.liquidationChan)
	le.wg.Wait()
	log.Println("Liquidation Engine stopped")
}

// monitorPositions continuously checks all open positions
func (le *LiquidationEngine) monitorPositions() {
	defer le.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-le.ctx.Done():
			return
		case <-ticker.C:
			// In production: query all open positions from database
			// For now, we'll simulate with in-memory tracking
			le.checkAllPositions()
		}
	}
}

// checkAllPositions validates margin for all open positions
func (le *LiquidationEngine) checkAllPositions() {
	// Placeholder: In production, fetch from positionRepo
	// For each position, check if it should be liquidated
}

// CheckPosition evaluates a single position for liquidation
func (le *LiquidationEngine) CheckPosition(position *domain.Position) {
	if position.Status != domain.PositionStatusOpen {
		return
	}

	markPrice, err := le.markPriceCalc.CalculateMarkPrice(
		le.ctx, position.Symbol, position.MarkPrice)
	if err != nil {
		log.Printf("Error calculating mark price: %v", err)
		return
	}

	if le.riskEngine.CheckLiquidation(position, markPrice) {
		log.Printf("Liquidation triggered for position %s at mark price %.2f",
			position.ID, markPrice)

		// Send to liquidation queue
		select {
		case le.liquidationChan <- position:
		default:
			log.Printf("WARNING: Liquidation queue full, dropping position %s", position.ID)
		}
	}
}

// executeLiquidations processes the liquidation queue
func (le *LiquidationEngine) executeLiquidations() {
	defer le.wg.Done()

	for position := range le.liquidationChan {
		if err := le.liquidatePosition(position); err != nil {
			log.Printf("ERROR liquidating position %s: %v", position.ID, err)
		}
	}
}

// liquidatePosition closes a position by submitting a market order
func (le *LiquidationEngine) liquidatePosition(position *domain.Position) error {
	log.Printf("Executing liquidation for position %s (user: %s, size: %.4f)",
		position.ID, position.UserID, position.Size)

	// 1. Create liquidation order (opposite side, market order)
	liquidationOrder := &domain.Order{
		ID:          utils.GenerateID("liq_order"),
		UserID:      "LIQUIDATION_ENGINE", // System account
		Symbol:      position.Symbol,
		Side:        domain.SideSell,
		Type:        domain.OrderTypeMarket,
		Quantity:    utils.Abs(position.Size),
		Status:      domain.OrderStatusOpen,
		Timestamp:   time.Now(),
		TimeInForce: "IOC", // Immediate or cancel
		ReduceOnly:  true,
	}

	if position.Side == domain.SideBuy {
		liquidationOrder.Side = domain.SideSell
	} else {
		liquidationOrder.Side = domain.SideBuy
	}

	// 2. Get account
	account, err := le.accountRepo.Get(le.ctx, position.UserID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// 3. Submit liquidation order to matching engine
	trades, err := le.matchingEngine.SubmitOrder(le.ctx, liquidationOrder, account)
	if err != nil {
		return fmt.Errorf("failed to execute liquidation: %w", err)
	}

	// 4. Update position status
	position.Status = domain.PositionStatusLiquidated

	// Calculate realized PnL from liquidation
	var totalPnL decimal.Decimal
	for _, trade := range trades {
		pnl := (trade.Price - position.EntryPrice) * trade.Quantity
		if position.Side == domain.SideSell {
			pnl = -pnl
		}
		totalPnL += pnl
	}
	position.RealizedPnL += totalPnL

	// 5. Save updated position
	if err := le.positionRepo.Save(le.ctx, position); err != nil {
		return fmt.Errorf("failed to save liquidated position: %w", err)
	}

	// 6. Update account balance
	account.Mu.Lock()
	account.Balance += position.RealizedPnL
	account.LockedBalance -= position.Margin
	delete(account.Positions, position.Symbol)
	account.Mu.Unlock()

	if err := le.accountRepo.Update(le.ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("Liquidation completed: position %s, realized PnL: %.2f",
		position.ID, position.RealizedPnL)

	return nil
}

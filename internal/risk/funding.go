package risk

import (
	"context"
	"log"
	"sync"
	"time"

	"vortex/internal/domain"
)

// FundingEngine manages funding rate calculations and payments
type FundingEngine struct {
	config        Config
	fundingCalc   *FundingRateCalculator
	positionRepo  domain.PositionRepository
	accountRepo   domain.AccountRepository
	markPriceCalc *MarkPriceCalculator
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewFundingEngine creates a new funding engine
func NewFundingEngine(
	config Config,
	fundingCalc *FundingRateCalculator,
	positionRepo domain.PositionRepository,
	accountRepo domain.AccountRepository,
	markPriceCalc *MarkPriceCalculator,
) *FundingEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &FundingEngine{
		config:        config,
		fundingCalc:   fundingCalc,
		positionRepo:  positionRepo,
		accountRepo:   accountRepo,
		markPriceCalc: markPriceCalc,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the funding rate payment cycle
func (fe *FundingEngine) Start() {
	fe.wg.Add(1)
	go fe.fundingCycle()
	log.Println("Funding Engine started")
}

// Stop gracefully stops the funding engine
func (fe *FundingEngine) Stop() {
	fe.cancel()
	fe.wg.Wait()
	log.Println("Funding Engine stopped")
}

// fundingCycle runs every 8 hours to apply funding payments
func (fe *FundingEngine) fundingCycle() {
	defer fe.wg.Done()

	ticker := time.NewTicker(fe.config.FundingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fe.ctx.Done():
			return
		case <-ticker.C:
			if err := fe.applyFundingToAllPositions(); err != nil {
				log.Printf("ERROR applying funding: %v", err)
			}
		}
	}
}

// applyFundingToAllPositions calculates and applies funding to all open positions
func (fe *FundingEngine) applyFundingToAllPositions() error {
	log.Println("Applying funding payments...")

	// In production: query all open positions from database
	// For each symbol, calculate funding rate and apply to positions

	return nil
}

package risk

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"vortex/internal/domain"
	"vortex/pkg/math"

	"github.com/shopspring/decimal"
)

// MarkPriceCalculator computes the fair price for margin calculations
type MarkPriceCalculator struct {
	marketDataProvider domain.MarketDataProvider
}

// NewMarkPriceCalculator creates a new mark price calculator
func NewMarkPriceCalculator(mdp domain.MarketDataProvider) *MarkPriceCalculator {
	return &MarkPriceCalculator{
		marketDataProvider: mdp,
	}
}

// CalculateMarkPrice computes mark price using Fair Price Marking
// Formula: MarkPrice = IndexPrice + EMA(Perpetual - Index)
func (mpc *MarkPriceCalculator) CalculateMarkPrice(ctx context.Context, symbol string, lastPrice decimal.Decimal) (decimal.Decimal, error) {
	indexPrice, err := mpc.marketDataProvider.GetIndexPrice(ctx, symbol)
	if err != nil {
		return math.Zero, fmt.Errorf("failed to get index price: %w", err)
	}

	// Calculate basis (premium/discount)
	basis := lastPrice.Sub(indexPrice)

	// Apply exponential moving average (EMA) to smooth basis
	// In production: maintain historical EMA state
	emaBasis := basis.Mul(decimal.NewFromFloat(0.3)) // Simplified: 30% weight on current basis

	markPrice := indexPrice.Add(emaBasis)

	// Sanity check: mark price shouldn't deviate too much
	maxDeviation := indexPrice.Mul(math.Point1) // 10% max deviation
	if math.Abs(markPrice.Sub(indexPrice)).GreaterThan(maxDeviation) {
		log.Printf("WARNING: Mark price %.2f deviates significantly from index %.2f",
			markPrice, indexPrice)
		markPrice = indexPrice // Fall back to index price
	}

	return markPrice, nil
}

// FundingRateCalculator computes the funding rate for perpetual swaps
type FundingRateCalculator struct {
	config             Config
	marketDataProvider domain.MarketDataProvider
}

// NewFundingRateCalculator creates a new funding rate calculator
func NewFundingRateCalculator(config Config, mdp domain.MarketDataProvider) *FundingRateCalculator {
	return &FundingRateCalculator{
		config:             config,
		marketDataProvider: mdp,
	}
}

// CalculateFundingRate computes the 8-hour funding rate
// Formula: FundingRate = Premium_Index + clamp(Interest_Rate - Premium_Index, 0.05%, -0.05%)
func (frc *FundingRateCalculator) CalculateFundingRate(ctx context.Context, symbol string, markPrice, indexPrice decimal.Decimal) (decimal.Decimal, error) {
	// Premium Index = (MarkPrice - IndexPrice) / IndexPrice
	premiumIndex := markPrice.Sub(indexPrice).Div(indexPrice)

	// Interest Rate (typically small, e.g., 0.01% per 8 hours)
	interestRate := decimal.NewFromFloat(0.0001)

	// Calculate funding rate
	fundingRate := premiumIndex.Add(math.Clamp(
		interestRate-premiumIndex,
		decimal.NewFromFloat(-0.0005), // -0.05%
		decimal.Decimal(0.0005),  // +0.05%
	),

	// Apply funding rate cap
	fundingRate == math.Clamp(fundingRate, -frc.config.FundingRateCap, frc.config.FundingRateCap)

	return fundingRate, nil
}

// ApplyFunding applies funding payment to a position
// Positive funding rate: longs pay shorts
// Negative funding rate: shorts pay longs
func (frc *FundingRateCalculator) ApplyFunding(position *domain.Position, fundingRate decimal.Decimal) decimal.Decimal {
	notionalValue := position.Size.Mul(position.MarkPrice)
	fundingPayment := notionalValue.Mul(fundingRate)

	// Longs pay when funding is positive
	if position.Side == domain.SideBuy {
		fundingPayment = -fundingPayment
	}

	position.RealizedPnL += fundingPayment
	position.LastFundingTime = time.Now()

	return fundingPayment
}

package routes

import (
	"net/http"

	"vortex/internal/handlers"
)

// APIRoutes registers all API routes
func APIRoutes(
	mux *http.ServeMux,
	tradingHandler *handlers.TradingHandler,
	marketDataHandler *handlers.MarketDataHandler,
	accountHandler *handlers.AccountHandler,
	riskHandler *handlers.RiskHandler,
) {
	// Trading endpoints - based on domain.Order model
	mux.HandleFunc("/api/v1/orders", tradingHandler.Orders)
	mux.HandleFunc("/api/v1/orders/", tradingHandler.Orders)
	mux.HandleFunc("/api/v1/order", tradingHandler.Order)
	mux.HandleFunc("/api/v1/order/", tradingHandler.Order)
	mux.HandleFunc("/api/v1/fills", tradingHandler.Fills)
	mux.HandleFunc("/api/v1/fills/", tradingHandler.Fills)

	// Market data endpoints
	mux.HandleFunc("/api/v1/ticker/24hr", marketDataHandler.Ticker24hr)
	mux.HandleFunc("/api/v1/ticker/24hr/", marketDataHandler.Ticker24hr)
	mux.HandleFunc("/api/v1/depth", marketDataHandler.OrderBook)
	mux.HandleFunc("/api/v1/depth/", marketDataHandler.OrderBook)
	mux.HandleFunc("/api/v1/trades", marketDataHandler.RecentTrades)
	mux.HandleFunc("/api/v1/trades/", marketDataHandler.RecentTrades)
	mux.HandleFunc("/api/v1/klines", marketDataHandler.Klines)
	mux.HandleFunc("/api/v1/klines/", marketDataHandler.Klines)

	// Account endpoints - based on domain.Account model
	mux.HandleFunc("/api/v1/account", accountHandler.Account)
	mux.HandleFunc("/api/v1/account/", accountHandler.Account)
	mux.HandleFunc("/api/v1/positions", accountHandler.Positions)
	mux.HandleFunc("/api/v1/positions/", accountHandler.Positions)

	// Risk management endpoints - based on risk engine
	mux.HandleFunc("/api/v1/risk/margin", riskHandler.Margin)
	mux.HandleFunc("/api/v1/risk/margin/", riskHandler.Margin)
	mux.HandleFunc("/api/v1/risk/liquidation", riskHandler.Liquidation)
	mux.HandleFunc("/api/v1/risk/liquidation/", riskHandler.Liquidation)
}

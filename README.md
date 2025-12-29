# Vortex - Perpetual Futures Exchange

A production-grade perpetual futures trading platform built in Go, featuring high-performance matching engine, comprehensive risk management, and REST API.

## Architecture Overview

Vortex is designed as a modular, scalable trading platform with the following core components:

### Core Systems

- **Trading Engine** (`internal/trading/`) - High-performance order matching with price-time priority
- **Risk Engine** (`internal/risk/`) - Real-time risk management, liquidation, and funding calculations
- **Domain Models** (`internal/domain/`) - Core business entities (Order, Position, Account, Trade)
- **API Layer** (`internal/server/`) - RESTful HTTP API with proper handler organization

### Key Features

- **Price-Time Priority Matching** - Heap-based order book for optimal performance
- **Fair Price Marking** - Prevents market manipulation through sophisticated mark price calculation
- **Automated Liquidation** - Real-time position monitoring and liquidation engine
- **Funding Rate Mechanism** - 8-hour funding payments for perpetual swap convergence
- **High-Precision Math** - Uses `github.com/shopspring/decimal` for all financial calculations
- **RESTful API** - Complete trading and market data endpoints

## Project Structure

```
vortex/
├── internal/
│   ├── domain/           # Core business models and interfaces
│   │   ├── models.go     # Order, Position, Account, Trade entities
│   │   ├── interfaces.go # Risk engine and repository interfaces
│   │   └── repositories.go # Data access interfaces
│   ├── trading/          # Trading engine and order book
│   │   ├── engine.go     # Matching engine implementation
│   │   └── orderbook.go  # Heap-based order book
│   ├── risk/             # Risk management system
│   │   ├── engine.go     # Risk validation and calculations
│   │   ├── pricing.go    # Mark price and funding rate calculator
│   │   ├── liquidation.go # Liquidation engine
│   │   └── funding.go    # Funding payment processor
│   └── server/           # HTTP API layer
│       ├── handlers/     # API endpoint handlers
│       │   ├── trading.go    # Order and fill endpoints
│       │   ├── marketdata.go # Market data endpoints
│       │   ├── account.go     # Account and position endpoints
│       │   └── risk.go        # Risk management endpoints
│       ├── routes/       # Route definitions
│       │   └── api.go        # API route registration
│       └── routes.go    # Main route setup
├── pkg/                  # Shared utilities
│   ├── id/              # ID generation utilities
│   └── math/            # Decimal math helpers and constants
└── README.md
```

## API Endpoints

### Trading Endpoints
- `GET/POST /api/v1/orders` - Order management
- `GET/PUT/DELETE /api/v1/order` - Single order operations
- `GET /api/v1/fills` - Trade fill history

### Market Data Endpoints
- `GET /api/v1/ticker/24hr` - 24-hour ticker statistics
- `GET /api/v1/depth` - Order book depth
- `GET /api/v1/trades` - Recent trade history
- `GET /api/v1/klines` - Candlestick data

### Account Endpoints
- `GET /api/v1/account` - Account information
- `GET/POST /api/v1/positions` - Position management

### Risk Management Endpoints
- `GET /api/v1/risk/margin` - Margin calculations
- `GET /api/v1/risk/liquidation` - Liquidation price calculations

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL (for production deployments)

### Getting Started

1. Clone the repository
2. Install dependencies: `go mod download`
3. Run the application: `make run`

### Makefile Commands

Build and test the application:
```bash
make all          # Build with tests
make build        # Build the application
make run          # Run the application
make test         # Run test suite
make itest        # Run integration tests
make watch        # Live reload during development
make clean        # Clean build artifacts
```

Database management:
```bash
make docker-run   # Start PostgreSQL container
make docker-down  # Stop PostgreSQL container
```

## Financial Domain

### Key Concepts

- **Perpetual Swaps**: Futures contracts without expiration date
- **Funding Rate**: Mechanism to keep perpetual price anchored to index price
- **Mark Price**: Fair price calculation for margin and liquidation
- **Liquidation Price**: Price at which a position is automatically closed

### Formulas

```
Funding Rate = Premium Index + clamp(Interest Rate - Premium Index, -0.05%, 0.05%)
Liquidation Price = Entry Price ± (Initial Margin - Maintenance Margin) / Position Size
Mark Price = Index Price + EMA(Perpetual Price - Index Price)
```

## Performance

- **Order Matching**: Sub-millisecond latency with heap-based order book
- **Risk Checks**: Real-time validation before order execution
- **Liquidation**: Automated monitoring with configurable thresholds
- **API**: RESTful endpoints with proper HTTP method handling

## Security

- **HMAC Authentication**: API key and signature-based authentication (planned)
- **Rate Limiting**: Request throttling to prevent abuse
- **Input Validation**: Comprehensive validation for all API inputs
- **Margin Protection**: Multiple layers of risk validation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make test` to ensure all tests pass
6. Submit a pull request

## License


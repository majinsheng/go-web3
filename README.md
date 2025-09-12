# Go Web3 API

A Go-based REST API server for interacting with the Ethereum blockchain using go-ethereum.


https://github.com/user-attachments/assets/a05aa8e3-cec6-440f-b972-d3e7144e87a9


## Features

- REST API for Ethereum blockchain interactions
- Support for wallet operations (balance check, transfers)
- Transaction management and querying
- Block information retrieval
- Real-time Ethereum event monitoring via WebSockets
- Contract event subscriptions
- Transaction monitoring for specific addresses and high-value transactions
- For development with hot reloading

## Project Structure

```
go-web3/
├── .github/                   # GitHub specific files
├── cmd/                       # Application entry points
│   └── api/                   # API server entry point
│       └── main.go            # Main application
├── docs/                      # Documentation
│   └── websocket.md           # WebSocket API documentation
├── internal/                  # Internal packages (not importable)
│   ├── api/                   # REST API implementation
│   │   ├── events_handler.go  # WebSocket and events handlers
│   │   ├── handler.go         # API request handlers
│   │   ├── server.go          # HTTP server implementation
│   │   └── transaction_handler.go # Transaction monitoring handlers
│   ├── config/                # Configuration management
│   │   └── config.go          # Config loading and parsing
│   ├── ethereum/              # Ethereum client implementation
│   │   └── client.go          # Ethereum client wrapper
│   └── events/                # Ethereum events system
│       ├── listener.go        # Event listener implementation
│       ├── service.go         # Event service management
│       ├── transaction.go     # Transaction processing and filtering
│       └── websocket.go       # WebSocket client management
├── pkg/                       # Public packages (importable)
├── static/                    # Static web files
│   └── index.html             # WebSocket testing interface
├── .air.toml                  # Configuration for hot reload
├── .env                       # Environment variables
├── .env.example               # Example environment variables
├── .gitignore                 # Git ignore rules
├── config.yaml                # Application configuration
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── Makefile                   # Build automation
├── README.md                  # This file
├── run.bat                    # Windows run script
└── run.sh                     # Unix run script
```

## Prerequisites

- Go 1.16 or higher
- Access to an Ethereum node (local or remote via services like Infura)

## Configuration

1. Copy the `.env.example` file to `.env` and update the values:
   ```
   PRIVATE_KEY=your_ethereum_private_key_here
   INFURA_API_KEY=your_infura_api_key_here
   ```

2. Modify `config.yaml` for your specific needs:
   ```yaml
   server:
     port: 8080
     host: localhost

   ethereum:
     provider: http://localhost:8545
     chainID: 1
     privateKey: "" # Will be loaded from environment variable
   ```

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/go-web3.git
cd go-web3

# Install dependencies
go mod tidy
```

## Running the Application

```bash
# Start the API server
go run cmd/api/main.go

# Alternatively, use the provided scripts
# On Windows
.\run.bat

# On Unix-based systems
./run.sh
```

## Development

### Local Ethereum Node Setup

For development and testing, you can run a local Ethereum node using tools like Ganache or Hardhat.

#### Using Ganache

[Ganache](https://trufflesuite.com/ganache/) provides a personal Ethereum blockchain for development.

```bash
# Install Ganache CLI
npm install -g ganache

# Start a local Ethereum node with 10 pre-funded accounts
ganache --deterministic --accounts 10 --gasLimit 10000000

# Connect to it by setting in your .env or config.yaml:
# ETH_NODE_URL=http://127.0.0.1:8545
```

#### Using Hardhat

[Hardhat](https://hardhat.org/) is a development environment for Ethereum software.

```bash
# Install Hardhat
npm install --save-dev hardhat

# Initialize a Hardhat project
npx hardhat init

# Start a local node
npx hardhat node

# Connect to it by setting in your .env or config.yaml:
# ETH_NODE_URL=http://127.0.0.1:8545
```

#### Using Geth (Go Ethereum)

For a more production-like environment, you can use Geth in development mode:

```bash
# Install Geth (instructions vary by OS)
# For Windows with Chocolatey:
choco install geth

# For macOS with Homebrew:
brew install ethereum

# Start a local development node with websocket
geth -dev --http --http.addr "0.0.0.0" --http.port "8545" --http.api "eth,net,web3" --ws --ws.addr "0.0.0.0" --ws.port "8546" --ws.api "eth,net,web3" --ws.origins "*"

# Connect to it by setting in your .env or config.yaml:
# ETH_NODE_URL=http://127.0.0.1:8545
```

### Hot Reloading

For development with hot reloading, you can use [Air](https://github.com/cosmtrek/air). A configuration file (`.air.toml`) is already included in the project.

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## Web Interface

The project includes a simple web interface for testing the WebSocket event system. You can access it by opening your browser to the root URL of the server (e.g., `http://localhost:8080/`). 

Features of the web interface:
- Connect to the WebSocket server
- Subscribe to contract events
- Real-time display of Ethereum events

## Component Details

### Ethereum Client

The Ethereum client provides a simplified interface to interact with the Ethereum blockchain:

- Connection to Ethereum nodes (local or remote via Infura)
- Account management
- Transaction creation and sending
- Balance checking
- Block and transaction querying

### Event System

The event system provides real-time notifications for Ethereum blockchain events:

- Block monitoring - receive notifications when new blocks are mined
- Transaction monitoring - receive notifications for new transactions in blocks
- Contract event monitoring - subscribe to specific contract events
- Transaction filtering - filter transactions by address, value, and more
- WebSocket server for real-time event delivery to clients
- Support for filtering events by type and contract address

### REST API

The REST API provides HTTP endpoints for interacting with the Ethereum blockchain:

- Account operations (balance checking)
- Transaction operations (sending, querying)
- Block operations (querying)
- Event subscription management
- Transaction monitoring endpoints for watching addresses and high-value transactions

### Transaction Monitoring

The transaction monitoring system allows for targeted watching of transactions with specific criteria:

- **Address monitoring**: Watch for transactions to or from specific Ethereum addresses
- **High-value monitoring**: Watch for transactions above a configured ETH value threshold
- **Real-time notifications**: Receive instant notifications via WebSocket when matching transactions are detected
- **Filtering capabilities**: Additional filtering options include contract calls and method signatures

API endpoints for transaction monitoring:
- `POST /api/v1/monitor/address` - Start monitoring a specific address
- `POST /api/v1/monitor/high-value` - Start monitoring for high-value transactions

## API Endpoints

### Ethereum Operations

- `GET /api/v1/eth/balance/:address` - Get the ETH balance for an address
- `POST /api/v1/eth/transfer` - Send ETH to an address
- `GET /api/v1/eth/tx/:hash` - Get transaction details
- `GET /api/v1/eth/tx/:hash/receipt` - Get transaction receipt
- `GET /api/v1/eth/block/latest` - Get the latest block info
- `GET /api/v1/eth/block/:number` - Get block info by number

### Ethereum Events

- `GET /api/v1/events/ws` - WebSocket endpoint for real-time Ethereum events
- `POST /api/v1/events/subscribe` - Subscribe to specific contract events
- `GET /api/v1/events/latest/:type` - Get latest events of a specific type

### Transaction Monitoring

- `POST /api/v1/monitor/address` - Start monitoring a specific address
- `POST /api/v1/monitor/high-value` - Start monitoring for high-value transactions

### Health Check

- `GET /api/v1/health` - Server health check

## Example Requests

### Get Balance

```bash
curl -X GET http://localhost:8080/api/v1/eth/balance/0x742d35Cc6634C0532925a3b844Bc454e4438f44e
```

### Send Transaction

```bash
curl -X POST http://localhost:8080/api/v1/eth/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    "amount": "1000000000000000"
  }'
```

### Monitor Address

```bash
curl -X POST http://localhost:8080/api/v1/monitor/address \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
  }'
```

### Monitor High-Value Transactions

```bash
curl -X POST http://localhost:8080/api/v1/monitor/high-value \
  -H "Content-Type: application/json" \
  -d '{
    "minValue": "1.5"
  }'
```

## License

MIT

### Using the Makefile

The project includes a Makefile with common commands:

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Clean up build artifacts
make clean

# Install dependencies
make deps

# Run with hot reload (requires Air)
make dev
```

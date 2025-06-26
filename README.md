# Guess-the-Number Blockchain Game

A simple blockchain-based game where players guess a number between 1 and 10. If they guess correctly, they win MTK tokens. Win three times in a row to receive a bonus prize!

Built with:
- Solidity (ERC20 token + game smart contract)
- Go backend (Gin + go-ethereum)
- Hardhat for local development
- Basic HTML frontend

---

## Features

- ERC20 Token: `MTK` with minting capabilities
- Game smart contract (`Game.sol`) where players guess numbers
- REST API in Go to interact with the smart contract
- Bonus system: 3 wins in a row = extra reward
- Game history tracking
- Minimal frontend to test functionality

---

## Getting Started

### Prerequisites

- [Go 1.20+](https://golang.org/)
- [Node.js + npm](https://nodejs.org/)
- [Hardhat](https://hardhat.org/)
- Infura account (for Sepolia)

---

## Installation

```bash
git clone https://github.com/exccrr/solidity-token-go-integration.git
cd solidity-token-go-integration
```

### 1. Install Node modules

```bash
npm install
```

### 2. Create `.env` file

Create a `.env` in the root with:

```
SEPOLIA_URL=wss://sepolia.infura.io/ws/v3/YOUR_INFURA_PROJECT_ID
PRIVATE_KEY=YOUR_WALLET_PRIVATE_KEY
```

> Make sure the wallet has Sepolia ETH to pay for gas.

---

## Deploy Contracts

First, deploy the Token:

```bash
npx hardhat run scripts/deploy-token.js --network sepolia
```

Then deploy the Game contract using the token address:

```bash
npx hardhat run scripts/deploy-game.js --network sepolia
```

Copy the deployed addresses into `main.go` (`tokenAddress` and `gameAddress`).

---

## ABI Binding (Go)

```bash
abigen --abi Game.abi.json --pkg game --out game-server/game/game.go
abigen --abi Token.abi.json --pkg token --out game-server/token/token.go
```

---

## Run Backend Server

```bash
go run game-server/main.go
```

Runs the API server on `http://localhost:8080`.

---

## API Endpoints

| Method | URL                      | Description              |
|--------|--------------------------|--------------------------|
| POST   | `/play`                  | Make a guess (JSON body) |
| GET    | `/mint`                  | Mint 1000 MTK            |
| GET    | `/balance/:address`      | Get MTK balance          |
| GET    | `/history`               | See game log             |
| GET    | `/`                      | Basic frontend           |

### Example:

```bash
curl -X POST http://localhost:8080/play \
  -H "Content-Type: application/json" \
  -d '{"address": "0xYourWallet", "guess": 7}'
```

---

## Frontend

Basic UI in `/frontend/index.html`.

You can open it directly or serve via backend (`GET /`).

---

## License

MIT License. Feel free to use, fork, or contribute!

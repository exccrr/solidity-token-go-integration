package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/exccrr/solidity-token-go-integration/game-server/game"
	"github.com/exccrr/solidity-token-go-integration/game-server/token"
)

const (
	tokenAddress = "0xa33239e13303Fe9586C25b70ABd4D5d65E7B368f"
	gameAddress  = "0x3726fef83444Ba54F925A5d2195f697234DfA30C"
	chainID      = 11155111
)

var (
	topicGameWin  = crypto.Keccak256Hash([]byte("Win(address,uint8,uint8)")).Hex()
	topicGameLoss = crypto.Keccak256Hash([]byte("Loss(address,uint8,uint8)")).Hex()
)

var (
	client        *ethclient.Client
	privateKey    *ecdsa.PrivateKey
	publicAddr    string
	tokenInstance *token.Token
	gameInstance  *game.Game
	winStreaks    = map[string]int{}
	bonusAmount   = new(big.Int).Mul(big.NewInt(50), big.NewInt(1e18))
)

type GameLog struct {
	Address   string    `json:"address"`
	Guess     int       `json:"guess"`
	Winning   int       `json:"winning"`
	Result    string    `json:"result"`
	TxHash    string    `json:"txHash"`
	Timestamp time.Time `json:"timestamp"`
}

var gameLogs []GameLog

type PlayRequest struct {
	Address string `json:"address"`
	Guess   int    `json:"guess"`
}

func main() {
	_ = godotenv.Load()

	infuraURL := os.Getenv("SEPOLIA_URL")
	priv := os.Getenv("PRIVATE_KEY")

	var err error
	privateKey, err = crypto.HexToECDSA(priv)
	if err != nil {
		log.Fatal("Invalid private key:", err)
	}

	client, err = ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum node:", err)
	}

	publicKey := privateKey.Public()
	publicAddr = crypto.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).Hex()
	log.Println("Owner address:", publicAddr)

	tokenInstance, err = token.NewToken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		log.Fatal("Failed to bind token contract:", err)
	}

	gameInstance, err = game.NewGame(common.HexToAddress(gameAddress), client)
	if err != nil {
		log.Fatal("Failed to bind game contract:", err)
	}

	go watchGameEvents()

	router := gin.Default()
	router.POST("/play", playHandler)
	router.GET("/mint", mintHandler)
	router.GET("/balance/:address", balanceHandler)
	router.GET("/history", func(c *gin.Context) {
		c.JSON(http.StatusOK, gameLogs)
	})

	router.StaticFile("/", "./frontend/index.html")
	router.Run(":8080")
}

func playHandler(c *gin.Context) {
	var req PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if req.Guess < 1 || req.Guess > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "guess must be between 1 and 10"})
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth error"})
		return
	}

	amount := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	approveTx, err := tokenInstance.Approve(auth, common.HexToAddress(gameAddress), amount)
	if err != nil {
		log.Println("Approve error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "approve failed"})
		return
	}
	log.Println("Approve tx hash:", approveTx.Hash().Hex())

	time.Sleep(2 * time.Second)

	allowance, err := tokenInstance.Allowance(nil, common.HexToAddress(publicAddr), common.HexToAddress(gameAddress))
	if err != nil {
		log.Println("Allowance check failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "allowance check failed"})
		return
	}
	log.Println("Allowance from sender to Game:", allowance.String())

	allowance, err = tokenInstance.Allowance(nil, common.HexToAddress(publicAddr), common.HexToAddress(gameAddress))
	if err != nil {
		log.Println("Allowance check failed:", err)
	}
	log.Println("Current allowance:", allowance)

	playTx, err := gameInstance.Play(auth, uint8(req.Guess))

	if err != nil {
		log.Println("Play error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "play failed"})
		return
	}
	log.Println("Play tx hash:", playTx.Hash().Hex())

	logEntry := GameLog{
		Address:   req.Address,
		Guess:     req.Guess,
		Winning:   -1,
		Result:    "submitted",
		TxHash:    playTx.Hash().Hex(),
		Timestamp: time.Now(),
	}

	if playTx != nil {
		logEntry.TxHash = playTx.Hash().Hex()
	}

	gameLogs = append(gameLogs, logEntry)

	c.JSON(http.StatusOK, gin.H{
		"result":    "submitted",
		"guess":     req.Guess,
		"approveTx": approveTx.Hash().Hex(),
		"playTx":    playTx.Hash().Hex(),
	})
}

func mintHandler(c *gin.Context) {
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth error"})
		return
	}

	amount := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))
	tx, err := tokenInstance.Mint(auth, common.HexToAddress(publicAddr), amount)
	if err != nil {
		log.Println("Mint failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mint failed"})
		return
	}
	log.Println("Mint tx hash:", tx.Hash().Hex())
	c.JSON(http.StatusOK, gin.H{
		"mintedTo": publicAddr,
		"amount":   "1000 MTK",
		"txHash":   tx.Hash().Hex(),
	})
}

func balanceHandler(c *gin.Context) {
	addr := c.Param("address")
	balance, err := tokenInstance.BalanceOf(nil, common.HexToAddress(addr))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "balance check failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"address": addr,
		"balance": balance.String(),
	})
}

func watchGameEvents() {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(gameAddress)},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Println("Event watcher disabled (likely HTTP endpoint)")
		return
	}

	log.Println("Listening for game events...")

	for {
		select {
		case err := <-sub.Err():
			log.Println("Subscription error:", err)
		case vLog := <-logs:
			log.Printf("New event log received: %+v\n", vLog)

			topic := vLog.Topics[0].Hex()

			switch topic {
			case topicGameWin:
				addr := extractAddressFromTopic(vLog.Topics[1])
				winStreaks[addr]++
				log.Println("Win for", addr, "- streak:", winStreaks[addr])

				if winStreaks[addr] == 3 {
					auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
					tx, err := tokenInstance.Mint(auth, common.HexToAddress(addr), bonusAmount)
					if err != nil {
						log.Println("Bonus mint error:", err)
					} else {
						log.Println("Bonus 50 MTK minted to", addr, "tx:", tx.Hash().Hex())
						winStreaks[addr] = 0
					}
				}

			case topicGameLoss:
				addr := extractAddressFromTopic(vLog.Topics[1])
				winStreaks[addr] = 0
				log.Println("Loss for", addr, "- streak reset")
			}
		}
	}
}

func extractAddressFromTopic(topic common.Hash) string {
	return common.BytesToAddress(topic.Bytes()[12:]).Hex()
}

package main

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/exccrr/solidity-token-go-integration/game-server/token"
)

const (
	tokenAddress = "0xa33239e13303Fe9586C25b70ABd4D5d65E7B368f"
)

var (
	client        *ethclient.Client
	privateKey    *ecdsa.PrivateKey
	publicAddr    string
	tokenInstance *token.Token
)

type PlayRequest struct {
	Address string `json:"address"`
	Guess   int    `json:"guess"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	infuraURL := os.Getenv("INFURA_URL")
	priv := os.Getenv("PRIVATE_KEY")

	var err error
	privateKey, err = crypto.HexToECDSA(priv)
	if err != nil {
		log.Fatal("Invalid private key:", err)
	}

	client, err = ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatal("Failed to connect to Infura:", err)
	}

	publicKey := privateKey.Public()
	publicAddr = crypto.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey)).Hex()
	log.Println("Owner address:", publicAddr)

	tokenInstance, err = token.NewToken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		log.Fatal("Failed to bind token:", err)
	}

	router := gin.Default()
	router.POST("/play", playHandler)
	router.Run(":8080")
}

func playHandler(c *gin.Context) {
	var req PlayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	winningNumber := rand.Intn(10) + 1
	log.Printf("Player %s guessed %d (winning %d)", req.Address, req.Guess, winningNumber)

	if req.Guess != winningNumber {
		c.JSON(http.StatusOK, gin.H{"result": "lose", "winningNumber": winningNumber})
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth error"})
		return
	}

	amount := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	tx, err := tokenInstance.Mint(auth, common.HexToAddress(req.Address), amount)
	if err != nil {
		log.Println("mint error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mint failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        "win",
		"txHash":        tx.Hash().Hex(),
		"winningNumber": winningNumber,
		"mintedTo":      req.Address,
		"amountMinted":  "10 MTK",
	})
}

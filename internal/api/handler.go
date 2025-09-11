package api

import (
	"context"
	"math/big"
	"net/http"
	"strconv"

	"github.com/em/go-web3/internal/ethereum"
	"github.com/em/go-web3/internal/events"
	"github.com/gin-gonic/gin"
)

// Handler handles the API requests
type Handler struct {
	ethClient    *ethereum.Client
	eventService *events.Service
}

// NewHandler creates a new API handler
func NewHandler(ethClient *ethereum.Client, eventService *events.Service) *Handler {
	return &Handler{
		ethClient:    ethClient,
		eventService: eventService,
	}
}

// SetupRoutes sets up the API routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Ethereum endpoints
		eth := v1.Group("/eth")
		{
			eth.GET("/balance/:address", h.GetBalance)
			eth.POST("/transfer", h.SendTransaction)
			eth.GET("/tx/:hash", h.GetTransaction)
			eth.GET("/tx/:hash/receipt", h.GetTransactionReceipt)
			eth.GET("/block/latest", h.GetLatestBlock)
			eth.GET("/block/:number", h.GetBlockByNumber)
		}

		// Events endpoints
		events := v1.Group("/events")
		{
			events.GET("/ws", h.EventsHandler)
			events.POST("/subscribe", h.SubscribeToContractEvents)
			events.GET("/latest/:type", h.GetLatestEvents)
		}

		// Transaction monitoring endpoints
		txMonitor := v1.Group("/monitor")
		{
			txMonitor.POST("/address", h.WatchAddressHandler)
			txMonitor.POST("/high-value", h.WatchHighValueTransactionsHandler)
		}

		// Health check
		v1.GET("/health", h.HealthCheck)
	}
}

// HealthCheck handles the health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GetBalance handles the get balance endpoint
func (h *Handler) GetBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "address is required",
		})
		return
	}

	balance, err := h.ethClient.GetBalance(context.Background(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": balance.String(),
	})
}

// TransactionRequest represents a transaction request
type TransactionRequest struct {
	To     string `json:"to" binding:"required"`
	Amount string `json:"amount" binding:"required"`
}

// SendTransaction handles the send transaction endpoint
func (h *Handler) SendTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid amount format",
		})
		return
	}

	txHash, err := h.ethClient.SendTransaction(context.Background(), req.To, amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"txHash": txHash,
	})
}

// GetTransaction handles the get transaction endpoint
func (h *Handler) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transaction hash is required",
		})
		return
	}

	tx, isPending, err := h.ethClient.GetTransactionByHash(context.Background(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// For simplicity, we won't try to get the sender address which requires more complex handling
	// In a production app, you would use a TransactionSigner to recover the sender

	// Check if To address is nil (contract creation)
	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	} else {
		to = "contract creation"
	}

	c.JSON(http.StatusOK, gin.H{
		"hash":      hash,
		"isPending": isPending,
		"to":        to,
		"value":     tx.Value().String(),
		"gasPrice":  tx.GasPrice().String(),
		"gas":       tx.Gas(),
		"nonce":     tx.Nonce(),
	})
}

// GetTransactionReceipt handles the get transaction receipt endpoint
func (h *Handler) GetTransactionReceipt(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transaction hash is required",
		})
		return
	}

	receipt, err := h.ethClient.GetTransactionReceipt(context.Background(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"txHash":          hash,
		"blockHash":       receipt.BlockHash.Hex(),
		"blockNumber":     receipt.BlockNumber.String(),
		"gasUsed":         receipt.GasUsed,
		"status":          receipt.Status,
		"contractAddress": receipt.ContractAddress.Hex(),
	})
}

// GetLatestBlock handles the get latest block endpoint
func (h *Handler) GetLatestBlock(c *gin.Context) {
	blockNumber, err := h.ethClient.GetLatestBlockNumber(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	block, err := h.ethClient.GetBlockByNumber(context.Background(), blockNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"number":     block.Number().String(),
		"hash":       block.Hash().Hex(),
		"parentHash": block.ParentHash().Hex(),
		"timestamp":  block.Time(),
		"txCount":    len(block.Transactions()),
	})
}

// GetBlockByNumber handles the get block by number endpoint
func (h *Handler) GetBlockByNumber(c *gin.Context) {
	numberStr := c.Param("number")
	if numberStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "block number is required",
		})
		return
	}

	number, err := strconv.ParseUint(numberStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid block number",
		})
		return
	}

	block, err := h.ethClient.GetBlockByNumber(context.Background(), number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"number":     block.Number().String(),
		"hash":       block.Hash().Hex(),
		"parentHash": block.ParentHash().Hex(),
		"timestamp":  block.Time(),
		"txCount":    len(block.Transactions()),
	})
}

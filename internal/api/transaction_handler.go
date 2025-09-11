package api

import (
	"math/big"
	"net/http"

	"github.com/em/go-web3/internal/events"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// WatchAddressRequest represents a request to watch a specific address
type WatchAddressRequest struct {
	Address string `json:"address" binding:"required"`
}

// WatchHighValueTransactionsRequest represents a request to watch for high-value transactions
type WatchHighValueTransactionsRequest struct {
	MinValue string `json:"minValue" binding:"required"` // In ETH as a string
}

// WatchAddressHandler handles adding an address to the watch list
func (h *Handler) WatchAddressHandler(c *gin.Context) {
	var req WatchAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate address
	if !common.IsHexAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address",
		})
		return
	}

	// Add address to watch list
	h.eventService.WatchAddress(req.Address)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address added to watch list",
		"address": req.Address,
	})
}

// WatchHighValueTransactionsHandler handles setting up a watch for high-value transactions
func (h *Handler) WatchHighValueTransactionsHandler(c *gin.Context) {
	var req WatchHighValueTransactionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse ETH value to wei
	ethValue, ok := new(big.Float).SetString(req.MinValue)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ETH value",
		})
		return
	}

	// Convert ETH to wei (1 ETH = 10^18 wei)
	weiValue := new(big.Float).Mul(ethValue, new(big.Float).SetInt(big.NewInt(1000000000000000000)))

	// Convert to integer
	minValue := new(big.Int)
	weiValue.Int(minValue)

	// Create filter for high-value transactions
	filter := &events.TransactionFilter{
		MinValue: minValue,
	}

	// Add filter to transaction processor
	h.eventService.AddTransactionFilter(filter)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Watching for high-value transactions",
		"minValue": req.MinValue + " ETH",
	})
}

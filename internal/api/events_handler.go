package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/user/go-web3/internal/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		return true
	},
}

// EventsHandler handles WebSocket connections for events
func (h *Handler) EventsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not upgrade connection to WebSocket",
		})
		return
	}

	// Create a new WebSocket client
	client := events.NewWebSocketClient(conn)

	// Register client with event service
	h.eventService.RegisterClient(client)

	// Start reading and writing goroutines
	client.StartReading(h.eventService)
	client.StartWriting()
}

// SubscribeToContractEvents handles contract event subscriptions
func (h *Handler) SubscribeToContractEvents(c *gin.Context) {
	var req struct {
		ContractAddress string   `json:"contractAddress" binding:"required"`
		EventSignatures []string `json:"eventSignatures"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.eventService.SubscribeToContract(req.ContractAddress, req.EventSignatures)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully subscribed to contract events",
	})
}

// GetLatestEvents gets the latest events of a specific type
func (h *Handler) GetLatestEvents(c *gin.Context) {
	eventType := c.Param("type")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "event type is required",
		})
		return
	}

	// In a real application, you would implement a cache or database to store recent events
	// For now, we'll just return a placeholder

	c.JSON(http.StatusOK, gin.H{
		"events": []gin.H{
			{
				"type":      eventType,
				"timestamp": time.Now().Unix(),
				"message":   "This is a placeholder for real event data",
			},
		},
	})
}

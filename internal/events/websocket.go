package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID      string
	conn    *websocket.Conn
	send    chan []byte
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	filters EventFilters
}

// EventFilters holds filters for events the client is interested in
type EventFilters struct {
	EventTypes      []EventType
	ContractAddress []string
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(conn *websocket.Conn) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketClient{
		ID:      uuid.New().String(),
		conn:    conn,
		send:    make(chan []byte, 256),
		ctx:     ctx,
		cancel:  cancel,
		filters: EventFilters{},
	}
}

// Send sends a message to the client
func (c *WebSocketClient) Send(message []byte) error {
	select {
	case c.send <- message:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client connection closed")
	default:
		// Buffer full, close the connection
		c.Close()
		return fmt.Errorf("client send buffer full")
	}
}

// SendPing sends a ping message to the client
func (c *WebSocketClient) SendPing() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
}

// Close closes the client connection
func (c *WebSocketClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()
	c.conn.Close()
}

// Done returns a channel that's closed when the client is done
func (c *WebSocketClient) Done() <-chan struct{} {
	return c.ctx.Done()
}

// StartReading starts reading messages from the client
func (c *WebSocketClient) StartReading(service *Service) {
	go func() {
		defer func() {
			service.UnregisterClient(c.ID)
		}()

		c.conn.SetReadLimit(512 * 1024) // 512KB
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.conn.SetPongHandler(func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket read error: %v\n", err)
				}
				break
			}

			// Process message (subscription request, filter change, etc.)
			c.handleMessage(message, service)
		}
	}()
}

// StartWriting starts writing messages to the client
func (c *WebSocketClient) StartWriting() {
	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer func() {
			ticker.Stop()
			c.conn.Close()
		}()

		for {
			select {
			case message, ok := <-c.send:
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if !ok {
					// Channel closed
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)

				// Add queued messages
				n := len(c.send)
				for i := 0; i < n; i++ {
					w.Write([]byte{'\n'})
					w.Write(<-c.send)
				}

				if err := w.Close(); err != nil {
					return
				}
			case <-ticker.C:
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

// handleMessage processes incoming WebSocket messages
func (c *WebSocketClient) handleMessage(message []byte, service *Service) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		fmt.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	// Check message type
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribe":
		// Handle subscription request
		if contract, ok := msg["contract"].(string); ok {
			var events []string
			if eventsArray, ok := msg["events"].([]interface{}); ok {
				for _, e := range eventsArray {
					if event, ok := e.(string); ok {
						events = append(events, event)
					}
				}
			}
			service.SubscribeToContract(contract, events)
		}

	case "filter":
		// Handle filter update
		if eventTypes, ok := msg["eventTypes"].([]interface{}); ok {
			var types []EventType
			for _, t := range eventTypes {
				if eventType, ok := t.(string); ok {
					types = append(types, EventType(eventType))
				}
			}
			c.filters.EventTypes = types
		}

		if contracts, ok := msg["contracts"].([]interface{}); ok {
			var addresses []string
			for _, addr := range contracts {
				if address, ok := addr.(string); ok {
					addresses = append(addresses, address)
				}
			}
			c.filters.ContractAddress = addresses
		}
	}
}

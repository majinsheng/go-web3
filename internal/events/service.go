package events

import (
	"encoding/json"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Service manages event subscriptions and broadcasting
type Service struct {
	listener    *Listener
	clients     map[string]*WebSocketClient
	txProcessor *TransactionProcessor
	mu          sync.RWMutex
}

// NewService creates a new event service
func NewService(ethClient *ethclient.Client) *Service {
	listener := NewListener(ethClient)
	return &Service{
		listener:    listener,
		clients:     make(map[string]*WebSocketClient),
		txProcessor: NewTransactionProcessor(listener),
	}
}

// Start starts the event service
func (s *Service) Start() error {
	// Start the event listener
	if err := s.listener.Start(); err != nil {
		return err
	}

	// Subscribe to different event types
	s.setupSubscriptions()

	return nil
}

// Stop stops the event service
func (s *Service) Stop() {
	// Stop the event listener
	s.listener.Stop()

	// Stop the transaction processor
	if s.txProcessor != nil {
		s.txProcessor.Stop()
	}

	// Close all WebSocket connections
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		client.Close()
	}
}

// setupSubscriptions sets up event handlers
func (s *Service) setupSubscriptions() {
	// Handle new blocks
	s.listener.Subscribe(EventTypeNewBlock, func(event Event) {
		s.broadcastEvent(event)
	})

	// Handle new transactions
	s.listener.Subscribe(EventTypeNewTransaction, func(event Event) {
		s.broadcastEvent(event)
	})

	// Handle contract events
	s.listener.Subscribe(EventTypeContractEvent, func(event Event) {
		s.broadcastEvent(event)
	})

	// Set up the transaction processor
	s.txProcessor.OnTransaction(func(info *TransactionInfo) {
		// Log high-value transactions
		// Note: This is just an example - in a real application, you'd handle this differently
		if info.Value.Cmp(big.NewInt(1000000000000000000)) > 0 { // > 1 ETH
			log.Printf("High-value transaction detected: %s, Value: %s ETH",
				info.Transaction.Hash().Hex(),
				new(big.Float).Quo(
					new(big.Float).SetInt(info.Value),
					new(big.Float).SetInt(big.NewInt(1000000000000000000)),
				).Text('f', 4),
			)
		}

		// Create a simplified event for WebSocket clients
		event := map[string]interface{}{
			"type":      "high_value_transaction",
			"hash":      info.Transaction.Hash().Hex(),
			"from":      info.From.Hex(),
			"to":        info.To.Hex(),
			"value":     info.Value.String(),
			"blockHash": info.BlockHash.Hex(),
		}

		// Convert to JSON
		eventJSON, err := json.Marshal(event)
		if err == nil {
			// Broadcast to interested clients
			s.broadcastRawEvent(eventJSON)
		}
	})

	// Start the transaction processor
	s.txProcessor.Start()
}

// SubscribeToContract subscribes to events from a specific contract
func (s *Service) SubscribeToContract(contractAddress string, eventSignatures []string) error {
	address := common.HexToAddress(contractAddress)

	// Convert event signatures to topic hashes if needed
	var topics [][]common.Hash
	if len(eventSignatures) > 0 {
		// This is a simplified example - in a real app, you would create proper topic filters
		topicSet := []common.Hash{}
		for _, sig := range eventSignatures {
			// In reality, you would hash the signature: keccak256(EventName(type1,type2))
			// For simplicity, we're just converting the string to a hash
			topicSet = append(topicSet, common.HexToHash(sig))
		}
		topics = [][]common.Hash{topicSet}
	}

	return s.listener.SubscribeToContractEvents(address, topics)
}

// AddTransactionFilter adds a filter for specific transaction types
func (s *Service) AddTransactionFilter(filter *TransactionFilter) {
	if s.txProcessor != nil {
		s.txProcessor.WithFilter(filter)
	}
}

// AddTransactionHandler adds a custom handler for transaction events
func (s *Service) AddTransactionHandler(handler TransactionHandlerFunc) {
	if s.txProcessor != nil {
		s.txProcessor.OnTransaction(handler)
	}
}

// WatchAddress creates a filter to watch transactions involving a specific address
func (s *Service) WatchAddress(address string) {
	addr := common.HexToAddress(address)
	filter := &TransactionFilter{
		FromAddress: &addr,
	}
	s.AddTransactionFilter(filter)
} // RegisterClient registers a new WebSocket client
func (s *Service) RegisterClient(client *WebSocketClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client.ID] = client

	// Set up a ping/pong to keep the connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := client.SendPing(); err != nil {
					log.Printf("Error sending ping to client %s: %v", client.ID, err)
					s.UnregisterClient(client.ID)
					return
				}
			case <-client.Done():
				return
			}
		}
	}()
}

// UnregisterClient unregisters a WebSocket client
func (s *Service) UnregisterClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[clientID]; ok {
		client.Close()
		delete(s.clients, clientID)
	}
}

// broadcastEvent broadcasts an event to all connected WebSocket clients
func (s *Service) broadcastEvent(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert event to JSON
	eventJSON, err := json.Marshal(map[string]interface{}{
		"type":      event.Type,
		"blockHash": event.BlockHash.Hex(),
		"blockNum":  event.BlockNum,
		"txHash":    event.TxHash.Hex(),
		"data":      event.Data,
	})
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	// Broadcast to all clients
	s.broadcastRawEvent(eventJSON)
}

// broadcastRawEvent broadcasts a raw JSON event to all connected WebSocket clients
func (s *Service) broadcastRawEvent(eventJSON []byte) {
	// Broadcast to all clients
	for _, client := range s.clients {
		err := client.Send(eventJSON)
		if err != nil {
			log.Printf("Error sending event to client %s: %v", client.ID, err)
			// Don't unregister here to avoid deadlock, let the ping/pong handle it
		}
	}
}

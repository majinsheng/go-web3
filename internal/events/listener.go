package events

import (
	"context"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EventType defines the type of Ethereum event
type EventType string

const (
	// EventTypeNewBlock is triggered when a new block is mined
	EventTypeNewBlock EventType = "new_block"
	// EventTypeNewTransaction is triggered when a new transaction is added to a block
	EventTypeNewTransaction EventType = "new_transaction"
	// EventTypeContractEvent is triggered when a contract event is emitted
	EventTypeContractEvent EventType = "contract_event"
)

// Event represents an Ethereum event
type Event struct {
	Type      EventType
	BlockHash common.Hash
	BlockNum  uint64
	TxHash    common.Hash
	Data      interface{}
}

// Handler defines a function that handles events
type Handler func(event Event)

// Listener listens for Ethereum events
type Listener struct {
	client        *ethclient.Client
	handlers      map[EventType][]Handler
	subscriptions []ethereum.Subscription
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewListener creates a new event listener
func NewListener(client *ethclient.Client) *Listener {
	ctx, cancel := context.WithCancel(context.Background())
	return &Listener{
		client:        client,
		handlers:      make(map[EventType][]Handler),
		subscriptions: []ethereum.Subscription{},
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Subscribe adds a handler for a specific event type
func (l *Listener) Subscribe(eventType EventType, handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.handlers[eventType] == nil {
		l.handlers[eventType] = []Handler{}
	}

	l.handlers[eventType] = append(l.handlers[eventType], handler)
}

// SubscribeToNewTransactions adds a specialized handler for new transaction events
// with access to transaction details for easier processing
type TransactionHandler func(tx *types.Transaction, blockHash common.Hash, blockNumber uint64)

// SubscribeToNewTransactions adds a specialized handler for new transaction events
func (l *Listener) SubscribeToNewTransactions(handler TransactionHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a wrapper that converts the transaction handler to a general event handler
	wrapperHandler := func(event Event) {
		if event.Type != EventTypeNewTransaction {
			return
		}

		tx, ok := event.Data.(*types.Transaction)
		if !ok {
			log.Printf("Warning: Expected transaction data but got %T", event.Data)
			return
		}

		// Call the specialized handler with transaction details
		handler(tx, event.BlockHash, event.BlockNum)
	}

	// Register the wrapper handler for transaction events
	if l.handlers[EventTypeNewTransaction] == nil {
		l.handlers[EventTypeNewTransaction] = []Handler{}
	}

	l.handlers[EventTypeNewTransaction] = append(l.handlers[EventTypeNewTransaction], wrapperHandler)
}

// Start begins listening for events
func (l *Listener) Start() error {
	log.Println("Starting Ethereum event listener")

	// Start listening for new blocks
	if err := l.subscribeToNewBlocks(); err != nil {
		return err
	}

	return nil
}

// Stop stops listening for events
func (l *Listener) Stop() {
	log.Println("Stopping Ethereum event listener")

	// Cancel the context
	l.cancel()

	// Unsubscribe from all subscriptions
	for _, sub := range l.subscriptions {
		sub.Unsubscribe()
	}
}

// subscribeToNewBlocks subscribes to new block events
func (l *Listener) subscribeToNewBlocks() error {
	headers := make(chan *types.Header)
	sub, err := l.client.SubscribeNewHead(l.ctx, headers)
	if err != nil {
		return err
	}

	l.subscriptions = append(l.subscriptions, sub)

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("Error in block subscription: %v", err)
				return
			case header := <-headers:
				// Fetch the full block
				block, err := l.client.BlockByHash(l.ctx, header.Hash())
				if err != nil {
					log.Printf("Error getting block: %v", err)
					continue
				}

				// Create an event
				event := Event{
					Type:      EventTypeNewBlock,
					BlockHash: block.Hash(),
					BlockNum:  block.NumberU64(),
					Data:      block,
				}

				// Notify handlers
				l.notifyHandlers(event)

				// Process transactions in the block
				for _, tx := range block.Transactions() {
					txEvent := Event{
						Type:      EventTypeNewTransaction,
						BlockHash: block.Hash(),
						BlockNum:  block.NumberU64(),
						TxHash:    tx.Hash(),
						Data:      tx,
					}
					l.notifyHandlers(txEvent)
				}
			case <-l.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// SubscribeToContractEvents subscribes to events from a specific contract
func (l *Listener) SubscribeToContractEvents(contractAddress common.Address, topics [][]common.Hash) error {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    topics,
	}

	logs := make(chan types.Log)
	sub, err := l.client.SubscribeFilterLogs(l.ctx, query, logs)
	if err != nil {
		return err
	}

	l.subscriptions = append(l.subscriptions, sub)

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("Error in contract event subscription: %v", err)
				return
			case vLog := <-logs:
				// Create an event
				event := Event{
					Type:      EventTypeContractEvent,
					BlockHash: vLog.BlockHash,
					BlockNum:  vLog.BlockNumber,
					TxHash:    vLog.TxHash,
					Data:      vLog,
				}

				// Notify handlers
				l.notifyHandlers(event)
			case <-l.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// notifyHandlers notifies all handlers for a specific event type
func (l *Listener) notifyHandlers(event Event) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, handler := range l.handlers[event.Type] {
		go handler(event)
	}
}

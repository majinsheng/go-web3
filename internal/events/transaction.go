package events

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionInfo represents enriched information about a transaction
type TransactionInfo struct {
	Transaction    *types.Transaction
	BlockHash      common.Hash
	BlockNumber    uint64
	From           common.Address
	To             common.Address
	Value          *big.Int
	GasPrice       *big.Int
	Gas            uint64
	Input          []byte
	IsContractCall bool
}

// TransactionHandlerFunc defines a function that processes transaction info
type TransactionHandlerFunc func(info *TransactionInfo)

// TransactionFilter defines criteria to filter transactions
type TransactionFilter struct {
	FromAddress     *common.Address
	ToAddress       *common.Address
	MinValue        *big.Int
	OnlyContractTxs bool
	MethodSignature string
}

// TransactionProcessor handles processing and filtering of transactions
type TransactionProcessor struct {
	listener *Listener
	ctx      context.Context
	cancel   context.CancelFunc
	handlers []TransactionHandlerFunc
	filter   *TransactionFilter
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(listener *Listener) *TransactionProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransactionProcessor{
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
		handlers: []TransactionHandlerFunc{},
	}
}

// WithFilter sets a filter for transactions
func (p *TransactionProcessor) WithFilter(filter *TransactionFilter) *TransactionProcessor {
	p.filter = filter
	return p
}

// OnTransaction adds a handler for transactions
func (p *TransactionProcessor) OnTransaction(handler TransactionHandlerFunc) *TransactionProcessor {
	p.handlers = append(p.handlers, handler)
	return p
}

// Start begins processing transactions
func (p *TransactionProcessor) Start() {
	p.listener.SubscribeToNewTransactions(func(tx *types.Transaction, blockHash common.Hash, blockNumber uint64) {
		// Create transaction info
		info := &TransactionInfo{
			Transaction: tx,
			BlockHash:   blockHash,
			BlockNumber: blockNumber,
			Value:       tx.Value(),
			GasPrice:    tx.GasPrice(),
			Gas:         tx.Gas(),
			Input:       tx.Data(),
		}

		// Set To address if available (will be nil for contract creation)
		if tx.To() != nil {
			info.To = *tx.To()
		}

		// Try to determine From address if possible
		// Note: Getting the sender requires the correct chain ID and might fail
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err == nil {
			info.From = from
		}

		// Determine if this is a contract call (data length > 0)
		info.IsContractCall = len(tx.Data()) > 0 && tx.To() != nil

		// Apply filter if set
		if p.filter != nil && !p.matchesFilter(info) {
			return
		}

		// Call all handlers
		for _, handler := range p.handlers {
			handler(info)
		}
	})
}

// Stop stops processing transactions
func (p *TransactionProcessor) Stop() {
	p.cancel()
}

// matchesFilter checks if a transaction matches the filter criteria
func (p *TransactionProcessor) matchesFilter(info *TransactionInfo) bool {
	// Check From address
	if p.filter.FromAddress != nil && info.From != *p.filter.FromAddress {
		return false
	}

	// Check To address
	if p.filter.ToAddress != nil && info.To != *p.filter.ToAddress {
		return false
	}

	// Check minimum value
	if p.filter.MinValue != nil && info.Value.Cmp(p.filter.MinValue) < 0 {
		return false
	}

	// Check if only contract transactions are wanted
	if p.filter.OnlyContractTxs && !info.IsContractCall {
		return false
	}

	// Check method signature if specified
	if p.filter.MethodSignature != "" && !matchesMethodSignature(info.Input, p.filter.MethodSignature) {
		return false
	}

	return true
}

// matchesMethodSignature checks if the transaction input data matches a method signature
func matchesMethodSignature(data []byte, signature string) bool {
	// Method signatures are the first 4 bytes (8 hex chars) of the keccak256 hash of the method signature
	// For simplicity, we're just checking if the hex representation of the first 4 bytes matches
	if len(data) < 4 {
		return false
	}

	// Remove 0x prefix if present
	signature = strings.TrimPrefix(signature, "0x")

	// Simple hex comparison of the first 4 bytes (8 hex chars)
	if len(signature) >= 8 {
		methodID := common.Bytes2Hex(data[:4])
		return strings.EqualFold(methodID, signature[:8])
	}

	return false
}

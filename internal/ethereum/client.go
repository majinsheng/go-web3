package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/user/go-web3/internal/config"
)

// Client wraps the Ethereum client with additional functionality
type Client struct {
	*ethclient.Client
	config      *config.EthereumConfig
	privateKey  *ecdsa.PrivateKey
	fromAddress common.Address
}

// NewClient creates a new Ethereum client
func NewClient(cfg *config.EthereumConfig) (*Client, error) {
	client, err := ethclient.Dial(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Client{
		Client:      client,
		config:      cfg,
		privateKey:  privateKey,
		fromAddress: fromAddress,
	}, nil
}

// GetBalance returns the balance of the given address
func (c *Client) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	account := common.HexToAddress(address)
	balance, err := c.Client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// SendTransaction sends a transaction to the given address with the specified amount
func (c *Client) SendTransaction(ctx context.Context, to string, amount *big.Int) (string, error) {
	toAddress := common.HexToAddress(to)

	// Get the nonce for the sender account
	nonce, err := c.Client.PendingNonceAt(ctx, c.fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get suggested gas price
	gasPrice, err := c.Client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to suggest gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		toAddress,
		amount,
		21000, // Gas limit for a standard ETH transfer
		gasPrice,
		nil, // Data
	)

	// Sign the transaction
	chainID := big.NewInt(c.config.ChainID)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = c.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// GetTransactionReceipt gets the receipt of a transaction
func (c *Client) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.Client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	return receipt, nil
}

// GetTransactionByHash gets a transaction by its hash
func (c *Client) GetTransactionByHash(ctx context.Context, txHash string) (*types.Transaction, bool, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := c.Client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get transaction: %w", err)
	}
	return tx, isPending, nil
}

// GetLatestBlockNumber gets the latest block number
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.Client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}
	return blockNumber, nil
}

// GetBlockByNumber gets a block by its number
func (c *Client) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	block, err := c.Client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	return block, nil
}

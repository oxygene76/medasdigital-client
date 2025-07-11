package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	comet "github.com/cometbft/cometbft/rpc/core/types"
	abci "github.com/cometbft/cometbft/abci/types"

	itypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// ClientBuilder helps create blockchain clients with proper configuration
type ClientBuilder struct {
	chainID         string
	rpcEndpoint     string
	keyringBackend  string
	keyringDir      string
	bech32Prefix    string
	addressCodec    AddressCodec
}

// NewClientBuilder creates a new client builder
func NewClientBuilder(config interface{}) *ClientBuilder {
	return &ClientBuilder{
		chainID:        "medasdigital-2",
		rpcEndpoint:    "https://rpc.medas-digital.io:26657",
		keyringBackend: keyring.BackendOS,
		keyringDir:     "",
		bech32Prefix:   "medas",
		addressCodec:   NewBech32AddressCodec("medas"),
	}
}

// WithChainID sets the chain ID
func (cb *ClientBuilder) WithChainID(chainID string) *ClientBuilder {
	cb.chainID = chainID
	return cb
}

// WithRPCEndpoint sets the RPC endpoint
func (cb *ClientBuilder) WithRPCEndpoint(endpoint string) *ClientBuilder {
	cb.rpcEndpoint = endpoint
	return cb
}

// WithKeyring sets keyring configuration
func (cb *ClientBuilder) WithKeyring(backend, dir string) *ClientBuilder {
	cb.keyringBackend = backend
	cb.keyringDir = dir
	return cb
}

// WithBech32Prefix sets the bech32 prefix and creates address codec
func (cb *ClientBuilder) WithBech32Prefix(prefix string) *ClientBuilder {
	cb.bech32Prefix = prefix
	cb.addressCodec = NewBech32AddressCodec(prefix)
	return cb
}

// BuildClient creates a configured blockchain client
func (cb *ClientBuilder) BuildClient() (*Client, error) {
	// Create RPC client
	rpcClient, err := comethttp.New(cb.rpcEndpoint, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	// Create codec
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// Create keyring (v0.50 compatible)
	realAddressCodec := NewBech32AddressCodec(cb.bech32Prefix)
	
	// Get the underlying SDK codec for keyring creation
	var sdkCodec address.Codec
	if bech32Codec, ok := realAddressCodec.(*Bech32AddressCodec); ok {
		sdkCodec = bech32Codec.GetSDKCodec()
	} else {
		return nil, fmt.Errorf("failed to get SDK codec")
	}
	
	kr, err := keyring.New(sdk.KeyringServiceName(), cb.keyringBackend, cb.keyringDir, nil, sdkCodec)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}

	// Create client context
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)).
		WithLegacyAmino(codec.NewLegacyAmino()).
		WithAccountRetriever(authtx.NewAccountRetriever(marshaler)).
		WithBroadcastMode("block").
		WithChainID(cb.chainID).
		WithKeyring(kr).
		WithClient(rpcClient).
		WithAddressCodec(sdkCodec).
		WithValidatorAddressCodec(NewBech32AddressCodec(cb.bech32Prefix + "valoper").(*Bech32AddressCodec).GetSDKCodec()).
		WithConsensusAddressCodec(NewBech32AddressCodec(cb.bech32Prefix + "valcons").(*Bech32AddressCodec).GetSDKCodec())

	return NewClient(clientCtx), nil
}

// KeyManager manages keyring operations
type KeyManager struct {
	keyring      keyring.Keyring
	addressCodec AddressCodec
}

// NewKeyManager creates a new key manager
func NewKeyManager(kr keyring.Keyring, codec AddressCodec) *KeyManager {
	return &KeyManager{
		keyring:      kr,
		addressCodec: codec,
	}
}

// CreateKey creates a new key with the given name and mnemonic
func (km *KeyManager) CreateKey(name, mnemonic string) (*keyring.Record, error) {
	if mnemonic == "" {
		// Generate new mnemonic
		entropy, err := keyring.NewEntropy()
		if err != nil {
			return nil, fmt.Errorf("failed to generate entropy: %w", err)
		}
		mnemonic, err = keyring.NewMnemonic(entropy)
		if err != nil {
			return nil, fmt.Errorf("failed to generate mnemonic: %w", err)
		}
	}

	record, err := km.keyring.NewAccount(name, mnemonic, "", "m/44'/118'/0'/0/0", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return record, nil
}

// GetKey retrieves a key by name
func (km *KeyManager) GetKey(name string) (*keyring.Record, error) {
	record, err := km.keyring.Key(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	return record, nil
}

// GetKeyByAddress retrieves a key by address
func (km *KeyManager) GetKeyByAddress(address string) (*keyring.Record, error) {
	addr, err := km.addressCodec.StringToBytes(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}

	record, err := km.keyring.KeyByAddress(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get key by address: %w", err)
	}
	return record, nil
}

// ListKeys returns all keys in the keyring
func (km *KeyManager) ListKeys() ([]*keyring.Record, error) {
	records, err := km.keyring.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	return records, nil
}

// DeleteKey deletes a key by name
func (km *KeyManager) DeleteKey(name string) error {
	err := km.keyring.Delete(name)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// ExportKey exports a key in armor format
func (km *KeyManager) ExportKey(name, passphrase string) (string, error) {
	armor, err := km.keyring.ExportPrivKeyArmor(name, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to export key: %w", err)
	}
	return armor, nil
}

// ImportKey imports a key from armor format
func (km *KeyManager) ImportKey(name, armor, passphrase string) error {
	err := km.keyring.ImportPrivKey(name, armor, passphrase)
	if err != nil {
		return fmt.Errorf("failed to import key: %w", err)
	}
	return nil
}

// BlockchainMonitor monitors blockchain events and status
type BlockchainMonitor struct {
	client   *Client
	ctx      context.Context
	cancel   context.CancelFunc
	eventCh  chan comet.ResultEvent
	blockCh  chan *comet.ResultBlock
	errorCh  chan error
}

// NewBlockchainMonitor creates a new blockchain monitor
func NewBlockchainMonitor(client *Client) *BlockchainMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockchainMonitor{
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
		eventCh: make(chan comet.ResultEvent, 100),
		blockCh: make(chan *comet.ResultBlock, 100),
		errorCh: make(chan error, 10),
	}
}

// Start starts monitoring blockchain events
func (bm *BlockchainMonitor) Start() error {
	// Subscribe to new blocks
	if err := bm.client.clientCtx.Client.Start(); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	// Start block monitoring goroutine
	go bm.monitorBlocks()

	return nil
}

// Stop stops the blockchain monitor
func (bm *BlockchainMonitor) Stop() {
	bm.cancel()
	close(bm.eventCh)
	close(bm.blockCh)
	close(bm.errorCh)
}

// GetEventChannel returns the event channel
func (bm *BlockchainMonitor) GetEventChannel() <-chan comet.ResultEvent {
	return bm.eventCh
}

// GetBlockChannel returns the block channel
func (bm *BlockchainMonitor) GetBlockChannel() <-chan *comet.ResultBlock {
	return bm.blockCh
}

// GetErrorChannel returns the error channel
func (bm *BlockchainMonitor) GetErrorChannel() <-chan error {
	return bm.errorCh
}

// monitorBlocks monitors new blocks
func (bm *BlockchainMonitor) monitorBlocks() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastHeight int64 = 0

	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			status, err := bm.client.clientCtx.Client.Status(bm.ctx)
			if err != nil {
				bm.errorCh <- fmt.Errorf("failed to get status: %w", err)
				continue
			}

			currentHeight := status.SyncInfo.LatestBlockHeight
			if currentHeight > lastHeight {
				// Fetch the latest block
				block, err := bm.client.clientCtx.Client.Block(bm.ctx, &currentHeight)
				if err != nil {
					bm.errorCh <- fmt.Errorf("failed to get block: %w", err)
					continue
				}

				select {
				case bm.blockCh <- block:
				default:
					// Channel is full, skip this block
				}

				lastHeight = currentHeight
			}
		}
	}
}

// TransactionHelper provides utilities for transaction handling
type TransactionHelper struct {
	client *Client
}

// NewTransactionHelper creates a new transaction helper
func NewTransactionHelper(client *Client) *TransactionHelper {
	return &TransactionHelper{
		client: client,
	}
}

// WaitForInclusion waits for a transaction to be included in a block
func (th *TransactionHelper) WaitForInclusion(txHash string, timeout time.Duration) (*sdk.TxResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction inclusion")
		case <-ticker.C:
			// Try to get the transaction
			txBytes, err := hex.DecodeString(txHash)
			if err != nil {
				return nil, fmt.Errorf("invalid transaction hash: %w", err)
			}

			res, err := th.client.clientCtx.Client.Tx(ctx, txBytes, false)
			if err == nil {
				// Transaction found
				txResponse := &sdk.TxResponse{
					Height:    res.Height,
					TxHash:    txHash,
					Code:      res.TxResult.Code,
					Data:      string(res.TxResult.Data),
					RawLog:    res.TxResult.Log,
					GasWanted: res.TxResult.GasWanted,
					GasUsed:   res.TxResult.GasUsed,
				}
				return txResponse, nil
			}
			
			// Continue polling if transaction not found
		}
	}
}

// BatchTransactions sends multiple transactions in sequence
func (th *TransactionHelper) BatchTransactions(msgs []sdk.Msg, signerName string) ([]*sdk.TxResponse, error) {
	var responses []*sdk.TxResponse

	for i, msg := range msgs {
		// Create transaction
		txBuilder := th.client.clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(msg); err != nil {
			return responses, fmt.Errorf("failed to set message %d: %w", i, err)
		}

		// Set gas limit
		txBuilder.SetGasLimit(200000)

		// Sign and broadcast
		res, err := tx.BroadcastTx(th.client.clientCtx, txBuilder.GetTx())
		if err != nil {
			return responses, fmt.Errorf("failed to broadcast transaction %d: %w", i, err)
		}

		responses = append(responses, res)

		// Wait a bit between transactions
		time.Sleep(1 * time.Second)
	}

	return responses, nil
}

// AddressValidator provides address validation and conversion utilities
type AddressValidator struct {
	addressCodec AddressCodec
}

// NewAddressValidator creates a new address validator
func NewAddressValidator(codec AddressCodec) *AddressValidator {
	return &AddressValidator{
		addressCodec: codec,
	}
}

// ValidateAddress validates a bech32 address
func (av *AddressValidator) ValidateAddress(addr string) error {
	_, err := av.addressCodec.StringToBytes(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	return nil
}

// StringToBytes converts address string to bytes
func (av *AddressValidator) StringToBytes(addr string) ([]byte, error) {
	return av.addressCodec.StringToBytes(addr)
}

// BytesToString converts address bytes to string
func (av *AddressValidator) BytesToString(addr []byte) (string, error) {
	return av.addressCodec.BytesToString(addr)
}

// IsValidBech32 checks if the address has valid bech32 format
func (av *AddressValidator) IsValidBech32(addr string) bool {
	return av.ValidateAddress(addr) == nil
}

// EventParser parses blockchain events
type EventParser struct{}

// NewEventParser creates a new event parser
func NewEventParser() *EventParser {
	return &EventParser{}
}

// ParseClientRegistrationEvent parses client registration events
func (ep *EventParser) ParseClientRegistrationEvent(events []abci.Event) (*itypes.RegisteredClient, error) {
	for _, event := range events {
		if event.Type == "client_registered" {
			var client itypes.RegisteredClient
			
			for _, attr := range event.Attributes {
				key := string(attr.Key)
				value := string(attr.Value)
				
				switch key {
				case "client_id":
					client.ID = value
				case "creator":
					client.Creator = value
				case "capabilities":
					client.Capabilities = strings.Split(value, ",")
				case "metadata":
					client.Metadata = value
				case "status":
					client.Status = value
				}
			}
			
			return &client, nil
		}
	}
	
	return nil, fmt.Errorf("client registration event not found")
}

// ParseAnalysisStoredEvent parses analysis stored events
func (ep *EventParser) ParseAnalysisStoredEvent(events []abci.Event) (*itypes.StoredAnalysis, error) {
	for _, event := range events {
		if event.Type == "analysis_stored" {
			var analysis itypes.StoredAnalysis
			
			for _, attr := range event.Attributes {
				key := string(attr.Key)
				value := string(attr.Value)
				
				switch key {
				case "analysis_id":
					analysis.ID = value
				case "client_id":
					analysis.ClientID = value
				case "creator":
					analysis.Creator = value
				case "analysis_type":
					analysis.AnalysisType = value
				case "tx_hash":
					analysis.TxHash = value
				case "status":
					analysis.Status = value
				}
			}
			
			return &analysis, nil
		}
	}
	
	return nil, fmt.Errorf("analysis stored event not found")
}

// ConnectionManager manages multiple RPC endpoints with failover
type ConnectionManager struct {
	endpoints    []string
	currentIndex int
	maxRetries   int
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(endpoints []string) *ConnectionManager {
	return &ConnectionManager{
		endpoints:    endpoints,
		currentIndex: 0,
		maxRetries:   3,
	}
}

// GetClient creates a client with automatic failover
func (cm *ConnectionManager) GetClient() (*comethttp.HTTP, error) {
	var lastErr error
	
	for i := 0; i < len(cm.endpoints); i++ {
		endpoint := cm.endpoints[cm.currentIndex]
		
		client, err := comethttp.New(endpoint, "/websocket")
		if err == nil {
			return client, nil
		}
		
		lastErr = err
		cm.currentIndex = (cm.currentIndex + 1) % len(cm.endpoints)
	}
	
	return nil, fmt.Errorf("all endpoints failed, last error: %w", lastErr)
}

// AddEndpoint adds a new RPC endpoint
func (cm *ConnectionManager) AddEndpoint(endpoint string) {
	cm.endpoints = append(cm.endpoints, endpoint)
}

// RemoveEndpoint removes an RPC endpoint
func (cm *ConnectionManager) RemoveEndpoint(endpoint string) {
	for i, ep := range cm.endpoints {
		if ep == endpoint {
			cm.endpoints = append(cm.endpoints[:i], cm.endpoints[i+1:]...)
			if cm.currentIndex >= len(cm.endpoints) {
				cm.currentIndex = 0
			}
			break
		}
	}
}

// GasEstimator provides gas estimation utilities
type GasEstimator struct {
	client *Client
}

// NewGasEstimator creates a new gas estimator
func NewGasEstimator(client *Client) *GasEstimator {
	return &GasEstimator{
		client: client,
	}
}

// EstimateGas estimates gas for a transaction
func (ge *GasEstimator) EstimateGas(msgs []sdk.Msg) (uint64, error) {
	txBuilder := ge.client.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Use a default gas limit for simulation
	txBuilder.SetGasLimit(1000000)

	// Simulate the transaction
	simRes, _, err := ge.client.clientCtx.Simulate(txBuilder.GetTx())
	if err != nil {
		return 0, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	// Add a buffer to the estimated gas
	estimatedGas := simRes.GasInfo.GasUsed
	gasWithBuffer := uint64(float64(estimatedGas) * 1.3) // 30% buffer

	return gasWithBuffer, nil
}

// DefaultClient creates a client with default settings
func DefaultClient() (*Client, error) {
	builder := NewClientBuilder(nil)
	return builder.BuildClient()
}

// CreateClientWithDefaults creates a client with default configuration
func CreateClientWithDefaults(chainID, rpcEndpoint string) (*Client, error) {
	builder := NewClientBuilder(nil).
		WithChainID(chainID).
		WithRPCEndpoint(rpcEndpoint).
		WithBech32Prefix("medas")
	
	return builder.BuildClient()
}

// GetSDKVersion returns the SDK version
func GetSDKVersion() string {
	return "v0.50.10"
}

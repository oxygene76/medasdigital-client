package blockchain

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	
	// CometBFT imports (ersetzt Tendermint)
	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/oxygene76/medasdigital-client/pkg/utils"
)

// ClientBuilder helps build blockchain clients with proper configuration (updated for v0.50)
type ClientBuilder struct {
	config *utils.Config
}

// NewClientBuilder creates a new client builder
func NewClientBuilder(config *utils.Config) *ClientBuilder {
	return &ClientBuilder{
		config: config,
	}
}

// BuildClient builds a configured blockchain client (updated for v0.50)
func (cb *ClientBuilder) BuildClient() (*Client, error) {
	// Create client context
	clientCtx, err := cb.createClientContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create client context: %w", err)
	}

	// Create blockchain client
	client := NewClient(clientCtx)

	// Initialize with endpoints
	if err := client.Initialize(
		cb.config.Chain.RPCEndpoint,
		cb.config.Chain.GRPCEndpoint,
		cb.config.Chain.RESTEndpoint,
	); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return client, nil
}

// createClientContext creates a properly configured client context (updated for v0.50)
func (cb *ClientBuilder) createClientContext() (client.Context, error) {
	// Create interface registry and codec (v0.50 style)
	interfaceRegistry := NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// Create address codec (new in v0.50)
	addressCodec := address.NewBech32Codec("medas")

	// Create keyring
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyring.BackendOS,
		cb.config.Client.KeyringDir,
		nil,
		addressCodec, // Address codec is now required
	)
	if err != nil {
		return client.Context{}, fmt.Errorf("failed to create keyring: %w", err)
	}

	// Create CometBFT RPC client (was Tendermint)
	rpcClient, err := http.New(cb.config.Chain.RPCEndpoint, "/websocket")
	if err != nil {
		return client.Context{}, fmt.Errorf("failed to create RPC client: %w", err)
	}

	// Create client context (updated for v0.50)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)).
		WithLegacyAmino(Amino).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync). // Updated broadcast mode
		WithHomeDir(cb.config.Client.DataDir).
		WithKeyring(kr).
		WithChainID(cb.config.Chain.ID).
		WithClient(rpcClient).
		WithAddressCodec(addressCodec). // New in v0.50
		WithValidatorAddressCodec(addressCodec). // New in v0.50
		WithConsensusAddressCodec(addressCodec)  // New in v0.50

	return clientCtx, nil
}

// KeyManager handles key management operations (updated for v0.50)
type KeyManager struct {
	keyring     keyring.Keyring
	addressCodec address.Codec
}

// NewKeyManager creates a new key manager (updated for v0.50)
func NewKeyManager(kr keyring.Keyring, addressCodec address.Codec) *KeyManager {
	return &KeyManager{
		keyring:     kr,
		addressCodec: addressCodec,
	}
}

// CreateKey creates a new key with the given name (updated for v0.50)
func (km *KeyManager) CreateKey(name, mnemonic string) (keyring.Info, error) {
	if mnemonic == "" {
		// Generate new mnemonic
		info, mnemonic, err := km.keyring.NewMnemonic(
			name,
			keyring.English,
			sdk.FullFundraiserPath,
			keyring.DefaultBIP39Passphrase,
			hd.Secp256k1,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create key: %w", err)
		}

		log.Printf("Created new key '%s' with mnemonic (save this!): %s", name, mnemonic)
		return info, nil
	} else {
		// Import from mnemonic
		info, err := km.keyring.NewAccount(
			name,
			mnemonic,
			keyring.DefaultBIP39Passphrase,
			sdk.FullFundraiserPath,
			hd.Secp256k1,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to import key: %w", err)
		}

		log.Printf("Imported key '%s' from mnemonic", name)
		return info, nil
	}
}

// GetKey retrieves a key by name
func (km *KeyManager) GetKey(name string) (keyring.Info, error) {
	info, err := km.keyring.Key(name)
	if err != nil {
		return nil, fmt.Errorf("key '%s' not found: %w", name, err)
	}
	return info, nil
}

// GetKeyByAddress retrieves a key by address (new helper for v0.50)
func (km *KeyManager) GetKeyByAddress(addr string) (keyring.Info, error) {
	// Convert string address to AccAddress
	accAddr, err := km.addressCodec.StringToBytes(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	info, err := km.keyring.KeyByAddress(accAddr)
	if err != nil {
		return nil, fmt.Errorf("key for address '%s' not found: %w", addr, err)
	}
	return info, nil
}

// ListKeys lists all available keys
func (km *KeyManager) ListKeys() ([]keyring.Info, error) {
	keys, err := km.keyring.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	return keys, nil
}

// DeleteKey deletes a key by name
func (km *KeyManager) DeleteKey(name string) error {
	if err := km.keyring.Delete(name); err != nil {
		return fmt.Errorf("failed to delete key '%s': %w", name, err)
	}
	log.Printf("Deleted key '%s'", name)
	return nil
}

// BlockchainMonitor monitors blockchain events and status (updated for v0.50)
type BlockchainMonitor struct {
	client   *Client
	stopCh   chan struct{}
	interval time.Duration
}

// NewBlockchainMonitor creates a new blockchain monitor
func NewBlockchainMonitor(client *Client, interval time.Duration) *BlockchainMonitor {
	return &BlockchainMonitor{
		client:   client,
		stopCh:   make(chan struct{}),
		interval: interval,
	}
}

// Start starts monitoring
func (bm *BlockchainMonitor) Start() {
	go bm.monitorStatus()
	go bm.monitorBlocks()
	log.Println("Blockchain monitor started")
}

// Stop stops monitoring
func (bm *BlockchainMonitor) Stop() {
	close(bm.stopCh)
	log.Println("Blockchain monitor stopped")
}

// monitorStatus periodically checks blockchain status
func (bm *BlockchainMonitor) monitorStatus() {
	ticker := time.NewTicker(bm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := bm.client.GetChainStatus()
			if err != nil {
				log.Printf("Error getting chain status: %v", err)
				continue
			}

			log.Printf("Chain status: height=%d, syncing=%t",
				status.SyncInfo.LatestBlockHeight,
				status.SyncInfo.CatchingUp,
			)

		case <-bm.stopCh:
			return
		}
	}
}

// monitorBlocks subscribes to and monitors new blocks (updated for CometBFT)
func (bm *BlockchainMonitor) monitorBlocks() {
	ctx := context.Background()
	
	// Subscribe to new block events using CometBFT
	eventsCh, err := bm.client.clientCtx.Client.Subscribe(
		ctx, 
		"medasdigital-client-monitor", 
		"tm.event='NewBlock'",
	)
	if err != nil {
		log.Printf("Error subscribing to blocks: %v", err)
		return
	}

	go func() {
		defer func() {
			if err := bm.client.clientCtx.Client.UnsubscribeAll(ctx, "medasdigital-client-monitor"); err != nil {
				log.Printf("Error unsubscribing: %v", err)
			}
		}()
		
		for {
			select {
			case event := <-eventsCh:
				// Handle new block event
				log.Printf("New block event received: height=%v", event.Data)
				
			case <-bm.stopCh:
				return
			}
		}
	}()
}

// TransactionHelper provides helper functions for transaction operations (updated for v0.50)
type TransactionHelper struct {
	client *Client
}

// NewTransactionHelper creates a new transaction helper
func NewTransactionHelper(client *Client) *TransactionHelper {
	return &TransactionHelper{
		client: client,
	}
}

// SendAndWait sends a transaction and waits for confirmation (updated for v0.50)
func (th *TransactionHelper) SendAndWait(msg sdk.Msg, timeout time.Duration) (*coretypes.ResultTx, error) {
	// Send transaction
	res, err := th.client.sendTransaction(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("Transaction sent: %s", res.TxHash)

	// Wait for confirmation
	tx, err := th.client.WaitForTransaction(res.TxHash, timeout)
	if err != nil {
		return nil, fmt.Errorf("transaction confirmation failed: %w", err)
	}

	log.Printf("Transaction confirmed in block %d", tx.Height)
	return tx, nil
}

// BatchSend sends multiple transactions in sequence
func (th *TransactionHelper) BatchSend(msgs []sdk.Msg, delay time.Duration) ([]*sdk.TxResponse, error) {
	var responses []*sdk.TxResponse

	for i, msg := range msgs {
		if i > 0 && delay > 0 {
			time.Sleep(delay)
		}

		res, err := th.client.sendTransaction(msg)
		if err != nil {
			return responses, fmt.Errorf("failed to send transaction %d: %w", i, err)
		}

		responses = append(responses, res)
		log.Printf("Batch transaction %d/%d sent: %s", i+1, len(msgs), res.TxHash)
	}

	return responses, nil
}

// AddressValidator provides address validation utilities (updated for v0.50)
type AddressValidator struct {
	addressCodec address.Codec
}

// NewAddressValidator creates a new address validator (updated for v0.50)
func NewAddressValidator(addressCodec address.Codec) *AddressValidator {
	return &AddressValidator{
		addressCodec: addressCodec,
	}
}

// ValidateBech32 validates a bech32 address (updated for v0.50)
func (av *AddressValidator) ValidateBech32(address string) error {
	_, err := av.addressCodec.StringToBytes(address)
	if err != nil {
		return fmt.Errorf("invalid bech32 address: %w", err)
	}
	return nil
}

// ValidatePrefix validates that an address has the expected prefix
func (av *AddressValidator) ValidatePrefix(address, expectedPrefix string) error {
	if !strings.HasPrefix(address, expectedPrefix) {
		return fmt.Errorf("address does not have expected prefix '%s'", expectedPrefix)
	}
	return nil
}

// GetAddressPrefix extracts the prefix from a bech32 address
func (av *AddressValidator) GetAddressPrefix(address string) (string, error) {
	parts := strings.Split(address, "1")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid bech32 address format")
	}
	return parts[0], nil
}

// ConvertAddressFormat converts address between string and bytes (new in v0.50)
func (av *AddressValidator) ConvertAddressFormat(address string) ([]byte, error) {
	return av.addressCodec.StringToBytes(address)
}

// FormatAddress converts address bytes to string (new in v0.50)
func (av *AddressValidator) FormatAddress(addressBytes []byte) (string, error) {
	return av.addressCodec.BytesToString(addressBytes)
}

// EventParser provides utilities for parsing blockchain events (updated for v0.50)
type EventParser struct{}

// NewEventParser creates a new event parser
func NewEventParser() *EventParser {
	return &EventParser{}
}

// ParseClientRegistrationEvent parses a client registration event (updated for v0.50)
func (ep *EventParser) ParseClientRegistrationEvent(events []sdk.Event) (string, error) {
	for _, event := range events {
		if event.Type == EventTypeRegisterClient {
			for _, attr := range event.Attributes {
				if attr.Key == AttributeKeyClientID {
					return attr.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("client registration event not found")
}

// ParseAnalysisStorageEvent parses an analysis storage event (updated for v0.50)
func (ep *EventParser) ParseAnalysisStorageEvent(events []sdk.Event) (string, string, error) {
	var clientID, analysisType string

	for _, event := range events {
		if event.Type == EventTypeStoreAnalysis {
			for _, attr := range event.Attributes {
				switch attr.Key {
				case AttributeKeyClientID:
					clientID = attr.Value
				case AttributeKeyAnalysisType:
					analysisType = attr.Value
				}
			}
		}
	}

	if clientID == "" || analysisType == "" {
		return "", "", fmt.Errorf("analysis storage event not found or incomplete")
	}

	return clientID, analysisType, nil
}

// ConnectionManager manages blockchain connections with retry logic (updated for v0.50)
type ConnectionManager struct {
	endpoints []string
	current   int
	maxRetries int
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(endpoints []string, maxRetries int) *ConnectionManager {
	return &ConnectionManager{
		endpoints:  endpoints,
		maxRetries: maxRetries,
	}
}

// GetActiveEndpoint returns the currently active endpoint
func (cm *ConnectionManager) GetActiveEndpoint() string {
	if len(cm.endpoints) == 0 {
		return ""
	}
	return cm.endpoints[cm.current]
}

// SwitchEndpoint switches to the next available endpoint
func (cm *ConnectionManager) SwitchEndpoint() bool {
	if len(cm.endpoints) <= 1 {
		return false
	}

	cm.current = (cm.current + 1) % len(cm.endpoints)
	log.Printf("Switched to endpoint: %s", cm.GetActiveEndpoint())
	return true
}

// TestConnection tests connection to current endpoint (updated for CometBFT)
func (cm *ConnectionManager) TestConnection(timeout time.Duration) error {
	endpoint := cm.GetActiveEndpoint()
	if endpoint == "" {
		return fmt.Errorf("no endpoints available")
	}

	client, err := http.New(endpoint, "/websocket")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = client.Status(ctx)
	return err
}

// FindWorkingEndpoint finds the first working endpoint
func (cm *ConnectionManager) FindWorkingEndpoint(timeout time.Duration) (string, error) {
	for i := 0; i < len(cm.endpoints); i++ {
		if err := cm.TestConnection(timeout); err == nil {
			return cm.GetActiveEndpoint(), nil
		}
		
		if !cm.SwitchEndpoint() {
			break
		}
	}

	return "", fmt.Errorf("no working endpoints found")
}

// GasEstimator provides gas estimation utilities (updated for v0.50)
type GasEstimator struct {
	client *Client
}

// NewGasEstimator creates a new gas estimator
func NewGasEstimator(client *Client) *GasEstimator {
	return &GasEstimator{
		client: client,
	}
}

// EstimateWithMargin estimates gas with a safety margin (updated for v0.50)
func (ge *GasEstimator) EstimateWithMargin(msg sdk.Msg, margin float64) (uint64, error) {
	baseGas, err := ge.client.EstimateGas(msg)
	if err != nil {
		return 0, err
	}

	// Add margin
	gasWithMargin := uint64(float64(baseGas) * (1.0 + margin))
	
	// Ensure minimum gas
	if gasWithMargin < 100000 {
		gasWithMargin = 100000
	}

	return gasWithMargin, nil
}

// EstimateBatch estimates gas for multiple messages
func (ge *GasEstimator) EstimateBatch(msgs []sdk.Msg, margin float64) (uint64, error) {
	totalGas := uint64(0)

	for _, msg := range msgs {
		gas, err := ge.EstimateWithMargin(msg, margin)
		if err != nil {
			return 0, err
		}
		totalGas += gas
	}

	return totalGas, nil
}

// V0.50 specific utility functions

// CreateClientWithDefaults creates a client with default v0.50 settings
func CreateClientWithDefaults(chainID, rpcEndpoint string) (*Client, error) {
	// Create default config
	config := &utils.Config{
		Chain: utils.ChainConfig{
			ID:          chainID,
			RPCEndpoint: rpcEndpoint,
		},
		Client: utils.ClientConfig{
			KeyringDir: "./.medasdigital/keyring",
			DataDir:    "./.medasdigital/data",
		},
	}

	// Build client
	builder := NewClientBuilder(config)
	return builder.BuildClient()
}

// GetSDKVersion returns the SDK version being used
func GetSDKVersion() string {
	return "v0.50.x"
}

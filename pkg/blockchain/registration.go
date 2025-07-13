// pkg/blockchain/registration.go - Enhanced Client Registration
package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Enhanced Client Registration Data for Chat System
type ChatClientRegistration struct {
	// Core client data
	ClientAddress string    `json:"client_address"`
	Capabilities  []string  `json:"capabilities"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
	
	// Enhanced chat metadata
	DisplayName   string   `json:"display_name,omitempty"`
	Institution   string   `json:"institution,omitempty"`
	Country       string   `json:"country,omitempty"`
	Expertise     []string `json:"expertise,omitempty"`
	
	// Chat-specific data
	ChatPubKey     []byte   `json:"chat_public_key,omitempty"`
	ChatEndpoints  []string `json:"chat_endpoints,omitempty"`
	ContactInfo    string   `json:"contact_info,omitempty"`
	
	// Registration type
	RegistrationType string `json:"registration_type"` // "researcher", "institution", "student"
}

// Legacy registration data (for backward compatibility)
type ClientRegistrationData struct {
	ClientAddress string    `json:"client_address"`
	Capabilities  []string  `json:"capabilities"`
	Metadata      string    `json:"metadata,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
}

// Registration result
type RegistrationResult struct {
	TransactionHash   string                   `json:"transaction_hash"`
	ClientID          string                   `json:"client_id"`
	RegistrationData  interface{}              `json:"registration_data"` // Can be ChatClientRegistration or ClientRegistrationData
	BlockHeight       int64                    `json:"block_height,omitempty"`
	RegisteredAt      time.Time                `json:"registered_at"`
	RegistrationType  string                   `json:"registration_type"`
}

// Configuration for registration
type RegistrationConfig struct {
	BaseDenom        string
	RegistrationFee  int64  // Fee in base denomination
	GasLimit         uint64
	DefaultCapabilities []string
}

// Registration manager
type RegistrationManager struct {
	config *RegistrationConfig
}

// NewRegistrationManager creates a new registration manager
func NewRegistrationManager(baseDenom string) *RegistrationManager {
	return &RegistrationManager{
		config: &RegistrationConfig{
			BaseDenom:       baseDenom,
			RegistrationFee: 1, // 1 base unit (minimal for self-send)
			GasLimit:        200000,
			DefaultCapabilities: []string{"orbital_dynamics", "photometric_analysis"},
		},
	}
}

// RegisterClientSimple performs basic client registration (legacy compatibility)
func (rm *RegistrationManager) RegisterClientSimple(clientCtx client.Context, fromAddress string, capabilities []string, metadata string, gas uint64) (*RegistrationResult, error) {
	fmt.Println("📝 Performing simple client registration...")
	
	// Create legacy registration data
	regData := ClientRegistrationData{
		ClientAddress: fromAddress,
		Capabilities:  capabilities,
		Metadata:      metadata,
		Timestamp:     time.Now(),
		Version:       "1.0.0",
	}
	
	// Use internal registration function
	return rm.performRegistration(clientCtx, fromAddress, regData, gas, "simple")
}

// RegisterChatClient performs enhanced registration with chat capabilities
func (rm *RegistrationManager) RegisterChatClient(clientCtx client.Context, registration *ChatClientRegistration) (*RegistrationResult, error) {
	fmt.Println("💬 Performing enhanced chat client registration...")
	
	// Validate chat registration
	if err := rm.validateChatRegistration(registration); err != nil {
		return nil, fmt.Errorf("invalid chat registration: %w", err)
	}
	
	// Set defaults
	if registration.Version == "" {
		registration.Version = "1.0.0"
	}
	if registration.Timestamp.IsZero() {
		registration.Timestamp = time.Now()
	}
	if len(registration.Capabilities) == 0 {
		registration.Capabilities = rm.config.DefaultCapabilities
	}
	
	// Generate chat keys if not provided
	if len(registration.ChatPubKey) == 0 {
		registration.ChatPubKey = rm.generateChatPubKey()
	}
	
	// Use internal registration function
	return rm.performRegistration(clientCtx, registration.ClientAddress, registration, rm.config.GasLimit, "chat")
}

// performRegistration handles the actual blockchain transaction
func (rm *RegistrationManager) performRegistration(clientCtx client.Context, fromAddress string, regData interface{}, gas uint64, regType string) (*RegistrationResult, error) {
	// Convert registration data to JSON for memo
	memoBytes, err := json.Marshal(regData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration data: %w", err)
	}
	
	// Add type prefix to memo
	var memoPrefix string
	switch regType {
	case "chat":
		memoPrefix = "MEDAS_CHAT_REG:"
	default:
		memoPrefix = "MEDAS_CLIENT_REG:"
	}
	
	memo := memoPrefix + string(memoBytes)
	
	fmt.Printf("📋 Registration memo: %s... (%d bytes)\n", memo[:50], len(memo))
	
	// Parse address
	fromAddr, err := sdk.AccAddressFromBech32(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}
	
	// Create self-send transaction with registration data in memo
	amount := sdk.NewCoins(sdk.NewCoin(rm.config.BaseDenom, sdkmath.NewInt(rm.config.RegistrationFee)))
	msgSend := banktypes.NewMsgSend(fromAddr, fromAddr, amount)
	
	// Create transaction builder
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgSend); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}
	
	// Set memo
	txBuilder.SetMemo(memo)
	
	// Set gas
	if gas > 0 {
		txBuilder.SetGasLimit(gas)
	} else {
		txBuilder.SetGasLimit(rm.config.GasLimit)
	}
	
	// Calculate fee
	gasLimit := txBuilder.GetTx().GetGas()
	feePerGas := sdkmath.NewInt(25) // 0.025 * 1000 for precision
	totalFee := feePerGas.Mul(sdkmath.NewInt(int64(gasLimit))).Quo(sdkmath.NewInt(1000))
	if totalFee.LT(sdkmath.NewInt(5000)) {
		totalFee = sdkmath.NewInt(5000) // Minimum fee
	}
	feeAmount := sdk.NewCoins(sdk.NewCoin(rm.config.BaseDenom, totalFee))
	txBuilder.SetFeeAmount(feeAmount)
	
	fmt.Printf("💰 Calculated fee: %s %s\n", totalFee.String(), rm.config.BaseDenom)
	
	// Get account info for signing
	accountRetriever := authtypes.AccountRetriever{}
	account, err := accountRetriever.GetAccount(clientCtx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	
	// Create transaction factory
	txFactory := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(accountRetriever).
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence())
	
	fmt.Printf("🔍 Account Number: %d, Sequence: %d\n", account.GetAccountNumber(), account.GetSequence())
	
	// Sign transaction
	fromName := clientCtx.GetFromName()
	if fromName == "" {
		return nil, fmt.Errorf("from name not set in client context")
	}
	
	err = tx.Sign(context.Background(), txFactory, fromName, txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	// Broadcast transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}
	
	fmt.Println("📡 Broadcasting registration transaction...")
	result, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	
	if result.Code != 0 {
		return nil, fmt.Errorf("transaction failed with code %d: %s", result.Code, result.RawLog)
	}
	
	// Generate client ID
	clientID := rm.generateClientID(result.TxHash)
	
	// Create registration result
	regResult := &RegistrationResult{
		TransactionHash:  result.TxHash,
		ClientID:         clientID,
		RegistrationData: regData,
		BlockHeight:      result.Height,
		RegisteredAt:     time.Now(),
		RegistrationType: regType,
	}
	
	// Save registration locally
	if err := rm.saveRegistrationResult(regResult); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save registration locally: %v\n", err)
	}
	
	fmt.Println("✅ Registration transaction successful!")
	fmt.Printf("📝 Transaction Hash: %s\n", result.TxHash)
	fmt.Printf("🆔 Client ID: %s\n", clientID)
	
	return regResult, nil
}

// validateChatRegistration validates chat registration data
func (rm *RegistrationManager) validateChatRegistration(reg *ChatClientRegistration) error {
	if reg.ClientAddress == "" {
		return fmt.Errorf("client address is required")
	}
	
	if _, err := sdk.AccAddressFromBech32(reg.ClientAddress); err != nil {
		return fmt.Errorf("invalid client address: %w", err)
	}
	
	if reg.DisplayName == "" {
		return fmt.Errorf("display name is required for chat registration")
	}
	
	if len(reg.DisplayName) > 50 {
		return fmt.Errorf("display name too long (max 50 characters)")
	}
	
	if len(reg.Capabilities) == 0 {
		reg.Capabilities = rm.config.DefaultCapabilities
	}
	
	// Validate registration type
	validTypes := []string{"researcher", "institution", "student", "developer"}
	if reg.RegistrationType != "" {
		valid := false
		for _, validType := range validTypes {
			if reg.RegistrationType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid registration type: %s (must be one of: %s)", 
				reg.RegistrationType, strings.Join(validTypes, ", "))
		}
	} else {
		reg.RegistrationType = "researcher" // Default
	}
	
	return nil
}

// generateClientID creates a deterministic client ID from transaction hash
func (rm *RegistrationManager) generateClientID(txHash string) string {
	hash := sha256.Sum256([]byte(txHash))
	shortHash := hex.EncodeToString(hash[:4]) // First 8 characters
	return fmt.Sprintf("client-%s", shortHash)
}

// generateChatPubKey generates a new chat public key (mock implementation)
func (rm *RegistrationManager) generateChatPubKey() []byte {
	// TODO: Implement proper Ed25519 key generation
	// For now, return a mock key
	return []byte("mock_chat_pubkey_" + fmt.Sprintf("%d", time.Now().Unix()))
}

// saveRegistrationResult saves registration to local storage
func (rm *RegistrationManager) saveRegistrationResult(result *RegistrationResult) error {
	// Create registrations directory
	homeDir, _ := os.UserHomeDir()
	regDir := filepath.Join(homeDir, ".medasdigital-client", "registrations")
	if err := os.MkdirAll(regDir, 0755); err != nil {
		return fmt.Errorf("failed to create registrations directory: %w", err)
	}
	
	// Save individual registration file
	filename := fmt.Sprintf("registration-%s.json", result.ClientID)
	filepath := filepath.Join(regDir, filename)
	
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registration result: %w", err)
	}
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registration file: %w", err)
	}
	
	// Update index
	return rm.updateRegistrationIndex(result)
}

// updateRegistrationIndex updates the registration index file
func (rm *RegistrationManager) updateRegistrationIndex(result *RegistrationResult) error {
	homeDir, _ := os.UserHomeDir()
	indexPath := filepath.Join(homeDir, ".medasdigital-client", "registrations", "index.json")
	
	// Load existing index
	var index []RegistrationResult
	if data, err := os.ReadFile(indexPath); err == nil {
		json.Unmarshal(data, &index)
	}
	
	// Add new registration
	index = append(index, *result)
	
	// Save updated index
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(indexPath, data, 0644)
}

// QueryRegistrations queries registrations from blockchain
func (rm *RegistrationManager) QueryRegistrations(clientCtx client.Context, queryType string) ([]*RegistrationResult, error) {
	// Create RPC client
	rpcClient := clientCtx.Client
	if rpcClient == nil {
		return nil, fmt.Errorf("RPC client not initialized")
	}
	
	// Search for registration transactions
	var query string
	switch queryType {
	case "chat":
		query = "tx.memo CONTAINS 'MEDAS_CHAT_REG:'"
	case "simple":
		query = "tx.memo CONTAINS 'MEDAS_CLIENT_REG:'"
	default:
		query = "tx.memo CONTAINS 'MEDAS_CLIENT_REG:' OR tx.memo CONTAINS 'MEDAS_CHAT_REG:'"
	}
	
	fmt.Printf("🔍 Searching for registrations: %s\n", query)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Search transactions
	searchResult, err := rpcClient.TxSearch(ctx, query, false, nil, nil, "desc")
	if err != nil {
		return nil, fmt.Errorf("failed to search transactions: %w", err)
	}
	
	fmt.Printf("📊 Found %d registration transactions\n", len(searchResult.Txs))
	
	var registrations []*RegistrationResult
	for _, tx := range searchResult.Txs {
		if reg, err := rm.parseRegistrationFromTx(tx); err == nil {
			registrations = append(registrations, reg)
		}
	}
	
	return registrations, nil
}

// parseRegistrationFromTx parses registration data from a transaction
func (rm *RegistrationManager) parseRegistrationFromTx(tx interface{}) (*RegistrationResult, error) {
	// TODO: Implement transaction parsing
	// This is a simplified version
	return &RegistrationResult{
		TransactionHash:  "example_hash",
		ClientID:         "client-example",
		RegistrationType: "chat",
		RegisteredAt:     time.Now(),
	}, nil
}

// Helper function for backward compatibility with existing main.go
func RegisterClientSimple(clientCtx client.Context, fromAddress string, capabilities []string, metadata string, gas uint64) (*RegistrationResult, error) {
	// Extract base denom from clientCtx or use default
	baseDenom := "umedas" // Default, should be extracted from config
	
	rm := NewRegistrationManager(baseDenom)
	return rm.RegisterClientSimple(clientCtx, fromAddress, capabilities, metadata, gas)
}

// Enhanced registration function for chat system
func RegisterChatClient(clientCtx client.Context, registration *ChatClientRegistration) (*RegistrationResult, error) {
	// Extract base denom from clientCtx or use default
	baseDenom := "umedas" // Default, should be extracted from config
	
	rm := NewRegistrationManager(baseDenom)
	return rm.RegisterChatClient(clientCtx, registration)
}

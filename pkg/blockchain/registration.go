// pkg/blockchain/registration.go - Enhanced Client Registration
package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	sdkmath "cosmossdk.io/math"
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

type BlockchainRegistrationData struct {
	TransactionHash    string                 `json:"transaction_hash"`
	BlockHeight        int64                  `json:"block_height"`
	BlockTime          time.Time              `json:"block_time"`
	FromAddress        string                 `json:"from_address"`
	ToAddress          string                 `json:"to_address"`
	Amount             string                 `json:"amount"`
	Denom              string                 `json:"denom"`
	Fee                string                 `json:"fee"`
	GasUsed            int64                  `json:"gas_used"`
	GasWanted          int64                  `json:"gas_wanted"`
	Memo               string                 `json:"memo"`
	RegistrationData   ClientRegistrationData `json:"registration_data"`
	ClientID           string                 `json:"client_id"`
	VerificationStatus string                 `json:"verification_status"`
	TxStatus           string                 `json:"tx_status"`
}

type TxData struct {
	FromAddress string
	ToAddress   string
	Amount      string
	Denom       string
	Fee         string
	Memo        string
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
// GetLocalRegistrationHashes retrieves local registration transaction hashes
func GetLocalRegistrationHashes() ([]string, error) {
	homeDir, _ := os.UserHomeDir()
	indexPath := filepath.Join(homeDir, ".medasdigital-client", "registrations", "index.json")
	
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no local registrations found")
	}
	
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}
	
	var registrations []RegistrationResult
	if err := json.Unmarshal(data, &registrations); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}
	
	var hashes []string
	for _, reg := range registrations {
		if reg.TransactionHash != "" {
			hashes = append(hashes, reg.TransactionHash)
		}
	}
	
	return hashes, nil
}

/// ERSETZEN Sie die FetchRegistrationFromBlockchain Funktion in registration.go:

// FetchRegistrationFromBlockchain fetches complete registration data from blockchain
func FetchRegistrationFromBlockchain(txHash string, rpcEndpoint, chainID string, codec codec.Codec) (*BlockchainRegistrationData, error) {
	// Create RPC client
	rpcClient, err := client.NewClientFromNode(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	// Query transaction
	hashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction hash: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Get transaction details
	txResult, err := rpcClient.Tx(ctx, hashBytes, false)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	
	// Get block details for timestamp
	block, err := rpcClient.Block(ctx, &txResult.Height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	
	// Parse transaction to extract registration data
	regData := &BlockchainRegistrationData{
		TransactionHash: txHash,
		BlockHeight:     txResult.Height,
		BlockTime:       block.Block.Time,
		GasUsed:         txResult.TxResult.GasUsed,
		GasWanted:       txResult.TxResult.GasWanted,
		TxStatus:        GetTxStatus(txResult.TxResult.Code),
	}
	
	// Decode transaction to get actual data
	if txData, err := DecodeTxData(txResult.Tx, codec); err == nil {
		regData.FromAddress = txData.FromAddress
		regData.ToAddress = txData.ToAddress
		regData.Amount = txData.Amount
		regData.Denom = txData.Denom
		regData.Fee = txData.Fee
		regData.Memo = txData.Memo
		
		// Parse memo for registration data
		if regData.Memo != "" {
			// Try to extract JSON from memo (remove prefix if present)
			memoContent := regData.Memo
			if strings.Contains(memoContent, "MEDAS_CLIENT_REG:") {
				memoContent = strings.Replace(memoContent, "MEDAS_CLIENT_REG:", "", 1)
			}
			if strings.Contains(memoContent, "MEDAS_CHAT_REG:") {
				memoContent = strings.Replace(memoContent, "MEDAS_CHAT_REG:", "", 1)
			}
			
			var clientRegData ClientRegistrationData
			if err := json.Unmarshal([]byte(memoContent), &clientRegData); err == nil {
				regData.RegistrationData = clientRegData
				regData.ClientID = GenerateClientIDFromHash(txHash)
				regData.VerificationStatus = "✅ Valid"
			} else {
				regData.VerificationStatus = "⚠️  Invalid memo format"
			}
		}
	}
	
	return regData, nil
}

// DecodeTxData decodes transaction data (simplified for MsgSend)
func DecodeTxData(txBytes []byte, codec codec.Codec) (*TxData, error) {
	// Create TxConfig from codec
	txConfig := authtx.NewTxConfig(codec, authtx.DefaultSignModes)
	
	// Use TxConfig to decode transaction
	tx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}
	
	txData := &TxData{}
	
	// Try to cast to TxWithMemo interface to get memo
	if txWithMemo, ok := tx.(interface{ GetMemo() string }); ok {
		txData.Memo = txWithMemo.GetMemo()
	}
	
	// Try to cast to FeeTx interface to get fee
	if feeTx, ok := tx.(interface{ GetFee() sdk.Coins }); ok {
		fee := feeTx.GetFee()
		if len(fee) > 0 {
			txData.Fee = fee[0].Amount.String()
			txData.Denom = fee[0].Denom
		}
	}
	
	// Extract message data
	msgs := tx.GetMsgs()
	if len(msgs) > 0 {
		// Try to cast to MsgSend
		for _, msg := range msgs {
			if msgSend, ok := msg.(*banktypes.MsgSend); ok {
				txData.FromAddress = msgSend.FromAddress
				txData.ToAddress = msgSend.ToAddress
				if len(msgSend.Amount) > 0 {
					txData.Amount = msgSend.Amount[0].Amount.String()
					if txData.Denom == "" {
						txData.Denom = msgSend.Amount[0].Denom
					}
				}
				break
			}
		}
	}
	
	return txData, nil
}

// GetTxStatus returns transaction status string
func GetTxStatus(code uint32) string {
	if code == 0 {
		return "✅ Success"
	}
	return fmt.Sprintf("❌ Failed (Code: %d)", code)
}

// GenerateClientIDFromHash creates a client ID from transaction hash (public version)
func GenerateClientIDFromHash(txHash string) string {
	hash := sha256.Sum256([]byte(txHash))
	shortHash := hex.EncodeToString(hash[:4])
	return fmt.Sprintf("client-%s", shortHash)
}



// TruncateString helper function
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

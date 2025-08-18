package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"      // ← NEU
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types" // ← NEU
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	comet "github.com/cometbft/cometbft/rpc/core/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec/types"

	itypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// Client handles blockchain communication for MedasDigital
type Client struct {
	clientCtx  client.Context
	txFactory  tx.Factory
	codec      *Codec
	monitoring bool
}

// NewClient creates a new blockchain client
func NewClient(clientCtx client.Context) *Client {
	// Create codec
	codec := NewCodec()
	
	// Create transaction factory with basic configuration
	// TxConfig wird vom clientCtx bereitgestellt, nicht vom codec
	txFactory := tx.Factory{}.
		WithKeybase(clientCtx.Keyring).
		WithChainID(clientCtx.ChainID).
		WithSimulateAndExecute(true).
		WithGas(200000)
	
	// Set TxConfig from clientCtx if available
	if clientCtx.TxConfig != nil {
		txFactory = txFactory.WithTxConfig(clientCtx.TxConfig)
	}
	
	// Set AccountRetriever from clientCtx if available
	if clientCtx.AccountRetriever != nil {
		txFactory = txFactory.WithAccountRetriever(clientCtx.AccountRetriever)
	}

	return &Client{
		clientCtx:  clientCtx,
		txFactory:  txFactory,
		codec:      codec,
		monitoring: false,
	}
}

// RegisterClient registers a new analysis client on the blockchain
func (c *Client) RegisterClient(creator string, capabilities []string, metadata map[string]interface{}) (string, error) {
	// Convert metadata to JSON
	metadataBytes, err := c.codec.MarshalJSON(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create registration message
	msg := &MsgRegisterClient{
		Creator:      creator,
		Capabilities: capabilities,
		Metadata:     string(metadataBytes),
	}

	// Send transaction
	res, err := c.sendTransaction(msg, creator)
	if err != nil {
		return "", fmt.Errorf("failed to register client: %w", err)
	}

	// Extract client ID from events
	clientID, err := c.extractClientIDFromEvents(res.Events)
	if err != nil {
		return "", fmt.Errorf("failed to extract client ID: %w", err)
	}

	return clientID, nil
}

// StoreAnalysisResult stores analysis results on the blockchain
func (c *Client) StoreAnalysisResult(creator, clientID, analysisType string, data []byte, height int64, txHash string) error {
	// Create analysis storage message
	msg := &MsgStoreAnalysis{
		Creator:      creator,
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         string(data), // Convert []byte to string
		BlockHeight:  height,
		TxHash:       txHash,
	}

	// Send transaction
	_, err := c.sendTransaction(msg, creator)
	if err != nil {
		return fmt.Errorf("failed to store analysis result: %w", err)
	}

	return nil
}

// UpdateClient updates client information
func (c *Client) UpdateClient(creator, clientID string, capabilities []string, metadata map[string]interface{}) error {
	// Convert metadata to JSON
	metadataBytes, err := c.codec.MarshalJSON(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create update message
	msg := &MsgUpdateClient{
		Creator:         creator,
		ClientID:        clientID,
		NewCapabilities: capabilities,
		NewMetadata:     string(metadataBytes),
	}

	// Send transaction
	_, err = c.sendTransaction(msg, creator)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}

// DeactivateClient deactivates a client
func (c *Client) DeactivateClient(creator, clientID string) error {
	// Create deactivation message
	msg := &MsgDeactivateClient{
		Creator:  creator,
		ClientID: clientID,
	}

	// Send transaction
	_, err := c.sendTransaction(msg, creator)
	if err != nil {
		return fmt.Errorf("failed to deactivate client: %w", err)
	}

	return nil
}

// sendTransaction signs and broadcasts a transaction
func (c *Client) sendTransaction(msg sdk.Msg, signerName string) (*sdk.TxResponse, error) {
	// Create transaction builder
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Estimate gas
	gasLimit, err := c.estimateGas([]sdk.Msg{msg})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Set gas limit with buffer
	txBuilder.SetGasLimit(gasLimit)

	// Set fee (optional - can be calculated from gas price)
	// For now, we'll let the node calculate the fee

	// Sign transaction - FIXED: Added context parameter for v0.50
	err = tx.Sign(context.Background(), c.txFactory, signerName, txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Encode transaction
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast transaction
	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return res, nil
}


// Ersetzen Sie die estimateGas Funktion in pkg/blockchain/client.go:

// estimateGas estimates gas for a transaction - clean version without fallback
func (c *Client) estimateGas(msgs []sdk.Msg) (uint64, error) {
	fmt.Println("🔧 Starting gas estimation...")
	
	// ✅ DEBUG: Keyring-Informationen anzeigen
	fmt.Printf("🔑 Keyring Info:\n")
	fmt.Printf("   Keyring Backend: %s\n", c.clientCtx.Keyring.Backend())
	
	// Testen ob der Keyring die benötigten Keys hat
	keys, err := c.clientCtx.Keyring.List()
	if err != nil {
		fmt.Printf("   ❌ Keyring List Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Available Keys: %d\n", len(keys))
		for _, key := range keys {
			fmt.Printf("     - %s\n", key.Name)
		}
	}
	
	// Create transaction builder for simulation
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set temporary gas limit for simulation
	txBuilder.SetGasLimit(1000000)
	
	// ✅ SICHER: Verwenden Sie explizit den gleichen Keyring
	simClientCtx := c.clientCtx.
		WithKeyring(c.clientCtx.Keyring). // Explizit gleicher Keyring
		WithSimulation(true).
		WithOffline(false).
		WithGenerateOnly(false)
	
	// ✅ SICHER: TxFactory mit explizit gleichem Keyring
	simFactory := tx.Factory{}.
		WithKeybase(c.clientCtx.Keyring). // Explizit gleicher Keyring
		WithTxConfig(c.clientCtx.TxConfig).
		WithAccountRetriever(c.clientCtx.AccountRetriever).
		WithChainID(c.clientCtx.ChainID).
		WithSimulateAndExecute(false).
		WithGas(1000000).
		WithGasAdjustment(1.0)
	
	fmt.Println("🔧 Calculating gas with proper keyring...")
	
	// Calculate gas - if this fails, return the error
	simRes, adjustedGas, err := tx.CalculateGas(simClientCtx, simFactory, msgs...)
	if err != nil {
		return 0, fmt.Errorf("gas calculation failed: %w", err)
	}

	fmt.Printf("✅ Gas estimation successful: %d\n", adjustedGas)
	_ = simRes
	
	// Add small buffer for safety
	gasWithBuffer := uint64(float64(adjustedGas) * 1.1) // 10% buffer
	fmt.Printf("📊 Gas with buffer: %d\n", gasWithBuffer)
	
	return gasWithBuffer, nil
}

// ENTFERNEN Sie die estimateGasFallback Funktion komplett
// GetClient retrieves client information
func (c *Client) GetClient(clientID string) (*itypes.RegisteredClient, error) {
	// Query client information from blockchain
	queryPath := fmt.Sprintf("/medas.client.v1.Query/Client")
	
	// Create query request (this would need proper protobuf message)
	reqBytes := []byte(fmt.Sprintf(`{"client_id":"%s"}`, clientID))
	
	// Execute query
	res, _, err := c.clientCtx.QueryWithData(queryPath, reqBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to query client: %w", err)
	}

	// Parse response
	var client itypes.RegisteredClient
	if err := c.codec.UnmarshalJSON(res, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}

	return &client, nil
}

// GetAnalysisResults retrieves analysis results
func (c *Client) GetAnalysisResults(clientID string, limit int) ([]*itypes.StoredAnalysis, error) {
	// Query analysis results from blockchain
	queryPath := fmt.Sprintf("/medas.analysis.v1.Query/AnalysisResults")
	
	// Create query request
	reqBytes := []byte(fmt.Sprintf(`{"client_id":"%s","limit":%d}`, clientID, limit))
	
	// Execute query
	res, _, err := c.clientCtx.QueryWithData(queryPath, reqBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis results: %w", err)
	}

	// Parse response
	var results []*itypes.StoredAnalysis
	if err := c.codec.UnmarshalJSON(res, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return results, nil
}

// extractClientIDFromEvents extracts client ID from transaction events
func (c *Client) extractClientIDFromEvents(events []abci.Event) (string, error) {
	for _, event := range events {
		if event.Type == "client_registered" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "client_id" {
					return string(attr.Value), nil
				}
			}
		}
	}
	return "", fmt.Errorf("client_id not found in events")
}

// StartEventMonitoring starts monitoring blockchain events
func (c *Client) StartEventMonitoring() error {
	if c.monitoring {
		return fmt.Errorf("event monitoring already started")
	}

	// Cast client to CometBFT HTTP client for event subscription
	if httpClient, ok := c.clientCtx.Client.(*comethttp.HTTP); ok {
		// Subscribe to events - FIXED: Handle 2 return values
		query := "tm.event='NewBlock'"
		_, err := httpClient.Subscribe(context.Background(), "medas-client", query, 100)
		if err != nil {
			return fmt.Errorf("failed to subscribe to events: %w", err)
		}
	} else {
		return fmt.Errorf("client does not support event subscription")
	}

	c.monitoring = true
	return nil
}

// StopEventMonitoring stops monitoring blockchain events
func (c *Client) StopEventMonitoring() error {
	if !c.monitoring {
		return fmt.Errorf("event monitoring not started")
	}

	// Cast client to CometBFT HTTP client for event unsubscription
	if httpClient, ok := c.clientCtx.Client.(*comethttp.HTTP); ok {
		query := "tm.event='NewBlock'"
		err := httpClient.Unsubscribe(context.Background(), "medas-client", query)
		if err != nil {
			return fmt.Errorf("failed to unsubscribe from events: %w", err)
		}
	}

	c.monitoring = false
	return nil
}

// GetChainStatus returns blockchain status information
func (c *Client) GetChainStatus() (*ChainStatus, error) {
	// Get node status
	status, err := c.clientCtx.Client.Status(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}

	// Get network info
	var networkInfo *comet.ResultNetInfo
	if httpClient, ok := c.clientCtx.Client.(*comethttp.HTTP); ok {
		networkInfo, err = httpClient.NetInfo(context.Background())
		if err != nil {
			// Don't fail if network info is not available
			networkInfo = nil
		}
	}

	chainStatus := &ChainStatus{
		ChainID:         status.NodeInfo.Network,
		LatestHeight:    status.SyncInfo.LatestBlockHeight,
		LatestBlockTime: status.SyncInfo.LatestBlockTime,
		CatchingUp:      status.SyncInfo.CatchingUp,
		NodeID:          string(status.NodeInfo.DefaultNodeID),
		NodeVersion:     status.NodeInfo.Version,
		Peers:           0,
	}

	if networkInfo != nil {
		chainStatus.Peers = networkInfo.NPeers
	}

	return chainStatus, nil
}

// GetLatestBlock returns the latest block information
func (c *Client) GetLatestBlock() (*BlockInfo, error) {
	// Get latest block
	block, err := c.clientCtx.Client.Block(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	blockInfo := &BlockInfo{
		Height:    block.Block.Height,
		Hash:      block.BlockID.Hash.String(),
		Time:      block.Block.Time,
		NumTxs:    len(block.Block.Txs),
		Proposer:  block.Block.ProposerAddress.String(),
	}

	return blockInfo, nil
}

// Health checks the health of the blockchain connection
func (c *Client) Health() error {
	// Try to get node status
	_, err := c.clientCtx.Client.Status(context.Background())
	if err != nil {
		return fmt.Errorf("blockchain connection unhealthy: %w", err)
	}

	return nil
}

// ===================================
// TRANSACTION QUERY METHODS (NEU)
// ===================================

// GetTx queries a transaction by hash using the Cosmos SDK
func (c *Client) GetTx(ctx context.Context, txHash string) (*txtypes.GetTxResponse, error) {
	// Create query client using cosmos-sdk/client
	queryClient := txtypes.NewServiceClient(c.clientCtx)
	
	// Query transaction
	req := &txtypes.GetTxRequest{Hash: txHash}
	resp, err := queryClient.GetTx(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction %s: %w", txHash, err)
	}
	
	return resp, nil
}

// GetStatus returns the current blockchain status (alias for existing method)
func (c *Client) GetStatus(ctx context.Context) (*comet.ResultStatus, error) {
	// Get status from CometBFT client
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain status: %w", err)
	}
	
	return status, nil
}

// QueryWithData performs a generic query with custom data (alias for existing method)
func (c *Client) QueryWithData(ctx context.Context, path string, data []byte) ([]byte, int64, error) {
	// Use the client context to perform the query
	result, height, err := c.clientCtx.QueryWithData(path, data)
	if err != nil {
		return nil, 0, fmt.Errorf("query failed for path %s: %w", path, err)
	}
	
	return result, height, nil
}

// ===================================
// PAYMENT VERIFICATION METHODS (NEU)
// ===================================

// VerifyPaymentTransaction verifies a blockchain payment transaction
func (c *Client) VerifyPaymentTransaction(ctx context.Context, txHash, senderAddr, recipientAddr string, expectedAmount float64, denom string) (bool, error) {
	// 1. Query transaction by hash
	txResponse, err := c.GetTx(ctx, txHash)
	if err != nil {
		return false, fmt.Errorf("failed to query transaction: %w", err)
	}
	
	if txResponse.TxResponse == nil {
		return false, fmt.Errorf("transaction not found")
	}
	
	// 2. Check transaction success
	if txResponse.TxResponse.Code != 0 {
		return false, fmt.Errorf("transaction failed with code %d", txResponse.TxResponse.Code)
	}
	
	// 3. Parse transaction messages
	tx, err := c.decodeTxFromAny(txResponse.TxResponse.Tx)
	if err != nil {
		return false, fmt.Errorf("failed to decode transaction: %w", err)
	}
	
	// 4. Verify payment details
	for _, msg := range tx.GetMsgs() {
		if bankMsg, ok := msg.(*banktypes.MsgSend); ok {
			// Check sender address
			if bankMsg.FromAddress != senderAddr {
				continue
			}
			
			// Check recipient address
			if bankMsg.ToAddress != recipientAddr {
				continue
			}
			
			// Check amount and denomination
			for _, coin := range bankMsg.Amount {
				if coin.Denom == denom {
					// Convert amount based on denomination
					var actualAmount float64
					if denom == "umedas" {
						actualAmount = float64(coin.Amount.Int64()) / 1000000.0 // 6 decimals
					} else {
						actualAmount = float64(coin.Amount.Int64())
					}
					
					// Allow small rounding differences (±0.1%)
					tolerance := expectedAmount * 0.001
					if actualAmount >= expectedAmount-tolerance && actualAmount <= expectedAmount+tolerance {
						return true, nil
					}
				}
			}
		}
	}
	
	return false, fmt.Errorf("no valid payment found in transaction")
}

// GetTransactionConfirmations calculates the number of confirmations for a transaction
func (c *Client) GetTransactionConfirmations(ctx context.Context, txHeight int64) (int64, error) {
	// Get current blockchain status
	status, err := c.GetStatus(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get blockchain status: %w", err)
	}
	
	latestHeight := status.SyncInfo.LatestBlockHeight
	confirmations := latestHeight - txHeight
	
	return confirmations, nil
}

// ===================================
// BALANCE QUERY METHODS (NEU)
// ===================================

// GetAccountBalance queries the balance of an account
func (c *Client) GetAccountBalance(ctx context.Context, address string) (sdk.Coins, error) {
	// Query balance using bank module
	queryClient := banktypes.NewQueryClient(c.clientCtx)
	
	req := &banktypes.QueryAllBalancesRequest{
		Address: address,
	}
	
	res, err := queryClient.AllBalances(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance for %s: %w", address, err)
	}
	
	return res.Balances, nil
}

// GetAccountBalanceByDenom queries the balance of a specific denomination
func (c *Client) GetAccountBalanceByDenom(ctx context.Context, address, denom string) (sdk.Coin, error) {
	// Query balance using bank module
	queryClient := banktypes.NewQueryClient(c.clientCtx)
	
	req := &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	}
	
	res, err := queryClient.Balance(ctx, req)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to query balance for %s/%s: %w", address, denom, err)
	}
	
	return *res.Balance, nil
}

// ===================================
// TRANSACTION CREATION METHODS (NEU)
// ===================================

// CreateSendTransaction creates a MsgSend transaction
func (c *Client) CreateSendTransaction(fromAddr, toAddr string, amount sdk.Coins, memo string) (*sdk.TxResponse, error) {
	// Convert addresses
	fromAddress, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}
	
	toAddress, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}
	
	// Create MsgSend
	msg := banktypes.NewMsgSend(fromAddress, toAddress, amount)
	
	// Create transaction
	return c.sendTransaction(msg, fromAddr)
}

// ===================================
// UTILITY METHODS (NEU)
// ===================================

// decodeTx decodes a transaction from bytes using the TxConfig
func (c *Client) decodeTx(txBytes []byte) (sdk.Tx, error) {
	// Use the TxConfig from client context to decode
	return c.clientCtx.TxConfig.TxDecoder()(txBytes)
}

// NACHHER (sicher):
	func (c *Client) decodeTxFromAny(txAny *types.Any) (sdk.Tx, error) {
    // NULL-CHECKS hinzufügen
    if txAny == nil {
        return nil, fmt.Errorf("transaction data is nil")
    }
    
    if txAny.Value == nil {
        return nil, fmt.Errorf("transaction value is nil") 
    }
    
    if c.clientCtx.TxConfig == nil {
        return nil, fmt.Errorf("TxConfig is not initialized")
    }
    
    txBytes := txAny.Value
    return c.clientCtx.TxConfig.TxDecoder()(txBytes)
	}

// ParseTransactionData parses transaction data for display
func (c *Client) ParseTransactionData(txResponse *txtypes.GetTxResponse) (*TransactionData, error) {
	if txResponse.TxResponse == nil {
		return nil, fmt.Errorf("no transaction response")
	}
	
	// Create transaction data - KORRIGIERT: TxHash statt Txhash
	txData := &TransactionData{
		Hash:      txResponse.TxResponse.TxHash, // ← KORRIGIERT
		Height:    txResponse.TxResponse.Height,
		Code:      txResponse.TxResponse.Code,
		Timestamp: txResponse.TxResponse.Timestamp,
		GasUsed:   txResponse.TxResponse.GasUsed,
		GasWanted: txResponse.TxResponse.GasWanted,
	}
	
	// Try to extract memo
	if txWithMemo, ok := tx.(interface{ GetMemo() string }); ok {
		txData.Memo = txWithMemo.GetMemo()
	}
	
	// Try to extract fee
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
// ChainStatus represents blockchain status
type ChainStatus struct {
	ChainID         string    `json:"chain_id"`
	LatestHeight    int64     `json:"latest_height"`
	LatestBlockTime time.Time `json:"latest_block_time"`
	CatchingUp      bool      `json:"catching_up"`
	NodeID          string    `json:"node_id"`
	NodeVersion     string    `json:"node_version"`
	Peers           int       `json:"peers"`
}

// BlockInfo represents block information
type BlockInfo struct {
	Height   int64     `json:"height"`
	Hash     string    `json:"hash"`
	Time     time.Time `json:"time"`
	NumTxs   int       `json:"num_txs"`
	Proposer string    `json:"proposer"`
}
// TransactionData represents parsed transaction data
type TransactionData struct {
	Hash        string `json:"hash"`
	Height      int64  `json:"height"`
	Code        uint32 `json:"code"`
	Timestamp   string `json:"timestamp"`
	GasUsed     int64  `json:"gas_used"`
	GasWanted   int64  `json:"gas_wanted"`
	Fee         string `json:"fee"`
	Memo        string `json:"memo"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Denom       string `json:"denom"`
}

package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	
	// CometBFT imports (ersetzt Tendermint)
	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	clienttypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// Client handles blockchain communication
type Client struct {
	clientCtx     client.Context
	chainID       string
	rpcEndpoint   string
	grpcEndpoint  string
	restEndpoint  string
	gasLimit      uint64
	gasPrice      string
	retryAttempts int
	retryDelay    time.Duration
}

// NewClient creates a new blockchain client
func NewClient(clientCtx client.Context) *Client {
	return &Client{
		clientCtx:     clientCtx,
		chainID:       clientCtx.ChainID,
		gasLimit:      200000,
		gasPrice:      "0.025medas",
		retryAttempts: 3,
		retryDelay:    2 * time.Second,
	}
}

// Initialize initializes the blockchain client with endpoints
func (c *Client) Initialize(rpcEndpoint, grpcEndpoint, restEndpoint string) error {
	c.rpcEndpoint = rpcEndpoint
	c.grpcEndpoint = grpcEndpoint
	c.restEndpoint = restEndpoint

	// Test connection
	if err := c.testConnection(); err != nil {
		return fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	log.Printf("Blockchain client initialized for chain: %s", c.chainID)
	return nil
}

// testConnection tests the connection to the blockchain
func (c *Client) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node status: %w", err)
	}

	if status.NodeInfo.Network != c.chainID {
		return fmt.Errorf("chain ID mismatch: expected %s, got %s", c.chainID, status.NodeInfo.Network)
	}

	log.Printf("Connected to blockchain: %s (block: %d)", status.NodeInfo.Network, status.SyncInfo.LatestBlockHeight)
	return nil
}

// RegisterClient registers a new client on the blockchain
func (c *Client) RegisterClient(creator string, capabilities []string, metadata string) (string, error) {
	log.Printf("Registering client with capabilities: %v", capabilities)

	msg := &clienttypes.MsgRegisterClient{
		Creator:      creator,
		Capabilities: capabilities,
		Metadata:     metadata,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return "", fmt.Errorf("invalid register client message: %w", err)
	}

	// Send transaction
	res, err := c.sendTransaction(msg)
	if err != nil {
		return "", fmt.Errorf("failed to register client: %w", err)
	}

	// Extract client ID from events
	clientID, err := c.extractClientIDFromEvents(res.Events)
	if err != nil {
		return "", fmt.Errorf("failed to extract client ID: %w", err)
	}

	log.Printf("Client registered successfully with ID: %s (tx: %s)", clientID, res.TxHash)
	return clientID, nil
}

// StoreAnalysisResult stores analysis results on the blockchain
func (c *Client) StoreAnalysisResult(creator, clientID, analysisType string, data map[string]interface{}, blockHeight int64, txHash string) error {
	log.Printf("Storing analysis result for client: %s, type: %s", clientID, analysisType)

	msg := &clienttypes.MsgStoreAnalysis{
		Creator:      creator,
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         data,
		BlockHeight:  blockHeight,
		TxHash:       txHash,
	}

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid store analysis message: %w", err)
	}

	// Send transaction
	res, err := c.sendTransaction(msg)
	if err != nil {
		return fmt.Errorf("failed to store analysis result: %w", err)
	}

	log.Printf("Analysis result stored successfully (tx: %s)", res.TxHash)
	return nil
}

// sendTransaction sends a transaction to the blockchain with retry logic
func (c *Client) sendTransaction(msg sdk.Msg) (*sdk.TxResponse, error) {
	var lastErr error

	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying transaction (attempt %d/%d)", attempt+1, c.retryAttempts)
			time.Sleep(c.retryDelay)
		}

		res, err := c.attemptTransaction(msg)
		if err == nil {
			return res, nil
		}

		lastErr = err
		
		// Don't retry on certain errors
		if strings.Contains(err.Error(), "account sequence mismatch") ||
		   strings.Contains(err.Error(), "insufficient funds") ||
		   strings.Contains(err.Error(), "invalid signature") {
			break
		}
	}

	return nil, fmt.Errorf("transaction failed after %d attempts: %w", c.retryAttempts, lastErr)
}

// attemptTransaction attempts to send a single transaction
func (c *Client) attemptTransaction(msg sdk.Msg) (*sdk.TxResponse, error) {
	// Get account info
	addr := c.clientCtx.GetFromAddress()
	if addr.Empty() {
		return nil, fmt.Errorf("from address not set")
	}

	// Create transaction factory (new in v0.50)
	txFactory := tx.Factory{}.
		WithAccountRetriever(c.clientCtx.AccountRetriever).
		WithChainID(c.clientCtx.ChainID).
		WithTxConfig(c.clientCtx.TxConfig).
		WithGasAdjustment(1.2).
		WithGasPrices(c.gasPrice).
		WithKeybase(c.clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Build transaction
	txBuilder, err := txFactory.BuildUnsignedTx(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to build unsigned transaction: %w", err)
	}

	// Set gas limit
	txBuilder.SetGasLimit(c.gasLimit)

	// Calculate fees
	fees := sdk.NewCoins(sdk.NewCoin("medas", math.NewInt(int64(c.gasLimit/100))))
	txBuilder.SetFeeAmount(fees)

	// Sign transaction
	err = tx.Sign(context.Background(), txFactory, c.clientCtx.GetFromName(), txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast transaction
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Send transaction (updated for v0.50)
	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Check for transaction errors
	if res.Code != 0 {
		return nil, fmt.Errorf("transaction failed with code %d: %s", res.Code, res.RawLog)
	}

	return res, nil
}

// extractClientIDFromEvents extracts client ID from transaction events
func (c *Client) extractClientIDFromEvents(events []sdk.Event) (string, error) {
	for _, event := range events {
		if event.Type == "register_client" {
			for _, attr := range event.Attributes {
				if attr.Key == "client_id" {
					return attr.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("client ID not found in transaction events")
}

// GetClient retrieves client information by ID
func (c *Client) GetClient(clientID string) (*clienttypes.RegisteredClient, error) {
	// Updated query path for v0.50
	queryPath := fmt.Sprintf("/custom/clientregistry/client/%s", clientID)
	
	res, _, err := c.clientCtx.QueryWithData(queryPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query client: %w", err)
	}

	var client clienttypes.RegisteredClient
	if err := json.Unmarshal(res, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client data: %w", err)
	}

	return &client, nil
}

// GetAnalysisResults retrieves analysis results for a client
func (c *Client) GetAnalysisResults(clientID string, limit int) ([]*clienttypes.StoredAnalysis, error) {
	queryPath := fmt.Sprintf("/custom/clientregistry/analysis/%s", clientID)
	queryData := map[string]interface{}{
		"limit": limit,
	}
	
	queryBytes, err := json.Marshal(queryData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query data: %w", err)
	}

	res, _, err := c.clientCtx.QueryWithData(queryPath, queryBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis results: %w", err)
	}

	var results []*clienttypes.StoredAnalysis
	if err := json.Unmarshal(res, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis results: %w", err)
	}

	return results, nil
}

// Query performs a generic query to the blockchain
func (c *Client) Query(queryType, queryID string) (interface{}, error) {
	switch queryType {
	case "client":
		return c.GetClient(queryID)
	case "analysis":
		return c.GetAnalysisById(queryID)
	case "block":
		return c.GetBlock(queryID)
	case "transaction":
		return c.GetTransaction(queryID)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}
}

// GetAnalysisById retrieves a specific analysis by ID
func (c *Client) GetAnalysisById(analysisID string) (*clienttypes.StoredAnalysis, error) {
	queryPath := fmt.Sprintf("/custom/clientregistry/analysis_by_id/%s", analysisID)
	
	res, _, err := c.clientCtx.QueryWithData(queryPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis: %w", err)
	}

	var analysis clienttypes.StoredAnalysis
	if err := json.Unmarshal(res, &analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis data: %w", err)
	}

	return &analysis, nil
}

// GetBlock retrieves block information
func (c *Client) GetBlock(heightStr string) (*coretypes.ResultBlock, error) {
	var height *int64
	if heightStr != "latest" {
		h, err := parseHeight(heightStr)
		if err != nil {
			return nil, fmt.Errorf("invalid block height: %w", err)
		}
		height = &h
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	block, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return block, nil
}

// GetTransaction retrieves transaction information
func (c *Client) GetTransaction(txHash string) (*coretypes.ResultTx, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert hex string to bytes
	hash, err := parseHash(txHash)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction hash: %w", err)
	}

	tx, err := c.clientCtx.Client.Tx(ctx, hash, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(address string, denom string) (*sdk.Coin, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	queryClient := banktypes.NewQueryClient(c.clientCtx)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := queryClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   denom,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query balance: %w", err)
	}

	return res.Balance, nil
}

// GetChainStatus retrieves current chain status
func (c *Client) GetChainStatus() (*coretypes.ResultStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain status: %w", err)
	}

	return status, nil
}

// WaitForTransaction waits for a transaction to be included in a block
func (c *Client) WaitForTransaction(txHash string, timeout time.Duration) (*coretypes.ResultTx, error) {
	start := time.Now()
	
	for time.Since(start) < timeout {
		tx, err := c.GetTransaction(txHash)
		if err == nil {
			return tx, nil
		}
		
		// Don't spam the node
		time.Sleep(1 * time.Second)
	}
	
	return nil, fmt.Errorf("transaction %s not found after %v", txHash, timeout)
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(msg sdk.Msg) (uint64, error) {
	// Create transaction factory for gas estimation
	txFactory := tx.Factory{}.
		WithAccountRetriever(c.clientCtx.AccountRetriever).
		WithChainID(c.clientCtx.ChainID).
		WithTxConfig(c.clientCtx.TxConfig).
		WithKeybase(c.clientCtx.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Build unsigned transaction
	_, adjusted, err := tx.CalculateGas(c.clientCtx, txFactory, msg)
	if err != nil {
		// Fallback to fixed estimates if simulation fails
		switch msg.Type() {
		case "register_client":
			return 150000, nil
		case "store_analysis":
			return 200000, nil
		default:
			return 100000, nil
		}
	}

	return adjusted, nil
}

// Helper functions
func parseHeight(heightStr string) (int64, error) {
	if heightStr == "" || heightStr == "latest" {
		return 0, nil
	}
	
	height := int64(0)
	if _, err := fmt.Sscanf(heightStr, "%d", &height); err != nil {
		return 0, err
	}
	
	return height, nil
}

func parseHash(hashStr string) ([]byte, error) {
	if len(hashStr) == 0 {
		return nil, fmt.Errorf("empty hash")
	}
	
	// Remove 0x prefix if present
	if strings.HasPrefix(hashStr, "0x") {
		hashStr = hashStr[2:]
	}
	
	// Convert hex string to bytes
	hash := make([]byte, len(hashStr)/2)
	for i := 0; i < len(hashStr); i += 2 {
		var b byte
		if _, err := fmt.Sscanf(hashStr[i:i+2], "%02x", &b); err != nil {
			return nil, fmt.Errorf("invalid hex character: %s", hashStr[i:i+2])
		}
		hash[i/2] = b
	}
	
	return hash, nil
}

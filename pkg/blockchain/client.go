package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	comet "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/oxygene76/medasdigital-client/internal/types"
)

// Client represents a blockchain client for the MedasDigital network
type Client struct {
	clientCtx     client.Context
	txFactory     tx.Factory
	retryAttempts int
	retryDelay    time.Duration
}

// NewClient creates a new blockchain client
func NewClient(clientCtx client.Context) *Client {
	// Create transaction factory for v0.50
	txFactory := tx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithKeybase(clientCtx.Keyring).
		WithChainID(clientCtx.ChainID).
		WithSimulateAndExecute(true).
		WithGas(200000)

	return &Client{
		clientCtx:     clientCtx,
		txFactory:     txFactory,
		retryAttempts: 3,
		retryDelay:    time.Second * 2,
	}
}

// Initialize sets up the blockchain client connection
func (c *Client) Initialize(endpoints []string) error {
	for _, endpoint := range endpoints {
		if err := c.testConnection(endpoint); err == nil {
			// Connection successful
			rpcClient, err := comethttp.New(endpoint, "/websocket")
			if err != nil {
				continue
			}
			
			c.clientCtx = c.clientCtx.WithClient(rpcClient)
			return nil
		}
	}
	
	return fmt.Errorf("failed to connect to any endpoint")
}

// testConnection tests connection to a specific endpoint
func (c *Client) testConnection(endpoint string) error {
	rpcClient, err := comethttp.New(endpoint, "/websocket")
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err = rpcClient.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	
	return nil
}

// RegisterClient registers a new client on the blockchain
func (c *Client) RegisterClient(creator string, capabilities []string, metadata string) (string, error) {
	msg := &MsgRegisterClient{
		Creator:      creator,
		Capabilities: capabilities,
		Metadata:     metadata,
	}

	// Send transaction with retry logic
	res, err := c.sendTransaction(msg)
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
func (c *Client) StoreAnalysisResult(creator, clientID, analysisType string, data map[string]interface{}, blockHeight int64, txHash string) error {
	// Convert data to JSON string
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	msg := &MsgStoreAnalysis{
		Creator:      creator,
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         string(dataBytes),
		BlockHeight:  blockHeight,
		TxHash:       txHash,
	}

	_, err = c.sendTransaction(msg)
	if err != nil {
		return fmt.Errorf("failed to store analysis result: %w", err)
	}

	return nil
}

// sendTransaction sends a transaction with retry logic
func (c *Client) sendTransaction(msg sdk.Msg) (*sdk.TxResponse, error) {
	var lastErr error

	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		res, err := c.attemptTransaction(msg)
		if err == nil {
			return res, nil
		}

		lastErr = err
		if attempt < c.retryAttempts-1 {
			time.Sleep(c.retryDelay)
		}
	}

	return nil, fmt.Errorf("transaction failed after %d attempts: %w", c.retryAttempts, lastErr)
}

// attemptTransaction makes a single transaction attempt
func (c *Client) attemptTransaction(msg sdk.Msg) (*sdk.TxResponse, error) {
	// Create transaction builder
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Estimate gas using the new v0.50 API
	gasUsed, err := c.EstimateGas([]sdk.Msg{msg})
	if err != nil {
		// Fall back to default gas if estimation fails
		gasUsed = 200000
	}

	// Set gas limit with buffer
	gasLimit := uint64(float64(gasUsed) * 1.3) // 30% buffer
	txBuilder.SetGasLimit(gasLimit)

	// Set fee
	feeAmount := math.NewInt(int64(gasLimit * 25)) // 25 units per gas
	fees := sdk.NewCoins(sdk.NewCoin("umedas", feeAmount))
	txBuilder.SetFeeAmount(fees)

	// Sign the transaction using v0.50 API
	err = tx.Sign(c.txFactory, c.clientCtx.GetFromName(), txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast transaction
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	return c.clientCtx.BroadcastTx(txBytes)
}

// EstimateGas estimates gas for messages using the new v0.50 API
func (c *Client) EstimateGas(msgs []sdk.Msg) (uint64, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set a high gas limit for simulation
	txBuilder.SetGasLimit(1000000)

	// Calculate gas using v0.50 API
	adjustedGas, err := tx.CalculateGas(c.clientCtx, c.txFactory, msgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate gas: %w", err)
	}

	return adjustedGas, nil
}

// WaitForTransaction waits for a transaction to be confirmed
func (c *Client) WaitForTransaction(txHash string, timeout time.Duration) (*sdk.TxResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction confirmation")
		case <-ticker.C:
			// Try to get the transaction
			txResponse, err := c.clientCtx.Client.Tx(ctx, []byte(txHash), false)
			if err == nil {
				// Transaction found and confirmed
				return &sdk.TxResponse{
					Height:    txResponse.Height,
					TxHash:    txHash,
					Code:      txResponse.TxResult.Code,
					Data:      string(txResponse.TxResult.Data),
					RawLog:    txResponse.TxResult.Log,
					Logs:      sdk.ABCIMessageLogs{},
					GasWanted: txResponse.TxResult.GasWanted,
					GasUsed:   txResponse.TxResult.GasUsed,
				}, nil
			}
			// Continue polling if transaction not found
		}
	}
}

// GetClient retrieves client information by ID
func (c *Client) GetClient(clientID string) (*types.RegisteredClient, error) {
	// This would query the x/clientregistry module
	// For now, return a placeholder implementation
	// In production, this would use ABCI query
	
	query := fmt.Sprintf("custom/clientregistry/client/%s", clientID)
	res, _, err := c.clientCtx.QueryWithData(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query client: %w", err)
	}
	
	var client types.RegisteredClient
	if err := json.Unmarshal(res, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}
	
	return &client, nil
}

// GetAnalysisResults retrieves analysis results for a client
func (c *Client) GetAnalysisResults(clientID string, limit int) ([]*types.StoredAnalysis, error) {
	// This would query the x/clientregistry module  
	// For now, return a placeholder implementation
	
	query := fmt.Sprintf("custom/clientregistry/analysis/%s?limit=%d", clientID, limit)
	res, _, err := c.clientCtx.QueryWithData(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis results: %w", err)
	}
	
	var results []*types.StoredAnalysis
	if err := json.Unmarshal(res, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}
	
	return results, nil
}

// Query performs a generic query on the blockchain
func (c *Client) Query(queryType, queryID string) (interface{}, error) {
	switch queryType {
	case "client":
		return c.GetClient(queryID)
	case "analysis": 
		return c.GetAnalysisResults(queryID, 10)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}
}

// GetAnalysisByID retrieves a specific analysis by ID  
func (c *Client) GetAnalysisByID(analysisID string) (*types.StoredAnalysis, error) {
	query := fmt.Sprintf("custom/clientregistry/analysis-by-id/%s", analysisID)
	res, _, err := c.clientCtx.QueryWithData(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis: %w", err)
	}
	
	var analysis types.StoredAnalysis
	if err := json.Unmarshal(res, &analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis: %w", err)
	}
	
	return &analysis, nil
}

// GetBlock retrieves a block by height
func (c *Client) GetBlock(height int64) (*comet.ResultBlock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.Block(ctx, &height)
}

// GetLatestBlock retrieves the latest block
func (c *Client) GetLatestBlock() (*comet.ResultBlock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.Block(ctx, nil)
}

// GetTransaction retrieves a transaction by hash
func (c *Client) GetTransaction(txHash string) (*comet.ResultTx, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.Tx(ctx, []byte(txHash), false)
}

// GetChainStatus retrieves the current chain status
func (c *Client) GetChainStatus() (*comet.ResultStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.Status(ctx)
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(address string) (sdk.Coins, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// This would use the bank module query
	// For now, return a placeholder
	return sdk.NewCoins(), nil
}

// SubscribeToEvents subscribes to blockchain events
func (c *Client) SubscribeToEvents(query string) (<-chan comet.ResultEvent, error) {
	ctx := context.Background()
	
	out, err := c.clientCtx.Client.Subscribe(ctx, "medasdigital-client", query, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}

	return out, nil
}

// UnsubscribeFromEvents unsubscribes from blockchain events
func (c *Client) UnsubscribeFromEvents(query string) error {
	ctx := context.Background()
	
	err := c.clientCtx.Client.Unsubscribe(ctx, "medasdigital-client", query)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from events: %w", err)
	}

	return nil
}

// extractClientIDFromEvents extracts client ID from transaction events
func (c *Client) extractClientIDFromEvents(events []sdk.Event) (string, error) {
	for _, event := range events {
		if event.Type == "client_registered" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "client_id" {
					return string(attr.Value), nil
				}
			}
		}
	}
	
	// Fallback: generate client ID if not found in events
	return fmt.Sprintf("client_%d", time.Now().Unix()), nil
}

// parseEvents parses transaction events into structured data
func (c *Client) parseEvents(events []sdk.Event) map[string]interface{} {
	parsed := make(map[string]interface{})
	
	for _, event := range events {
		eventData := make(map[string]string)
		for _, attr := range event.Attributes {
			// In v0.50, attributes are already strings
			eventData[string(attr.Key)] = string(attr.Value)
		}
		parsed[event.Type] = eventData
	}
	
	return parsed
}

// GetNetworkInfo retrieves network information
func (c *Client) GetNetworkInfo() (*comet.ResultNetInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.NetInfo(ctx)
}

// GetValidators retrieves the current validator set
func (c *Client) GetValidators(height int64) (*comet.ResultValidators, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.clientCtx.Client.Validators(ctx, &height, nil, nil)
}

// IsHealthy checks if the blockchain client is healthy
func (c *Client) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.clientCtx.Client.Status(ctx)
	return err == nil
}

// GetClientContext returns the underlying client context
func (c *Client) GetClientContext() client.Context {
	return c.clientCtx
}

// SetRetryConfig sets retry configuration
func (c *Client) SetRetryConfig(attempts int, delay time.Duration) {
	c.retryAttempts = attempts
	c.retryDelay = delay
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() string {
	return c.clientCtx.ChainID
}

// GetFromAddress returns the from address
func (c *Client) GetFromAddress() sdk.AccAddress {
	return c.clientCtx.GetFromAddress()
}

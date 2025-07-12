package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	comet "github.com/cometbft/cometbft/rpc/core/types"
	abci "github.com/cometbft/cometbft/abci/types"

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

// Fügen Sie Debug-Informationen in sendTransaction hinzu:

func (c *Client) sendTransaction(msg sdk.Msg, signerName string) (*sdk.TxResponse, error) {
	fmt.Printf("🔧 sendTransaction called with signer: %s\n", signerName)
	
	// Create transaction builder
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}
	
	fmt.Println("🔧 Setting fixed gas limit (no estimation)")
	// Set fixed gas limit - NO ESTIMATION
	txBuilder.SetGasLimit(200000)
	
	fmt.Println("🔧 Setting fee amount")
	// Set fee
	fees := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(5000)))
	txBuilder.SetFeeAmount(fees)

	fmt.Printf("🔧 Signing transaction with: %s\n", signerName)
	// Sign transaction
	err := tx.Sign(context.Background(), c.txFactory, signerName, txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	fmt.Println("🔧 Encoding transaction")
	// Encode transaction
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	fmt.Println("🔧 Broadcasting transaction")
	// Broadcast transaction
	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	fmt.Println("✅ Transaction sent successfully!")
	return res, nil
}


// estimateGas estimates gas for a transaction - FIXED: Handle 3 return values
func (c *Client) estimateGas(msgs []sdk.Msg) (uint64, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set temporary gas limit for simulation
	txBuilder.SetGasLimit(1000000)

	// Calculate gas - FIXED: v0.50 returns 3 values: simRes, adjustedGas, error
	simRes, adjustedGas, err := tx.CalculateGas(c.clientCtx, c.txFactory, msgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate gas: %w", err)
	}

	// Use the adjusted gas value (uint64) - simRes contains detailed simulation info
	_ = simRes // We can use this for detailed gas info if needed in the future
	return adjustedGas, nil
}

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

package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/query"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/rpc/client/http"

	clienttypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// Client handles blockchain communication
type Client struct {
	clientCtx    client.Context
	queryClient  *QueryClient
	txClient     *TxClient
	chainID      string
	rpcEndpoint  string
}

// QueryClient handles blockchain queries
type QueryClient struct {
	clientCtx client.Context
}

// TxClient handles transaction broadcasting
type TxClient struct {
	clientCtx client.Context
}

// NewClient creates a new blockchain client
func NewClient(clientCtx client.Context) *Client {
	return &Client{
		clientCtx:   clientCtx,
		queryClient: &QueryClient{clientCtx: clientCtx},
		txClient:    &TxClient{clientCtx: clientCtx},
		chainID:     clientCtx.ChainID,
	}
}

// GetStatus returns the current blockchain status
func (c *Client) GetStatus() (*BlockchainStatus, error) {
	status, err := c.clientCtx.Client.Status(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return &BlockchainStatus{
		ChainID:         status.NodeInfo.Network,
		LatestHeight:    status.SyncInfo.LatestBlockHeight,
		LatestBlockTime: status.SyncInfo.LatestBlockTime,
		Syncing:         status.SyncInfo.CatchingUp,
		NodeInfo: NodeInfo{
			ID:      string(status.NodeInfo.DefaultNodeID),
			Moniker: status.NodeInfo.Moniker,
			Version: status.NodeInfo.Version,
		},
	}, nil
}

// GetBlock returns block information for a specific height
func (c *Client) GetBlock(height int64) (*BlockInfo, error) {
	block, err := c.clientCtx.Client.Block(context.Background(), &height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	blockInfo := &BlockInfo{
		Height:    block.Block.Height,
		Hash:      block.BlockID.Hash.String(),
		Time:      block.Block.Time,
		NumTxs:    len(block.Block.Data.Txs),
		Proposer:  block.Block.Header.ProposerAddress.String(),
		ChainID:   block.Block.ChainID,
	}

	// Parse transactions
	for i, txBytes := range block.Block.Data.Txs {
		txInfo := TxInfo{
			Hash:   fmt.Sprintf("%X", txBytes.Hash()),
			Index:  i,
			Height: block.Block.Height,
		}

		// Try to decode transaction
		tx, err := c.clientCtx.TxConfig.TxDecoder()(txBytes)
		if err == nil {
			txInfo.Messages = len(tx.GetMsgs())
			txInfo.Fee = tx.GetFee().String()
			txInfo.Memo = tx.GetMemo()
		}

		blockInfo.Transactions = append(blockInfo.Transactions, txInfo)
	}

	return blockInfo, nil
}

// GetTransaction returns transaction information
func (c *Client) GetTransaction(txHash string) (*TxResponse, error) {
	res, err := c.clientCtx.Client.Tx(context.Background(), []byte(txHash), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &TxResponse{
		Hash:   res.Hash.String(),
		Height: res.Height,
		Code:   res.TxResult.Code,
		Log:    res.TxResult.Log,
		Events: parseEvents(res.TxResult.Events),
	}, nil
}

// BroadcastTx broadcasts a transaction to the network
func (c *Client) BroadcastTx(msgs []sdk.Msg, memo string, fees sdk.Coins) (*sdk.TxResponse, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()

	// Set messages
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set memo
	txBuilder.SetMemo(memo)

	// Set fees
	txBuilder.SetFeeAmount(fees)

	// Set gas limit (estimate or use default)
	gasLimit := uint64(200000) // Default gas limit
	txBuilder.SetGasLimit(gasLimit)

	// Sign and broadcast
	return tx.BroadcastTx(c.clientCtx, txBuilder.GetTx())
}

// RegisterClient registers a new client on the blockchain
func (c *Client) RegisterClient(capabilities []string, metadata string) (*sdk.TxResponse, error) {
	msg := &clienttypes.MsgRegisterClient{
		Creator:      c.clientCtx.GetFromAddress().String(),
		Capabilities: capabilities,
		Metadata:     metadata,
	}

	log.Printf("Registering client with capabilities: %v", capabilities)

	return c.BroadcastTx([]sdk.Msg{msg}, "Client registration", sdk.NewCoins())
}

// StoreAnalysisResult stores analysis results on the blockchain
func (c *Client) StoreAnalysisResult(clientID, analysisType string, data map[string]interface{}, blockHeight int64, txHash string) (*sdk.TxResponse, error) {
	msg := &clienttypes.MsgStoreAnalysis{
		Creator:      c.clientCtx.GetFromAddress().String(),
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         data,
		BlockHeight:  blockHeight,
		TxHash:       txHash,
	}

	log.Printf("Storing analysis result for client %s, type: %s", clientID, analysisType)

	memo := fmt.Sprintf("Analysis result: %s", analysisType)
	return c.BroadcastTx([]sdk.Msg{msg}, memo, sdk.NewCoins())
}

// GetAnalysisResults retrieves analysis results for a specific client
func (c *Client) GetAnalysisResults(clientID string, limit int) ([]clienttypes.StoredAnalysis, error) {
	// This would query the custom clientregistry module
	// For now, we'll simulate the query
	
	log.Printf("Querying analysis results for client: %s (limit: %d)", clientID, limit)

	// Simulate some results
	var results []clienttypes.StoredAnalysis
	for i := 0; i < min(limit, 3); i++ {
		result := clienttypes.StoredAnalysis{
			ID:           fmt.Sprintf("analysis_%s_%d", clientID, i),
			ClientID:     clientID,
			Creator:      c.clientCtx.GetFromAddress().String(),
			AnalysisType: "orbital_dynamics",
			Data: map[string]interface{}{
				"planet9_probability":      0.75 + float64(i)*0.05,
				"clustering_significance":  3.2 + float64(i)*0.1,
				"objects_analyzed":         150 + i*10,
			},
			BlockHeight: 1000 + int64(i),
			TxHash:      fmt.Sprintf("tx_hash_%d", i),
			Timestamp:   time.Now().Add(-time.Duration(i*24) * time.Hour),
		}
		results = append(results, result)
	}

	return results, nil
}

// GetRegisteredClients retrieves all registered clients
func (c *Client) GetRegisteredClients() ([]clienttypes.RegisteredClient, error) {
	log.Println("Querying registered clients...")

	// Simulate registered clients
	clients := []clienttypes.RegisteredClient{
		{
			ID:       "client_001",
			Creator:  "medas1abc123...",
			Capabilities: []string{"orbital_dynamics", "photometric_analysis"},
			Metadata: `{"institution": "Observatory A", "gpu": "RTX 4090"}`,
			RegisteredAt: time.Now().Add(-72 * time.Hour),
			Status:   "active",
			LastSeen: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:       "client_002",
			Creator:  "medas1def456...",
			Capabilities: []string{"ai_training", "gpu_compute"},
			Metadata: `{"institution": "University B", "gpu": "RTX 3090"}`,
			RegisteredAt: time.Now().Add(-48 * time.Hour),
			Status:   "active",
			LastSeen: time.Now().Add(-30 * time.Minute),
		},
	}

	return clients, nil
}

// GetClientByID retrieves a specific client by ID
func (c *Client) GetClientByID(clientID string) (*clienttypes.RegisteredClient, error) {
	log.Printf("Querying client: %s", clientID)

	clients, err := c.GetRegisteredClients()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if client.ID == clientID {
			return &client, nil
		}
	}

	return nil, fmt.Errorf("client not found: %s", clientID)
}

// Query performs a generic query
func (c *Client) Query(queryType, queryID string) (interface{}, error) {
	log.Printf("Performing query: type=%s, id=%s", queryType, queryID)

	switch queryType {
	case "analysis":
		return c.getAnalysisByID(queryID)
	case "client":
		return c.GetClientByID(queryID)
	case "block":
		height, err := strconv.ParseInt(queryID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid block height: %s", queryID)
		}
		return c.GetBlock(height)
	case "transaction":
		return c.GetTransaction(queryID)
	default:
		return nil, fmt.Errorf("unknown query type: %s", queryType)
	}
}

// getAnalysisByID retrieves a specific analysis by ID
func (c *Client) getAnalysisByID(analysisID string) (*clienttypes.StoredAnalysis, error) {
	// Simulate analysis lookup
	analysis := &clienttypes.StoredAnalysis{
		ID:           analysisID,
		ClientID:     "client_001",
		Creator:      c.clientCtx.GetFromAddress().String(),
		AnalysisType: "orbital_dynamics",
		Data: map[string]interface{}{
			"planet9_probability":      0.82,
			"clustering_significance":  4.1,
			"objects_analyzed":         175,
			"gravitational_effects":    15,
			"recommendations":          3,
		},
		BlockHeight: 1234,
		TxHash:      "ABC123DEF456",
		Timestamp:   time.Now().Add(-12 * time.Hour),
	}

	return analysis, nil
}

// GetAccountInfo retrieves account information
func (c *Client) GetAccountInfo(address string) (*AccountInfo, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Query account
	authQueryClient := types.NewQueryClient(c.clientCtx)
	accountRes, err := authQueryClient.Account(context.Background(), &types.QueryAccountRequest{
		Address: addr.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	// Query balances
	bankQueryClient := banktypes.NewQueryClient(c.clientCtx)
	balanceRes, err := bankQueryClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: addr.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}

	var account types.AccountI
	if err := c.clientCtx.Codec.UnpackAny(accountRes.Account, &account); err != nil {
		return nil, fmt.Errorf("failed to unpack account: %w", err)
	}

	return &AccountInfo{
		Address:       address,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
		Balances:      balanceRes.Balances,
	}, nil
}

// GetValidators retrieves validator information
func (c *Client) GetValidators() ([]ValidatorInfo, error) {
	validators, err := c.clientCtx.Client.Validators(context.Background(), nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	var validatorInfos []ValidatorInfo
	for _, val := range validators.Validators {
		info := ValidatorInfo{
			Address:     val.Address.String(),
			PubKey:      val.PubKey.String(),
			VotingPower: val.VotingPower,
			Proposer:    val.ProposerPriority,
		}
		validatorInfos = append(validatorInfos, info)
	}

	return validatorInfos, nil
}

// SubscribeToBlocks subscribes to new blocks
func (c *Client) SubscribeToBlocks(ctx context.Context, callback func(*BlockInfo)) error {
	log.Println("Subscribing to new blocks...")

	// Subscribe to new block events
	subscriber := "medasdigital-client"
	eventCh, err := c.clientCtx.Client.Subscribe(ctx, subscriber, "tm.event = 'NewBlock'")
	if err != nil {
		return fmt.Errorf("failed to subscribe to blocks: %w", err)
	}

	go func() {
		defer func() {
			if err := c.clientCtx.Client.Unsubscribe(ctx, subscriber, "tm.event = 'NewBlock'"); err != nil {
				log.Printf("Error unsubscribing: %v", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				if event.Data != nil {
					// Parse block from event
					height := event.Data.(map[string]interface{})["height"].(int64)
					blockInfo, err := c.GetBlock(height)
					if err != nil {
						log.Printf("Error getting block %d: %v", height, err)
						continue
					}
					callback(blockInfo)
				}
			}
		}
	}()

	return nil
}

// SubscribeToTransactions subscribes to transactions
func (c *Client) SubscribeToTransactions(ctx context.Context, callback func(*TxResponse)) error {
	log.Println("Subscribing to new transactions...")

	subscriber := "medasdigital-client-tx"
	eventCh, err := c.clientCtx.Client.Subscribe(ctx, subscriber, "tm.event = 'Tx'")
	if err != nil {
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}

	go func() {
		defer func() {
			if err := c.clientCtx.Client.Unsubscribe(ctx, subscriber, "tm.event = 'Tx'"); err != nil {
				log.Printf("Error unsubscribing: %v", err)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				if event.Data != nil {
					// Parse transaction from event
					data := event.Data.(map[string]interface{})
					if txHash, ok := data["hash"].(string); ok {
						txResponse, err := c.GetTransaction(txHash)
						if err != nil {
							log.Printf("Error getting transaction %s: %v", txHash, err)
							continue
						}
						callback(txResponse)
					}
				}
			}
		}
	}()

	return nil
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(msgs []sdk.Msg) (uint64, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Simulate transaction to estimate gas
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, fmt.Errorf("failed to encode transaction: %w", err)
	}

	simRes, err := c.clientCtx.Client.ABCIQuery(context.Background(), "/cosmos.tx.v1beta1.Service/Simulate", txBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	// Parse simulation result
	// In a real implementation, you would decode the simulation response
	// For now, return a default gas estimate
	return 200000, nil
}

// Helper functions

func parseEvents(events []interface{}) []Event {
	var parsedEvents []Event
	for _, event := range events {
		if eventMap, ok := event.(map[string]interface{}); ok {
			parsedEvent := Event{
				Type: eventMap["type"].(string),
			}
			
			if attributes, ok := eventMap["attributes"].([]interface{}); ok {
				for _, attr := range attributes {
					if attrMap, ok := attr.(map[string]interface{}); ok {
						parsedEvent.Attributes = append(parsedEvent.Attributes, EventAttribute{
							Key:   attrMap["key"].(string),
							Value: attrMap["value"].(string),
						})
					}
				}
			}
			
			parsedEvents = append(parsedEvents, parsedEvent)
		}
	}
	return parsedEvents
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Data structures

type BlockchainStatus struct {
	ChainID         string    `json:"chain_id"`
	LatestHeight    int64     `json:"latest_height"`
	LatestBlockTime time.Time `json:"latest_block_time"`
	Syncing         bool      `json:"syncing"`
	NodeInfo        NodeInfo  `json:"node_info"`
}

type NodeInfo struct {
	ID      string `json:"id"`
	Moniker string `json:"moniker"`
	Version string `json:"version"`
}

type BlockInfo struct {
	Height       int64     `json:"height"`
	Hash         string    `json:"hash"`
	Time         time.Time `json:"time"`
	NumTxs       int       `json:"num_txs"`
	Proposer     string    `json:"proposer"`
	ChainID      string    `json:"chain_id"`
	Transactions []TxInfo  `json:"transactions"`
}

type TxInfo struct {
	Hash     string `json:"hash"`
	Index    int    `json:"index"`
	Height   int64  `json:"height"`
	Messages int    `json:"messages"`
	Fee      string `json:"fee"`
	Memo     string `json:"memo"`
}

type TxResponse struct {
	Hash   string  `json:"hash"`
	Height int64   `json:"height"`
	Code   uint32  `json:"code"`
	Log    string  `json:"log"`
	Events []Event `json:"events"`
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AccountInfo struct {
	Address       string     `json:"address"`
	AccountNumber uint64     `json:"account_number"`
	Sequence      uint64     `json:"sequence"`
	Balances      sdk.Coins  `json:"balances"`
}

type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      string `json:"pub_key"`
	VotingPower int64  `json:"voting_power"`
	Proposer    int64  `json:"proposer"`
}

package blockchain

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	clienttypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// TxManager handles transaction creation, signing, and broadcasting
type TxManager struct {
	clientCtx client.Context
	config    TxConfig
}

// TxConfig contains transaction configuration
type TxConfig struct {
	GasLimit    uint64
	GasPrice    sdk.DecCoins
	Memo        string
	TimeoutTime time.Duration
}

// NewTxManager creates a new transaction manager
func NewTxManager(clientCtx client.Context, config TxConfig) *TxManager {
	return &TxManager{
		clientCtx: clientCtx,
		config:    config,
	}
}

// DefaultTxConfig returns default transaction configuration
func DefaultTxConfig() TxConfig {
	return TxConfig{
		GasLimit:    200000,
		GasPrice:    sdk.NewDecCoins(sdk.NewDecCoin("umedas", sdk.NewInt(1000))),
		Memo:        "",
		TimeoutTime: 10 * time.Minute,
	}
}

// BuildTx builds a transaction with the given messages
func (tm *TxManager) BuildTx(msgs []sdk.Msg, memo string) (authsigning.Tx, error) {
	txBuilder := tm.clientCtx.TxConfig.NewTxBuilder()

	// Set messages
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set memo
	if memo != "" {
		txBuilder.SetMemo(memo)
	} else {
		txBuilder.SetMemo(tm.config.Memo)
	}

	// Set gas limit
	txBuilder.SetGasLimit(tm.config.GasLimit)

	// Set fee amount
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("umedas", sdk.NewInt(int64(tm.config.GasLimit)))))

	return txBuilder.GetTx(), nil
}

// EstimateGas estimates gas for a transaction
func (tm *TxManager) EstimateGas(msgs []sdk.Msg) (uint64, error) {
	txBuilder := tm.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return 0, fmt.Errorf("failed to set messages: %w", err)
	}

	// Create simulation request
	simReq := tx.SimulateReq{
		TxBytes: nil, // Will be set by the simulation
	}

	// Simulate transaction
	_, gas, err := tx.CalculateGas(tm.clientCtx, simReq)
	if err != nil {
		// If simulation fails, return default gas
		log.Printf("Gas estimation failed, using default: %v", err)
		return tm.config.GasLimit, nil
	}

	// Add 20% buffer to estimated gas
	estimatedGas := uint64(float64(gas) * 1.2)
	return estimatedGas, nil
}

// SignTx signs a transaction
func (tm *TxManager) SignTx(tx authsigning.Tx) (authsigning.Tx, error) {
	// Get account info for sequence number
	account, err := tm.clientCtx.AccountRetriever.GetAccount(tm.clientCtx, tm.clientCtx.GetFromAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Create signing data
	signerData := authsigning.SignerData{
		ChainID:       tm.clientCtx.ChainID,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}

	// Sign transaction
	signedTx, err := tx.AuthInfoBuilder().SetSignatures(signing.SignatureV2{
		PubKey: tm.clientCtx.GetFromAddress().Bytes(),
		Data: &signing.SingleSignatureData{
			SignMode:  tm.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: signerData.Sequence,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	return signedTx, nil
}

// BroadcastTx broadcasts a transaction
func (tm *TxManager) BroadcastTx(tx authsigning.Tx) (*sdk.TxResponse, error) {
	// Encode transaction
	txBytes, err := tm.clientCtx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast transaction
	res, err := tm.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return res, nil
}

// SendTx builds, signs, and broadcasts a transaction
func (tm *TxManager) SendTx(msgs []sdk.Msg, memo string) (*sdk.TxResponse, error) {
	log.Printf("Sending transaction with %d messages", len(msgs))

	// Estimate gas
	estimatedGas, err := tm.EstimateGas(msgs)
	if err != nil {
		log.Printf("Warning: gas estimation failed: %v", err)
		estimatedGas = tm.config.GasLimit
	}

	// Update gas limit
	originalGasLimit := tm.config.GasLimit
	tm.config.GasLimit = estimatedGas
	defer func() {
		tm.config.GasLimit = originalGasLimit
	}()

	// Build transaction
	tx, err := tm.BuildTx(msgs, memo)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Use the simplified broadcasting approach
	txBuilder := tm.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	if memo != "" {
		txBuilder.SetMemo(memo)
	}

	txBuilder.SetGasLimit(estimatedGas)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("umedas", sdk.NewInt(int64(estimatedGas)))))

	// Broadcast using the client context
	return tx.BroadcastTx(tm.clientCtx, txBuilder.GetTx())
}

// RegisterClientTx creates and sends a client registration transaction
func (tm *TxManager) RegisterClientTx(capabilities []string, metadata map[string]interface{}, gpuInfo map[string]interface{}) (*sdk.TxResponse, error) {
	log.Printf("Creating client registration transaction")

	msg, err := CreateRegistrationMessage(
		tm.clientCtx.GetFromAddress().String(),
		capabilities,
		metadata,
		gpuInfo,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration message: %w", err)
	}

	memo := "MedasDigital Client Registration"
	return tm.SendTx([]sdk.Msg{msg}, memo)
}

// StoreAnalysisTx creates and sends an analysis storage transaction
func (tm *TxManager) StoreAnalysisTx(clientID string, result *clienttypes.AnalysisResult) (*sdk.TxResponse, error) {
	log.Printf("Creating analysis storage transaction for client %s", clientID)

	// Prepare metadata
	metadata := map[string]interface{}{
		"duration":      result.Duration.String(),
		"gpu_used":      result.Metadata.GPUUsed,
		"gpu_devices":   result.Metadata.GPUDevices,
		"cpu_cores":     result.Metadata.CPUCores,
		"memory_used":   result.Metadata.MemoryUsedMB,
		"version":       result.Metadata.Version,
		"input_files":   result.Metadata.InputFiles,
		"output_files":  result.Metadata.OutputFiles,
		"parameters":    result.Metadata.Parameters,
	}

	msg, err := CreateAnalysisMessage(
		tm.clientCtx.GetFromAddress().String(),
		clientID,
		result.Type,
		result.Results,
		metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis message: %w", err)
	}

	memo := fmt.Sprintf("Analysis Result: %s", result.Type)
	return tm.SendTx([]sdk.Msg{msg}, memo)
}

// UpdateClientTx creates and sends a client update transaction
func (tm *TxManager) UpdateClientTx(clientID string, capabilities []string, metadata map[string]interface{}, status string) (*sdk.TxResponse, error) {
	log.Printf("Creating client update transaction for client %s", clientID)

	// Marshal metadata
	var metadataStr string
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = string(metadataBytes)
	}

	msg := NewMsgUpdateClient(
		tm.clientCtx.GetFromAddress().String(),
		clientID,
		capabilities,
		metadataStr,
		status,
	)

	memo := fmt.Sprintf("Client Update: %s", clientID)
	return tm.SendTx([]sdk.Msg{msg}, memo)
}

// DeactivateClientTx creates and sends a client deactivation transaction
func (tm *TxManager) DeactivateClientTx(clientID, reason string) (*sdk.TxResponse, error) {
	log.Printf("Creating client deactivation transaction for client %s", clientID)

	msg := NewMsgDeactivateClient(
		tm.clientCtx.GetFromAddress().String(),
		clientID,
		reason,
	)

	memo := fmt.Sprintf("Client Deactivation: %s", clientID)
	return tm.SendTx([]sdk.Msg{msg}, memo)
}

// WaitForTx waits for a transaction to be included in a block
func (tm *TxManager) WaitForTx(txHash string, timeout time.Duration) (*TxResponse, error) {
	log.Printf("Waiting for transaction %s to be included in block", txHash)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction %s", txHash)
		case <-ticker.C:
			// Query transaction
			res, err := tm.clientCtx.Client.Tx(ctx, []byte(txHash), true)
			if err != nil {
				continue // Transaction not yet included
			}

			return &TxResponse{
				Hash:   res.Hash.String(),
				Height: res.Height,
				Code:   res.TxResult.Code,
				Log:    res.TxResult.Log,
				Events: parseEvents(res.TxResult.Events),
			}, nil
		}
	}
}

// GetTxStatus returns the status of a transaction
func (tm *TxManager) GetTxStatus(txHash string) (*TxStatus, error) {
	res, err := tm.clientCtx.Client.Tx(context.Background(), []byte(txHash), true)
	if err != nil {
		return &TxStatus{
			Hash:   txHash,
			Status: "not_found",
			Error:  err.Error(),
		}, nil
	}

	status := &TxStatus{
		Hash:        res.Hash.String(),
		Height:      res.Height,
		Code:        res.TxResult.Code,
		Log:         res.TxResult.Log,
		GasWanted:   res.TxResult.GasWanted,
		GasUsed:     res.TxResult.GasUsed,
		Events:      parseEvents(res.TxResult.Events),
	}

	if res.TxResult.Code == 0 {
		status.Status = "success"
	} else {
		status.Status = "failed"
		status.Error = res.TxResult.Log
	}

	return status, nil
}

// BatchTx sends multiple transactions in sequence
func (tm *TxManager) BatchTx(txRequests []TxRequest, waitForInclusion bool) ([]*sdk.TxResponse, error) {
	log.Printf("Sending batch of %d transactions", len(txRequests))

	var responses []*sdk.TxResponse
	var errors []error

	for i, req := range txRequests {
		log.Printf("Sending transaction %d/%d: %s", i+1, len(txRequests), req.Memo)

		res, err := tm.SendTx(req.Messages, req.Memo)
		if err != nil {
			errors = append(errors, fmt.Errorf("transaction %d failed: %w", i, err))
			responses = append(responses, nil)
			continue
		}

		responses = append(responses, res)

		// Wait for inclusion if requested
		if waitForInclusion && res.Code == 0 {
			_, err := tm.WaitForTx(res.TxHash, 30*time.Second)
			if err != nil {
				log.Printf("Warning: failed to wait for transaction %s: %v", res.TxHash, err)
			}
		}

		// Small delay between transactions to avoid sequence conflicts
		time.Sleep(100 * time.Millisecond)
	}

	if len(errors) > 0 {
		return responses, fmt.Errorf("batch transaction errors: %v", errors)
	}

	return responses, nil
}

// Data structures

// TxRequest represents a transaction request
type TxRequest struct {
	Messages []sdk.Msg
	Memo     string
}

// TxStatus represents transaction status information
type TxStatus struct {
	Hash      string  `json:"hash"`
	Height    int64   `json:"height"`
	Code      uint32  `json:"code"`
	Status    string  `json:"status"`
	Log       string  `json:"log"`
	Error     string  `json:"error,omitempty"`
	GasWanted int64   `json:"gas_wanted"`
	GasUsed   int64   `json:"gas_used"`
	Events    []Event `json:"events"`
}

// Helper functions

// CreateBatchRegistrations creates a batch of registration transactions
func CreateBatchRegistrations(clients []ClientRegistration) []TxRequest {
	var requests []TxRequest

	for _, client := range clients {
		msg, err := CreateRegistrationMessage(
			client.Creator,
			client.Capabilities,
			client.Metadata,
			client.GPUInfo,
		)
		if err != nil {
			log.Printf("Warning: failed to create registration for %s: %v", client.Creator, err)
			continue
		}

		requests = append(requests, TxRequest{
			Messages: []sdk.Msg{msg},
			Memo:     fmt.Sprintf("Batch registration for %s", client.Creator),
		})
	}

	return requests
}

// CreateBatchAnalysis creates a batch of analysis storage transactions
func CreateBatchAnalysis(results []AnalysisStorage) []TxRequest {
	var requests []TxRequest

	for _, result := range results {
		msg, err := CreateAnalysisMessage(
			result.Creator,
			result.ClientID,
			result.AnalysisType,
			result.Data,
			result.Metadata,
		)
		if err != nil {
			log.Printf("Warning: failed to create analysis storage for %s: %v", result.ClientID, err)
			continue
		}

		requests = append(requests, TxRequest{
			Messages: []sdk.Msg{msg},
			Memo:     fmt.Sprintf("Analysis: %s", result.AnalysisType),
		})
	}

	return requests
}

// Data structures for batch operations

type ClientRegistration struct {
	Creator      string                 `json:"creator"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
	GPUInfo      map[string]interface{} `json:"gpu_info"`
}

type AnalysisStorage struct {
	Creator      string                 `json:"creator"`
	ClientID     string                 `json:"client_id"`
	AnalysisType string                 `json:"analysis_type"`
	Data         map[string]interface{} `json:"data"`
	Metadata     map[string]interface{} `json:"metadata"`
}

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"

	"github.com/oxygene76/medasdigital-client/internal/types"
	"github.com/oxygene76/medasdigital-client/pkg/analysis"
	"github.com/oxygene76/medasdigital-client/pkg/blockchain"
	"github.com/oxygene76/medasdigital-client/pkg/gpu"
	"github.com/oxygene76/medasdigital-client/pkg/utils"
)

// MedasDigitalClient represents the main client for astronomical analysis
type MedasDigitalClient struct {
	config       *utils.Config
	clientCtx    client.Context
	clientID     string
	capabilities []string
	isRegistered bool
	gpuManager   *gpu.Manager
	analyzer     *analysis.Manager
	blockchain   *blockchain.Client
}

// NewMedasDigitalClient creates a new MedasDigital client instance
func NewMedasDigitalClient() (*MedasDigitalClient, error) {
	config, err := utils.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := &MedasDigitalClient{
		config:       config,
		capabilities: config.Client.Capabilities,
		isRegistered: false,
	}

	if err := client.initializeBlockchainClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain client: %w", err)
	}

	if config.GPU.Enabled {
		client.gpuManager, err = gpu.NewManager(config.GPU)
		if err != nil {
			log.Printf("Warning: Failed to initialize GPU manager: %v", err)
		}
	}

	client.analyzer = analysis.NewManager(client.gpuManager)
	client.blockchain = blockchain.NewClient(client.clientCtx)

	return client, nil
}

func (c *MedasDigitalClient) initializeBlockchainClient() error {
	// Codec setup
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// Address codecs - required for 0.50
	addressCodec := address.NewBech32Codec(c.config.Chain.Bech32Prefix)
	validatorAddressCodec := address.NewBech32Codec(c.config.Chain.Bech32Prefix + "valoper")
	consensusAddressCodec := address.NewBech32Codec(c.config.Chain.Bech32Prefix + "valcons")

	// Keyring setup
	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendOS, c.config.Client.KeyringDir, nil, addressCodec)
	if err != nil {
		return fmt.Errorf("failed to create keyring: %w", err)
	}

	// RPC Client setup - Updated to use CometBFT
	rpcClient, err := comethttp.New(c.config.Chain.RPCEndpoint, "/websocket")
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	// Client Context setup with address codecs
	c.clientCtx = client.Context{}.
		WithCodec(marshaler).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)).
		WithLegacyAmino(codec.NewLegacyAmino()).
		WithAccountRetriever(authtx.NewAccountRetriever(marshaler)).
		WithBroadcastMode("block").
		WithChainID(c.config.Chain.ID).
		WithKeyring(kr).
		WithClient(rpcClient).
		WithAddressCodec(addressCodec).
		WithValidatorAddressCodec(validatorAddressCodec).
		WithConsensusAddressCodec(consensusAddressCodec)

	return nil
}

// Register registers the client on the blockchain
func (c *MedasDigitalClient) Register(capabilities []string, metadata, from string) error {
	log.Println("Registering client on blockchain...")

	// Set from address
	info, err := c.clientCtx.Keyring.Key(from)
	if err != nil {
		return fmt.Errorf("failed to get key info: %w", err)
	}

	addr, err := info.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	c.clientCtx = c.clientCtx.WithFromAddress(addr).WithFromName(from)

	// Create registration message
	msg := &types.MsgRegisterClient{
		Creator:      addr.String(),
		Capabilities: capabilities,
		Metadata:     c.generateMetadata(metadata),
	}

	// Send transaction
	res, err := c.sendTransaction(msg)
	if err != nil {
		return fmt.Errorf("failed to send registration transaction: %w", err)
	}

	// Extract client ID from response
	c.clientID = c.extractClientID(res)
	c.capabilities = capabilities
	c.isRegistered = true

	log.Printf("Client registered successfully with ID: %s", c.clientID)
	return nil
}

func (c *MedasDigitalClient) generateMetadata(userMetadata string) string {
	metadata := map[string]interface{}{
		"version":       "1.0.0",
		"capabilities":  c.capabilities,
		"created_at":    time.Now().Format(time.RFC3339),
		"user_metadata": userMetadata,
		"gpu_enabled":   c.config.GPU.Enabled,
	}

	if c.gpuManager != nil {
		gpuInfo, _ := c.gpuManager.GetInfo()
		metadata["gpu_info"] = gpuInfo
	}

	data, _ := json.Marshal(metadata)
	return string(data)
}

func (c *MedasDigitalClient) sendTransaction(msg sdk.Msg) (*sdk.TxResponse, error) {
	txBuilder := c.clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, err
	}

	// Set gas and fees (simplified)
	txBuilder.SetGasLimit(200000)

	return tx.BroadcastTx(c.clientCtx, txBuilder.GetTx())
}

func (c *MedasDigitalClient) extractClientID(res *sdk.TxResponse) string {
	// Extract client ID from transaction events
	// This is a simplified implementation
	return fmt.Sprintf("client_%s_%d", res.TxHash[:8], time.Now().Unix())
}

// AnalyzeOrbitalDynamics performs orbital dynamics analysis
func (c *MedasDigitalClient) AnalyzeOrbitalDynamics(inputFile, outputFile string) error {
	if !c.hasCapability("orbital_dynamics") {
		return fmt.Errorf("client does not have orbital_dynamics capability")
	}

	log.Printf("Starting orbital dynamics analysis on file: %s", inputFile)

	result, err := c.analyzer.AnalyzeOrbitalDynamics(inputFile)
	if err != nil {
		return fmt.Errorf("orbital dynamics analysis failed: %w", err)
	}

	// Store results on blockchain
	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store results: %w", err)
	}

	// Save local results
	if outputFile != "" {
		if err := c.saveResults(result, outputFile); err != nil {
			return fmt.Errorf("failed to save results locally: %w", err)
		}
	}

	log.Printf("Orbital dynamics analysis completed successfully")
	return nil
}

// AnalyzePhotometric performs photometric analysis
func (c *MedasDigitalClient) AnalyzePhotometric(surveyData, targetList string) error {
	if !c.hasCapability("photometric_analysis") {
		return fmt.Errorf("client does not have photometric_analysis capability")
	}

	log.Printf("Starting photometric analysis on survey data: %s", surveyData)

	result, err := c.analyzer.AnalyzePhotometric(surveyData, targetList)
	if err != nil {
		return fmt.Errorf("photometric analysis failed: %w", err)
	}

	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("Photometric analysis completed successfully")
	return nil
}

// AnalyzeClustering performs clustering analysis
func (c *MedasDigitalClient) AnalyzeClustering() error {
	if !c.hasCapability("clustering_analysis") {
		return fmt.Errorf("client does not have clustering_analysis capability")
	}

	log.Printf("Starting clustering analysis")

	result, err := c.analyzer.AnalyzeClustering()
	if err != nil {
		return fmt.Errorf("clustering analysis failed: %w", err)
	}

	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("Clustering analysis completed successfully")
	return nil
}

// AIDetection performs AI-powered object detection
func (c *MedasDigitalClient) AIDetection(modelPath, surveyImages string, gpuAccel bool) error {
	if !c.hasCapability("ai_training") {
		return fmt.Errorf("client does not have ai_training capability")
	}

	if gpuAccel && c.gpuManager == nil {
		return fmt.Errorf("GPU acceleration requested but no GPU available")
	}

	log.Printf("Starting AI detection on survey images: %s", surveyImages)

	result, err := c.analyzer.AIDetection(modelPath, surveyImages, gpuAccel)
	if err != nil {
		return fmt.Errorf("AI detection failed: %w", err)
	}

	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store results: %w", err)
	}

	log.Printf("AI detection completed successfully")
	return nil
}

// TrainDeepDetector trains a deep learning detector
func (c *MedasDigitalClient) TrainDeepDetector(trainingData, architecture string, gpuDevices []int, batchSize, epochs int) error {
	if !c.hasCapability("ai_training") {
		return fmt.Errorf("client does not have ai_training capability")
	}

	if c.gpuManager == nil {
		return fmt.Errorf("GPU training requested but no GPU available")
	}

	log.Printf("Starting deep detector training with architecture: %s", architecture)

	result, err := c.analyzer.TrainDeepDetector(trainingData, architecture, gpuDevices, batchSize, epochs)
	if err != nil {
		return fmt.Errorf("training failed: %w", err)
	}

	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store training results: %w", err)
	}

	log.Printf("Deep detector training completed successfully")
	return nil
}

// TrainAnomalyDetector trains an anomaly detection model
func (c *MedasDigitalClient) TrainAnomalyDetector() error {
	log.Printf("Starting anomaly detector training")

	result, err := c.analyzer.TrainAnomalyDetector()
	if err != nil {
		return fmt.Errorf("anomaly detector training failed: %w", err)
	}

	if err := c.storeAnalysisResult(result); err != nil {
		return fmt.Errorf("failed to store training results: %w", err)
	}

	log.Printf("Anomaly detector training completed successfully")
	return nil
}

// Status returns the current client status
func (c *MedasDigitalClient) Status() error {
	fmt.Printf("=== MedasDigital Client Status ===\n")
	fmt.Printf("Client ID: %s\n", c.clientID)
	fmt.Printf("Registered: %t\n", c.isRegistered)
	fmt.Printf("Capabilities: %v\n", c.capabilities)
	fmt.Printf("Chain ID: %s\n", c.config.Chain.ID)
	fmt.Printf("RPC Endpoint: %s\n", c.config.Chain.RPCEndpoint)

	// Blockchain status
	status, err := c.clientCtx.Client.Status(context.Background())
	if err != nil {
		fmt.Printf("Blockchain Status: ERROR - %v\n", err)
	} else {
		fmt.Printf("Blockchain Status: Connected (Block: %d)\n", status.SyncInfo.LatestBlockHeight)
	}

	// GPU status
	if c.gpuManager != nil {
		fmt.Printf("GPU Status: Available\n")
		gpuInfo, err := c.gpuManager.GetInfo()
		if err == nil {
			data, _ := json.MarshalIndent(gpuInfo, "", "  ")
			fmt.Printf("GPU Info:\n%s\n", string(data))
		}
	} else {
		fmt.Printf("GPU Status: Not Available\n")
	}

	return nil
}

// Results retrieves recent analysis results
func (c *MedasDigitalClient) Results(limit int) error {
	fmt.Printf("=== Recent Analysis Results (limit: %d) ===\n", limit)

	results, err := c.blockchain.GetAnalysisResults(c.clientID, limit)
	if err != nil {
		return fmt.Errorf("failed to retrieve results: %w", err)
	}

	for i, result := range results {
		fmt.Printf("\n--- Result %d ---\n", i+1)
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("%s\n", string(data))
	}

	return nil
}

// Query queries blockchain data
func (c *MedasDigitalClient) Query(queryType, queryID string) error {
	fmt.Printf("=== Querying %s: %s ===\n", queryType, queryID)

	result, err := c.blockchain.Query(queryType, queryID)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("%s\n", string(data))

	return nil
}

// GPUStatus returns GPU status information
func (c *MedasDigitalClient) GPUStatus() error {
	if c.gpuManager == nil {
		fmt.Println("GPU not available or not enabled")
		return nil
	}

	return c.gpuManager.PrintStatus()
}

// GPUBenchmark runs a GPU benchmark
func (c *MedasDigitalClient) GPUBenchmark() error {
	if c.gpuManager == nil {
		return fmt.Errorf("GPU not available or not enabled")
	}

	return c.gpuManager.RunBenchmark()
}

func (c *MedasDigitalClient) hasCapability(capability string) bool {
	for _, cap := range c.capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

func (c *MedasDigitalClient) storeAnalysisResult(result *types.AnalysisResult) error {
	if !c.isRegistered {
		return fmt.Errorf("client not registered")
	}

	msg := &types.MsgStoreAnalysis{
		Creator:      c.clientCtx.GetFromAddress().String(),
		ClientID:     c.clientID,
		AnalysisType: result.Type,
		Data:         result.Results,
		BlockHeight:  result.BlockHeight,
		TxHash:       result.TxHash,
	}

	_, err := c.sendTransaction(msg)
	return err
}

func (c *MedasDigitalClient) saveResults(result *types.AnalysisResult, outputFile string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return utils.WriteFile(outputFile, data)
}

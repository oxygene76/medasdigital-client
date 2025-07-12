package main

import (
	"context"
    	"fmt"
	"os"
	"path/filepath"
	"time"
	"strings"
	sdkmath "cosmossdk.io/math"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	
	"github.com/cosmos/cosmos-sdk/client/tx"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"


	
	// ✅ KORREKTE v0.50 IMPORTS für echte Blockchain-Kommunikation:
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"           // Für TxConfig
	"github.com/cosmos/cosmos-sdk/client/flags"              // Für BroadcastMode
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"    // Für AccountRetriever

	blockchain "github.com/oxygene76/medasdigital-client/pkg/blockchain"  // Wieder hinzufügen
	medasClient "github.com/oxygene76/medasdigital-client/pkg/client"
)

const (
	// Application constants
	appName = "medasdigital-client"
	version = "v1.0.0"
	
	// Default chain configuration
	defaultChainID     = "medasdigital-2"
	defaultRPCEndpoint = "https://rpc.medas-digital.io:26657"
	defaultBech32Prefix = "medas"
	defaultBaseDenom    = "umedas"        // ← NEU HINZUFÜGEN
)

var (
	// Global client instance
	globalClient *medasClient.MedasDigitalClient
	
	// Configuration
	cfgFile string
	homeDir string
	
	// ✅ NEU: Globale Registry-Instanzen um Konflikte zu vermeiden
	globalInterfaceRegistry types.InterfaceRegistry
	globalCodec             codec.Codec
)

// Client Registration Data Structure
type ClientRegistrationData struct {
	ClientAddress string    `json:"client_address"`
	Capabilities  []string  `json:"capabilities"`
	Metadata      string    `json:"metadata,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
}

// Registration Result Storage
type RegistrationResult struct {
	TransactionHash  string                 `json:"transaction_hash"`
	ClientID         string                 `json:"client_id"`
	RegistrationData ClientRegistrationData `json:"registration_data"`
	BlockHeight      int64                  `json:"block_height,omitempty"`
	RegisteredAt     time.Time              `json:"registered_at"`
}

// Config represents the application configuration
type Config struct {
	Chain struct {
		ID           string `yaml:"chain_id"`
		RPCEndpoint  string `yaml:"rpc_endpoint"`
		Bech32Prefix string `yaml:"bech32_prefix"`
		BaseDenom    string `yaml:"base_denom"` 
	} `yaml:"chain"`
	Client struct {
		KeyringDir     string   `yaml:"keyring_dir"`
		KeyringBackend string   `yaml:"keyring_backend"`  // ← NEU HINZUFÜGEN
		Capabilities   []string `yaml:"capabilities"`
	} `yaml:"client"`
	GPU struct {
		Enabled     bool `yaml:"enabled"`
		DeviceID    int  `yaml:"device_id"`
		MemoryLimit int  `yaml:"memory_limit"`
	} `yaml:"gpu"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "MedasDigital Client for astronomical analysis",
	Long: `MedasDigital Client is a command-line tool for conducting astronomical 
analysis on the MedasDigital blockchain network. It supports various analysis 
types including orbital dynamics, photometric analysis, clustering, and 
AI-powered object detection with GPU acceleration.

The client integrates with the MedasDigital blockchain to store analysis 
results and collaborate with other researchers in the search for Planet 9 
and other astronomical objects.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		if err := initConfig(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		
		// Initialize client context for blockchain commands
		if cmd.Name() != "init" && cmd.Name() != "version" && cmd.Name() != "help" {
			if err := initializeClient(); err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}
		}
		
		return nil
	},
}

// initCmd initializes the client configuration
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize client configuration",
	Long: `Initialize the MedasDigital client configuration. This creates the 
default configuration file and sets up the local directories needed for 
operation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Initializing MedasDigital Client v%s\n", version)
		
		// Create home directory
		if err := os.MkdirAll(homeDir, 0755); err != nil {
			return fmt.Errorf("failed to create home directory: %w", err)
		}
		
		// Create default configuration
config := &Config{
    Chain: struct {
        ID           string `yaml:"chain_id"`
        RPCEndpoint  string `yaml:"rpc_endpoint"`
        Bech32Prefix string `yaml:"bech32_prefix"`
        BaseDenom    string `yaml:"base_denom"`     // ← NEU HINZUFÜGEN
    }{
        ID:           defaultChainID,
        RPCEndpoint:  defaultRPCEndpoint,
        Bech32Prefix: defaultBech32Prefix,
        BaseDenom:    defaultBaseDenom,                       // ← NEU HINZUFÜGEN
    },
			Client: struct {
			KeyringDir     string   `yaml:"keyring_dir"`
			KeyringBackend string   `yaml:"keyring_backend"`  // ← NEU
			Capabilities   []string `yaml:"capabilities"`
		}{
			KeyringDir:     filepath.Join(homeDir, "keyring"),
			KeyringBackend: "test",  // ← NEU HINZUFÜGEN
			Capabilities:   []string{"orbital_dynamics", "photometric_analysis"},
		},
			GPU: struct {
				Enabled     bool `yaml:"enabled"`
				DeviceID    int  `yaml:"device_id"`
				MemoryLimit int  `yaml:"memory_limit"`
			}{
				Enabled:     false,
				DeviceID:    0,
				MemoryLimit: 8192, // 8GB default
			},
		}
		
		// Save configuration using viper
		viper.Set("chain", config.Chain)
		viper.Set("client", config.Client)
		viper.Set("gpu", config.GPU)
		
		if err := viper.WriteConfigAs(cfgFile); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		
		fmt.Printf("Configuration initialized at: %s\n", cfgFile)
		fmt.Printf("Home directory: %s\n", homeDir)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Configure your settings in the config file")
		fmt.Println("2. Register your client: medasdigital-client register")
		fmt.Println("3. Check status: medasdigital-client status")
		
		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register client on the blockchain",
	Long: `Register this client on the MedasDigital blockchain. This assigns 
a unique client ID and registers the client's capabilities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		capabilities, _ := cmd.Flags().GetStringSlice("capabilities")
		metadata, _ := cmd.Flags().GetString("metadata")
		from, _ := cmd.Flags().GetString("from")
		keyringBackend, _ := cmd.Flags().GetString("keyring-backend")
		gas, _ := cmd.Flags().GetUint64("gas")

		 if gas > 0 {
     		   fmt.Printf("⛽ Using manual gas limit: %d\n", gas)
    		} else {
   		fmt.Println("⛽ Using automatic gas estimation")
    		}
		
		if from == "" {
			return fmt.Errorf("--from flag is required")
		}
		
		if len(capabilities) == 0 {
			// Use default capabilities from config
			capabilities = viper.GetStringSlice("client.capabilities")
			if len(capabilities) == 0 {
				capabilities = []string{"orbital_dynamics", "photometric_analysis"}
			}
		}
		
		fmt.Printf("Registering client with capabilities: %v\n", capabilities)
		
		// Use our custom keyring context
		clientCtx, err := initKeysClientContextWithBackend(keyringBackend)
		if err != nil {
			return fmt.Errorf("failed to initialize client context: %w", err)
		}
		
		// Get key info to verify it exists
		keyInfo, err := clientCtx.Keyring.Key(from)
		if err != nil {
			fmt.Printf("Key '%s' not found. Create it first with:\n", from)
			fmt.Printf("  ./bin/medasdigital-client keys add %s --keyring-backend %s\n", from, keyringBackend)
			return fmt.Errorf("failed to get key info for '%s': %v", from, err)
		}
		
		addr, err := keyInfo.GetAddress()
		if err != nil {
			return fmt.Errorf("failed to get address from key: %w", err)
		}
		
		fmt.Printf("Using key '%s' with address: %s\n", from, addr.String())
		
		// ✅ SICHERER ANSATZ: Test connection first, dann entscheiden
		cfg := loadConfig()
		fmt.Printf("🔍 Testing connection to %s...\n", cfg.Chain.RPCEndpoint)
		
		if err := testBlockchainConnection(cfg.Chain.RPCEndpoint); err != nil {
    fmt.Printf("⚠️  Blockchain connection failed: %v\n", err)
    fmt.Println("💡 Running in simulation mode...")
    return simulateRegistration(from, addr.String(), capabilities, metadata)
}

fmt.Println("✅ Blockchain connection successful!")
fmt.Println("🔗 Connected to:", cfg.Chain.RPCEndpoint)
fmt.Println("⛓️  Chain ID:", cfg.Chain.ID)

// ✅ ECHTE BLOCKCHAIN-REGISTRIERUNG
fmt.Println("📡 Creating blockchain client for real transaction...")

// Create blockchain client with complete context

// Prepare metadata
metadataMap := make(map[string]interface{})
if metadata != "" {
    metadataMap["description"] = metadata
}
metadataMap["timestamp"] = time.Now().Unix()
metadataMap["client_version"] = version
metadataMap["registration_type"] = "client_registration"

fmt.Println("📝 Registering client using standard blockchain transaction...")

// Set FromName in clientCtx for signing
// Set FromName in clientCtx for signing and ensure TxConfig is set
if globalInterfaceRegistry == nil {
    globalInterfaceRegistry = getInterfaceRegistry()
}
if globalCodec == nil {
    globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
}

// Create TxConfig
txConfig := authtx.NewTxConfig(globalCodec, authtx.DefaultSignModes)

// Create RPC client
rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
if err != nil {
    fmt.Printf("❌ Failed to create RPC client: %v\n", err)
    return simulateRegistration(from, addr.String(), capabilities, metadata)
}

// Complete clientCtx with all required fields
clientCtx = clientCtx.
    WithFromName(from).
    WithFromAddress(addr).
    WithTxConfig(txConfig).
    WithClient(rpcClient).
    WithChainID(cfg.Chain.ID).
    WithCodec(globalCodec).
    WithInterfaceRegistry(globalInterfaceRegistry).
    WithBroadcastMode(flags.BroadcastSync)

// Use simple registration method
result, err := registerClientSimple(clientCtx, addr.String(), capabilities, metadata, gas)
if err != nil {
    fmt.Printf("❌ Blockchain registration failed: %v\n", err)
    fmt.Println("💡 This might be due to:")
    fmt.Println("   • Insufficient funds for transaction fees")
    fmt.Println("   • Chain not accepting transactions")
    fmt.Println("   • Network connectivity issues")
    fmt.Println("\n💡 Falling back to simulation...")
    return simulateRegistration(from, addr.String(), capabilities, metadata)
}

// ✅ SUCCESS!
fmt.Println("\n🎉 CLIENT SUCCESSFULLY REGISTERED ON BLOCKCHAIN!")
fmt.Println("=" + strings.Repeat("=", 50))
fmt.Printf("🆔 Client ID: %s\n", result.ClientID)
fmt.Printf("📍 Address: %s\n", result.RegistrationData.ClientAddress)
fmt.Printf("⛓️  Chain: %s\n", cfg.Chain.ID)
fmt.Printf("🔧 Capabilities: %v\n", result.RegistrationData.Capabilities)
fmt.Printf("📊 Transaction Hash: %s\n", result.TransactionHash)
fmt.Printf("🏔️  Block Height: %d\n", result.BlockHeight)
fmt.Printf("🕒 Registered: %s\n", result.RegisteredAt.Format("2006-01-02 15:04:05"))
fmt.Printf("💾 Registration saved to: ~/.medasdigital-client/registrations/\n")

if metadata != "" {
    fmt.Printf("📋 Metadata: %s\n", metadata)
}

fmt.Println("=" + strings.Repeat("=", 50))
fmt.Println("✅ Your client is now active on the MedasDigital network!")
fmt.Println("\n💡 To verify registration on blockchain:")
fmt.Printf("   ./bin/medasdigital-client query tx %s\n", result.TransactionHash)

return nil

	},
}

// statusCmd shows the client status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show client status",
	Long:  "Display the current status of the MedasDigital client including blockchain connection and GPU status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return globalClient.Status()
	},
}

var checkAccountCmd = &cobra.Command{
	Use:   "check-account [address]",
	Short: "Check account status on blockchain",
	Long:  "Check if an account exists on the blockchain and show its details",
	Args:  cobra.RangeArgs(0, 1),  // ✅ FIXED: RangeArgs statt MaxArgs
	RunE: func(cmd *cobra.Command, args []string) error {
		var address string
		
		if len(args) > 0 {
			address = args[0]
		} else {
			// Use default key
			from, _ := cmd.Flags().GetString("from")
			if from == "" {
				return fmt.Errorf("please provide address or use --from flag")
			}
			
			clientCtx, err := initKeysClientContext()
			if err != nil {
				return fmt.Errorf("failed to initialize client context: %w", err)
			}
			
			keyInfo, err := clientCtx.Keyring.Key(from)
			if err != nil {
				return fmt.Errorf("key not found: %w", err)
			}
			
			addr, err := keyInfo.GetAddress()
			if err != nil {
				return fmt.Errorf("failed to get address: %w", err)
			}
			
			address = addr.String()
		}
		
		fmt.Printf("🔍 Checking account: %s\n", address)
		
		// Load config
		cfg := loadConfig()
		
		// Test connection first
		fmt.Printf("🔗 Connecting to: %s\n", cfg.Chain.RPCEndpoint)
		if err := testBlockchainConnection(cfg.Chain.RPCEndpoint); err != nil {
			return fmt.Errorf("blockchain connection failed: %w", err)
		}
		
		fmt.Println("✅ Blockchain connection successful!")
		
		// Create RPC client for account query
		rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
		if err != nil {
			return fmt.Errorf("failed to create RPC client: %w", err)
		}
		
		// Create minimal client context for query
		if globalInterfaceRegistry == nil {
			globalInterfaceRegistry = getInterfaceRegistry()
		}
		if globalCodec == nil {
			globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
		}
		
		queryCtx := client.Context{}.
			WithClient(rpcClient).
			WithChainID(cfg.Chain.ID).
			WithCodec(globalCodec).
			WithInterfaceRegistry(globalInterfaceRegistry)
		
		// Parse address for validation
		_, err = sdk.AccAddressFromBech32(address)
		if err != nil {
			return fmt.Errorf("invalid address format: %w", err)
		}
		
		// Try to query account using v0.50.10 methods
fmt.Println("📊 Querying account information (Cosmos SDK v0.50.10)...")

// Parse address for proper encoding
addr, err := sdk.AccAddressFromBech32(address)
if err != nil {
    return fmt.Errorf("invalid address format: %w", err)
}

fmt.Printf("🔍 Address Details:\n")
fmt.Printf("   Bech32: %s\n", address)
fmt.Printf("   Hex: %X\n", addr.Bytes())
fmt.Printf("   Length: %d bytes\n", len(addr.Bytes()))
fmt.Println()

// TEST 1: Direct gRPC-style query (v0.50.10 method)
fmt.Println("🔍 Testing gRPC Query (v0.50.10 method):")

// Create proper query request
queryReq := fmt.Sprintf(`{"address":"%s"}`, address)
fmt.Printf("   Query Request: %s\n", queryReq)

// Try the auth module query
authQueryPath := "/cosmos.auth.v1beta1.Query/Account"
fmt.Printf("   Query Path: %s\n", authQueryPath)

authRes, authHeight, authErr := queryCtx.QueryWithData(authQueryPath, []byte(queryReq))
fmt.Printf("   Auth Query Result:\n")
fmt.Printf("     Error: %v\n", authErr)
fmt.Printf("     Height: %d\n", authHeight)
fmt.Printf("     Response Length: %d bytes\n", len(authRes))

if authErr == nil && len(authRes) > 0 {
    fmt.Printf("     Response Data: %s\n", string(authRes))
} else if authErr != nil {
    fmt.Printf("     Error Type: %T\n", authErr)
    fmt.Printf("     Error Details: %s\n", authErr.Error())
}
fmt.Println()

// TEST 2: Bank balance query (v0.50.10)
fmt.Println("🔍 Testing Bank Balance Query (v0.50.10):")

// Try different denom names that might be used
denoms := []string{"stake", "token", "medas", "umedas", "atom", "uatom"}

for _, denom := range denoms {
    balanceQueryReq := fmt.Sprintf(`{"address":"%s","denom":"%s"}`, address, denom)
    balanceQueryPath := "/cosmos.bank.v1beta1.Query/Balance"
    
    fmt.Printf("   Testing denom '%s':\n", denom)
    fmt.Printf("     Query: %s\n", balanceQueryReq)
    
    balRes, balHeight, balErr := queryCtx.QueryWithData(balanceQueryPath, []byte(balanceQueryReq))
    fmt.Printf("     Result: ")
    if balErr != nil {
        fmt.Printf("Error - %v\n", balErr)
    } else {
        fmt.Printf("Success - %d bytes, height %d\n", len(balRes), balHeight)
        if len(balRes) > 0 && len(balRes) < 1000 {
            fmt.Printf("     Data: %s\n", string(balRes))
        }
    }
}
fmt.Println()

// TEST 3: All balances query
fmt.Println("🔍 Testing All Balances Query (v0.50.10):")
allBalancesReq := fmt.Sprintf(`{"address":"%s"}`, address)
allBalancesPath := "/cosmos.bank.v1beta1.Query/AllBalances"
fmt.Printf("   Query: %s\n", allBalancesReq)

allBalRes, allBalHeight, allBalErr := queryCtx.QueryWithData(allBalancesPath, []byte(allBalancesReq))
fmt.Printf("   Result:\n")
fmt.Printf("     Error: %v\n", allBalErr)
fmt.Printf("     Height: %d\n", allBalHeight)
fmt.Printf("     Response Length: %d bytes\n", len(allBalRes))

if allBalErr == nil && len(allBalRes) > 0 {
    fmt.Printf("     Response Data: %s\n", string(allBalRes))
}
fmt.Println()

// TEST 4: Chain information that works
fmt.Println("🔍 Working Chain Information:")
status, err := queryCtx.Client.Status(context.Background())
if err != nil {
    fmt.Printf("   Status Error: %v\n", err)
} else {
    fmt.Printf("   Chain ID: %s\n", status.NodeInfo.Network)
    fmt.Printf("   Latest Height: %d\n", status.SyncInfo.LatestBlockHeight)
    fmt.Printf("   Latest Block Time: %s\n", status.SyncInfo.LatestBlockTime)
    fmt.Printf("   App Version: %s\n", status.NodeInfo.Version)
    fmt.Printf("   Catching Up: %t\n", status.SyncInfo.CatchingUp)
}
fmt.Println()

// TEST 5: Try AccountRetriever (v0.50.10 way)
fmt.Println("🔍 Testing AccountRetriever (v0.50.10 method):")
accountRetriever := authtypes.AccountRetriever{}

account, accErr := accountRetriever.GetAccount(queryCtx, addr)
fmt.Printf("   AccountRetriever Result:\n")
fmt.Printf("     Error: %v\n", accErr)

if accErr == nil && account != nil {
    fmt.Printf("     Account Found: ✅\n")
    fmt.Printf("     Account Number: %d\n", account.GetAccountNumber())
    fmt.Printf("     Sequence: %d\n", account.GetSequence())
    fmt.Printf("     Address: %s\n", account.GetAddress().String())
    fmt.Printf("     PubKey: %v\n", account.GetPubKey())
} else if accErr != nil {
    fmt.Printf("     Error Type: %T\n", accErr)
    fmt.Printf("     Error Details: %s\n", accErr.Error())
}
fmt.Println()

// SUMMARY
fmt.Println("📋 Summary (Cosmos SDK v0.50.10):")
fmt.Printf("   Address: %s\n", address)
fmt.Printf("   Chain: %s\n", cfg.Chain.ID)
fmt.Printf("   SDK Version: v0.50.10\n")
fmt.Printf("   RPC Connection: ✅ Working\n")

// Determine what we actually found
if authErr == nil || allBalErr == nil || accErr == nil {
    fmt.Printf("   Account Status: ✅ Found via at least one method\n")
} else {
    fmt.Printf("   Account Status: ❓ Not found via tested methods\n")
    fmt.Printf("   Note: Account may exist but use different query format\n")
}

return nil
		
	},
}

// analyzeCmd represents the analyze command group
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Perform astronomical analysis",
	Long:  "Perform various types of astronomical analysis including orbital dynamics, photometric analysis, and clustering.",
}

// analyzeOrbitalCmd performs orbital dynamics analysis
var analyzeOrbitalCmd = &cobra.Command{
	Use:   "orbital-dynamics [input-file]",
	Short: "Perform orbital dynamics analysis",
	Long: `Analyze orbital dynamics of Trans-Neptunian Objects (TNOs) and 
calculate Planet 9 influence probabilities.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		outputFile, _ := cmd.Flags().GetString("output")
		
		fmt.Printf("Starting orbital dynamics analysis on: %s\n", inputFile)
		
		if err := globalClient.AnalyzeOrbitalDynamics(inputFile, outputFile); err != nil {
			return fmt.Errorf("orbital dynamics analysis failed: %w", err)
		}
		
		fmt.Println("Orbital dynamics analysis completed!")
		return nil
	},
}

// analyzePhotometricCmd performs photometric analysis
var analyzePhotometricCmd = &cobra.Command{
	Use:   "photometric [survey-data]",
	Short: "Perform photometric analysis",
	Long:  "Analyze photometric survey data to identify variable objects and light curves.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		surveyData := args[0]
		targetList, _ := cmd.Flags().GetString("targets")
		
		fmt.Printf("Starting photometric analysis on: %s\n", surveyData)
		
		if err := globalClient.AnalyzePhotometric(surveyData, targetList); err != nil {
			return fmt.Errorf("photometric analysis failed: %w", err)
		}
		
		fmt.Println("Photometric analysis completed!")
		return nil
	},
}

// analyzeClusteringCmd performs clustering analysis
var analyzeClusteringCmd = &cobra.Command{
	Use:   "clustering",
	Short: "Perform clustering analysis",
	Long:  "Perform clustering analysis to identify groups of similar objects in orbital parameter space.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting clustering analysis...")
		
		if err := globalClient.AnalyzeClustering(); err != nil {
			return fmt.Errorf("clustering analysis failed: %w", err)
		}
		
		fmt.Println("Clustering analysis completed!")
		return nil
	},
}

// aiCmd represents the AI command group
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered analysis commands",
	Long:  "Commands for AI-powered object detection and model training using GPU acceleration.",
}

// aiTrainCmd trains AI models
var aiTrainCmd = &cobra.Command{
	Use:   "train [training-data] [architecture]",
	Short: "Train AI detection models",
	Long:  "Train deep learning models for astronomical object detection using GPU acceleration.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		trainingData := args[0]
		architecture := args[1]
		
		gpuDevices, _ := cmd.Flags().GetIntSlice("gpu-devices")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		epochs, _ := cmd.Flags().GetInt("epochs")
		
		fmt.Printf("Starting AI training with architecture: %s\n", architecture)
		
		if err := globalClient.TrainDeepDetector(trainingData, architecture, gpuDevices, batchSize, epochs); err != nil {
			return fmt.Errorf("AI training failed: %w", err)
		}
		
		fmt.Println("AI training completed!")
		return nil
	},
}

// aiDetectCmd performs AI detection
var aiDetectCmd = &cobra.Command{
	Use:   "detect [model-path] [survey-images]",
	Short: "Perform AI-powered object detection",
	Long:  "Use trained AI models to detect objects in astronomical survey images.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelPath := args[0]
		surveyImages := args[1]
		
		gpuAccel, _ := cmd.Flags().GetBool("gpu")
		
		fmt.Printf("Starting AI detection on: %s\n", surveyImages)
		
		if err := globalClient.AIDetection(modelPath, surveyImages, gpuAccel); err != nil {
			return fmt.Errorf("AI detection failed: %w", err)
		}
		
		fmt.Println("AI detection completed!")
		return nil
	},
}

// gpuCmd represents the GPU command group
var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "GPU management commands",
	Long:  "Commands for GPU status monitoring and benchmarking.",
}

// gpuStatusCmd shows GPU status
var gpuStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GPU status",
	Long:  "Display current GPU status including memory usage, temperature, and utilization.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return globalClient.GPUStatus()
	},
}

// gpuBenchmarkCmd runs GPU benchmark
var gpuBenchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run GPU benchmark",
	Long:  "Run a performance benchmark to test GPU computational capabilities.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return globalClient.GPUBenchmark()
	},
}

// resultsCmd retrieves analysis results
var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Retrieve analysis results",
	Long:  "Retrieve recent analysis results from the blockchain.",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		return globalClient.Results(limit)
	},
}

// queryCmd queries blockchain data
var queryCmd = &cobra.Command{
	Use:   "query [type] [id]",
	Short: "Query blockchain data",
	Long:  "Query specific data from the blockchain by type and ID.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		queryType := args[0]
		queryID := args[1]
		return globalClient.Query(queryType, queryID)
	},
}

func init() {
	cobra.OnInitialize(initViper)
	
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.medasdigital-client/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "home directory (default is $HOME/.medasdigital-client)")

	addKeysCommands()
	checkAccountCmd.Flags().String("from", "", "Key name to check")
	
	
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(checkAccountCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(aiCmd)
	rootCmd.AddCommand(gpuCmd)
	rootCmd.AddCommand(resultsCmd)
	rootCmd.AddCommand(queryCmd)
	
	// Analyze subcommands
	analyzeCmd.AddCommand(analyzeOrbitalCmd)
	analyzeCmd.AddCommand(analyzePhotometricCmd)
	analyzeCmd.AddCommand(analyzeClusteringCmd)
	
	// AI subcommands
	aiCmd.AddCommand(aiTrainCmd)
	aiCmd.AddCommand(aiDetectCmd)
	
	// GPU subcommands
	gpuCmd.AddCommand(gpuStatusCmd)
	gpuCmd.AddCommand(gpuBenchmarkCmd)
	
	// Register command flags
	registerCmd.Flags().StringSlice("capabilities", []string{}, "Client capabilities")
	registerCmd.Flags().String("metadata", "", "Additional metadata")
	registerCmd.Flags().String("from", "", "Key name to sign transaction")
	registerCmd.Flags().String("keyring-backend", "test", "Keyring backend (test|file|os)")
	registerCmd.Flags().Uint64("gas", 0, "Manual gas limit (0 = auto estimation)")  // ← NEU
	registerCmd.MarkFlagRequired("from")



	
	// Analyze orbital flags
	analyzeOrbitalCmd.Flags().String("output", "", "Output file for results")
	
	// Analyze photometric flags
	analyzePhotometricCmd.Flags().String("targets", "", "Target list file")
	
	// AI train flags
	aiTrainCmd.Flags().IntSlice("gpu-devices", []int{0}, "GPU device IDs to use")
	aiTrainCmd.Flags().Int("batch-size", 32, "Training batch size")
	aiTrainCmd.Flags().Int("epochs", 100, "Number of training epochs")
	
	// AI detect flags
	aiDetectCmd.Flags().Bool("gpu", true, "Use GPU acceleration")
	
	// Results flags
	resultsCmd.Flags().Int("limit", 10, "Maximum number of results to retrieve")
	

}
func addKeysCommands() {
	// Create keys command with proper client context
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage keyring",
		Long:  "Commands for managing the keyring and keys",
	}

	// Add key command
	addKeyCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize client context for keys
			clientCtx, err := initKeysClientContext()
			if err != nil {
				return fmt.Errorf("failed to initialize client context: %w", err)
			}

			keyName := args[0]
			
			// Check if --recover flag is set
			recover, _ := cmd.Flags().GetBool("recover")
			
			if recover {
				fmt.Print("Enter your mnemonic: ")
				var mnemonic string
				fmt.Scanln(&mnemonic)
				
				// Recover key from mnemonic
				keyInfo, err := clientCtx.Keyring.NewAccount(keyName, mnemonic, "", sdk.FullFundraiserPath, hd.Secp256k1)
				if err != nil {
					return fmt.Errorf("failed to recover key: %w", err)
				}
				
				addr, err := keyInfo.GetAddress()
				if err != nil {
					return fmt.Errorf("failed to get address: %w", err)
				}
				
				fmt.Printf("Key '%s' recovered successfully\n", keyName)
				fmt.Printf("Address: %s\n", addr.String())
			} else {
				// Generate new key
				keyInfo, mnemonic, err := clientCtx.Keyring.NewMnemonic(keyName, keyring.English, sdk.FullFundraiserPath, "", hd.Secp256k1)
				if err != nil {
					return fmt.Errorf("failed to create key: %w", err)
				}
				
				addr, err := keyInfo.GetAddress()
				if err != nil {
					return fmt.Errorf("failed to get address: %w", err)
				}
				
				fmt.Printf("Key '%s' created successfully\n", keyName)
				fmt.Printf("Address: %s\n", addr.String())
				fmt.Printf("Mnemonic: %s\n", mnemonic)
				fmt.Println("\n**Important**: Save the mnemonic phrase securely!")
			}
			
			return nil
		},
	}
	
	// Add flags to add command
	addKeyCmd.Flags().Bool("recover", false, "Recover key from mnemonic")
	
	// List keys command
	listKeysCmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize client context for keys
			clientCtx, err := initKeysClientContext()
			if err != nil {
				return fmt.Errorf("failed to initialize client context: %w", err)
			}

			keys, err := clientCtx.Keyring.List()
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}
			
			if len(keys) == 0 {
				fmt.Println("No keys found")
				return nil
			}
			
			fmt.Println("Keys:")
			for _, key := range keys {
				addr, err := key.GetAddress()
				if err != nil {
					fmt.Printf("- %s: (error getting address: %v)\n", key.Name, err)
					continue
				}
				fmt.Printf("- %s: %s\n", key.Name, addr.String())
			}
			
			return nil
		},
	}
	
	// Show key command
	showKeyCmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show key information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := initKeysClientContext()
			if err != nil {
				return fmt.Errorf("failed to initialize client context: %w", err)
			}

			keyName := args[0]
			keyInfo, err := clientCtx.Keyring.Key(keyName)
			if err != nil {
				return fmt.Errorf("key '%s' not found: %w", keyName, err)
			}
			
			addr, err := keyInfo.GetAddress()
			if err != nil {
				return fmt.Errorf("failed to get address: %w", err)
			}
			
			fmt.Printf("Name: %s\n", keyInfo.Name)
			fmt.Printf("Address: %s\n", addr.String())
			fmt.Printf("Type: %s\n", keyInfo.GetType())
			
			return nil
		},
	}
	
	// Delete key command
	deleteKeyCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := initKeysClientContext()
			if err != nil {
				return fmt.Errorf("failed to initialize client context: %w", err)
			}

			keyName := args[0]
			
			fmt.Printf("Are you sure you want to delete key '%s'? (y/N): ", keyName)
			var response string
			fmt.Scanln(&response)
			
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return nil
			}
			
			err = clientCtx.Keyring.Delete(keyName)
			if err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}
			
			fmt.Printf("Key '%s' deleted successfully\n", keyName)
			return nil
		},
	}
	
	// Add subcommands
	keysCmd.AddCommand(addKeyCmd)
	keysCmd.AddCommand(listKeysCmd)
	keysCmd.AddCommand(showKeyCmd)
	keysCmd.AddCommand(deleteKeyCmd)
	
	// Add to root command
	rootCmd.AddCommand(keysCmd)
}

func initViper() {
	if homeDir == "" {
		homeDir = filepath.Join(os.Getenv("HOME"), ".medasdigital-client")
	}
	
	if cfgFile == "" {
		cfgFile = filepath.Join(homeDir, "config.yaml")
	}
	
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func initConfig() error {
	// Create home directory if it doesn't exist
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}
	
	return nil
}

func initializeClient() error {
	// Initialize SDK config with default values
	sdkConfig := sdk.GetConfig()
	
	// Get bech32 prefix from config or use default
	bech32Prefix := viper.GetString("chain.bech32_prefix")
	if bech32Prefix == "" {
		bech32Prefix = defaultBech32Prefix
	}
	
	sdkConfig.SetBech32PrefixForAccount(bech32Prefix, bech32Prefix+"pub")
	sdkConfig.SetBech32PrefixForValidator(bech32Prefix+"valoper", bech32Prefix+"valoperpub")
	sdkConfig.SetBech32PrefixForConsensusNode(bech32Prefix+"valcons", bech32Prefix+"valconspub")
	sdkConfig.Seal()
	
	// Initialize global registry and codec ONCE
	if globalInterfaceRegistry == nil {
		globalInterfaceRegistry = getInterfaceRegistry()
	}
	if globalCodec == nil {
		globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
	}
	
	// Initialize global client
	var err error
	globalClient, err = medasClient.NewMedasDigitalClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	return nil
}

// Helper function to initialize client context for keys operations
func initKeysClientContext() (client.Context, error) {
	// Load config first
	cfg := loadConfig()
	
	// Use global codec instances to avoid conflicts
	if globalInterfaceRegistry == nil {
		globalInterfaceRegistry = getInterfaceRegistry()
	}
	if globalCodec == nil {
		globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
	}
	
	clientCtx := client.Context{}.
		WithKeyringDir(cfg.Client.KeyringDir).
		WithCodec(globalCodec).
		WithInterfaceRegistry(globalInterfaceRegistry)
	
	// Initialize keyring with proper backend
	keyringBackend := keyring.BackendTest // Use test backend as default
	if cfg.Client.KeyringBackend != "" {
		keyringBackend = cfg.Client.KeyringBackend
	}
	
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyringBackend,
		cfg.Client.KeyringDir,
		nil, // no input
		globalCodec,
	)
	if err != nil {
		return client.Context{}, fmt.Errorf("failed to create keyring: %w", err)
	}
	
	clientCtx = clientCtx.WithKeyring(kr)
	
	return clientCtx, nil
}

// Helper function to load configuration
func loadConfig() *Config {
	config := &Config{}
	
	// Set defaults if not in config
	config.Chain.ID = viper.GetString("chain.id")
	if config.Chain.ID == "" {
		config.Chain.ID = defaultChainID
	}
	
	config.Chain.RPCEndpoint = viper.GetString("chain.rpc_endpoint")
	if config.Chain.RPCEndpoint == "" {
		config.Chain.RPCEndpoint = defaultRPCEndpoint
	}
	
	config.Chain.Bech32Prefix = viper.GetString("chain.bech32_prefix")
	if config.Chain.Bech32Prefix == "" {
		config.Chain.Bech32Prefix = defaultBech32Prefix
	}
	
	config.Chain.BaseDenom = viper.GetString("chain.base_denom")
	if config.Chain.BaseDenom == "" {
		config.Chain.BaseDenom = defaultBaseDenom
	}
	
	config.Client.KeyringDir = viper.GetString("client.keyring_dir")
	if config.Client.KeyringDir == "" {
		config.Client.KeyringDir = filepath.Join(homeDir, "keyring")
	}
	
	config.Client.KeyringBackend = viper.GetString("client.keyring_backend")
	if config.Client.KeyringBackend == "" {
		config.Client.KeyringBackend = "test" // Safe default
	}
	
	return config
}

// Helper functions for codec
func getInterfaceRegistry() types.InterfaceRegistry {
	// Only create once to avoid conflicts
	if globalInterfaceRegistry != nil {
		return globalInterfaceRegistry
	}
	
	interfaceRegistry := types.NewInterfaceRegistry()
	
	// Register standard Cosmos SDK interfaces
	std.RegisterInterfaces(interfaceRegistry)
	
	// ✅ WICHTIG: Register Account interfaces für v0.50.10
	authtypes.RegisterInterfaces(interfaceRegistry)
	
	// ✅ NEU: Register bank interfaces
	banktypes.RegisterInterfaces(interfaceRegistry)
	
	// ✅ NEU: Register base account implementations
	interfaceRegistry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)
	
	// ✅ Register our blockchain messages
	blockchain.RegisterInterfaces(interfaceRegistry)
	
	return interfaceRegistry
}

func initKeysClientContextWithBackend(keyringBackend string) (client.Context, error) {
	// Load config first
	cfg := loadConfig()
	
	// Use provided backend or fall back to config
	if keyringBackend == "" {
		keyringBackend = cfg.Client.KeyringBackend
	}
	if keyringBackend == "" {
		keyringBackend = "test" // Safe default
	}
	
	// Use global codec instances to avoid conflicts
	if globalInterfaceRegistry == nil {
		globalInterfaceRegistry = getInterfaceRegistry()
	}
	if globalCodec == nil {
		globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
	}
	
	clientCtx := client.Context{}.
		WithKeyringDir(cfg.Client.KeyringDir).
		WithCodec(globalCodec).
		WithInterfaceRegistry(globalInterfaceRegistry)
	
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyringBackend,
		cfg.Client.KeyringDir,
		nil, // no input
		globalCodec,
	)
	if err != nil {
		return client.Context{}, fmt.Errorf("failed to create keyring with backend '%s': %w", keyringBackend, err)
	}
	
	clientCtx = clientCtx.WithKeyring(kr)
	
	return clientCtx, nil
}

// Vollständige createFullBlockchainClient Funktion für main.go:

// Ersetzen Sie die createFullBlockchainClient Funktion in main.go:

func createFullBlockchainClient(clientCtx client.Context, cfg *Config) (*blockchain.Client, error) {
	// Create RPC client
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	// Use the global codec instances that we already created
	if globalInterfaceRegistry == nil {
		globalInterfaceRegistry = getInterfaceRegistry()
	}
	if globalCodec == nil {
		globalCodec = codec.NewProtoCodec(globalInterfaceRegistry)
	}
	
	// Create TxConfig using v0.50 API
	txConfig := authtx.NewTxConfig(globalCodec, authtx.DefaultSignModes)
	
	// Create AccountRetriever
	accountRetriever := authtypes.AccountRetriever{}
	
	// ✅ WICHTIG: Verwenden Sie das GLEICHE Keyring wie bei check-account!
	fullClientCtx := clientCtx.
		WithClient(rpcClient).
		WithChainID(cfg.Chain.ID).
		WithCodec(globalCodec).
		WithInterfaceRegistry(globalInterfaceRegistry).
		WithTxConfig(txConfig).
		WithAccountRetriever(accountRetriever).
		WithNodeURI(cfg.Chain.RPCEndpoint).
		WithKeyring(clientCtx.Keyring).              // ✅ GLEICHER KEYRING!
		WithFromName(clientCtx.GetFromName()).       // ✅ FROM NAME
		WithFromAddress(clientCtx.GetFromAddress()). // ✅ FROM ADDRESS
		WithOffline(false).
		WithGenerateOnly(false).
		WithSimulation(false).
		WithUseLedger(false).
		WithBroadcastMode(flags.BroadcastSync)
	
	// Create blockchain client
	blockchainClient := blockchain.NewClient(fullClientCtx)
	
	return blockchainClient, nil
}

// Neue sichere Connection-Test Funktion:
func testBlockchainConnection(rpcEndpoint string) error {
	// Einfacher Connection-Test ohne vollständigen Client Context
	rpcClient, err := client.NewClientFromNode(rpcEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	// Test simple status call
	_, err = rpcClient.Status(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	
	return nil
}

// Fallback simulation function
func simulateRegistration(keyName, address string, capabilities []string, metadata string) error {
	fmt.Println("🧪 Running registration simulation...")
	fmt.Printf("✅ Client registration simulated successfully!\n")
	fmt.Printf("🆔 Client ID: client-%s\n", address[:8])
	fmt.Printf("📍 Address: %s\n", address)
	fmt.Printf("🔧 Capabilities: %v\n", capabilities)
	
	if metadata != "" {
		fmt.Printf("📋 Metadata: %s\n", metadata)
	}
	
	fmt.Println("\n💡 Note: This was a simulation. For real blockchain registration,")
	fmt.Println("   ensure the MedasDigital chain is running and accessible.")
	
	return nil
}

// Fügen Sie einen neuen Command zur main.go hinzu:

// Generate client ID from transaction hash
func generateClientID(txHash string) string {
	// Create deterministic client ID from transaction hash
	hash := sha256.Sum256([]byte(txHash))
	shortHash := hex.EncodeToString(hash[:4]) // First 8 characters
	return fmt.Sprintf("client-%s", shortHash)
}

// Save registration result locally
func saveRegistrationResult(result *RegistrationResult) error {
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

	// Update index file
	return updateRegistrationIndex(result)
}

// Update registration index
func updateRegistrationIndex(result *RegistrationResult) error {
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

// Register client using standard MsgSend with memo
func registerClientSimple(clientCtx client.Context, fromAddress string, capabilities []string, metadata string, gas uint64) (*RegistrationResult, error) {
	// Create registration data
	regData := ClientRegistrationData{
		ClientAddress: fromAddress,
		Capabilities:  capabilities,
		Metadata:      metadata,
		Timestamp:     time.Now(),
		Version:       "1.0.0",
	}

	// Convert to JSON for memo
	memoBytes, err := json.Marshal(regData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration data: %w", err)
	}

	// Create a minimal self-send transaction with registration data in memo
	fromAddr, err := sdk.AccAddressFromBech32(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	// Load config to get base denom
	cfg := loadConfig()
baseDenom := cfg.Chain.BaseDenom
fmt.Printf("🔍 Config BaseDenom: '%s'\n", baseDenom)
if baseDenom == "" {
    baseDenom = "stake" // Fallback
    fmt.Println("⚠️  Using fallback denom: stake")
} else {
    fmt.Printf("✅ Using config denom: %s\n", baseDenom)
}

	// Create MsgSend (send 1 smallest unit to self)
	amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(1)))
	msgSend := banktypes.NewMsgSend(fromAddr, fromAddr, amount)

	// Create transaction
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgSend); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set memo with registration data
	txBuilder.SetMemo(string(memoBytes))

	// Set gas and fees
	if gas > 0 {
		txBuilder.SetGasLimit(gas)
	} else {
		txBuilder.SetGasLimit(200000) // Reasonable default for MsgSend
	}

	// Calculate fee (simple fee calculation)
	gasLimit := txBuilder.GetTx().GetGas()
	feeAmount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(int64(gasLimit*1000))))
	txBuilder.SetFeeAmount(feeAmount)

	// Sign transaction
	fromName := clientCtx.GetFromName()
	if fromName == "" {
		return nil, fmt.Errorf("from name not set in client context")
	}

	// Create a proper tx factory
txFactory := tx.Factory{}.
    WithChainID(clientCtx.ChainID).
    WithKeybase(clientCtx.Keyring).
    WithTxConfig(clientCtx.TxConfig).
    WithAccountRetriever(authtypes.AccountRetriever{})

err = tx.Sign(context.Background(), txFactory, fromName, txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast
	result, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("transaction failed with code %d: %s", result.Code, result.RawLog)
	}

	// Generate client ID from transaction hash
	clientID := generateClientID(result.TxHash)

	// Create result
	regResult := &RegistrationResult{
		TransactionHash:  result.TxHash,
		ClientID:         clientID,
		RegistrationData: regData,
		BlockHeight:      result.Height,
		RegisteredAt:     time.Now(),
	}

	// Save registration locally
	if err := saveRegistrationResult(regResult); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save registration locally: %v\n", err)
		// Don't fail the entire operation for this
	}

	return regResult, nil
}



func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

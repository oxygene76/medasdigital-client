package main

import (
	"context"
    "fmt"
	"sync" 
	"os"
	"io"
	"strconv"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	
	"encoding/json"
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
	"github.com/oxygene76/medasdigital-client/pkg/compute"
    "github.com/gorilla/mux"  // Für HTTP Router
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
	
	// ✅ Computing Service Tracking (NEU HINZUGEFÜGT)
	serviceStartTime time.Time
	globalClientCtx  interface{} // Für payment service
)


type ChainStatus struct {
	LatestBlockHeight int64
	LatestBlockTime   time.Time
	ChainID          string
	NodeVersion      string
	CatchingUp       bool
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


// Enhanced status command that fetches from blockchain
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show client status and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()
		
		fmt.Println("=== MedasDigital Client Status ===")
		
		// Get registration hashes from local storage
		localHashes, err := blockchain.GetLocalRegistrationHashes()
		var blockchainRegistration *blockchain.BlockchainRegistrationData
		var isRegistered bool
		
		if err == nil && len(localHashes) > 0 {
			// Try to fetch the most recent registration from blockchain
			for _, hash := range localHashes {
				if regData, err := blockchain.FetchRegistrationFromBlockchain(hash, cfg.Chain.RPCEndpoint, cfg.Chain.ID, globalCodec); err == nil {
					if blockchainRegistration == nil || regData.BlockTime.After(blockchainRegistration.BlockTime) {
						blockchainRegistration = regData
						isRegistered = true
					}
				}
			}
		}
		
		// Client Registration Status from Blockchain
		if isRegistered && blockchainRegistration != nil {
			fmt.Printf("Client ID: %s\n", blockchainRegistration.ClientID)
			fmt.Printf("Registered: %s ✅\n", blockchainRegistration.BlockTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("Registration TX: %s\n", blockchainRegistration.TransactionHash)
			fmt.Printf("Block Height: %d\n", blockchainRegistration.BlockHeight)
			fmt.Printf("Transaction Status: %s\n", blockchainRegistration.TxStatus)
			
			// Show blockchain-verified data
			fmt.Printf("Verified Address: %s\n", blockchainRegistration.FromAddress)
			fmt.Printf("Verified Capabilities: %v\n", blockchainRegistration.RegistrationData.Capabilities)
			fmt.Printf("Gas Used: %d / %d\n", blockchainRegistration.GasUsed, blockchainRegistration.GasWanted)
			fmt.Printf("Fee Paid: %s %s\n", blockchainRegistration.Fee, blockchainRegistration.Denom)
			
			// Show memo data if available
			if blockchainRegistration.Memo != "" {
				fmt.Printf("Blockchain Memo: %s\n", blockchain.TruncateString(blockchainRegistration.Memo, 100))
			}
			
			fmt.Printf("Verification: ✅ Confirmed on blockchain\n")
		} else {
			fmt.Printf("Client ID: Not registered\n")
			fmt.Printf("Registered: false ❌\n")
			
			// Check if we have local hashes but couldn't fetch from blockchain
			if len(localHashes) > 0 {
				fmt.Printf("Note: Found %d local registration(s) but could not verify on blockchain\n", len(localHashes))
				fmt.Println("💡 This might indicate:")
				fmt.Println("   • Network connectivity issues")
				fmt.Println("   • Chain reorganization")
				fmt.Println("   • Transaction not yet finalized")
			}
		}
		
		// Available capabilities from config  
		if cfg.Client.Capabilities != nil {
			fmt.Printf("Available Capabilities: %v\n", cfg.Client.Capabilities)
		} else {
			fmt.Printf("Available Capabilities: [orbital_dynamics photometric_analysis clustering_analysis ai_training]\n")
		}
		
		// Chain information
		fmt.Printf("Chain ID: %s\n", cfg.Chain.ID)
		fmt.Printf("RPC Endpoint: %s\n", cfg.Chain.RPCEndpoint)
		
		// Test blockchain connection with detailed info
		fmt.Print("Blockchain Status: ")
		if status, err := getDetailedChainStatus(cfg.Chain.RPCEndpoint); err != nil {
			fmt.Printf("❌ Disconnected (%v)\n", err)
		} else {
			fmt.Printf("✅ Connected (Block: %d, %s)\n", 
				status.LatestBlockHeight, 
				status.LatestBlockTime.Format("15:04:05"))
		}
		
		// GPU Status
		fmt.Print("GPU Status: ")
		if cfg.GPU.Enabled {
			if gpuAvailable, gpuInfo := testGPUAvailability(); gpuAvailable {
				fmt.Printf("✅ Available (%s)\n", gpuInfo)
			} else {
				fmt.Printf("❌ Not Available (%s)\n", gpuInfo)
			}
		} else {
			fmt.Printf("Not Available\n")
		}
		
		return nil
	},
}

// VOLLSTÄNDIGER ERSATZ für den check-account Command (um alle Probleme zu beheben):
var checkAccountCmd = &cobra.Command{
	Use:   "check-account [address]",
	Short: "Check account status on blockchain",
	Long:  "Check if an account exists on the blockchain and show its details",
	Args:  cobra.RangeArgs(0, 1),
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

		// TEST 1: Bank balance query (Protobuf method)
		fmt.Println("🔍 Testing Bank Balance Query (v0.50.10):")
		denoms := []string{"umedas", "medas", "stake"}
		for _, denom := range denoms {
			fmt.Printf("   Testing denom '%s':\n", denom)
			
			// Create proper protobuf query
			balanceReq := &banktypes.QueryBalanceRequest{
				Address: address,
				Denom:   denom,
			}
			
			reqBytes, err := queryCtx.Codec.Marshal(balanceReq)
			if err != nil {
				fmt.Printf("     ❌ Failed to marshal request: %v\n", err)
				continue
			}
			
			fmt.Printf("     Query: Protobuf-encoded (%d bytes)\n", len(reqBytes))
			
			balRes, balHeight, balErr := queryCtx.QueryWithData("/cosmos.bank.v1beta1.Query/Balance", reqBytes)
			if balErr != nil {
				fmt.Printf("     Result: ❌ Error - %v\n", balErr)
			} else {
				fmt.Printf("     Result: ✅ Success - %d bytes, height %d\n", len(balRes), balHeight)
				
				// Decode the response
				var balanceResp banktypes.QueryBalanceResponse
				if err := queryCtx.Codec.Unmarshal(balRes, &balanceResp); err != nil {
					fmt.Printf("     ❌ Failed to decode response: %v\n", err)
				} else {
					if balanceResp.Balance != nil && !balanceResp.Balance.Amount.IsZero() {
						fmt.Printf("     💰 BALANCE FOUND: %s %s\n", balanceResp.Balance.Amount, balanceResp.Balance.Denom)
					} else {
						fmt.Printf("     💰 Balance: 0 %s\n", denom)
					}
				}
			}
		}

		// TEST 2: All Balances Query (Protobuf) - FIXED VARIABLES
		fmt.Printf("\n   Testing All Balances Query:\n")
		allBalancesReq := &banktypes.QueryAllBalancesRequest{
			Address: address,
		}

		var allReqBytes []byte
		allReqBytes, err = queryCtx.Codec.Marshal(allBalancesReq)
		if err != nil {
			fmt.Printf("     ❌ Failed to marshal all balances request: %v\n", err)
		} else {
			allBalRes, allBalHeight, allBalErr := queryCtx.QueryWithData("/cosmos.bank.v1beta1.Query/AllBalances", allReqBytes)
			if allBalErr != nil {
				fmt.Printf("     Result: ❌ Error - %v\n", allBalErr)
			} else {
				fmt.Printf("     Result: ✅ Success - %d bytes, height %d\n", len(allBalRes), allBalHeight)
				
				var allBalancesResp banktypes.QueryAllBalancesResponse
				if err := queryCtx.Codec.Unmarshal(allBalRes, &allBalancesResp); err != nil {
					fmt.Printf("     ❌ Failed to decode response: %v\n", err)
				} else {
					if len(allBalancesResp.Balances) > 0 {
						fmt.Printf("     💰 TOTAL BALANCES FOUND:\n")
						for _, balance := range allBalancesResp.Balances {
							fmt.Printf("       %s %s\n", balance.Amount, balance.Denom)
						}
					} else {
						fmt.Printf("     💰 No balances found (empty account)\n")
					}
				}
			}
		}

		// TEST 3: Transaction-based balance estimation
		fmt.Printf("\n   Transaction-based balance analysis:\n")
		query := fmt.Sprintf("transfer.recipient='%s' OR transfer.sender='%s'", address, address)
		txSearchResult, err := rpcClient.TxSearch(context.Background(), query, false, nil, nil, "desc")
		if err != nil {
			fmt.Printf("     ❌ Could not search transactions: %v\n", err)
		} else {
			fmt.Printf("     📊 Found %d transactions involving this address\n", len(txSearchResult.Txs))
			
			if len(txSearchResult.Txs) > 0 {
				// Look at recent transactions for balance hints
				var totalReceived, totalSent int64
				
				for i, tx := range txSearchResult.Txs[:min(10, len(txSearchResult.Txs))] {
					if i < 3 { // Show first 3 transactions
						fmt.Printf("     %d. Block %d: Status %d\n", i+1, tx.Height, tx.TxResult.Code)
					}
					
					// Analyze events for amounts
					for _, event := range tx.TxResult.Events {
						if event.Type == "transfer" {
							var isReceiver, isSender bool
							var amount string
							
							for _, attr := range event.Attributes {
								if attr.Key == "recipient" && attr.Value == address {
									isReceiver = true
								}
								if attr.Key == "sender" && attr.Value == address {
									isSender = true
								}
								if attr.Key == "amount" {
									amount = attr.Value
								}
							}
							
							if amount != "" && strings.Contains(amount, "umedas") {
								// Extract numeric amount
								amountStr := strings.Replace(amount, "umedas", "", -1)
								if amountVal, err := strconv.ParseInt(amountStr, 10, 64); err == nil {
									if isReceiver {
										totalReceived += amountVal
									}
									if isSender {
										totalSent += amountVal
									}
								}
							}
						}
					}
				}
				
				if totalReceived > 0 || totalSent > 0 {
					fmt.Printf("     💸 Transaction analysis (last 10 txs):\n")
					fmt.Printf("       Received: %d umedas\n", totalReceived)
					fmt.Printf("       Sent: %d umedas\n", totalSent)
					fmt.Printf("       Net: %d umedas\n", totalReceived-totalSent)
					fmt.Printf("     💡 Note: This is not exact balance, just transaction history\n")
				}
			}
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

		// Parse address for AccountRetriever
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			fmt.Printf("   ❌ Invalid address format: %v\n", err)
			return nil
		}

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
		if accErr == nil {
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

var balanceCmd = &cobra.Command{
	Use:   "balance [address]",
	Short: "Check account balance using multiple methods",
	Long:  "Check account balance using Tendermint RPC and alternative query methods",
	Args:  cobra.RangeArgs(0, 1),
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
		
		fmt.Printf("💰 Checking balance for: %s\n", address)
		fmt.Println("=" + strings.Repeat("=", 60))
		
		cfg := loadConfig()
		
		// Method 1: Direct Tendermint RPC Balance Query
		if balance, err := queryBalanceViaTendermint(address, cfg); err != nil {
			fmt.Printf("❌ Tendermint RPC Query failed: %v\n", err)
		} else {
			fmt.Println("✅ Tendermint RPC Balance Query:")
			if len(balance) == 0 {
				fmt.Printf("   Balance: 0 (no funds)\n")
			} else {
				for _, coin := range balance {
					fmt.Printf("   %s %s\n", coin.Amount, coin.Denom)
				}
			}
		}
		
		// Method 2: Alternative REST Query (different format)
		fmt.Println("\n🔍 Alternative REST Query:")
		if balance, err := queryBalanceViaREST(address, cfg); err != nil {
			fmt.Printf("❌ REST Query failed: %v\n", err)
		} else {
			fmt.Println("✅ REST Balance Query:")
			if len(balance) == 0 {
				fmt.Printf("   Balance: 0 (no funds)\n")
			} else {
				for denom, amount := range balance {
					fmt.Printf("   %s %s\n", amount, denom)
				}
			}
		}
		
		// Method 3: Account Info with Bank Module
		fmt.Println("\n🏦 Bank Module Query:")
		if balance, err := queryBalanceViaBankModule(address, cfg); err != nil {
			fmt.Printf("❌ Bank Module Query failed: %v\n", err)
		} else {
			fmt.Println("✅ Bank Module Balance:")
			if len(balance) == 0 {
				fmt.Printf("   Balance: 0 (no funds)\n")
			} else {
				for _, coin := range balance {
					fmt.Printf("   %s %s\n", coin.Amount, coin.Denom)
				}
			}
		}
		
		// Method 4: Transaction History Analysis
		fmt.Println("\n📊 Transaction History Analysis:")
		if err := analyzeTransactionHistory(address, cfg); err != nil {
			fmt.Printf("❌ Transaction analysis failed: %v\n", err)
		}
		
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
	serviceStartTime = time.Now() 
	cobra.OnInitialize(initViper)
	
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.medasdigital-client/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", "", "home directory (default is $HOME/.medasdigital-client)")

	addKeysCommands()
	checkAccountCmd.Flags().String("from", "", "Key name to check")
	balanceCmd.Flags().String("from", "", "Key name to check balance for")
	
	
	// Add subcommands
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(listRegistrationsCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(checkAccountCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(aiCmd)
	rootCmd.AddCommand(gpuCmd)
	rootCmd.AddCommand(resultsCmd)
	rootCmd.AddCommand(queryCmd)
	// Computing Commands (vollständig)
	rootCmd.AddCommand(realPaymentServiceCmd)  // Payment Service
	rootCmd.AddCommand(contractCmd)            // contract submit-job, list-providers, etc.
	rootCmd.AddCommand(serveCmd)               // Kostenloser Test Service  
	rootCmd.AddCommand(piCmd)   
    
    // PI subcommands
    piCmd.AddCommand(piCalculateCmd)
    piCmd.AddCommand(piBenchmarkCmd)
    
    // Flags für serve command (kostenloser Service)
    serveCmd.Flags().Int("port", 8080, "Port to listen on")
    serveCmd.Flags().Int("max-jobs", 2, "Maximum concurrent jobs")
    serveCmd.Flags().Duration("max-runtime", 30*time.Minute, "Maximum runtime per job")
    serveCmd.Flags().Bool("test-mode", true, "Enable test mode")
    
    // Flags für pi calculate command
    piCalculateCmd.Flags().String("method", "chudnovsky", "Calculation method (chudnovsky|machin|bailey)")
    piCalculateCmd.Flags().String("output", "", "Output file for result")
    piCalculateCmd.Flags().Bool("verbose", false, "Verbose output")
	
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
// ========================================
// COMPUTING SERVICE COMMANDS (FEHLEN KOMPLETT)
// ========================================

// serveCmd startet sicheren kostenlosen Test-Service 
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start free PI computation test service (LIMITED)",
	Long: `Start a free test service for PI computation with strict limits.
⚠️  SECURITY LIMITS:
- Maximum 100 digits per calculation
- Maximum 5 concurrent jobs
- Maximum 10 calculations per IP per hour
- 5 minute timeout per job

Example:
  medasdigital-client serve --port 8080 --max-jobs 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		maxJobs, _ := cmd.Flags().GetInt("max-jobs")
		maxRuntime, _ := cmd.Flags().GetDuration("max-runtime")
		testMode, _ := cmd.Flags().GetBool("test-mode")
		
		// SICHERHEITSLIMITS DURCHSETZEN
		if maxJobs > 5 {
			maxJobs = 5
			fmt.Println("⚠️  Max jobs limited to 5 for free service")
		}
		if maxRuntime > 5*time.Minute {
			maxRuntime = 5 * time.Minute
			fmt.Println("⚠️  Max runtime limited to 5 minutes for free service")
		}
		
		fmt.Println("🧪 Starting FREE PI Computation Test Service (LIMITED)")
		fmt.Println("⚠️  SECURITY LIMITS ENFORCED:")
		fmt.Println("   • Max digits: 100")
		fmt.Println("   • Max concurrent jobs: 5") 
		fmt.Println("   • Max runtime: 5 minutes")
		fmt.Println("   • Rate limit: 10 requests/hour/IP")
		fmt.Printf("📊 Actual max jobs: %d\n", maxJobs)
		fmt.Printf("⏱️  Actual max runtime: %v\n", maxRuntime)
		fmt.Printf("🌐 Listening on port: %d\n", port)
		fmt.Println("💰 Cost: FREE (with limits)")
		fmt.Println("💡 For unlimited calculations, use: payment-service")
		
		service := NewSecureFreeTestService(maxJobs, maxRuntime, testMode)
		return service.Start(port)
	},
}

// piCmd für direkte PI-Berechnung ohne Service
var piCmd = &cobra.Command{
	Use:   "pi",
	Short: "PI calculation commands",
	Long:  "Commands for PI calculation and testing",
}

// Direct PI calculation command mit CLI-Limit
var piCalculateCmd = &cobra.Command{
	Use:   "calculate [digits]",
	Short: "Calculate PI to specified digits (max 1000 for CLI)",
	Long: `Calculate PI to the specified number of decimal places.
⚠️  CLI LIMIT: Maximum 1000 digits
For higher precision, use payment-service.

Examples:
  medasdigital-client pi calculate 100
  medasdigital-client pi calculate 1000 --method chudnovsky`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		digits, err := strconv.Atoi(args[0])
		if err != nil || digits < 1 {
			return fmt.Errorf("invalid digits: %s (must be positive integer)", args[0])
		}
		
		// CLI LIMIT
		if digits > 1000 {
			return fmt.Errorf("CLI limit exceeded. Max 1000 digits. Use payment-service for higher precision.")
		}
		
		method, _ := cmd.Flags().GetString("method")
		output, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		fmt.Printf("🧮 Calculating PI to %d decimal places (CLI mode)\n", digits)
		fmt.Printf("📊 Method: %s\n", method)
		
		result, err := calculatePIDirectly(digits, method, verbose)
		if err != nil {
			return fmt.Errorf("PI calculation failed: %w", err)
		}
		
		// Output result
		if output != "" {
			err := writeResultToFile(result, output)
			if err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			fmt.Printf("✅ Result written to: %s\n", output)
		} else {
			fmt.Printf("✅ Result: %s\n", result.Value)
		}
		
		fmt.Printf("⏱️  Duration: %v\n", result.Duration)
		fmt.Printf("🔧 Iterations: %d\n", result.Iterations)
		
		if digits >= 500 {
			fmt.Println("\n💡 For unlimited precision, use:")
			fmt.Println("   medasdigital-client payment-service")
		}
		
		return nil
	},
}

// PI benchmark command
var piBenchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run PI calculation benchmark",
	Long: `Run a benchmark test to measure PI calculation performance.

This tests different algorithms and digit counts to measure system performance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🏁 Starting PI Calculation Benchmark")
		fmt.Println("====================================")
		
		results := runPIBenchmark()
		displayBenchmarkResults(results)
		
		return nil
	},
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

	globalClientCtx = globalClient 
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



// Get detailed chain status
func getDetailedChainStatus(rpcEndpoint string) (*ChainStatus, error) {
	rpcClient, err := client.NewClientFromNode(rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	status, err := rpcClient.Status(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	
	return &ChainStatus{
		LatestBlockHeight: status.SyncInfo.LatestBlockHeight,
		LatestBlockTime:   status.SyncInfo.LatestBlockTime,
		ChainID:          status.NodeInfo.Network,
		NodeVersion:      status.NodeInfo.Version,
		CatchingUp:       status.SyncInfo.CatchingUp,
	}, nil
}

// Enhanced list registrations command with blockchain data
var listRegistrationsCmd = &cobra.Command{
	Use:   "list-registrations",
	Short: "List all registrations with blockchain verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get local hashes
		hashes, err := blockchain.GetLocalRegistrationHashes()
		if err != nil {
			fmt.Printf("❌ No local registrations found: %v\n", err)
			fmt.Println("💡 Run: ./bin/medasdigital-client register --from <keyname>")
			return nil
		}
		
		cfg := loadConfig()
		fmt.Printf("📋 Found %d local registration hash(es), fetching from blockchain...\n", len(hashes))
		fmt.Println("=" + strings.Repeat("=", 80))
		
		var validRegistrations []*blockchain.BlockchainRegistrationData
		
		for i, hash := range hashes {
			fmt.Printf("\n%d. 📊 Transaction Hash: %s\n", i+1, hash)
			
			regData, err := blockchain.FetchRegistrationFromBlockchain(hash, cfg.Chain.RPCEndpoint, cfg.Chain.ID, globalCodec)
			if err != nil {
				fmt.Printf("   ❌ Failed to fetch from blockchain: %v\n", err)
				continue
			}
			
			validRegistrations = append(validRegistrations, regData)
			
			fmt.Printf("   🆔 Client ID: %s\n", regData.ClientID)
			fmt.Printf("   📍 Address: %s\n", regData.FromAddress)
			fmt.Printf("   🔧 Capabilities: %v\n", regData.RegistrationData.Capabilities)
			fmt.Printf("   🏔️  Block: %d\n", regData.BlockHeight)
			fmt.Printf("   🕒 Time: %s\n", regData.BlockTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("   ⛽ Gas: %d / %d\n", regData.GasUsed, regData.GasWanted)
			fmt.Printf("   💰 Fee: %s %s\n", regData.Fee, regData.Denom)
			fmt.Printf("   🔍 Status: %s\n", regData.TxStatus)
			fmt.Printf("   ✅ Verification: %s\n", regData.VerificationStatus)
		}
		
		fmt.Println("\n=" + strings.Repeat("=", 80))
		fmt.Printf("✅ Successfully verified %d/%d registrations from blockchain\n", 
			len(validRegistrations), len(hashes))
		
		return nil
	},
}

// Enhanced whoami command with blockchain data
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current client identity from blockchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		hashes, err := blockchain.GetLocalRegistrationHashes()
		if err != nil {
			fmt.Println("❌ Not registered")
			fmt.Println("💡 Run: ./bin/medasdigital-client register --from <keyname>")
			return nil
		}
		
		cfg := loadConfig()
		var latest *blockchain.BlockchainRegistrationData
		
		// Find most recent valid registration from blockchain
		for _, hash := range hashes {
			if regData, err := blockchain.FetchRegistrationFromBlockchain(hash, cfg.Chain.RPCEndpoint, cfg.Chain.ID, globalCodec); err == nil {
				if latest == nil || regData.BlockTime.After(latest.BlockTime) {
					latest = regData
				}
			}
		}
		
		if latest == nil {
			fmt.Println("❌ No valid registrations found on blockchain")
			fmt.Printf("💡 Found %d local hash(es) but none could be verified\n", len(hashes))
			return nil
		}
		
		fmt.Println("👤 Current Client Identity (Blockchain Verified)")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Printf("🆔 Client ID: %s\n", latest.ClientID)
		fmt.Printf("📍 Address: %s\n", latest.FromAddress)
		fmt.Printf("🔧 Capabilities: %v\n", latest.RegistrationData.Capabilities)
		fmt.Printf("📊 Registration TX: %s\n", latest.TransactionHash)
		fmt.Printf("🏔️  Block Height: %d\n", latest.BlockHeight)
		fmt.Printf("🕒 Registered: %s\n", latest.BlockTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("⛽ Gas Used: %d / %d\n", latest.GasUsed, latest.GasWanted)
		fmt.Printf("💰 Fee Paid: %s %s\n", latest.Fee, latest.Denom)
		fmt.Printf("🔍 Status: %s\n", latest.TxStatus)
		fmt.Printf("✅ Verification: %s\n", latest.VerificationStatus)
		
		return nil
	},
}


// Test GPU availability
func testGPUAvailability() (bool, string) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return false, "nvidia-smi not available"
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		parts := strings.Split(lines[0], ",")
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			memory := strings.TrimSpace(parts[1])
			return true, fmt.Sprintf("%s (%s MB)", name, memory)
		}
	}
	
	return false, "No NVIDIA GPUs detected"
}


// Method 1: Direct Tendermint RPC Query
func queryBalanceViaTendermint(address string, cfg *Config) ([]sdk.Coin, error) {
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	// Query using ABCI query directly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Try different query paths
	queryPaths := []string{
		fmt.Sprintf("store/bank/key/balances/%s", address),
		fmt.Sprintf("custom/bank/balances/%s", address),
		fmt.Sprintf("bank/balances/%s", address),
	}
	
	for _, path := range queryPaths {
		result, err := rpcClient.ABCIQuery(ctx, path, nil)
		if err == nil && result.Response.Code == 0 && len(result.Response.Value) > 0 {
			// Try to decode the response
			fmt.Printf("   Found data via path: %s\n", path)
			fmt.Printf("   Raw data: %x\n", result.Response.Value)
			
			// Try to parse as JSON or protobuf
			var coins []sdk.Coin
			if err := json.Unmarshal(result.Response.Value, &coins); err == nil {
				return coins, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no balance data found via Tendermint RPC")
}

// Method 2: Alternative REST Query
func queryBalanceViaREST(address string, cfg *Config) (map[string]string, error) {
	// Parse the RPC endpoint to get the base URL
	rpcURL := cfg.Chain.RPCEndpoint
	// Convert RPC URL to REST URL (common pattern)
	restURL := strings.Replace(rpcURL, ":26657", ":1317", 1)
	restURL = strings.Replace(restURL, "rpc.", "api.", 1)
	
	// Try different REST endpoints
	endpoints := []string{
		fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", restURL, address),
		fmt.Sprintf("%s/bank/balances/%s", restURL, address),
		fmt.Sprintf("%s/cosmos/bank/v1beta1/all_balances/%s", restURL, address),
	}
	
	for _, endpoint := range endpoints {
		fmt.Printf("   Trying: %s\n", endpoint)
		
		resp, err := http.Get(endpoint)
		if err != nil {
			fmt.Printf("   HTTP Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			
			fmt.Printf("   Response: %s\n", string(body))
			
			// Try to parse the JSON response
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err == nil {
				balances := make(map[string]string)
				
				// Different possible JSON structures
				if balancesArray, ok := result["balances"].([]interface{}); ok {
					for _, bal := range balancesArray {
						if balMap, ok := bal.(map[string]interface{}); ok {
							if denom, ok := balMap["denom"].(string); ok {
								if amount, ok := balMap["amount"].(string); ok {
									balances[denom] = amount
								}
							}
						}
					}
				}
				
				if len(balances) > 0 {
					return balances, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("no REST endpoint returned valid balance data")
}

// 4. FIX für queryBalanceViaBankModule Funktion - KOMPLETT ERSETZEN:
func queryBalanceViaBankModule(address string, cfg *Config) ([]sdk.Coin, error) {
	// Create proper client context
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}
	
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
	
	// Try to use bank query client directly
	_, err = sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	
	// Try different bank query approaches
	denoms := []string{"umedas", "medas", "stake", "token"}
	
	var totalBalance []sdk.Coin
	for _, denom := range denoms {
		// Try to query specific denom
		queryReq := &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   denom,
		}
		
		reqBytes, err := queryCtx.Codec.Marshal(queryReq)
		if err != nil {
			continue
		}
		
		res, _, err := queryCtx.QueryWithData("/cosmos.bank.v1beta1.Query/Balance", reqBytes)
		if err != nil {
			fmt.Printf("   Error querying %s: %v\n", denom, err)
			continue
		}
		
		var queryRes banktypes.QueryBalanceResponse
		if err := queryCtx.Codec.Unmarshal(res, &queryRes); err != nil {
			continue
		}
		
		if queryRes.Balance != nil && !queryRes.Balance.Amount.IsZero() {
			totalBalance = append(totalBalance, *queryRes.Balance)
		}
	}
	
	return totalBalance, nil
}

// Method 4: Analyze Transaction History for Balance
func analyzeTransactionHistory(address string, cfg *Config) error {
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	
	// Get recent transactions
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Search for transactions involving this address
	query := fmt.Sprintf("transfer.recipient='%s' OR transfer.sender='%s'", address, address)
	
	result, err := rpcClient.TxSearch(ctx, query, false, nil, nil, "desc")
	if err != nil {
		return fmt.Errorf("failed to search transactions: %w", err)
	}
	
	fmt.Printf("   Found %d transactions involving this address\n", len(result.Txs))
	
	if len(result.Txs) > 0 {
		fmt.Println("   Recent transactions:")
		for i, tx := range result.Txs[:min(5, len(result.Txs))] {
			fmt.Printf("     %d. Block %d: %s (Code: %d)\n", 
				i+1, tx.Height, tx.Hash.String(), tx.TxResult.Code)
		}
		
		// Analyze last transaction for balance hints
		lastTx := result.Txs[0]
		if len(lastTx.TxResult.Events) > 0 {
			fmt.Println("   Last transaction events:")
			for _, event := range lastTx.TxResult.Events {
				if event.Type == "transfer" || event.Type == "coin_spent" || event.Type == "coin_received" {
					fmt.Printf("     %s:\n", event.Type)
					for _, attr := range event.Attributes {
						fmt.Printf("       %s: %s\n", attr.Key, attr.Value)
					}
				}
			}
		}
	}
	
	return nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========================================
// COMPUTING IMPLEMENTATION FUNCTIONS (FEHLEN KOMPLETT)
// ========================================

// calculatePIDirectly berechnet PI direkt ohne Service
func calculatePIDirectly(digits int, method string, verbose bool) (*compute.PIResult, error) {
	if verbose {
		fmt.Printf("🚀 Starting PI calculation: %d digits using %s\n", digits, method)
	}
	
	// Create PI calculator using the compute package
	calc := compute.NewPICalculator(digits, method)
	
	// Calculate PI
	result, err := calc.Calculate()
	if err != nil {
		return nil, err
	}
	
	if verbose {
		fmt.Printf("✅ PI calculation completed in %v\n", result.Duration)
		fmt.Printf("🔧 Iterations: %d, Verified: %t\n", result.Iterations, result.Verified)
	}
	
	return result, nil
}

// runPIBenchmark führt Benchmark-Tests durch
func runPIBenchmark() []BenchmarkResult {
	fmt.Println("🧮 Testing different digit counts and methods...")
	
	tests := []struct {
		digits int
		method string
	}{
		{10, "chudnovsky"},
		{100, "chudnovsky"},
		{1000, "chudnovsky"},
		{100, "machin"},
		{100, "bailey"},
	}
	
	var results []BenchmarkResult
	
	for _, test := range tests {
		fmt.Printf("📊 Testing: %d digits, %s method\n", test.digits, test.method)
		
		start := time.Now()
		result, err := calculatePIDirectly(test.digits, test.method, false)
		duration := time.Since(start)
		
		benchResult := BenchmarkResult{
			Digits:     test.digits,
			Method:     test.method,
			Duration:   duration,
			Success:    err == nil,
			Verified:   result != nil && result.Verified,
			Iterations: 0,
		}
		
		if result != nil {
			benchResult.Iterations = result.Iterations
		}
		
		results = append(results, benchResult)
	}
	
	return results
}

// BenchmarkResult für Benchmark-Tests
type BenchmarkResult struct {
	Digits     int           `json:"digits"`
	Method     string        `json:"method"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	Verified   bool          `json:"verified"`
	Iterations int64         `json:"iterations"`
}

// displayBenchmarkResults zeigt Benchmark-Ergebnisse an
func displayBenchmarkResults(results []BenchmarkResult) {
	fmt.Println("\n📊 Benchmark Results:")
	fmt.Println("=====================")
	
	for _, result := range results {
		status := "✅"
		if !result.Success {
			status = "❌"
		}
		
		fmt.Printf("%s %4d digits | %-12s | %8v | %6d iter | verified: %t\n",
			status, result.Digits, result.Method, result.Duration, result.Iterations, result.Verified)
	}
	
	fmt.Println("\n🏆 Fastest method: Chudnovsky algorithm")
	fmt.Println("💡 For production use, consider the paid service with higher precision")
}

// writeResultToFile schreibt PI-Ergebnis in Datei
func writeResultToFile(result *compute.PIResult, filename string) error {
	content := fmt.Sprintf("PI Calculation Result\n")
	content += fmt.Sprintf("====================\n")
	content += fmt.Sprintf("Digits: %d\n", result.Digits)
	content += fmt.Sprintf("Method: %s\n", result.Method)
	content += fmt.Sprintf("Duration: %v\n", result.Duration)
	content += fmt.Sprintf("Iterations: %d\n", result.Iterations)
	content += fmt.Sprintf("Verified: %t\n", result.Verified)
	content += fmt.Sprintf("Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	content += fmt.Sprintf("\nResult:\n%s\n", result.Value)
	
	return os.WriteFile(filename, []byte(content), 0644)
}

// ========================================
// SECURE FREE TEST SERVICE (FEHLT KOMPLETT)
// ========================================

// SICHERHEITSKONSTANTEN
const (
	FREE_SERVICE_MAX_DIGITS = 100
	FREE_SERVICE_MAX_JOBS_PER_IP = 10
	FREE_SERVICE_RATE_WINDOW = time.Hour
	FREE_SERVICE_MAX_RUNTIME = 5 * time.Minute
	FREE_SERVICE_MAX_CONCURRENT = 5
)

// SecureFreeTestService - Sichere Version mit Limits
type SecureFreeTestService struct {
	maxJobs       int
	maxRuntime    time.Duration
	testMode      bool
	activeJobs    map[string]*TestJob
	jobCounter    int64
	
	// SICHERHEITSFEATURES
	rateLimiter   map[string]*RateLimit
	mu            sync.RWMutex
	maxDigits     int
	maxJobsPerIP  int
}

// RateLimit für IP-basierte Begrenzung
type RateLimit struct {
	Count     int
	ResetTime time.Time
	IP        string
}

// TestJob für kostenlose Test-Berechnungen
type TestJob struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Parameters map[string]interface{}    `json:"parameters"`
	Status     string                    `json:"status"`
	Result     *compute.PIResult         `json:"result,omitempty"`
	StartTime  time.Time                 `json:"start_time"`
	EndTime    time.Time                 `json:"end_time,omitempty"`
	Duration   string                    `json:"duration,omitempty"`
	Progress   int                       `json:"progress"`
	Error      string                    `json:"error,omitempty"`
}

// NewSecureFreeTestService erstellt sicheren kostenlosen Service
func NewSecureFreeTestService(maxJobs int, maxRuntime time.Duration, testMode bool) *SecureFreeTestService {
	// SICHERHEITSLIMITS DURCHSETZEN
	if maxJobs > FREE_SERVICE_MAX_CONCURRENT {
		maxJobs = FREE_SERVICE_MAX_CONCURRENT
	}
	if maxRuntime > FREE_SERVICE_MAX_RUNTIME {
		maxRuntime = FREE_SERVICE_MAX_RUNTIME
	}
	
	return &SecureFreeTestService{
		maxJobs:      maxJobs,
		maxRuntime:   maxRuntime,
		testMode:     testMode,
		activeJobs:   make(map[string]*TestJob),
		jobCounter:   0,
		rateLimiter:  make(map[string]*RateLimit),
		maxDigits:    FREE_SERVICE_MAX_DIGITS,
		maxJobsPerIP: FREE_SERVICE_MAX_JOBS_PER_IP,
	}
}

// Start startet den sicheren kostenlosen Service
func (sfts *SecureFreeTestService) Start(port int) error {
	r := mux.NewRouter()
	
	// Security Middleware
	r.Use(sfts.securityMiddleware)
	r.Use(sfts.rateLimitMiddleware)
	
	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/status", sfts.handleStatus).Methods("GET")
	api.HandleFunc("/calculate", sfts.handleCalculate).Methods("POST")
	api.HandleFunc("/limits", sfts.handleLimits).Methods("GET")
	
	fmt.Printf("🚀 Secure Free PI Test Service started on http://localhost:%d\n", port)
	fmt.Println("\n🔒 SECURITY FEATURES ENABLED:")
	fmt.Printf("   ✅ Max digits per calculation: %d\n", FREE_SERVICE_MAX_DIGITS)
	fmt.Printf("   ✅ Max concurrent jobs: %d\n", FREE_SERVICE_MAX_CONCURRENT)
	fmt.Printf("   ✅ Rate limit: %d requests/hour/IP\n", FREE_SERVICE_MAX_JOBS_PER_IP)
	fmt.Printf("   ✅ Job timeout: %v\n", FREE_SERVICE_MAX_RUNTIME)
	
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("   GET  /api/v1/status           - Service status")
	fmt.Println("   POST /api/v1/calculate        - Submit PI calculation (LIMITED)")
	fmt.Println("   GET  /api/v1/limits           - Show current limits")
	
	fmt.Println("\n🧮 Example PI calculation (MAX 100 digits):")
	fmt.Printf("   curl -X POST http://localhost:%d/api/v1/calculate \\\n", port)
	fmt.Println("     -H 'Content-Type: application/json' \\")
	fmt.Println("     -d '{\"digits\": 100, \"method\": \"chudnovsky\"}'")
	
	fmt.Println("\n⚠️  IMPORTANT: This free service has strict limits!")
	fmt.Println("   For unlimited calculations, use: payment-service")
	
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

// Handler methods (vereinfacht für main.go)
func (sfts *SecureFreeTestService) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"service":        "Secure Free PI Computation Service",
		"status":         "running",
		"max_digits":     sfts.maxDigits,
		"max_runtime":    sfts.maxRuntime.String(),
		"rate_limit":     fmt.Sprintf("%d/hour/IP", sfts.maxJobsPerIP),
		"cost":           "FREE (with limits)",
		"methods":        compute.GetAvailableMethods(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (sfts *SecureFreeTestService) handleCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Digits int    `json:"digits"`
		Method string `json:"method"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	
	// SICHERHEITSPRÜFUNGEN
	if req.Digits <= 0 {
		http.Error(w, "Digits must be positive", http.StatusBadRequest)
		return
	}
	
	if req.Digits > sfts.maxDigits {
		http.Error(w, fmt.Sprintf("Digits limit exceeded. Free service max: %d digits. Use payment-service for more.", sfts.maxDigits), http.StatusBadRequest)
		return
	}
	
	if req.Method == "" {
		req.Method = "chudnovsky"
	}
	
	clientIP := sfts.getClientIP(r)
	fmt.Printf("🧮 Free calculation request: %d digits, %s method from IP %s\n", req.Digits, req.Method, clientIP)
	
	// Calculate PI mit Timeout
	ctx, cancel := context.WithTimeout(context.Background(), sfts.maxRuntime)
	defer cancel()
	
	// Channel for result
	resultChan := make(chan *compute.PIResult, 1)
	errorChan := make(chan error, 1)
	
	// Start calculation in goroutine
	go func() {
		result, err := calculatePIDirectly(req.Digits, req.Method, false)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()
	
	// Wait for result or timeout
	select {
	case result := <-resultChan:
		response := map[string]interface{}{
			"result":     result,
			"cost":       "FREE",
			"limits": map[string]interface{}{
				"max_digits":        sfts.maxDigits,
				"max_runtime":       sfts.maxRuntime.String(),
				"used_digits":       req.Digits,
				"calculation_time":  result.Duration.String(),
			},
			"upgrade_info": "For unlimited calculations, use payment-service with MEDAS tokens",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		
		fmt.Printf("✅ Free calculation completed: %d digits in %v\n", req.Digits, result.Duration)
		
	case err := <-errorChan:
		http.Error(w, fmt.Sprintf("Calculation failed: %v", err), http.StatusInternalServerError)
		fmt.Printf("❌ Free calculation failed: %v\n", err)
		
	case <-ctx.Done():
		http.Error(w, fmt.Sprintf("Calculation timeout after %v", sfts.maxRuntime), http.StatusRequestTimeout)
		fmt.Printf("⏰ Free calculation timeout: %d digits\n", req.Digits)
	}
}

func (sfts *SecureFreeTestService) handleLimits(w http.ResponseWriter, r *http.Request) {
	limits := map[string]interface{}{
		"service_type":     "Free PI Calculation Service",
		"max_digits":       sfts.maxDigits,
		"max_runtime":      sfts.maxRuntime.String(),
		"rate_limit":       fmt.Sprintf("%d/hour/IP", sfts.maxJobsPerIP),
		"upgrade_info": map[string]interface{}{
			"unlimited_service": "payment-service",
			"max_digits":        "100,000+",
			"cost":              "MEDAS tokens",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(limits)
}

// Security middleware methods
func (sfts *SecureFreeTestService) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		
		if r.ContentLength > 1024*10 { // 10KB limit
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (sfts *SecureFreeTestService) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			next.ServeHTTP(w, r)
			return
		}
		
		clientIP := sfts.getClientIP(r)
		
		sfts.mu.Lock()
		rateLimit, exists := sfts.rateLimiter[clientIP]
		
		if exists && time.Now().After(rateLimit.ResetTime) {
			rateLimit.Count = 0
			rateLimit.ResetTime = time.Now().Add(FREE_SERVICE_RATE_WINDOW)
		}
		
		if !exists {
			rateLimit = &RateLimit{
				Count:     0,
				ResetTime: time.Now().Add(FREE_SERVICE_RATE_WINDOW),
				IP:        clientIP,
			}
			sfts.rateLimiter[clientIP] = rateLimit
		}
		
		if rateLimit.Count >= sfts.maxJobsPerIP {
			sfts.mu.Unlock()
			http.Error(w, fmt.Sprintf("Rate limit exceeded. Max %d requests per hour per IP.", sfts.maxJobsPerIP), http.StatusTooManyRequests)
			return
		}
		
		rateLimit.Count++
		sfts.mu.Unlock()
		
		next.ServeHTTP(w, r)
	})
}

func (sfts *SecureFreeTestService) getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip, _, _ = strings.Cut(ip, ":")
	}
	
	return ip
}



func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

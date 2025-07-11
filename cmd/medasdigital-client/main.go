package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"

	medasClient "github.com/oxygene76/medasdigital-client/pkg/client"
	"github.com/oxygene76/medasdigital-client/pkg/utils"
)

const (
	// Application constants
	appName = "medasdigital-client"
	version = "v1.0.0"
	
	// Default chain configuration
	defaultChainID     = "medasdigital-2"
	defaultRPCEndpoint = "https://rpc.medas-digital.io:26657"
	defaultBech32Prefix = "medas"
)

var (
	// Global client instance
	globalClient *medasClient.MedasDigitalClient
	
	// Configuration
	cfgFile string
	homeDir string
)

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
		config := &utils.Config{
			Chain: utils.ChainConfig{
				ID:           defaultChainID,
				RPCEndpoint:  defaultRPCEndpoint,
				Bech32Prefix: defaultBech32Prefix,
			},
			Client: utils.ClientConfig{
				KeyringDir:   filepath.Join(homeDir, "keyring"),
				Capabilities: []string{"orbital_dynamics", "photometric_analysis"},
			},
			GPU: utils.GPUConfig{
				Enabled:     false,
				DeviceID:    0,
				MemoryLimit: 8192, // 8GB default
			},
		}
		
		// Save configuration
		if err := utils.SaveConfig(cfgFile, config); err != nil {
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

// registerCmd registers the client on the blockchain
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register client on the blockchain",
	Long: `Register this client on the MedasDigital blockchain. This assigns 
a unique client ID and registers the client's capabilities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		capabilities, _ := cmd.Flags().GetStringSlice("capabilities")
		metadata, _ := cmd.Flags().GetString("metadata")
		from, _ := cmd.Flags().GetString("from")
		
		if from == "" {
			return fmt.Errorf("--from flag is required")
		}
		
		if len(capabilities) == 0 {
			// Use default capabilities from config
			config, err := utils.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			capabilities = config.Client.Capabilities
		}
		
		fmt.Printf("Registering client with capabilities: %v\n", capabilities)
		
		if err := globalClient.Register(capabilities, metadata, from); err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
		
		fmt.Println("Client registered successfully!")
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
	
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(statusCmd)
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
	
	// Add standard cosmos flags
	flags.AddTxFlagsToCmd(registerCmd)
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
	// Load configuration
	config, err := utils.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Initialize SDK config
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(config.Chain.Bech32Prefix, config.Chain.Bech32Prefix+"pub")
	sdkConfig.SetBech32PrefixForValidator(config.Chain.Bech32Prefix+"valoper", config.Chain.Bech32Prefix+"valoperpub")
	sdkConfig.SetBech32PrefixForConsensusNode(config.Chain.Bech32Prefix+"valcons", config.Chain.Bech32Prefix+"valconspub")
	sdkConfig.Seal()
	
	// Initialize global client
	globalClient, err = medasClient.NewMedasDigitalClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

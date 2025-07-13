// cmd/medasdigital-client/register.go - Enhanced Registration Commands
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	
	blockchain "github.com/oxygene76/medasdigital-client/pkg/blockchain"
)

// registerCmd represents the register command with enhanced features
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register client on the MedasDigital blockchain",
	Long: `Register this client on the MedasDigital blockchain with enhanced capabilities
for scientific collaboration and chat functionality.

Basic registration:
  medasdigital-client register --from mykey

Enhanced chat registration:
  medasdigital-client register --from mykey --chat \
    --display-name "Dr. Jane Smith" \
    --institution "Stanford University" \
    --expertise "photometry,orbital_dynamics"`,
}

// Simple registration command (legacy compatibility)
var registerSimpleCmd = &cobra.Command{
	Use:   "simple",
	Short: "Simple client registration (legacy)",
	Long:  "Perform basic client registration compatible with older versions.",
	RunE:  runSimpleRegistration,
}

// Enhanced chat registration command
var registerChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Enhanced registration with chat capabilities",
	Long:  "Register client with enhanced chat and collaboration features.",
	RunE:  runChatRegistration,
}

func init() {
	// Add subcommands to register
	registerCmd.AddCommand(registerSimpleCmd)
	registerCmd.AddCommand(registerChatCmd)
	
	// Global register flags
	registerCmd.PersistentFlags().String("from", "", "Key name to sign transaction (required)")
	registerCmd.PersistentFlags().String("keyring-backend", "test", "Keyring backend (test|file|os)")
	registerCmd.PersistentFlags().Uint64("gas", 0, "Manual gas limit (0 = auto estimation)")
	registerCmd.PersistentFlags().StringSlice("capabilities", []string{}, "Client capabilities")
	registerCmd.PersistentFlags().String("metadata", "", "Additional metadata (legacy)")
	
	// Mark required flags
	registerCmd.MarkPersistentFlagRequired("from")
	
	// Simple registration flags (minimal)
	// Uses only the global flags
	
	// Chat registration flags (enhanced)
	registerChatCmd.Flags().String("display-name", "", "Your display name (required for chat)")
	registerChatCmd.Flags().String("institution", "", "Your institution or organization")
	registerChatCmd.Flags().String("country", "", "Your country")
	registerChatCmd.Flags().StringSlice("expertise", []string{}, "Your research expertise areas")
	registerChatCmd.Flags().String("contact", "", "Contact information (optional)")
	registerChatCmd.Flags().String("type", "researcher", "Registration type (researcher|institution|student|developer)")
	
	// Mark required flags for chat registration
	registerChatCmd.MarkFlagRequired("display-name")
	
	// Set default run function for register command (runs chat by default)
	registerCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Check if --chat flag is provided or any chat-specific flags
		displayName, _ := cmd.Flags().GetString("display-name")
		if displayName != "" {
			return runChatRegistration(cmd, args)
		}
		
		// Default to simple registration
		return runSimpleRegistration(cmd, args)
	}
}

// runSimpleRegistration performs basic client registration
func runSimpleRegistration(cmd *cobra.Command, args []string) error {
	fmt.Println("📝 Starting Simple Client Registration")
	fmt.Println("═══════════════════════════════════════")
	
	// Get flags
	from, _ := cmd.Flags().GetString("from")
	keyringBackend, _ := cmd.Flags().GetString("keyring-backend")
	gas, _ := cmd.Flags().GetUint64("gas")
	capabilities, _ := cmd.Flags().GetStringSlice("capabilities")
	metadata, _ := cmd.Flags().GetString("metadata")
	
	// Validate required flags
	if from == "" {
		return fmt.Errorf("--from flag is required")
	}
	
	// Use default capabilities if none provided
	if len(capabilities) == 0 {
		capabilities = []string{"orbital_dynamics", "photometric_analysis"}
		fmt.Printf("🔧 Using default capabilities: %v\n", capabilities)
	}
	
	fmt.Printf("👤 Key: %s\n", from)
	fmt.Printf("🔧 Capabilities: %v\n", capabilities)
	if metadata != "" {
		fmt.Printf("📋 Metadata: %s\n", metadata)
	}
	
	// Initialize client context
	clientCtx, err := initKeysClientContextWithBackend(keyringBackend)
	if err != nil {
		return fmt.Errorf("failed to initialize client context: %w", err)
	}
	
	// Get key info and address
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
	
	fmt.Printf("📍 Address: %s\n", addr.String())
	
	// Test blockchain connection
	cfg := loadConfig()
	fmt.Printf("🔍 Testing connection to %s...\n", cfg.Chain.RPCEndpoint)
	
	if err := testBlockchainConnection(cfg.Chain.RPCEndpoint); err != nil {
		fmt.Printf("⚠️  Blockchain connection failed: %v\n", err)
		fmt.Println("💡 Running in simulation mode...")
		return simulateRegistration(from, addr.String(), capabilities, metadata)
	}
	
	fmt.Println("✅ Blockchain connection successful!")
	
	// Setup full client context for transaction
	cfg := loadConfig()
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
    		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	txConfig := authtx.NewTxConfig(globalCodec, authtx.DefaultSignModes)
	fullClientCtx := clientCtx.
    	WithFromName(fromName).
    	WithFromAddress(fromAddr).
    	WithTxConfig(txConfig).
    	WithClient(rpcClient).
    	WithChainID(cfg.Chain.ID).
    	WithCodec(globalCodec).
    	WithInterfaceRegistry(globalInterfaceRegistry).
    	WithBroadcastMode(flags.BroadcastSync)
	
	if err != nil {
		return fmt.Errorf("failed to setup client context: %w", err)
	}
	
	// Perform simple registration using new package
	result, err := blockchain.RegisterClientSimple(clientCtx, addr.String(), capabilities, metadata, gas)
	if err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		fmt.Println("💡 Falling back to simulation...")
		return simulateRegistration(from, addr.String(), capabilities, metadata)
	}
	
	// Display success
	return displayRegistrationSuccess(result, cfg.Chain.ID)
}

// runChatRegistration performs enhanced chat registration
func runChatRegistration(cmd *cobra.Command, args []string) error {
	fmt.Println("💬 Starting Enhanced Chat Registration")
	fmt.Println("═════════════════════════════════════")
	
	// Get flags
	from, _ := cmd.Flags().GetString("from")
	keyringBackend, _ := cmd.Flags().GetString("keyring-backend")
	gas, _ := cmd.Flags().GetUint64("gas")
	capabilities, _ := cmd.Flags().GetStringSlice("capabilities")
	
	// Chat-specific flags
	displayName, _ := cmd.Flags().GetString("display-name")
	institution, _ := cmd.Flags().GetString("institution")
	country, _ := cmd.Flags().GetString("country")
	expertise, _ := cmd.Flags().GetStringSlice("expertise")
	contact, _ := cmd.Flags().GetString("contact")
	regType, _ := cmd.Flags().GetString("type")
	
	// Validate required flags
	if from == "" {
		return fmt.Errorf("--from flag is required")
	}
	if displayName == "" {
		return fmt.Errorf("--display-name flag is required for chat registration")
	}
	
	// Use default capabilities if none provided
	if len(capabilities) == 0 {
		capabilities = []string{"orbital_dynamics", "photometric_analysis", "chat", "collaboration"}
		fmt.Printf("🔧 Using default chat capabilities: %v\n", capabilities)
	}
	
	fmt.Printf("👤 Key: %s\n", from)
	fmt.Printf("📛 Display Name: %s\n", displayName)
	fmt.Printf("🏛️  Institution: %s\n", institution)
	fmt.Printf("🌍 Country: %s\n", country)
	fmt.Printf("🔬 Expertise: %v\n", expertise)
	fmt.Printf("📊 Type: %s\n", regType)
	fmt.Printf("🔧 Capabilities: %v\n", capabilities)
	
	// Initialize client context
	clientCtx, err := initKeysClientContextWithBackend(keyringBackend)
	if err != nil {
		return fmt.Errorf("failed to initialize client context: %w", err)
	}
	
	// Get key info and address
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
	
	fmt.Printf("📍 Address: %s\n", addr.String())
	
	// Test blockchain connection
	cfg := loadConfig()
	fmt.Printf("🔍 Testing connection to %s...\n", cfg.Chain.RPCEndpoint)
	
	if err := testBlockchainConnection(cfg.Chain.RPCEndpoint); err != nil {
		fmt.Printf("⚠️  Blockchain connection failed: %v\n", err)
		fmt.Println("💡 Running in simulation mode...")
		return simulateChatRegistration(from, addr.String(), displayName, institution, capabilities)
	}
	
	fmt.Println("✅ Blockchain connection successful!")
	
	// Setup full client context for transaction
	
	
	cfg := loadConfig()
	rpcClient, err := client.NewClientFromNode(cfg.Chain.RPCEndpoint)
	if err != nil {
    		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	txConfig := authtx.NewTxConfig(globalCodec, authtx.DefaultSignModes)
	fullClientCtx := clientCtx.
    	WithFromName(fromName).
    	WithFromAddress(fromAddr).
    	WithTxConfig(txConfig).
    	WithClient(rpcClient).
    	WithChainID(cfg.Chain.ID).
    	WithCodec(globalCodec).
    	WithInterfaceRegistry(globalInterfaceRegistry).
    	WithBroadcastMode(flags.BroadcastSync)

	if err != nil {
		return fmt.Errorf("failed to setup client context: %w", err)
	}
	
	// Create enhanced registration data
	registration := &blockchain.ChatClientRegistration{
		ClientAddress:    addr.String(),
		Capabilities:     capabilities,
		DisplayName:      displayName,
		Institution:      institution,
		Country:          country,
		Expertise:        expertise,
		ContactInfo:      contact,
		RegistrationType: regType,
		Timestamp:        time.Now(),
		Version:          "1.0.0",
	}
	
	// Perform enhanced registration
	result, err := blockchain.RegisterChatClient(clientCtx, registration)
	if err != nil {
		fmt.Printf("❌ Chat registration failed: %v\n", err)
		fmt.Println("💡 Falling back to simulation...")
		return simulateChatRegistration(from, addr.String(), displayName, institution, capabilities)
	}
	
	// Display success with chat-specific information
	return displayChatRegistrationSuccess(result, cfg.Chain.ID)
}

// displayRegistrationSuccess shows registration success information
func displayRegistrationSuccess(result *blockchain.RegistrationResult, chainID string) error {
	fmt.Println("\n🎉 CLIENT SUCCESSFULLY REGISTERED ON BLOCKCHAIN!")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("🆔 Client ID: %s\n", result.ClientID)
	
	// Handle different registration data types
	if regData, ok := result.RegistrationData.(blockchain.ClientRegistrationData); ok {
		fmt.Printf("📍 Address: %s\n", regData.ClientAddress)
		fmt.Printf("🔧 Capabilities: %v\n", regData.Capabilities)
		if regData.Metadata != "" {
			fmt.Printf("📋 Metadata: %s\n", regData.Metadata)
		}
	}
	
	fmt.Printf("⛓️  Chain: %s\n", chainID)
	fmt.Printf("📊 Transaction Hash: %s\n", result.TransactionHash)
	fmt.Printf("🏔️  Block Height: %d\n", result.BlockHeight)
	fmt.Printf("🕒 Registered: %s\n", result.RegisteredAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("💾 Registration saved to: ~/.medasdigital-client/registrations/\n")
	
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("✅ Your client is now active on the MedasDigital network!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   1. Check status: ./bin/medasdigital-client status")
	fmt.Println("   2. Verify registration: ./bin/medasdigital-client query tx", result.TransactionHash)
	
	return nil
}

// displayChatRegistrationSuccess shows enhanced chat registration success
func displayChatRegistrationSuccess(result *blockchain.RegistrationResult, chainID string) error {
	fmt.Println("\n🎉 CHAT CLIENT SUCCESSFULLY REGISTERED!")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("🆔 Client ID: %s\n", result.ClientID)
	
	// Handle chat registration data
	if regData, ok := result.RegistrationData.(*blockchain.ChatClientRegistration); ok {
		fmt.Printf("📍 Address: %s\n", regData.ClientAddress)
		fmt.Printf("📛 Display Name: %s\n", regData.DisplayName)
		fmt.Printf("🏛️  Institution: %s\n", regData.Institution)
		fmt.Printf("🌍 Country: %s\n", regData.Country)
		fmt.Printf("🔬 Expertise: %v\n", regData.Expertise)
		fmt.Printf("📊 Type: %s\n", regData.RegistrationType)
		fmt.Printf("🔧 Capabilities: %v\n", regData.Capabilities)
		if len(regData.ChatPubKey) > 0 {
			fmt.Printf("🔑 Chat Key: Generated (%d bytes)\n", len(regData.ChatPubKey))
		}
	}
	
	fmt.Printf("⛓️  Chain: %s\n", chainID)
	fmt.Printf("📊 Transaction Hash: %s\n", result.TransactionHash)
	fmt.Printf("🏔️  Block Height: %d\n", result.BlockHeight)
	fmt.Printf("🕒 Registered: %s\n", result.RegisteredAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("💾 Registration saved to: ~/.medasdigital-client/registrations/\n")
	
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("✅ Your chat client is now ready for scientific collaboration!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   1. Check status: ./bin/medasdigital-client status")
	fmt.Println("   2. Discover peers: ./bin/medasdigital-client chat peers")
	fmt.Println("   3. Start chatting: ./bin/medasdigital-client chat start")
	
	return nil
}

// simulateChatRegistration simulates chat registration when blockchain is unavailable
func simulateChatRegistration(keyName, address, displayName, institution string, capabilities []string) error {
	fmt.Println("🧪 Running enhanced chat registration simulation...")
	fmt.Printf("✅ Chat client registration simulated successfully!\n")
	fmt.Printf("🆔 Client ID: client-%s\n", address[:8])
	fmt.Printf("📍 Address: %s\n", address)
	fmt.Printf("📛 Display Name: %s\n", displayName)
	fmt.Printf("🏛️  Institution: %s\n", institution)
	fmt.Printf("🔧 Capabilities: %v\n", capabilities)
	fmt.Printf("🔑 Chat Key: Generated (simulation)\n")
	
	fmt.Println("\n💡 Note: This was a simulation. For real blockchain registration,")
	fmt.Println("   ensure the MedasDigital chain is running and accessible.")
	
	return nil
}

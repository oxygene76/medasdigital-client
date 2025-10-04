package main

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
    
    "github.com/spf13/cobra"
    "github.com/oxygene76/medasdigital-client/pkg/contract"
)

var contractCmd = &cobra.Command{
    Use:   "contract",
    Short: "Interact with MEDAS computing smart contract",
}

var contractListProvidersCmd = &cobra.Command{
    Use:   "list-providers",
    Short: "List all available providers",
    RunE: func(cmd *cobra.Command, args []string) error {
        contractAddr, _ := cmd.Flags().GetString("contract")
        
        client := contract.NewClient(contract.Config{
            ContractAddress: contractAddr,
            RPCEndpoint:     defaultRPCEndpoint,
            ChainID:         defaultChainID,
        }, "", "", "") 
        
        providers, err := client.ListProviders(context.Background())
        if err != nil {
            return err
        }
        
        if len(providers) == 0 {
            fmt.Println("No providers registered")
            return nil
        }
        
        fmt.Println("Available Computing Providers")
        fmt.Println(strings.Repeat("=", 80))
        
        for i, p := range providers {
            statusIcon := "✅"
            if !p.Active {
                statusIcon = "❌"
            }
            
            capacity := float64(p.Capacity-p.ActiveJobs) / float64(p.Capacity) * 100
            
            fmt.Printf("\n%d. %s %s\n", i+1, statusIcon, p.Name)
            fmt.Printf("   Address: %s\n", p.Address)
            fmt.Printf("   Endpoint: %s\n", p.Endpoint)
            fmt.Printf("   Capacity: %d/%d (%.0f%% free)\n", p.ActiveJobs, p.Capacity, capacity)
            fmt.Printf("   Completed: %d | Reputation: %s\n", p.TotalCompleted, p.Reputation)
            
            fmt.Printf("   Services:\n")
            for _, cap := range p.Capabilities {
                price := "N/A"
                if pInfo, ok := p.Pricing[cap.ServiceType]; ok {
                    price = fmt.Sprintf("%s/%s", pInfo.BasePrice, pInfo.Unit)
                }
                fmt.Printf("     - %s: %d max, ~%ds, %s MEDAS\n",
                    cap.ServiceType, cap.MaxComplexity, cap.AvgCompletionTime, price)
            }
        }
        
        return nil
    },
}

var contractSubmitJobCmd = &cobra.Command{
    Use:   "submit-job",
    Short: "Submit computing job",
    RunE: func(cmd *cobra.Command, args []string) error {
        contractAddr, _ := cmd.Flags().GetString("contract")
        clientKey, _ := cmd.Flags().GetString("from")
        jobType, _ := cmd.Flags().GetString("type")
        digits, _ := cmd.Flags().GetInt("digits")
        method, _ := cmd.Flags().GetString("method")
        criteria, _ := cmd.Flags().GetString("criteria")
        payment, _ := cmd.Flags().GetString("payment")
        simulate, _ := cmd.Flags().GetBool("simulate")
        
        // Adresse vom Keyring holen
        clientCtx, err := initKeysClientContext()
        if err != nil {
            return fmt.Errorf("failed to init keyring: %w", err)
        }
        
        keyInfo, err := clientCtx.Keyring.Key(clientKey)
        if err != nil {
            return fmt.Errorf("key not found: %w", err)
        }
        
        clientAddrSDK, err := keyInfo.GetAddress()
        if err != nil {
            return fmt.Errorf("failed to get address: %w", err)
        }
        
        clientAddrStr := clientAddrSDK.String()
        
        client := contract.NewClient(contract.Config{
            ContractAddress: contractAddr,
            RPCEndpoint:     defaultRPCEndpoint,
            ChainID:         defaultChainID,
        }, clientKey, clientAddrStr, cfg.Client.KeyringBackend)  
        
        fmt.Println("Finding best provider...")
        
        provider, err := client.FindBestProvider(context.Background(), jobType, digits, criteria)
        if err != nil {
            return err
        }
        
        fmt.Printf("Selected: %s\n", provider.Name)
        fmt.Printf("  Price: %s MEDAS/digit\n", provider.Pricing[jobType].BasePrice)
        
        if simulate {
            fmt.Println("Simulation mode - not submitting")
            return nil
        }
        
        params := map[string]interface{}{
            "digits": digits,
            "method": method,
        }
        
        fmt.Println("Submitting job...")
        
        jobID, txHash, err := client.SubmitJob(
            context.Background(),
            provider.Address,
            jobType,
            params,
            payment,
        )
        if err != nil {
            return err
        }
        
        fmt.Printf("\nJob submitted!\n")
        fmt.Printf("  Job ID: %d\n", jobID)
        fmt.Printf("  TX Hash: %s\n", txHash)
        fmt.Println("\nWaiting for completion...")
        
        completedJob, err := client.WaitForCompletion(context.Background(), jobID, 10*time.Minute)
        if err != nil {
            fmt.Printf("Check status: contract get-job --job-id %d\n", jobID)
            return err
        }
        
        fmt.Printf("\nCompleted!\n")
        fmt.Printf("  Result: %s\n", completedJob.ResultURL)
        
        return nil
    },
}
var contractGetJobCmd = &cobra.Command{
    Use:   "get-job",
    Short: "Get job status",
    RunE: func(cmd *cobra.Command, args []string) error {
        contractAddr, _ := cmd.Flags().GetString("contract")
        jobID, _ := cmd.Flags().GetUint64("job-id")
        
        client := contract.NewClient(contract.Config{
            ContractAddress: contractAddr,
            RPCEndpoint:     defaultRPCEndpoint,
            ChainID:         defaultChainID,
        }, "", "", "") 
        
        job, err := client.GetJob(context.Background(), jobID)
        if err != nil {
            return err
        }
        
        fmt.Printf("Job #%d\n", job.ID)
        fmt.Println(strings.Repeat("=", 60))
        fmt.Printf("Status: %s\n", job.Status)
        fmt.Printf("Provider: %s\n", job.Provider)
        fmt.Printf("Type: %s\n", job.JobType)
        fmt.Printf("Payment: %s umedas\n", job.PaymentAmount)
        
        if job.Status == "completed" {
            fmt.Printf("Result: %s\n", job.ResultURL)
        }
        
        return nil
    },
}
var contractProviderNodeCmd = &cobra.Command{
    Use:   "provider-node",
    Short: "Run provider node",
    Long:  "Start provider node to process computing jobs",
    RunE: func(cmd *cobra.Command, args []string) error {
        contractAddr, _ := cmd.Flags().GetString("contract")
        register, _ := cmd.Flags().GetBool("register")
        
        // Load config
        cfg := loadConfig()
        
        if !cfg.Provider.Enabled {
            return fmt.Errorf("provider not enabled in config. Set provider.enabled: true")
        }
        
        if cfg.Provider.FundingAddress == "" {
            fmt.Println("⚠️  Warning: No funding_address set - auto-harvest disabled")
        }
        
        // Get provider address from key
        addrCmd := exec.Command(
            "medasdigitald", "keys", "show", cfg.Provider.KeyName, "-a",
            "--keyring-backend", cfg.Provider.KeyringBackend,
        )
        addrOutput, err := addrCmd.Output()
        if err != nil {
            return fmt.Errorf("failed to get provider address: %w", err)
        }
        providerAddr := strings.TrimSpace(string(addrOutput))
        
        fmt.Printf("Provider Address: %s\n", providerAddr)
        
        if register {
            fmt.Println("Registering provider...")
            if err := registerProvider(cfg, contractAddr, providerAddr); err != nil {
                return fmt.Errorf("registration failed: %w", err)
            }
            fmt.Println("✅ Provider registered")
            time.Sleep(5 * time.Second)
        }
        
        // Create provider node with config values
        node := contract.NewProviderNode(
            contractAddr,
            providerAddr,
            cfg.Provider.KeyName,
            cfg.Chain.RPCEndpoint,
            cfg.Chain.ID,
            "MEDAS Provider Node",
            cfg.Provider.Endpoint,
            cfg.Provider.Port,
            cfg.Provider.Workers,
            cfg.Provider.FundingAddress,
            cfg.Provider.MinBalance,
            cfg.Provider.MaxBalance,
            cfg.Provider.HarvestIntervalHours,
        )
        
        return node.Start(context.Background())
    },
}

func getProviderAddressFromKey(keyName string) (string, error) {
    clientCtx, err := initKeysClientContext()
    if err != nil {
        return "", err
    }
    
    keyInfo, err := clientCtx.Keyring.Key(keyName)
    if err != nil {
        return "", err
    }
    
    addr, err := keyInfo.GetAddress()
    if err != nil {
        return "", err
    }
    
    return addr.String(), nil
}

func registerProvider(cfg *Config, contractAddr, providerAddr string) error {
    msg := fmt.Sprintf(`{
        "register_provider": {
            "name": "MEDAS Provider Node",
            "capabilities": [{
                "service_type": "pi_calculation",
                "max_complexity": 100000,
                "avg_completion_time": 180
            }],
            "pricing": {
                "pi_calculation": {
                    "base_price": "0.0001",
                    "unit": "digit"
                }
            },
            "endpoint": "%s"
        }
    }`, cfg.Provider.Endpoint)
    
    cmd := exec.Command(
        "medasdigitald", "tx", "wasm", "execute",
        contractAddr, msg,
        "--from", cfg.Provider.KeyName,
        "--keyring-backend", cfg.Provider.KeyringBackend,
        "--gas", "300000",
        "--fees", "80000umedas",
        "--node", cfg.Chain.RPCEndpoint,
        "--chain-id", cfg.Chain.ID,
        "-y",
    )
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("%w\noutput: %s", err, output)
    }
    
    return nil
}

func init() {
    rootCmd.AddCommand(contractCmd)
    contractCmd.AddCommand(contractListProvidersCmd)
    contractCmd.AddCommand(contractSubmitJobCmd)
    contractCmd.AddCommand(contractGetJobCmd)
    contractCmd.AddCommand(contractProviderNodeCmd)
    
    contractCmd.PersistentFlags().String("contract",
        "medas1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrswl7kpn",
        "Contract address")
    
    contractSubmitJobCmd.Flags().String("from", "", "Client key (required)")
    contractSubmitJobCmd.Flags().String("type", "pi_calculation", "Job type")
    contractSubmitJobCmd.Flags().Int("digits", 1000, "Digits")
    contractSubmitJobCmd.Flags().String("method", "chudnovsky", "Method")
    contractSubmitJobCmd.Flags().String("criteria", "price", "Selection criteria")
    contractSubmitJobCmd.Flags().String("payment", "1000000umedas", "Payment")
    contractSubmitJobCmd.Flags().Bool("simulate", false, "Simulate only")
    contractSubmitJobCmd.MarkFlagRequired("from")
    
    contractGetJobCmd.Flags().Uint64("job-id", 0, "Job ID (required)")
    contractGetJobCmd.MarkFlagRequired("job-id")

    // contractProviderNodeCmd.Flags().String("provider-key", "", "Provider key name (required)")
    // contractProviderNodeCmd.Flags().String("name", "MEDAS Provider", "Provider name")
    // contractProviderNodeCmd.Flags().String("endpoint", "", "Provider endpoint URL (required)")
    // contractProviderNodeCmd.Flags().Int("port", 8080, "HTTP port")
    // contractProviderNodeCmd.Flags().Int("workers", 4, "Worker threads")
    // contractProviderNodeCmd.Flags().Bool("register", false, "Register provider first")
    
    // contractProviderNodeCmd.MarkFlagRequired("provider-key")
    // contractProviderNodeCmd.MarkFlagRequired("endpoint")

    contractProviderNodeCmd.Flags().Bool("register", false, "Register provider first")
}

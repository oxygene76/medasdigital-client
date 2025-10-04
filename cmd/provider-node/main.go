package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "os/signal"
    "strings"
    
    "github.com/oxygene76/medasdigital-client/pkg/contract"
)

func main() {
    var (
        contractAddr = flag.String("contract", "", "Contract address")
        providerKey  = flag.String("provider-key", "", "Provider key name")
        providerName = flag.String("name", "MEDAS Provider", "Provider name")
        endpoint     = flag.String("endpoint", "", "Provider endpoint URL")
        httpPort     = flag.Int("port", 8080, "HTTP port")
        workers      = flag.Int("workers", 4, "Worker threads")
        register     = flag.Bool("register", false, "Register first")
    )
    
    flag.Parse()
    
    if *contractAddr == "" || *providerKey == "" || *endpoint == "" {
        fmt.Println("Usage:")
        fmt.Println("  provider-node --contract <addr> --provider-key <key> --endpoint <url>")
        os.Exit(1)
    }
    
    providerAddr, err := getProviderAddress(*providerKey)
    if err != nil {
        log.Fatalf("Get address failed: %v", err)
    }
    
    node := contract.NewProviderNode(
        *contractAddr,
        providerAddr,
        *providerKey,
        "https://rpc.medas-digital.io:26657",
        "medasdigital-2",
        *providerName,
        *endpoint,
        *httpPort,
        *workers,
    )
    
    if *register {
        if err := node.RegisterProvider(*endpoint); err != nil {
            log.Fatalf("Registration failed: %v", err)
        }
        log.Printf("Provider registered")
    }
    
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    
    if err := node.Start(ctx); err != nil {
        log.Fatalf("Node failed: %v", err)
    }
}

func getProviderAddress(keyName string) (string, error) {
    cmd := exec.Command("medasdigitald", "keys", "show", keyName, "-a")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

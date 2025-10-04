package contract

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/oxygene76/medasdigital-client/pkg/compute"
)

type ProviderNode struct {
    contractAddr         string
    providerAddr         string
    providerKey          string
    rpcURL               string
    chainID              string
    providerName         string
    endpointURL          string
    httpPort             int
    workers              int
    fundingAddress       string
    minBalance           uint64
    maxBalance           uint64
    harvestInterval      time.Duration
    jobManager           *compute.JobManager
    wsClient             *websocket.Conn
    results              map[string]*compute.ComputeJob  // NEW: Store results
    resultsMu            sync.RWMutex                     // NEW: Mutex for thread-safe access
}

func NewProviderNode(
    contractAddr, providerAddr, providerKey, rpcURL, chainID string,
    providerName, endpointURL string,
    httpPort, workers int,
    fundingAddress string,
    minBalance, maxBalance uint64,
    harvestIntervalHours int,
) *ProviderNode {
    return &ProviderNode{
        contractAddr:    contractAddr,
        providerAddr:    providerAddr,
        providerKey:     providerKey,
        rpcURL:          rpcURL,
        chainID:         chainID,
        providerName:    providerName,
        endpointURL:     endpointURL,
        httpPort:        httpPort,
        workers:         workers,
        fundingAddress:  fundingAddress,
        minBalance:      minBalance,
        maxBalance:      maxBalance,
        harvestInterval: time.Duration(harvestIntervalHours) * time.Hour,
        jobManager: compute.NewJobManager(workers, 100, compute.NewPricingManager("medas1kc7lctfazdpd8y6ecapdfv3d6ch97prc58qaem")),
        results:         make(map[string]*compute.ComputeJob), // NEW: Initialize results map
    }
}

func (p *ProviderNode) Start(ctx context.Context) error {
    log.Printf("Provider Node Started")
    log.Printf("  Name: %s", p.providerName)
    log.Printf("  Address: %s", p.providerAddr)
    log.Printf("  Endpoint: %s", p.endpointURL)
    log.Printf("  Listening for jobs...")
    
    if p.fundingAddress != "" {
        log.Printf("  Auto-Harvest enabled:")
        log.Printf("    Funding Address: %s", p.fundingAddress)
        log.Printf("    Min Balance: %d umedas", p.minBalance)
        log.Printf("    Max Balance: %d umedas", p.maxBalance)
        log.Printf("    Check Interval: %v", p.harvestInterval)
        go p.autoHarvest(ctx)
    } else {
        log.Printf("  Auto-Harvest disabled (no funding_address set)")
    }

    go p.startHTTPServer(ctx)
    
    return p.subscribeToJobs(ctx)
}

func (p *ProviderNode) autoHarvest(ctx context.Context) {
    ticker := time.NewTicker(p.harvestInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            log.Println("Auto-harvest stopped")
            return
        case <-ticker.C:
            p.harvestExcessBalance()
        }
    }
}

func (p *ProviderNode) harvestExcessBalance() {
    balance, err := p.getProviderBalance()
    if err != nil {
        log.Printf("Failed to get balance: %v", err)
        return
    }
    
    if balance <= p.maxBalance {
        log.Printf("Balance check: %d umedas (below threshold)", balance)
        return
    }
    
    transfer := balance - p.minBalance
    log.Printf("💰 Harvesting %d umedas to funding wallet", transfer)
    
    cmd := exec.Command(
        "medasdigitald", "tx", "bank", "send",
        p.providerKey, p.fundingAddress, fmt.Sprintf("%dumedas", transfer),
        "--keyring-backend", "test",
        "--gas", "200000",
        "--fees", "5000umedas",
        "--node", p.rpcURL,
        "--chain-id", p.chainID,
        "-y",
        "--output", "json",
    )
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        log.Printf("❌ Harvest failed: %v\nstderr: %s", err, stderr.String())
    } else {
        log.Printf("✅ Successfully harvested %d umedas", transfer)
    }
}

func (p *ProviderNode) getProviderBalance() (uint64, error) {
    cmd := exec.Command(
        "medasdigitald", "query", "bank", "balances", p.providerAddr,
        "--node", p.rpcURL,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }
    
    var result struct {
        Balances []struct {
            Denom  string `json:"denom"`
            Amount string `json:"amount"`
        } `json:"balances"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return 0, err
    }
    
    for _, balance := range result.Balances {
        if balance.Denom == "umedas" {
            amount, err := strconv.ParseUint(balance.Amount, 10, 64)
            if err != nil {
                return 0, err
            }
            return amount, nil
        }
    }
    
    return 0, nil
}

func (p *ProviderNode) subscribeToJobs(ctx context.Context) error {
    wsURL := "wss://rpc.medas-digital.io:26657/websocket"
    
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    if err != nil {
        return fmt.Errorf("websocket dial failed: %w", err)
    }
    p.wsClient = conn
    defer conn.Close()
    
    query := fmt.Sprintf(
        "wasm._contract_address='%s' AND wasm.action='submit_job' AND wasm.provider='%s'",
        p.contractAddr,
        p.providerAddr,
    )
    
    log.Printf("🔍 Subscribing with query: %s", query)
    
    subscribeMsg := map[string]interface{}{
        "jsonrpc": "2.0",
        "method":  "subscribe",
        "id":      1,
        "params": map[string]interface{}{
            "query": query,
        },
    }
    
    if err := conn.WriteJSON(subscribeMsg); err != nil {
        return fmt.Errorf("subscribe failed: %w", err)
    }
    
    log.Printf("✅ WebSocket connected and subscribed")
    
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            var msg map[string]interface{}
            if err := conn.ReadJSON(&msg); err != nil {
                log.Printf("Read error: %v", err)
                continue
            }
            
            if result, ok := msg["result"].(map[string]interface{}); ok {
                if events, ok := result["events"].(map[string]interface{}); ok {
                    p.handleJobEvent(ctx, events)
                } else if data, ok := result["data"].(map[string]interface{}); ok {
                    if value, ok := data["value"].(map[string]interface{}); ok {
                        if txResult, ok := value["TxResult"].(map[string]interface{}); ok {
                            if result, ok := txResult["result"].(map[string]interface{}); ok {
                                if evts, ok := result["events"].([]interface{}); ok {
                                    p.handleJobEventArray(ctx, evts)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

func (p *ProviderNode) handleJobEventArray(ctx context.Context, events []interface{}) {
    for _, evt := range events {
        if event, ok := evt.(map[string]interface{}); ok {
            if eventType, ok := event["type"].(string); ok && eventType == "wasm" {
                if attrs, ok := event["attributes"].([]interface{}); ok {
                    var jobID uint64
                    for _, attr := range attrs {
                        if a, ok := attr.(map[string]interface{}); ok {
                            if key, _ := a["key"].(string); key == "job_id" {
                                if value, ok := a["value"].(string); ok {
                                    jobID, _ = strconv.ParseUint(value, 10, 64)
                                    break
                                }
                            }
                        }
                    }
                    if jobID > 0 {
                        log.Printf("📥 New job received: %d", jobID)
                        go p.processJob(ctx, jobID)
                    }
                }
            }
        }
    }
}

func (p *ProviderNode) handleJobEvent(ctx context.Context, events map[string]interface{}) {
    wasmEvents, ok := events["wasm.job_id"].([]interface{})
    if !ok || len(wasmEvents) == 0 {
        return
    }
    
    jobIDStr := wasmEvents[0].(string)
    jobID, _ := strconv.ParseUint(jobIDStr, 10, 64)
    
    log.Printf("📥 New job received: %d", jobID)
    
    go p.processJob(ctx, jobID)
}

func (p *ProviderNode) processJob(ctx context.Context, contractJobID uint64) {
    cj, err := p.getContractJob(ctx, contractJobID)
    if err != nil {
        log.Printf("Failed to get job: %v", err)
        return
    }
    
    var params map[string]interface{}
    if err := json.Unmarshal([]byte(cj.Parameters), &params); err != nil {
        log.Printf("Failed to parse parameters: %v", err)
        return
    }
    
    log.Printf("Processing job %d: %s", contractJobID, cj.JobType)
    
    job, err := p.jobManager.SubmitJob(
        compute.JobTypePICalculation,
        params,
        cj.Client,
        compute.TierStandard,
        "",
    )
    if err != nil {
        log.Printf("Failed to submit job: %v", err)
        return
    }
    
    // Wait for completion and get final job state
    var completedJob *compute.ComputeJob
    for {
        time.Sleep(1 * time.Second)
        currentJob, _ := p.jobManager.GetJob(job.ID)
        if currentJob.Status == compute.StatusCompleted {
            completedJob = currentJob // NEW: Save the completed job
            break
        }
        if currentJob.Status == compute.StatusFailed {
            log.Printf("Job failed")
            return
        }
    }
    
    // NEW: Store the result for HTTP retrieval
    p.resultsMu.Lock()
    p.results[job.ID] = completedJob
    p.resultsMu.Unlock()
    
    resultURL := fmt.Sprintf("%s/results/%s.json", p.endpointURL, job.ID)
    
    // NEW: Calculate real hash from result
    resultData, _ := json.Marshal(completedJob.Result)
    hash := sha256.Sum256(resultData)
    resultHash := hex.EncodeToString(hash[:])
    
    log.Printf("✅ Job completed, marking as complete in contract")
    
    if err := p.completeContractJob(ctx, contractJobID, resultHash, resultURL); err != nil {
        log.Printf("Failed to complete job in contract: %v", err)
        return
    }
    
    log.Printf("Job %d completed successfully", contractJobID)
}

func (p *ProviderNode) completeContractJob(ctx context.Context, jobID uint64, hash, url string) error {
    msg := fmt.Sprintf(`{"complete_job":{"job_id":%d,"result_hash":"%s","result_url":"%s"}}`,
        jobID, hash, url)
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "tx", "wasm", "execute",
        p.contractAddr, msg,
        "--from", p.providerKey,
        "--keyring-backend", "test",
        "--gas", "220000",
        "--fees", "5500umedas",
        "-y",
        "--node", p.rpcURL,
        "--chain-id", p.chainID,
        "--output", "json",
    )
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("complete job tx failed: %w\nstderr: %s", err, stderr.String())
    }
    
    return nil
}

func (p *ProviderNode) getContractJob(ctx context.Context, jobID uint64) (*ContractJob, error) {
    query := fmt.Sprintf(`{"get_job":{"job_id":%d}}`, jobID)
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "wasm", "contract-state", "smart",
        p.contractAddr, query,
        "--node", p.rpcURL,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var result struct {
        Data ContractJob `json:"data"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }
    
    return &result.Data, nil
}

func (p *ProviderNode) startHTTPServer(ctx context.Context) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
    })
    
    // NEW: Enhanced results handler that returns real PI results
    http.HandleFunc("/results/", func(w http.ResponseWriter, r *http.Request) {
        // Extract job ID from URL: /results/pi_calculation-1.json
        path := strings.TrimPrefix(r.URL.Path, "/results/")
        jobID := strings.TrimSuffix(path, ".json")
        
        p.resultsMu.RLock()
        job, exists := p.results[jobID]
        p.resultsMu.RUnlock()
        
        w.Header().Set("Content-Type", "application/json")
        
        if !exists {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Result not found",
                "job_id": jobID,
            })
            return
        }
        
        // Return the actual computation result
        json.NewEncoder(w).Encode(map[string]interface{}{
            "job_id":       job.ID,
            "status":       job.Status,
            "result":       job.Result,
            "duration":     job.Duration,
            "completed_at": job.CompletedAt,
            "tier":         job.Tier,
            "parameters":   job.Parameters,
        })
    })
    
    addr := fmt.Sprintf(":%d", p.httpPort)
    log.Printf("HTTP server on port %d", p.httpPort)
    
    server := &http.Server{Addr: addr}
    
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()
    
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Printf("HTTP server error: %v", err)
    }
}

package contract

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "strconv"
    "time"
    
    "github.com/gorilla/mux"
    tmrpc "github.com/cometbft/cometbft/rpc/client/http"
    coretypes "github.com/cometbft/cometbft/rpc/core/types"
    
    "github.com/oxygene76/medasdigital-client/pkg/compute"
)

type ProviderNode struct {
    contractAddr  string
    providerAddr  string
    providerKey   string
    rpcURL        string
    chainID       string
    providerName  string
    endpointURL   string
    httpPort      int
    
    wsClient      *tmrpc.HTTP
    jobManager    *compute.JobManager
}

func NewProviderNode(
    contractAddr, providerAddr, providerKey, rpcURL, chainID string,
    providerName, endpointURL string,
    httpPort, workers int,
) *ProviderNode {
    pricingManager := compute.NewPricingManager("")
    jobManager := compute.NewJobManager(10, workers, pricingManager)
    
    return &ProviderNode{
        contractAddr: contractAddr,
        providerAddr: providerAddr,
        providerKey:  providerKey,
        rpcURL:       rpcURL,
        chainID:      chainID,
        providerName: providerName,
        endpointURL:  endpointURL,
        httpPort:     httpPort,
        jobManager:   jobManager,
    }
}

func (p *ProviderNode) Start(ctx context.Context) error {
    // HTTP-Server starten
    go p.startHTTPServer(ctx)
    
    // WebSocket-Client
    wsClient, err := tmrpc.New(p.rpcURL, "/websocket")
    if err != nil {
        return fmt.Errorf("websocket failed: %w", err)
    }
    p.wsClient = wsClient
    
    if err := p.wsClient.Start(); err != nil {
        return err
    }
    defer p.wsClient.Stop()
    
    // Subscribe auf Jobs
    query := fmt.Sprintf(
        "wasm._contract_address='%s' AND wasm.action='submit_job' AND wasm.provider='%s'",
        p.contractAddr,
        p.providerAddr,
    )
    
    eventCh, err := p.wsClient.Subscribe(ctx, "provider", query)
    if err != nil {
        return err
    }
    
    log.Printf("Provider Node Started")
    log.Printf("  Name: %s", p.providerName)
    log.Printf("  Address: %s", p.providerAddr)
    log.Printf("  Endpoint: %s", p.endpointURL)
    log.Printf("  Listening for jobs...")
    
    for {
        select {
        case <-ctx.Done():
            return nil
        case event := <-eventCh:
            go p.handleJobEvent(ctx, event)
        }
    }
}

func (p *ProviderNode) handleJobEvent(ctx context.Context, event coretypes.ResultEvent) {
    jobIDStr := event.Events["wasm.job_id"]
    if len(jobIDStr) == 0 {
        return
    }
    
    jobID, _ := strconv.ParseUint(jobIDStr[0], 10, 64)
    log.Printf("New job: %d", jobID)
    
    contractJob, err := p.queryContractJob(ctx, jobID)
    if err != nil {
        log.Printf("Query failed: %v", err)
        return
    }
    
    go p.processJob(ctx, jobID, contractJob)
}

func (p *ProviderNode) processJob(ctx context.Context, contractJobID uint64, cj ContractJob) {
    var params map[string]interface{}
    json.Unmarshal([]byte(cj.Parameters), &params)
    
    job, err := p.jobManager.SubmitJob(
        compute.JobTypePICalculation,
        params,
        cj.Client,
        compute.TierStandard,
        "",
    )
    if err != nil {
        log.Printf("Job failed: %v", err)
        return
    }
    
    job.PaymentVerified = true
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            currentJob, _ := p.jobManager.GetJob(job.ID)
            
            if currentJob.Status == compute.StatusCompleted {
                resultURL := fmt.Sprintf("%s/results/%s.json", p.endpointURL, job.ID)
                resultHash := fmt.Sprintf("hash_%s", job.ID)
                
                if err := p.completeContractJob(ctx, contractJobID, resultHash, resultURL); err != nil {
                    log.Printf("Complete failed: %v", err)
                    return
                }
                
                log.Printf("Job %d completed", contractJobID)
                return
            }
            
            if currentJob.Status == compute.StatusFailed {
                log.Printf("Job %d failed: %s", contractJobID, currentJob.Error)
                return
            }
        }
    }
}

func (p *ProviderNode) RegisterProvider(endpoint string) error {
    msg := map[string]interface{}{
        "register_provider": map[string]interface{}{
            "name": p.providerName,
            "capabilities": []map[string]interface{}{
                {
                    "service_type":       "pi_calculation",
                    "max_complexity":     100000,
                    "avg_completion_time": 180,
                },
            },
            "pricing": map[string]interface{}{
                "pi_calculation": map[string]interface{}{
                    "base_price": "0.0001",
                    "unit":       "digit",
                },
            },
            "endpoint": endpoint,
        },
    }
    
    msgJSON, _ := json.Marshal(msg)
    
    gasEst, err := EstimateGas(
        context.Background(),
        p.contractAddr,
        string(msgJSON),
        p.providerKey,
        "",
        p.rpcURL,
        p.chainID,
    )
    if err != nil {
        return fmt.Errorf("gas estimation failed: %w", err)
    }
    
    cmd := exec.Command(
        "medasdigitald", "tx", "wasm", "execute",
        p.contractAddr, string(msgJSON),
        "--from", p.providerKey,
        "--gas", fmt.Sprintf("%d", gasEst.GasWanted),
        "--fees", gasEst.Fees,
        "-y",
        "--node", p.rpcURL,
        "--chain-id", p.chainID,
    )
    
    return cmd.Run()
}

func (p *ProviderNode) completeContractJob(ctx context.Context, jobID uint64, hash, url string) error {
    msg := fmt.Sprintf(`{"complete_job":{"job_id":%d,"result_hash":"%s","result_url":"%s"}}`,
        jobID, hash, url)
    
    gasEst, err := EstimateGas(
        ctx,
        p.contractAddr,
        msg,
        p.providerKey,
        "",
        p.rpcURL,
        p.chainID,
    )
    if err != nil {
        return err
    }
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "tx", "wasm", "execute",
        p.contractAddr, msg,
        "--from", p.providerKey,
        "--gas", fmt.Sprintf("%d", gasEst.GasWanted),
        "--fees", gasEst.Fees,
        "-y",
        "--node", p.rpcURL,
        "--chain-id", p.chainID,
    )
    
    return cmd.Run()
}

func (p *ProviderNode) queryContractJob(ctx context.Context, jobID uint64) (ContractJob, error) {
    query := fmt.Sprintf(`{"get_job":{"job_id":%d}}`, jobID)
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "wasm", "contract-state", "smart",
        p.contractAddr, query,
        "--node", p.rpcURL,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return ContractJob{}, err
    }
    
    var result struct {
        Data ContractJob `json:"data"`
    }
    
    json.Unmarshal(output, &result)
    return result.Data, nil
}

func (p *ProviderNode) startHTTPServer(ctx context.Context) error {
    r := mux.NewRouter()
    r.HandleFunc("/results/{jobID}.json", p.handleGetResult).Methods("GET")
    r.HandleFunc("/health", p.handleHealth).Methods("GET")
    
    server := &http.Server{Addr: fmt.Sprintf(":%d", p.httpPort), Handler: r}
    
    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()
    
    log.Printf("HTTP server on port %d", p.httpPort)
    return server.ListenAndServe()
}

func (p *ProviderNode) handleGetResult(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    jobID := vars["jobID"]
    
    job, err := p.jobManager.GetJob(jobID)
    if err != nil {
        http.Error(w, "Result not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(job.Result)
}

func (p *ProviderNode) handleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":   "healthy",
        "provider": p.providerName,
        "address":  p.providerAddr,
    })
}

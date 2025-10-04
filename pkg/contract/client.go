package contract

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "sort"
    "strconv"
    "strings"
    "time"
)

type Client struct {
    config     Config
    clientKey  string
    clientAddr string
    keyringBackend string
}

func NewClient(config Config, clientKey string, clientAddr string, keyringBackend string) *Client {
    return &Client{
        config:         config,
        clientKey:      clientKey,
        clientAddr:     clientAddr,
        keyringBackend: keyringBackend,
    }
}

// ListProviders holt alle Provider vom Contract
func (c *Client) ListProviders(ctx context.Context) ([]Provider, error) {
    query := `{"list_providers":{}}`
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "wasm", "contract-state", "smart",
        c.config.ContractAddress, query,
        "--node", c.config.RPCEndpoint,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("query failed: %w, output: %s", err, output)
    }
    
    var result struct {
        Data struct {
            Providers []Provider `json:"providers"`
        } `json:"data"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("parse failed: %w", err)
    }
    
    return result.Data.Providers, nil
}

// FindBestProvider wählt Provider basierend auf Kriterien
func (c *Client) FindBestProvider(
    ctx context.Context,
    jobType string,
    complexity int,
    criteria string,
) (*Provider, error) {
    providers, err := c.ListProviders(ctx)
    if err != nil {
        return nil, err
    }
    
    var suitable []Provider
    for _, p := range providers {
        if !p.Active {
            continue
        }
        
        for _, cap := range p.Capabilities {
            if cap.ServiceType == jobType && cap.MaxComplexity >= complexity {
                if p.ActiveJobs < p.Capacity {
                    suitable = append(suitable, p)
                }
                break
            }
        }
    }
    
    if len(suitable) == 0 {
        return nil, fmt.Errorf("no suitable provider found")
    }
    
    switch criteria {
    case "price":
        sort.Slice(suitable, func(i, j int) bool {
            return getPrice(suitable[i], jobType) < getPrice(suitable[j], jobType)
        })
    case "speed":
        sort.Slice(suitable, func(i, j int) bool {
            return getAvgTime(suitable[i], jobType) < getAvgTime(suitable[j], jobType)
        })
    case "reputation":
        sort.Slice(suitable, func(i, j int) bool {
            return getReputationScore(suitable[i]) > getReputationScore(suitable[j])
        })
    case "availability":
        sort.Slice(suitable, func(i, j int) bool {
            availI := float64(suitable[i].Capacity - suitable[i].ActiveJobs)
            availJ := float64(suitable[j].Capacity - suitable[j].ActiveJobs)
            return availI > availJ
        })
    default:
        return nil, fmt.Errorf("unknown criteria: %s", criteria)
    }
    
    return &suitable[0], nil
}

// SubmitJob submitted Job mit Auto-Gas
func (c *Client) SubmitJob(
    ctx context.Context,
    providerAddr string,
    jobType string,
    parameters map[string]interface{},
    paymentAmount string,
) (uint64, string, error) {
    paramsJSON, _ := json.Marshal(parameters)
    paramsStr := strings.ReplaceAll(string(paramsJSON), `"`, `\"`)
    
    msg := fmt.Sprintf(`{"submit_job":{"provider":"%s","job_type":"%s","parameters":"%s"}}`,
        providerAddr, jobType, paramsStr)
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "tx", "wasm", "execute",
        c.config.ContractAddress, msg,
        "--amount", paymentAmount,
        "--from", c.clientKey,
        "--keyring-backend", c.keyringBackend,
        "--gas", "auto",
        "--gas-adjustment", "1.3",
        "--fees", "6000umedas",
        "-y",
        "--node", c.config.RPCEndpoint,
        "--chain-id", c.config.ChainID,
        "--output", "json",
    )
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return 0, "", fmt.Errorf("submit failed: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
    }
    
    var txResp struct {
        TxHash string `json:"txhash"`
    }
    if err := json.Unmarshal(stdout.Bytes(), &txResp); err != nil {
        return 0, "", fmt.Errorf("parse tx response failed: %w", err)
    }
    
    fmt.Printf("TX Hash: %s\n", txResp.TxHash)  // ← DEBUG: TX Hash ausgeben
    fmt.Println("Waiting for TX finalization...")
    
    time.Sleep(6 * time.Second)
    jobID, err := c.getJobIDFromTx(ctx, txResp.TxHash)
    if err != nil {
        return 0, txResp.TxHash, fmt.Errorf("get job_id failed: %w", err)
    }
    
    return jobID, txResp.TxHash, nil
}
// GetJob holt Job-Details
func (c *Client) GetJob(ctx context.Context, jobID uint64) (*ContractJob, error) {
    query := fmt.Sprintf(`{"get_job":{"job_id":%d}}`, jobID)
    
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "wasm", "contract-state", "smart",
        c.config.ContractAddress, query,
        "--node", c.config.RPCEndpoint,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    var result struct {
        Data ContractJob `json:"data"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }
    
    return &result.Data, nil
}

// WaitForCompletion wartet auf Job-Completion
func (c *Client) WaitForCompletion(ctx context.Context, jobID uint64, timeout time.Duration) (*ContractJob, error) {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for time.Now().Before(deadline) {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-ticker.C:
            job, err := c.GetJob(ctx, jobID)
            if err != nil {
                continue
            }
            
            if job.Status == JobStatusCompleted {
                return job, nil
            }
            
            if job.Status == JobStatusFailed {
                return nil, fmt.Errorf("job failed")
            }
        }
    }
    
    return nil, fmt.Errorf("job timeout after %v", timeout)
}

// Helper: Job-ID aus TX extrahieren
func (c *Client) getJobIDFromTx(ctx context.Context, txHash string) (uint64, error) {
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "tx", txHash,
        "--node", c.config.RPCEndpoint,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }
    
    var txResp struct {
        Logs []struct {
            Events []struct {
                Type       string `json:"type"`
                Attributes []struct {
                    Key   string `json:"key"`
                    Value string `json:"value"`
                } `json:"attributes"`
            } `json:"events"`
        } `json:"logs"`
    }
    
    if err := json.Unmarshal(output, &txResp); err != nil {
        return 0, err
    }
    
    for _, log := range txResp.Logs {
        for _, event := range log.Events {
            if event.Type == "wasm" {
                for _, attr := range event.Attributes {
                    if attr.Key == "job_id" {
                        return strconv.ParseUint(attr.Value, 10, 64)
                    }
                }
            }
        }
    }
    
    return 0, fmt.Errorf("job_id not found in tx")
}

// Helper functions
func getPrice(p Provider, jobType string) float64 {
    if price, ok := p.Pricing[jobType]; ok {
        var val float64
        fmt.Sscanf(price.BasePrice, "%f", &val)
        return val
    }
    return 999999.0
}

func getAvgTime(p Provider, jobType string) int {
    for _, cap := range p.Capabilities {
        if cap.ServiceType == jobType {
            return cap.AvgCompletionTime
        }
    }
    return 999999
}

func getReputationScore(p Provider) float64 {
    var score float64
    fmt.Sscanf(p.Reputation, "%f", &score)
    return score * float64(p.TotalCompleted)
}

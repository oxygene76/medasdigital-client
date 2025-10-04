package contract

// Config für Contract-Interaktion
type Config struct {
    ContractAddress string
    RPCEndpoint     string
    ChainID         string
}

// Provider aus Smart Contract
type Provider struct {
    Address        string                 `json:"address"`
    Name           string                 `json:"name"`
    Capabilities   []Capability           `json:"capabilities"`
    Pricing        map[string]PriceInfo   `json:"pricing"`
    Endpoint       string                 `json:"endpoint"`
    Capacity       int                    `json:"capacity"`
    ActiveJobs     int                    `json:"active_jobs"`
    TotalCompleted int                    `json:"total_completed"`
    Reputation     string                 `json:"reputation"`
    Active         bool                   `json:"active"`
    RegisteredAt   string                 `json:"registered_at"`
}

type Capability struct {
    ServiceType       string `json:"service_type"`
    MaxComplexity     int    `json:"max_complexity"`
    AvgCompletionTime int    `json:"avg_completion_time"`
}

type PriceInfo struct {
    BasePrice string `json:"base_price"`
    Unit      string `json:"unit"`
}

// Job aus Smart Contract
type ContractJob struct {
    ID            uint64 `json:"id"`
    Client        string `json:"client"`
    Provider      string `json:"provider"`
    JobType       string `json:"job_type"`
    Parameters    string `json:"parameters"`
    PaymentAmount string `json:"payment_amount"`
    Status        string `json:"status"`
    ResultHash    string `json:"result_hash,omitempty"`
    ResultURL     string `json:"result_url,omitempty"`
    CreatedAt     string `json:"created_at"`
    CompletedAt   string `json:"completed_at,omitempty"`
}

// JobStatus constants
const (
    JobStatusSubmitted = "submitted"
    JobStatusCompleted = "completed"
    JobStatusFailed    = "failed"
    JobStatusCancelled = "cancelled"
)

// GasEstimation für Transaktionen
type GasEstimation struct {
    GasWanted uint64
    GasUsed   uint64
    Fees      string
}

// ProviderRegistration für neue Provider
type ProviderRegistration struct {
    Name         string
    Endpoint     string
    Capabilities []Capability
    Pricing      map[string]PriceInfo
}

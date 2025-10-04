package contract

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

// EstimateGas führt Gas-Simulation durch
func EstimateGas(
    ctx context.Context,
    contractAddr string,
    msg string,
    fromAddr string,
    amount string,
    rpcEndpoint string,
    chainID string,
) (*GasEstimation, error) {
    
    args := []string{
        "tx", "wasm", "execute",
        contractAddr, msg,
        "--from", fromAddr,
        "--gas", "auto",
        "--gas-adjustment", "1.3",
        "--generate-only",
        "--offline",
        "--chain-id", chainID,
        "--account-number", "0",
        "--sequence", "0",
        "--output", "json",
    }
    
    if amount != "" {
        args = append(args, "--amount", amount)
    }
    
    cmd := exec.CommandContext(ctx, "medasdigitald", args...)
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("gas estimation failed: %w, output: %s", err, output)
    }
    
    var result struct {
        AuthInfo struct {
            Fee struct {
                GasLimit string `json:"gas_limit"`
            } `json:"fee"`
        } `json:"auth_info"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("parse gas estimation: %w", err)
    }
    
    gasWanted, _ := strconv.ParseUint(result.AuthInfo.Fee.GasLimit, 10, 64)
    gasUsed := uint64(float64(gasWanted) / 1.3)
    
    feePerGas := 0.025
    totalFee := uint64(float64(gasWanted) * feePerGas)
    
    return &GasEstimation{
        GasWanted: gasWanted,
        GasUsed:   gasUsed,
        Fees:      fmt.Sprintf("%dumedas", totalFee),
    }, nil
}

// GetMinimumGasPrice holt minimalen Gas-Preis vom Node
func GetMinimumGasPrice(ctx context.Context, rpcEndpoint string) (string, error) {
    cmd := exec.CommandContext(ctx,
        "medasdigitald", "query", "txfees", "minimum-gas-prices",
        "--node", rpcEndpoint,
        "--output", "json",
    )
    
    output, err := cmd.Output()
    if err != nil {
        return "0.025umedas", nil
    }
    
    var result struct {
        MinimumGasPrices []struct {
            Denom  string `json:"denom"`
            Amount string `json:"amount"`
        } `json:"minimum_gas_prices"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return "0.025umedas", nil
    }
    
    for _, price := range result.MinimumGasPrices {
        if price.Denom == "umedas" {
            return price.Amount + price.Denom, nil
        }
    }
    
    return "0.025umedas", nil
}

// EstimateGasWithAdjustment erlaubt custom gas-adjustment
func EstimateGasWithAdjustment(
    ctx context.Context,
    contractAddr string,
    msg string,
    fromKey string,
    amount string,
    rpcEndpoint string,
    chainID string,
    gasAdjustment float64,
) (*GasEstimation, error) {
    
    addrCmd := exec.CommandContext(ctx, "medasdigitald", "keys", "show", fromKey, "-a")
    addrOutput, err := addrCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get address from key: %w", err)
    }
    fromAddr := strings.TrimSpace(string(addrOutput))
    
    args := []string{
        "tx", "wasm", "execute",
        contractAddr, msg,
        "--from", fromAddr,
        "--gas", "auto",
        "--gas-adjustment", fmt.Sprintf("%.2f", gasAdjustment),
        "--dry-run",
        "--node", rpcEndpoint,
        "--chain-id", chainID,
        "--output", "json",
    }
    
    if amount != "" {
        args = append(args, "--amount", amount)
    }
    
    cmd := exec.CommandContext(ctx, "medasdigitald", args...)
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("gas estimation failed: %w", err)
    }
    
    var result struct {
        GasInfo struct {
            GasWanted string `json:"gas_wanted"`
            GasUsed   string `json:"gas_used"`
        } `json:"gas_info"`
    }
    
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("parse gas estimation: %w", err)
    }
    
    gasWanted, _ := strconv.ParseUint(result.GasInfo.GasWanted, 10, 64)
    gasUsed, _ := strconv.ParseUint(result.GasInfo.GasUsed, 10, 64)
    
    feePerGas := 0.025
    totalFee := uint64(float64(gasWanted) * feePerGas)
    
    return &GasEstimation{
        GasWanted: gasWanted,
        GasUsed:   gasUsed,
        Fees:      fmt.Sprintf("%dumedas", totalFee),
    }, nil
}

package compute

import (
	"fmt"
	"math"
	"time"
)

// ServiceTier represents different service levels
type ServiceTier string

const (
	TierBasic    ServiceTier = "basic"
	TierStandard ServiceTier = "standard"
	TierPremium  ServiceTier = "premium"
)

// PricingTier defines pricing structure for a service tier
type PricingTier struct {
	Name                ServiceTier   `json:"name"`
	PricePerDigit       float64       `json:"price_per_digit"`
	MaxDigits           int           `json:"max_digits"`
	MaxRuntimeMinutes   int           `json:"max_runtime_minutes"`
	CommunityFeePercent float64       `json:"community_fee_percent"`
	Features            []string      `json:"features"`
	Priority            int           `json:"priority"`
	Description         string        `json:"description"`
}

// PricingManager handles all pricing calculations
type PricingManager struct {
	tiers              map[ServiceTier]*PricingTier
	communityPoolAddr  string
	baseCurrency       string
}

// PriceBreakdown represents detailed cost breakdown
type PriceBreakdown struct {
	Tier         ServiceTier `json:"tier"`
	Digits       int         `json:"digits"`
	Method       string      `json:"method"`
	BaseCost     float64     `json:"base_cost"`
	ServiceFee   float64     `json:"service_fee"`
	CommunityFee float64     `json:"community_fee"`
	TotalCost    float64     `json:"total_cost"`
	Currency     string      `json:"currency"`
	Breakdown    string      `json:"breakdown"`
	Features     []string    `json:"features"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// NewPricingManager creates a new pricing manager
func NewPricingManager(communityPoolAddr string) *PricingManager {
	pm := &PricingManager{
		tiers:             make(map[ServiceTier]*PricingTier),
		communityPoolAddr: communityPoolAddr,
		baseCurrency:      "MEDAS",
	}
	
	// Initialize default tiers
	pm.initializeDefaultTiers()
	return pm
}

// initializeDefaultTiers sets up default pricing tiers
func (pm *PricingManager) initializeDefaultTiers() {
	pm.tiers[TierBasic] = &PricingTier{
		Name:                TierBasic,
		PricePerDigit:       0.0001, // 0.0001 MEDAS per digit
		MaxDigits:           1000,
		MaxRuntimeMinutes:   5,
		CommunityFeePercent: 0.15, // 15%
		Priority:            1,
		Description:         "Basic PI calculation for testing and learning",
		Features: []string{
			"Standard precision calculation",
			"Basic result verification",
			"Single algorithm (Chudnovsky)",
			"Up to 1,000 digits",
		},
	}
	
	pm.tiers[TierStandard] = &PricingTier{
		Name:                TierStandard,
		PricePerDigit:       0.00025,
		MaxDigits:           10000,
		MaxRuntimeMinutes:   30,
		CommunityFeePercent: 0.15,
		Priority:            2,
		Description:         "Standard service with progress monitoring",
		Features: []string{
			"Real-time progress updates",
			"Multiple algorithms available",
			"Advanced result verification",
			"Up to 10,000 digits",
			"Job status monitoring",
		},
	}
	
	pm.tiers[TierPremium] = &PricingTier{
		Name:                TierPremium,
		PricePerDigit:       0.0005,
		MaxDigits:           100000,
		MaxRuntimeMinutes:   120,
		CommunityFeePercent: 0.15,
		Priority:            3,
		Description:         "Premium service with highest priority and guarantees",
		Features: []string{
			"Highest priority processing",
			"All calculation algorithms",
			"Guaranteed completion",
			"Real-time monitoring dashboard",
			"Up to 100,000 digits",
			"Performance analytics",
			"Dedicated support",
		},
	}
}

// CalculatePrice calculates total price for a computation job
func (pm *PricingManager) CalculatePrice(digits int, tier ServiceTier, method string) (*PriceBreakdown, error) {
	tierConfig, exists := pm.tiers[tier]
	if !exists {
		return nil, fmt.Errorf("unknown tier: %s", tier)
	}
	
	if digits <= 0 {
		return nil, fmt.Errorf("digits must be positive")
	}
	
	if digits > tierConfig.MaxDigits {
		return nil, fmt.Errorf("digits (%d) exceed tier limit (%d)", digits, tierConfig.MaxDigits)
	}
	
	// Base cost calculation
	baseCost := float64(digits) * tierConfig.PricePerDigit
	
	// Apply method multiplier
	methodMultiplier := pm.getMethodMultiplier(method)
	baseCost *= methodMultiplier
	
	// Community fee
	communityFee := baseCost * tierConfig.CommunityFeePercent
	
	// Service provider fee (remainder)
	serviceFee := baseCost - communityFee
	
	// Total cost
	totalCost := baseCost
	
	// Estimate calculation time
	estimatedTime := EstimateCalculationTime(digits, method)
	
	breakdown := &PriceBreakdown{
		Tier:          tier,
		Digits:        digits,
		Method:        method,
		BaseCost:      baseCost,
		ServiceFee:    serviceFee,
		CommunityFee:  communityFee,
		TotalCost:     totalCost,
		Currency:      pm.baseCurrency,
		Features:      tierConfig.Features,
		EstimatedTime: estimatedTime,
		Breakdown: fmt.Sprintf(
			"%.6f %s (%.1f%% service provider + %.1f%% community pool)", 
			totalCost, 
			pm.baseCurrency,
			(1-tierConfig.CommunityFeePercent)*100, 
			tierConfig.CommunityFeePercent*100,
		),
	}
	
	return breakdown, nil
}

// getMethodMultiplier returns pricing multiplier based on calculation method
func (pm *PricingManager) getMethodMultiplier(method string) float64 {
	switch PIMethod(method) {
	case MethodChudnovsky:
		return 1.0 // Base price (fastest algorithm)
	case MethodMachin:
		return 1.3 // 30% more (slower convergence)
	case MethodBailey:
		return 1.2 // 20% more (moderate speed)
	default:
		return 1.0 // Default to base price
	}
}

// GetTier returns configuration for a specific tier
func (pm *PricingManager) GetTier(tier ServiceTier) (*PricingTier, error) {
	tierConfig, exists := pm.tiers[tier]
	if !exists {
		return nil, fmt.Errorf("tier not found: %s", tier)
	}
	return tierConfig, nil
}

// GetAllTiers returns all available pricing tiers
func (pm *PricingManager) GetAllTiers() map[ServiceTier]*PricingTier {
	// Return a copy to prevent external modification
	tiers := make(map[ServiceTier]*PricingTier)
	for k, v := range pm.tiers {
		tierCopy := *v
		tiers[k] = &tierCopy
	}
	return tiers
}

// ValidateTierLimits checks if requested parameters are within tier limits
func (pm *PricingManager) ValidateTierLimits(digits int, tier ServiceTier) error {
	tierConfig, exists := pm.tiers[tier]
	if !exists {
		return fmt.Errorf("invalid tier: %s", tier)
	}
	
	if digits <= 0 {
		return fmt.Errorf("digits must be positive")
	}
	
	if digits > tierConfig.MaxDigits {
		return fmt.Errorf("digits (%d) exceed tier limit (%d)", digits, tierConfig.MaxDigits)
	}
	
	return nil
}

// EstimateResourceUsage estimates computational resources needed
func (pm *PricingManager) EstimateResourceUsage(digits int, method string) ResourceEstimate {
	// Base resource calculation
	baseCPU := math.Log(float64(digits)) * 10 // Logarithmic scaling
	baseMemory := float64(digits) * 0.001     // Linear scaling (MB)
	
	// Method-specific adjustments
	var cpuMultiplier, memoryMultiplier float64
	switch PIMethod(method) {
	case MethodChudnovsky:
		cpuMultiplier = 1.0
		memoryMultiplier = 1.2 // Needs more memory for precision
	case MethodMachin:
		cpuMultiplier = 1.5 // More CPU intensive
		memoryMultiplier = 1.0
	case MethodBailey:
		cpuMultiplier = 1.3
		memoryMultiplier = 1.1
	default:
		cpuMultiplier = 1.0
		memoryMultiplier = 1.0
	}
	
	estimatedCPU := baseCPU * cpuMultiplier
	estimatedMemory := baseMemory * memoryMultiplier
	estimatedTime := EstimateCalculationTime(digits, method)
	
	// Apply reasonable bounds
	if estimatedCPU > 100 {
		estimatedCPU = 100
	}
	if estimatedMemory > 1024 { // 1GB max
		estimatedMemory = 1024
	}
	
	return ResourceEstimate{
		CPUPercent:    estimatedCPU,
		MemoryMB:      estimatedMemory,
		EstimatedTime: estimatedTime,
		Method:        method,
		Digits:        digits,
	}
}

// ResourceEstimate represents estimated resource usage
type ResourceEstimate struct {
	CPUPercent    float64       `json:"cpu_percent"`
	MemoryMB      float64       `json:"memory_mb"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Method        string        `json:"method"`
	Digits        int           `json:"digits"`
}

// PricingInfo provides comprehensive pricing information
type PricingInfo struct {
	Tiers             map[ServiceTier]*PricingTier `json:"tiers"`
	Currency          string                       `json:"currency"`
	CommunityPoolAddr string                       `json:"community_pool_address"`
	MethodMultipliers map[string]float64           `json:"method_multipliers"`
	LastUpdated       time.Time                    `json:"last_updated"`
}

// GetPricingInfo returns comprehensive pricing information
func (pm *PricingManager) GetPricingInfo() *PricingInfo {
	methodMultipliers := make(map[string]float64)
	for _, method := range GetAvailableMethods() {
		methodMultipliers[method] = pm.getMethodMultiplier(method)
	}
	
	return &PricingInfo{
		Tiers:             pm.GetAllTiers(),
		Currency:          pm.baseCurrency,
		CommunityPoolAddr: pm.communityPoolAddr,
		MethodMultipliers: methodMultipliers,
		LastUpdated:       time.Now(),
	}
}

// CalculateBulkPricing calculates pricing for multiple configurations
func (pm *PricingManager) CalculateBulkPricing(requests []PricingRequest) ([]PriceBreakdown, []error) {
	results := make([]PriceBreakdown, len(requests))
	errors := make([]error, len(requests))
	
	for i, req := range requests {
		breakdown, err := pm.CalculatePrice(req.Digits, req.Tier, req.Method)
		if err != nil {
			errors[i] = err
		} else {
			results[i] = *breakdown
		}
	}
	
	return results, errors
}

// PricingRequest represents a pricing calculation request
type PricingRequest struct {
	Digits int         `json:"digits"`
	Tier   ServiceTier `json:"tier"`
	Method string      `json:"method"`
}

// CompareServiceTiers compares all tiers for given parameters
func (pm *PricingManager) CompareServiceTiers(digits int, method string) ([]PriceBreakdown, error) {
	var comparisons []PriceBreakdown
	
	for _, tier := range []ServiceTier{TierBasic, TierStandard, TierPremium} {
		breakdown, err := pm.CalculatePrice(digits, tier, method)
		if err != nil {
			// Skip tiers that can't handle the request
			continue
		}
		comparisons = append(comparisons, *breakdown)
	}
	
	if len(comparisons) == 0 {
		return nil, fmt.Errorf("no tiers can handle request for %d digits", digits)
	}
	
	return comparisons, nil
}

// ValidatePaymentAmount checks if payment amount matches expected cost
func (pm *PricingManager) ValidatePaymentAmount(expectedCost, actualPayment float64, tolerancePercent float64) bool {
	if tolerancePercent <= 0 {
		tolerancePercent = 1.0 // Default 1% tolerance
	}
	
	tolerance := expectedCost * tolerancePercent / 100.0
	lowerBound := expectedCost - tolerance
	upperBound := expectedCost + tolerance
	
	return actualPayment >= lowerBound && actualPayment <= upperBound
}

// GetCommunityPoolAddress returns the community pool address
func (pm *PricingManager) GetCommunityPoolAddress() string {
	return pm.communityPoolAddr
}

// SetCommunityPoolAddress updates the community pool address
func (pm *PricingManager) SetCommunityPoolAddress(address string) {
	pm.communityPoolAddr = address
}

// GetTierForDigits suggests the best tier for given digit count
func (pm *PricingManager) GetTierForDigits(digits int) ServiceTier {
	if digits <= pm.tiers[TierBasic].MaxDigits {
		return TierBasic
	}
	if digits <= pm.tiers[TierStandard].MaxDigits {
		return TierStandard
	}
	return TierPremium
}

// CalculateDiscountedPrice applies discount for bulk calculations
func (pm *PricingManager) CalculateDiscountedPrice(breakdown *PriceBreakdown, discountPercent float64) *PriceBreakdown {
	if discountPercent <= 0 || discountPercent >= 100 {
		return breakdown
	}
	
	discountedBreakdown := *breakdown
	discount := breakdown.TotalCost * discountPercent / 100.0
	
	discountedBreakdown.TotalCost -= discount
	discountedBreakdown.ServiceFee -= discount * (1 - breakdown.CommunityFee/breakdown.BaseCost)
	discountedBreakdown.Breakdown = fmt.Sprintf(
		"%.6f %s (%.1f%% discount applied) - %s",
		discountedBreakdown.TotalCost,
		breakdown.Currency,
		discountPercent,
		breakdown.Breakdown,
	)
	
	return &discountedBreakdown
}

// ExamplePricingScenarios returns example pricing for documentation
func (pm *PricingManager) ExamplePricingScenarios() []PriceBreakdown {
	examples := []struct {
		digits int
		tier   ServiceTier
		method string
		desc   string
	}{
		{100, TierBasic, "chudnovsky", "Basic 100-digit calculation"},
		{1000, TierStandard, "chudnovsky", "Standard 1,000-digit calculation"},
		{10000, TierPremium, "chudnovsky", "Premium 10,000-digit calculation"},
		{500, TierStandard, "machin", "Historical Machin's formula"},
	}
	
	var scenarios []PriceBreakdown
	for _, example := range examples {
		breakdown, err := pm.CalculatePrice(example.digits, example.tier, example.method)
		if err == nil {
			breakdown.Breakdown = example.desc + " - " + breakdown.Breakdown
			scenarios = append(scenarios, *breakdown)
		}
	}
	
	return scenarios
}

package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/errors"
)

// Error definitions
var (
	ErrInvalidAnalysisResult = errors.Register("analysis", 1, "invalid analysis result")
	ErrInvalidAnalysisType   = errors.Register("analysis", 2, "invalid analysis type")
	ErrInvalidCapability     = errors.Register("analysis", 3, "invalid capability")
	ErrInvalidAddress        = errors.Register("analysis", 4, "invalid address")
	ErrInvalidMessage        = errors.Register("analysis", 5, "invalid message")
)

// AnalysisResult represents the result of an astronomical analysis
type AnalysisResult struct {
	AnalysisType string                 `json:"analysis_type"`
	Data         map[string]interface{} `json:"data"`
	Metadata     map[string]string      `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
	ClientID     string                 `json:"client_id"`
	BlockHeight  int64                  `json:"block_height"`
	TxHash       string                 `json:"tx_hash"`
}

// OrbitalDynamicsResult represents orbital dynamics analysis results
type OrbitalDynamicsResult struct {
	*AnalysisResult
	OrbitalElements []OrbitalElements   `json:"orbital_elements"`
	Predictions     []OrbitPrediction   `json:"predictions"`
	Targets         []ObservationTarget `json:"observation_targets"`
	Confidence      float64             `json:"confidence"`
	ModelVersion    string              `json:"model_version"`
}

// PhotometricResult represents photometric analysis results
type PhotometricResult struct {
	*AnalysisResult
	LightCurves     []LightCurve           `json:"light_curves"`
	Magnitudes      []float64              `json:"magnitudes"`
	Colors          map[string]float64     `json:"colors"`
	Variability     map[string]interface{} `json:"variability"`
	Classification  string                 `json:"classification"`
}

// ClusteringResult represents object clustering analysis results
type ClusteringResult struct {
	*AnalysisResult
	Clusters      []ObjectCluster        `json:"clusters"`
	Statistics    map[string]interface{} `json:"statistics"`
	Algorithm     string                 `json:"algorithm"`
	Parameters    map[string]float64     `json:"parameters"`
	QualityScore  float64                `json:"quality_score"`
}

// AITrainingResult represents AI training results with GPU statistics
type AITrainingResult struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	DeviceID        int                    `json:"device_id"`
	Epochs          int                    `json:"epochs"`
	BatchSize       int                    `json:"batch_size"`
	LearningRate    float64                `json:"learning_rate"`
	ModelType       string                 `json:"model_type"`
	DatasetSize     int                    `json:"dataset_size"`
	Progress        float64                `json:"progress"`
	Loss            float64                `json:"loss"`
	Accuracy        float64                `json:"accuracy"`
	GPUStats        GPUInfo                `json:"gpu_stats"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GPUDevice represents a single GPU device
type GPUDevice struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Memory             int64   `json:"memory"`              // Total memory in bytes
	MemoryUsed         int64   `json:"memory_used"`         // Used memory in bytes
	MemoryFree         int64   `json:"memory_free"`         // Free memory in bytes
	MemoryGB           float64 `json:"memory_gb"`           // Total memory in GB
	Temperature        float64 `json:"temperature"`         // Temperature in Celsius
	Utilization        float64 `json:"utilization"`         // GPU utilization percentage
	MemoryUtilization  float64 `json:"memory_utilization"`  // Memory utilization percentage
	PowerDraw          float64 `json:"power_draw"`          // Power draw in watts
	PowerUsage         float64 `json:"power_usage"`         // Power usage in watts (alias for PowerDraw)
	MaxPowerDraw       float64 `json:"max_power_draw"`      // Maximum power draw in watts
	ClockSpeed         int     `json:"clock_speed"`         // Core clock speed in MHz
	MemoryClockSpeed   int     `json:"memory_clock_speed"`  // Memory clock speed in MHz
	ComputeCapability  string  `json:"compute_capability"`  // CUDA compute capability
	IsAvailable        bool    `json:"is_available"`        // Whether the device is available
}

// SetMemoryFromGB sets both Memory (bytes) and MemoryGB from GB value
func (d *GPUDevice) SetMemoryFromGB(gb float64) {
	d.MemoryGB = gb
	d.Memory = int64(gb * 1024 * 1024 * 1024)
}

// SetMemoryFromBytes sets both Memory (bytes) and MemoryGB from bytes value
func (d *GPUDevice) SetMemoryFromBytes(bytes int64) {
	d.Memory = bytes
	d.MemoryGB = float64(bytes) / (1024 * 1024 * 1024)
}

// GetMemoryUsagePercent returns memory usage as percentage
func (d *GPUDevice) GetMemoryUsagePercent() float64 {
	if d.Memory == 0 {
		return 0.0
	}
	return (float64(d.MemoryUsed) / float64(d.Memory)) * 100.0
}

// SetPowerUsage sets both PowerDraw and PowerUsage to the same value
func (d *GPUDevice) SetPowerUsage(watts float64) {
	d.PowerDraw = watts
	d.PowerUsage = watts
}

// IsOverheating returns true if temperature is above 85°C
func (d *GPUDevice) IsOverheating() bool {
	return d.Temperature > 85.0
}

// IsHighUtilization returns true if GPU utilization is above 90%
func (d *GPUDevice) IsHighUtilization() bool {
	return d.Utilization > 90.0
}

// GPUInfo represents GPU information
type GPUInfo struct {
	DeviceCount         int         `json:"device_count"`
	Devices            []GPUDevice `json:"devices"`
	TotalMemoryGB      float64     `json:"total_memory_gb"`
	AvailableMemoryGB  float64     `json:"available_memory_gb"`
	CUDAVersion        string      `json:"cuda_version"`
	DriverVersion      string      `json:"driver_version"`
	IsInitialized      bool        `json:"is_initialized"`
	Timestamp          time.Time   `json:"timestamp"`
}

// UpdateTotalMemory calculates total and available memory from devices
func (gi *GPUInfo) UpdateTotalMemory() {
	gi.TotalMemoryGB = 0
	gi.AvailableMemoryGB = 0
	
	for _, device := range gi.Devices {
		gi.TotalMemoryGB += device.MemoryGB
		availableGB := float64(device.MemoryFree) / (1024 * 1024 * 1024)
		gi.AvailableMemoryGB += availableGB
	}
}

// GetAverageTemperature returns average temperature across all devices
func (gi *GPUInfo) GetAverageTemperature() float64 {
	if len(gi.Devices) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, device := range gi.Devices {
		total += device.Temperature
	}
	return total / float64(len(gi.Devices))
}

// GetAverageUtilization returns average utilization across all devices
func (gi *GPUInfo) GetAverageUtilization() float64 {
	if len(gi.Devices) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, device := range gi.Devices {
		total += device.Utilization
	}
	return total / float64(len(gi.Devices))
}

// GetAvailableDevices returns number of available devices
func (gi *GPUInfo) GetAvailableDevices() int {
	count := 0
	for _, device := range gi.Devices {
		if device.IsAvailable {
			count++
		}
	}
	return count
}

// OrbitalElements represents Keplerian orbital elements
type OrbitalElements struct {
	SemiMajorAxis      float64   `json:"semi_major_axis"`      // a (AU)
	Eccentricity       float64   `json:"eccentricity"`         // e
	Inclination        float64   `json:"inclination"`          // i (degrees)
	LongitudeAscending float64   `json:"longitude_ascending"`  // Ω (degrees)
	ArgumentPeriapsis  float64   `json:"argument_periapsis"`   // ω (degrees)
	MeanAnomaly        float64   `json:"mean_anomaly"`         // M (degrees)
	Epoch              time.Time `json:"epoch"`                // Reference epoch
	Period             float64   `json:"period"`               // Orbital period (years)
	Uncertainty        map[string]float64 `json:"uncertainty"` // Parameter uncertainties
}

// OrbitPrediction represents predicted positions for an object
type OrbitPrediction struct {
	Time        time.Time `json:"time"`
	RA          float64   `json:"ra"`           // Right Ascension (degrees)
	Dec         float64   `json:"dec"`          // Declination (degrees)
	Distance    float64   `json:"distance"`     // Distance from Earth (AU)
	Magnitude   float64   `json:"magnitude"`    // Predicted magnitude
	Phase       float64   `json:"phase"`        // Phase angle (degrees)
	Uncertainty float64   `json:"uncertainty"`  // Position uncertainty (arcseconds)
}

// ObservationTarget represents recommended observation parameters
type ObservationTarget struct {
	TargetID       string    `json:"target_id"`
	Name           string    `json:"name"`
	RA             float64   `json:"ra"`
	Dec            float64   `json:"dec"`
	Magnitude      float64   `json:"magnitude"`
	OptimalTime    time.Time `json:"optimal_time"`
	Duration       int       `json:"duration"`        // Recommended observation duration (seconds)
	Filter         string    `json:"filter"`          // Recommended filter
	Priority       int       `json:"priority"`        // Priority (1-10, 10 highest)
	Observability  float64   `json:"observability"`   // Observability score (0-1)
	Notes          string    `json:"notes"`
}

// LightCurve represents photometric time series data
type LightCurve struct {
	ObjectID      string    `json:"object_id"`
	Filter        string    `json:"filter"`           // Photometric filter
	Times         []float64 `json:"times"`            // Time series (JD)
	Magnitudes    []float64 `json:"magnitudes"`       // Magnitude measurements
	Errors        []float64 `json:"errors"`           // Measurement errors
	Flags         []int     `json:"flags"`            // Quality flags
	Period        float64   `json:"period"`           // Detected period (days)
	Amplitude     float64   `json:"amplitude"`        // Variability amplitude
	PhaseZero     float64   `json:"phase_zero"`       // Phase zero point
	Classification string   `json:"classification"`    // Variable star type
	Confidence    float64   `json:"confidence"`       // Classification confidence
}

// ObjectCluster represents a cluster of astronomical objects
type ObjectCluster struct {
	ClusterID      int                    `json:"cluster_id"`
	CenterRA       float64                `json:"center_ra"`       // Cluster center RA
	CenterDec      float64                `json:"center_dec"`      // Cluster center Dec
	CenterDistance float64                `json:"center_distance"` // Cluster center distance
	Radius         float64                `json:"radius"`          // Cluster radius (arcmin)
	MemberCount    int                    `json:"member_count"`    // Number of member objects
	Members        []string               `json:"members"`         // Object IDs in cluster
	Properties     map[string]interface{} `json:"properties"`      // Cluster properties
	Confidence     float64                `json:"confidence"`      // Cluster confidence score
}

// RegisteredClient represents a registered analysis client
type RegisteredClient struct {
	ID           string    `json:"id"`
	Creator      string    `json:"creator"`
	Capabilities []string  `json:"capabilities"`
	Metadata     string    `json:"metadata"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StoredAnalysis represents stored analysis data
type StoredAnalysis struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	Creator      string    `json:"creator"`
	AnalysisType string    `json:"analysis_type"`
	Data         string    `json:"data"`
	TxHash       string    `json:"tx_hash"`
	BlockHeight  int64     `json:"block_height"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validation functions
func IsValidAnalysisType(analysisType string) bool {
	validTypes := []string{
		"orbital_dynamics",
		"photometric",
		"clustering",
		"ai_training",
		"planet9_search",
		"asteroid_tracking",
		"variable_star_analysis",
		"exoplanet_detection",
	}
	
	for _, valid := range validTypes {
		if analysisType == valid {
			return true
		}
	}
	return false
}

func IsValidCapability(capability string) bool {
	validCapabilities := []string{
		"orbital_dynamics",
		"photometric_analysis",
		"object_clustering",
		"ai_training",
		"gpu_acceleration",
		"planet9_search",
		"asteroid_tracking",
		"variable_star_detection",
		"exoplanet_analysis",
		"high_precision_astrometry",
		"multi_wavelength_analysis",
	}
	
	for _, valid := range validCapabilities {
		if capability == valid {
			return true
		}
	}
	return false
}

func IsValidStatus(status string) bool {
	validStatuses := []string{
		"active",
		"inactive",
		"suspended",
		"pending",
		"completed",
		"failed",
		"processing",
	}
	
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// JSON marshaling methods
func (ar *AnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(ar)
}

func (ar *AnalysisResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, ar)
}

func (ar *AnalysisResult) GetMetadata(key string) (string, bool) {
	value, exists := ar.Metadata[key]
	return value, exists
}

func (ar *AnalysisResult) SetMetadata(key, value string) {
	if ar.Metadata == nil {
		ar.Metadata = make(map[string]string)
	}
	ar.Metadata[key] = value
}

func (ar *AnalysisResult) GetDataField(key string) (interface{}, bool) {
	value, exists := ar.Data[key]
	return value, exists
}

func (ar *AnalysisResult) SetDataField(key string, value interface{}) {
	if ar.Data == nil {
		ar.Data = make(map[string]interface{})
	}
	ar.Data[key] = value
}

// String representations
func (ar *AnalysisResult) String() string {
	return fmt.Sprintf("AnalysisResult{Type: %s, ClientID: %s, BlockHeight: %d, TxHash: %s}",
		ar.AnalysisType, ar.ClientID, ar.BlockHeight, ar.TxHash)
}

func (od *OrbitalDynamicsResult) String() string {
	return fmt.Sprintf("OrbitalDynamicsResult{Elements: %d, Predictions: %d, Confidence: %.2f}",
		len(od.OrbitalElements), len(od.Predictions), od.Confidence)
}

func (pr *PhotometricResult) String() string {
	return fmt.Sprintf("PhotometricResult{LightCurves: %d, Classification: %s}",
		len(pr.LightCurves), pr.Classification)
}

func (cr *ClusteringResult) String() string {
	return fmt.Sprintf("ClusteringResult{Clusters: %d, Algorithm: %s, QualityScore: %.2f}",
		len(cr.Clusters), cr.Algorithm, cr.QualityScore)
}

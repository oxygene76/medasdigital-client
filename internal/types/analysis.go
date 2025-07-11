package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// AnalysisResult represents the result of an astronomical analysis
type AnalysisResult struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Type        string                 `json:"type"`
	Results     map[string]interface{} `json:"results"`
	BlockHeight int64                  `json:"block_height,omitempty"`
	TxHash      string                 `json:"tx_hash,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	Confidence  float64                `json:"confidence,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OrbitalDynamicsResult represents results from orbital dynamics analysis
type OrbitalDynamicsResult struct {
	*AnalysisResult
	Elements          OrbitalElements         `json:"orbital_elements"`
	Predictions       []OrbitPrediction       `json:"predictions"`
	Planet9Probability float64                `json:"planet9_probability"`
	Recommendations   []ObservationTarget     `json:"recommendations"`
}

// OrbitalElements represents orbital elements of a celestial object
type OrbitalElements struct {
	SemiMajorAxis     float64 `json:"semi_major_axis"`     // a (AU)
	Eccentricity      float64 `json:"eccentricity"`        // e
	Inclination       float64 `json:"inclination"`         // i (degrees)
	LongitudeOfNode   float64 `json:"longitude_of_node"`   // Ω (degrees)
	ArgumentOfPeriaps float64 `json:"argument_of_periaps"` // ω (degrees)
	MeanAnomaly       float64 `json:"mean_anomaly"`        // M (degrees)
	Epoch             float64 `json:"epoch"`               // Julian Date
	Period            float64 `json:"period"`              // years
}

// OrbitPrediction represents predicted positions
type OrbitPrediction struct {
	Time     time.Time `json:"time"`
	RA       float64   `json:"ra"`       // Right Ascension (degrees)
	Dec      float64   `json:"dec"`      // Declination (degrees)
	Distance float64   `json:"distance"` // Distance from Sun (AU)
	Magnitude float64  `json:"magnitude"` // Apparent magnitude
}

// ObservationTarget represents recommended observation targets
type ObservationTarget struct {
	Name         string    `json:"name"`
	RA           float64   `json:"ra"`
	Dec          float64   `json:"dec"`
	OptimalTime  time.Time `json:"optimal_time"`
	Priority     int       `json:"priority"`
	Telescope    string    `json:"telescope"`
	ExpectedMag  float64   `json:"expected_magnitude"`
	SearchRadius float64   `json:"search_radius"` // arcminutes
}

// PhotometricResult represents results from photometric analysis
type PhotometricResult struct {
	*AnalysisResult
	Objects         []PhotometricObject `json:"objects"`
	LightCurves     []LightCurve        `json:"light_curves"`
	VariabilityType string              `json:"variability_type"`
	Period          float64             `json:"period,omitempty"` // days
}

// PhotometricObject represents a photometric object
type PhotometricObject struct {
	ID          string  `json:"id"`
	RA          float64 `json:"ra"`
	Dec         float64 `json:"dec"`
	Magnitude   float64 `json:"magnitude"`
	Color       float64 `json:"color"`
	Variability bool    `json:"variability"`
	Class       string  `json:"classification"`
}

// LightCurve represents photometric time series data
type LightCurve struct {
	ObjectID    string                `json:"object_id"`
	Filter      string                `json:"filter"`
	Observations []PhotometricPoint   `json:"observations"`
	Statistics  LightCurveStatistics `json:"statistics"`
}

// PhotometricPoint represents a single photometric measurement
type PhotometricPoint struct {
	Time      float64 `json:"time"`      // Julian Date
	Magnitude float64 `json:"magnitude"`
	Error     float64 `json:"error"`
	Flag      string  `json:"flag,omitempty"`
}

// LightCurveStatistics contains statistical analysis of light curves
type LightCurveStatistics struct {
	Mean       float64 `json:"mean"`
	StdDev     float64 `json:"std_dev"`
	Amplitude  float64 `json:"amplitude"`
	Period     float64 `json:"period,omitempty"`
	PeriodError float64 `json:"period_error,omitempty"`
}

// ClusteringResult represents results from clustering analysis
type ClusteringResult struct {
	*AnalysisResult
	Clusters      []ObjectCluster `json:"clusters"`
	NoiseObjects  []string        `json:"noise_objects"`
	ClusterMethod string          `json:"cluster_method"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// ObjectCluster represents a cluster of celestial objects
type ObjectCluster struct {
	ID      int      `json:"id"`
	Members []string `json:"members"`
	Center  Point3D  `json:"center"`
	Radius  float64  `json:"radius"`
	Density float64  `json:"density"`
}

// Point3D represents a 3D coordinate
type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// AITrainingResult represents results from AI training
type AITrainingResult struct {
	*AnalysisResult
	ModelType     string                 `json:"model_type"`
	Architecture  string                 `json:"architecture"`
	TrainingStats TrainingStatistics     `json:"training_stats"`
	Validation    ValidationMetrics      `json:"validation"`
	ModelPath     string                 `json:"model_path"`
	GPUUsed       bool                   `json:"gpu_used"`
	Performance   map[string]interface{} `json:"performance"`
}

// TrainingStatistics contains training process statistics
type TrainingStatistics struct {
	Epochs         int     `json:"epochs"`
	BatchSize      int     `json:"batch_size"`
	LearningRate   float64 `json:"learning_rate"`
	TrainLoss      float64 `json:"train_loss"`
	ValidationLoss float64 `json:"validation_loss"`
	TrainingTime   float64 `json:"training_time"` // seconds
	GPUMemoryUsed  int64   `json:"gpu_memory_used,omitempty"` // bytes
}

// ValidationMetrics contains model validation metrics
type ValidationMetrics struct {
	Accuracy    float64 `json:"accuracy"`
	Precision   float64 `json:"precision"`
	Recall      float64 `json:"recall"`
	F1Score     float64 `json:"f1_score"`
	AUC         float64 `json:"auc,omitempty"`
	ConfusionMatrix [][]int `json:"confusion_matrix,omitempty"`
}

// GPUInfo represents GPU information and capabilities
type GPUInfo struct {
	Name            string  `json:"name"`
	Memory          int64   `json:"memory"`          // bytes
	MemoryUsed      int64   `json:"memory_used"`     // bytes
	Temperature     float64 `json:"temperature"`     // Celsius
	Utilization     float64 `json:"utilization"`     // percentage
	CUDAVersion     string  `json:"cuda_version"`
	DriverVersion   string  `json:"driver_version"`
	ComputeCapability string `json:"compute_capability"`
}

// Capabilities represents client capabilities
const (
	CapabilityOrbitalDynamics     = "orbital_dynamics"
	CapabilityPhotometricAnalysis = "photometric_analysis"
	CapabilityAstrometricValidation = "astrometric_validation"
	CapabilityClusteringAnalysis  = "clustering_analysis"
	CapabilitySurveyProcessing    = "survey_processing"
	CapabilityAnomalyDetection    = "anomaly_detection"
	CapabilityAITraining         = "ai_training"
	CapabilityGPUCompute         = "gpu_compute"
)

// Analysis types
const (
	AnalysisTypeOrbitalDynamics  = "orbital_dynamics"
	AnalysisTypePhotometric      = "photometric"
	AnalysisTypeClustering       = "clustering"
	AnalysisTypeAITraining       = "ai_training"
	AnalysisTypeAnomalyDetection = "anomaly_detection"
	AnalysisTypeSurveyProcessing = "survey_processing"
)

// Status constants
const (
	StatusPending    = "pending"
	StatusRunning    = "running"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
)

// ValidCapabilities returns a list of valid client capabilities
func ValidCapabilities() []string {
	return []string{
		CapabilityOrbitalDynamics,
		CapabilityPhotometricAnalysis,
		CapabilityAstrometricValidation,
		CapabilityClusteringAnalysis,
		CapabilitySurveyProcessing,
		CapabilityAnomalyDetection,
		CapabilityAITraining,
		CapabilityGPUCompute,
	}
}

// IsValidCapability checks if a capability is valid
func IsValidCapability(capability string) bool {
	for _, valid := range ValidCapabilities() {
		if capability == valid {
			return true
		}
	}
	return false
}

// ValidAnalysisTypes returns a list of valid analysis types
func ValidAnalysisTypes() []string {
	return []string{
		AnalysisTypeOrbitalDynamics,
		AnalysisTypePhotometric,
		AnalysisTypeClustering,
		AnalysisTypeAITraining,
		AnalysisTypeAnomalyDetection,
		AnalysisTypeSurveyProcessing,
	}
}

// IsValidAnalysisType checks if an analysis type is valid
func IsValidAnalysisType(analysisType string) bool {
	for _, valid := range ValidAnalysisTypes() {
		if analysisType == valid {
			return true
		}
	}
	return false
}

// NewAnalysisResult creates a new analysis result
func NewAnalysisResult(clientID, analysisType string) *AnalysisResult {
	return &AnalysisResult{
		ID:        generateResultID(),
		ClientID:  clientID,
		Type:      analysisType,
		Results:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Status:    StatusPending,
		Metadata:  make(map[string]interface{}),
	}
}

// Validate validates an analysis result
func (ar *AnalysisResult) Validate() error {
	if ar.ClientID == "" {
		return errors.Wrap(ErrInvalidAnalysisResult, "client_id cannot be empty")
	}
	
	if ar.Type == "" {
		return errors.Wrap(ErrInvalidAnalysisResult, "type cannot be empty")
	}
	
	if !IsValidAnalysisType(ar.Type) {
		return errors.Wrapf(ErrInvalidAnalysisResult, "invalid analysis type: %s", ar.Type)
	}
	
	if ar.Status == "" {
		return errors.Wrap(ErrInvalidAnalysisResult, "status cannot be empty")
	}
	
	validStatuses := []string{StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled}
	isValidStatus := false
	for _, valid := range validStatuses {
		if ar.Status == valid {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return errors.Wrapf(ErrInvalidAnalysisResult, "invalid status: %s", ar.Status)
	}
	
	return nil
}

// SetResults sets the analysis results
func (ar *AnalysisResult) SetResults(results map[string]interface{}) {
	ar.Results = results
}

// AddMetadata adds metadata to the analysis result
func (ar *AnalysisResult) AddMetadata(key string, value interface{}) {
	if ar.Metadata == nil {
		ar.Metadata = make(map[string]interface{})
	}
	ar.Metadata[key] = value
}

// SetStatus sets the analysis status
func (ar *AnalysisResult) SetStatus(status string) error {
	validStatuses := []string{StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled}
	for _, valid := range validStatuses {
		if status == valid {
			ar.Status = status
			return nil
		}
	}
	return errors.Wrapf(ErrInvalidAnalysisResult, "invalid status: %s", status)
}

// ToJSON converts the analysis result to JSON
func (ar *AnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(ar)
}

// FromJSON creates an analysis result from JSON
func FromJSON(data []byte) (*AnalysisResult, error) {
	var result AnalysisResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrap(ErrInvalidAnalysisResult, err.Error())
	}
	return &result, nil
}

// Error definitions compatible with Cosmos SDK v0.50
var (
	ErrInvalidAnalysisResult = errors.Register("analysis", 1, "invalid analysis result")
	ErrInvalidCapability     = errors.Register("analysis", 2, "invalid capability")
	ErrInvalidAnalysisType   = errors.Register("analysis", 3, "invalid analysis type")
	ErrAnalysisNotFound      = errors.Register("analysis", 4, "analysis not found")
	ErrInsufficientData      = errors.Register("analysis", 5, "insufficient data for analysis")
	ErrGPUNotAvailable       = errors.Register("analysis", 6, "GPU not available")
	ErrAnalysisTimeout       = errors.Register("analysis", 7, "analysis timeout")
)

// generateResultID generates a unique result ID
func generateResultID() string {
	return fmt.Sprintf("result_%d_%s", 
		time.Now().Unix(), 
		strings.ToLower(fmt.Sprintf("%x", time.Now().UnixNano())[:8]))
}

// Interface implementations for Cosmos SDK v0.50 compatibility
var _ sdk.Msg = (*AnalysisResult)(nil)

// Route implements sdk.Msg interface (legacy)
func (ar *AnalysisResult) Route() string {
	return "analysis"
}

// Type implements sdk.Msg interface (legacy)
func (ar *AnalysisResult) Type() string {
	return "store_analysis_result"
}

// GetSigners implements sdk.Msg interface
func (ar *AnalysisResult) GetSigners() []sdk.AccAddress {
	// For analysis results, no specific signers required
	return []sdk.AccAddress{}
}

// GetSignBytes implements sdk.Msg interface (legacy)
func (ar *AnalysisResult) GetSignBytes() []byte {
	bz, err := json.Marshal(ar)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg interface
func (ar *AnalysisResult) ValidateBasic() error {
	return ar.Validate()
}

// GetSignersStr returns signers as strings (v0.50 requirement)
func (ar *AnalysisResult) GetSignersStr() ([]string, error) {
	// For analysis results, return empty slice
	return []string{}, nil
}

// RegisterInterfaces registers interfaces for protobuf
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&AnalysisResult{},
	)
}

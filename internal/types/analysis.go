package types

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AnalysisResult represents the result of an analysis operation
type AnalysisResult struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	BlockHeight int64                  `json:"block_height,omitempty"`
	TxHash      string                 `json:"tx_hash,omitempty"`
	Results     map[string]interface{} `json:"results"`
	Metadata    AnalysisMetadata       `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
}

// AnalysisMetadata contains metadata about the analysis
type AnalysisMetadata struct {
	InputFiles    []string               `json:"input_files"`
	OutputFiles   []string               `json:"output_files"`
	Parameters    map[string]interface{} `json:"parameters"`
	GPUUsed       bool                   `json:"gpu_used"`
	GPUDevices    []int                  `json:"gpu_devices,omitempty"`
	CPUCores      int                    `json:"cpu_cores"`
	MemoryUsedMB  int64                  `json:"memory_used_mb"`
	Version       string                 `json:"version"`
}

// OrbitalDynamicsResult represents orbital dynamics analysis results
type OrbitalDynamicsResult struct {
	Objects              []TNOObject       `json:"objects"`
	Planet9Probability   float64          `json:"planet9_probability"`
	ClusteringSignificance float64        `json:"clustering_significance"`
	GravitationalEffects []GravEffect     `json:"gravitational_effects"`
	Recommendations      []Recommendation `json:"recommendations"`
}

// TNOObject represents a Trans-Neptunian Object
type TNOObject struct {
	Designation      string  `json:"designation"`
	SemimajorAxis    float64 `json:"semimajor_axis"`    // AU
	Eccentricity     float64 `json:"eccentricity"`
	Inclination      float64 `json:"inclination"`       // degrees
	LongitudeNode    float64 `json:"longitude_node"`    // degrees
	ArgumentPeriapsis float64 `json:"argument_periapsis"` // degrees
	MeanAnomaly      float64 `json:"mean_anomaly"`      // degrees
	Epoch            float64 `json:"epoch"`             // MJD
	AbsoluteMagnitude float64 `json:"absolute_magnitude"`
	AlbedoEstimate   float64 `json:"albedo_estimate"`
	DiameterKM       float64 `json:"diameter_km"`
}

// GravEffect represents gravitational effects on objects
type GravEffect struct {
	ObjectID          string  `json:"object_id"`
	DeltaSemimajor    float64 `json:"delta_semimajor"`    // AU change
	DeltaEccentricity float64 `json:"delta_eccentricity"` // change
	DeltaInclination  float64 `json:"delta_inclination"`  // degrees change
	Significance      float64 `json:"significance"`       // sigma
}

// PhotometricResult represents photometric analysis results
type PhotometricResult struct {
	Objects           []PhotometricObject `json:"objects"`
	VariabilityAnalysis []VariabilityData `json:"variability_analysis"`
	ColorIndices      []ColorData         `json:"color_indices"`
	LightCurves       []LightCurve        `json:"light_curves"`
	RotationPeriods   []RotationData      `json:"rotation_periods"`
}

// PhotometricObject represents photometric measurements of an object
type PhotometricObject struct {
	ObjectID     string              `json:"object_id"`
	Observations []PhotometricObs    `json:"observations"`
	MeanMagnitude map[string]float64 `json:"mean_magnitude"` // filter -> magnitude
	Variability  float64             `json:"variability"`
	Classification string            `json:"classification"`
}

// PhotometricObs represents a single photometric observation
type PhotometricObs struct {
	MJD         float64 `json:"mjd"`
	Magnitude   float64 `json:"magnitude"`
	Uncertainty float64 `json:"uncertainty"`
	Filter      string  `json:"filter"`
	Observatory string  `json:"observatory"`
	Airmass     float64 `json:"airmass"`
}

// VariabilityData represents variability analysis results
type VariabilityData struct {
	ObjectID      string  `json:"object_id"`
	Amplitude     float64 `json:"amplitude"`
	Period        float64 `json:"period"`        // hours
	Confidence    float64 `json:"confidence"`    // 0-1
	VariabilityType string `json:"variability_type"`
}

// ColorData represents color index measurements
type ColorData struct {
	ObjectID   string             `json:"object_id"`
	ColorIndex map[string]float64 `json:"color_index"` // e.g., "g-r", "r-i"
	Uncertainty map[string]float64 `json:"uncertainty"`
	Taxonomy   string             `json:"taxonomy"`
}

// LightCurve represents a light curve
type LightCurve struct {
	ObjectID     string           `json:"object_id"`
	Observations []PhotometricObs `json:"observations"`
	Period       float64          `json:"period"`       // hours
	Amplitude    float64          `json:"amplitude"`    // magnitude
	Quality      float64          `json:"quality"`      // 0-1
}

// RotationData represents rotation period analysis
type RotationData struct {
	ObjectID     string  `json:"object_id"`
	Period       float64 `json:"period"`       // hours
	Uncertainty  float64 `json:"uncertainty"`  // hours
	Amplitude    float64 `json:"amplitude"`    // magnitude
	Method       string  `json:"method"`
	Confidence   float64 `json:"confidence"`   // 0-1
}

// AIDetectionResult represents AI detection results
type AIDetectionResult struct {
	Detections    []Detection     `json:"detections"`
	ModelInfo     ModelInfo       `json:"model_info"`
	ProcessingInfo ProcessingInfo  `json:"processing_info"`
	Statistics    DetectionStats  `json:"statistics"`
}

// Detection represents a single object detection
type Detection struct {
	ID            string     `json:"id"`
	RA            float64    `json:"ra"`             // degrees
	Dec           float64    `json:"dec"`            // degrees
	Magnitude     float64    `json:"magnitude"`
	Confidence    float64    `json:"confidence"`     // 0-1
	BoundingBox   BoundingBox `json:"bounding_box"`
	Classification string     `json:"classification"`
	Motion        MotionData `json:"motion"`
	ImageInfo     ImageInfo  `json:"image_info"`
}

// BoundingBox represents object bounding box in image
type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// MotionData represents object motion information
type MotionData struct {
	RA_Rate   float64 `json:"ra_rate"`    // arcsec/hour
	Dec_Rate  float64 `json:"dec_rate"`   // arcsec/hour
	Distance  float64 `json:"distance"`   // AU (estimated)
	Velocity  float64 `json:"velocity"`   // km/s
}

// ImageInfo represents information about the source image
type ImageInfo struct {
	Filename    string    `json:"filename"`
	Exposure    float64   `json:"exposure"`    // seconds
	Filter      string    `json:"filter"`
	Observatory string    `json:"observatory"`
	Timestamp   time.Time `json:"timestamp"`
	Seeing      float64   `json:"seeing"`      // arcsec
}

// ModelInfo represents AI model information
type ModelInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	TrainingDate time.Time `json:"training_date"`
	Accuracy     float64 `json:"accuracy"`
}

// ProcessingInfo represents processing information
type ProcessingInfo struct {
	GPUUsed       bool          `json:"gpu_used"`
	GPUDevices    []int         `json:"gpu_devices"`
	ProcessingTime time.Duration `json:"processing_time"`
	ImagesProcessed int          `json:"images_processed"`
	BatchSize     int           `json:"batch_size"`
}

// DetectionStats represents detection statistics
type DetectionStats struct {
	TotalDetections   int     `json:"total_detections"`
	HighConfidence    int     `json:"high_confidence"`    // > 0.8
	MediumConfidence  int     `json:"medium_confidence"`  // 0.5-0.8
	LowConfidence     int     `json:"low_confidence"`     // < 0.5
	MovingObjects     int     `json:"moving_objects"`
	StationaryObjects int     `json:"stationary_objects"`
	MeanConfidence    float64 `json:"mean_confidence"`
}

// ClusteringResult represents clustering analysis results
type ClusteringResult struct {
	Clusters       []Cluster       `json:"clusters"`
	Statistics     ClusterStats    `json:"statistics"`
	Significance   float64         `json:"significance"`
	Method         string          `json:"method"`
	Parameters     map[string]interface{} `json:"parameters"`
}

// Cluster represents a cluster of objects
type Cluster struct {
	ID          int        `json:"id"`
	Objects     []string   `json:"objects"`     // object IDs
	Centroid    Centroid   `json:"centroid"`
	Radius      float64    `json:"radius"`
	Coherence   float64    `json:"coherence"`   // 0-1
	Significance float64   `json:"significance"` // sigma
}

// Centroid represents cluster centroid in orbital element space
type Centroid struct {
	SemimajorAxis     float64 `json:"semimajor_axis"`
	Eccentricity      float64 `json:"eccentricity"`
	Inclination       float64 `json:"inclination"`
	LongitudeNode     float64 `json:"longitude_node"`
	ArgumentPeriapsis float64 `json:"argument_periapsis"`
}

// ClusterStats represents clustering statistics
type ClusterStats struct {
	NumClusters      int     `json:"num_clusters"`
	SilhouetteScore  float64 `json:"silhouette_score"`
	CalinskiHarabasz float64 `json:"calinski_harabasz"`
	DaviesBouldin    float64 `json:"davies_bouldin"`
	Inertia          float64 `json:"inertia"`
}

// Recommendation represents an observation recommendation
type Recommendation struct {
	Priority     string  `json:"priority"`       // "high", "medium", "low"
	RA           float64 `json:"ra"`             // degrees
	Dec          float64 `json:"dec"`            // degrees
	MagnitudeEst float64 `json:"magnitude_est"`
	Urgency      string  `json:"urgency"`        // "immediate", "next_month", "opportunity"
	Reason       string  `json:"reason"`
	ValidFrom    time.Time `json:"valid_from"`
	ValidUntil   time.Time `json:"valid_until"`
}

// Blockchain message types

// MsgRegisterClient represents a client registration message
type MsgRegisterClient struct {
	Creator      string   `json:"creator"`
	Capabilities []string `json:"capabilities"`
	Metadata     string   `json:"metadata"`
}

// Route returns the message route
func (msg MsgRegisterClient) Route() string { return "clientregistry" }

// Type returns the message type
func (msg MsgRegisterClient) Type() string { return "register_client" }

// ValidateBasic validates the message
func (msg MsgRegisterClient) ValidateBasic() error {
	if len(msg.Creator) == 0 {
		return fmt.Errorf("creator cannot be empty")
	}
	if len(msg.Capabilities) == 0 {
		return fmt.Errorf("capabilities cannot be empty")
	}
	return nil
}

// GetSignBytes returns the message bytes for signing
func (msg MsgRegisterClient) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return bz
}

// GetSigners returns the signers
func (msg MsgRegisterClient) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgStoreAnalysis represents an analysis storage message
type MsgStoreAnalysis struct {
	Creator      string                 `json:"creator"`
	ClientID     string                 `json:"client_id"`
	AnalysisType string                 `json:"analysis_type"`
	Data         map[string]interface{} `json:"data"`
	BlockHeight  int64                  `json:"block_height,omitempty"`
	TxHash       string                 `json:"tx_hash,omitempty"`
}

// Route returns the message route
func (msg MsgStoreAnalysis) Route() string { return "clientregistry" }

// Type returns the message type
func (msg MsgStoreAnalysis) Type() string { return "store_analysis" }

// ValidateBasic validates the message
func (msg MsgStoreAnalysis) ValidateBasic() error {
	if len(msg.Creator) == 0 {
		return fmt.Errorf("creator cannot be empty")
	}
	if len(msg.ClientID) == 0 {
		return fmt.Errorf("client_id cannot be empty")
	}
	if len(msg.AnalysisType) == 0 {
		return fmt.Errorf("analysis_type cannot be empty")
	}
	return nil
}

// GetSignBytes returns the message bytes for signing
func (msg MsgStoreAnalysis) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return bz
}

// GetSigners returns the signers
func (msg MsgStoreAnalysis) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// RegisteredClient represents a registered client
type RegisteredClient struct {
	ID           string    `json:"id"`
	Creator      string    `json:"creator"`
	Capabilities []string  `json:"capabilities"`
	Metadata     string    `json:"metadata"`
	RegisteredAt time.Time `json:"registered_at"`
	Status       string    `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
}

// StoredAnalysis represents stored analysis data
type StoredAnalysis struct {
	ID           string                 `json:"id"`
	ClientID     string                 `json:"client_id"`
	Creator      string                 `json:"creator"`
	AnalysisType string                 `json:"analysis_type"`
	Data         map[string]interface{} `json:"data"`
	BlockHeight  int64                  `json:"block_height"`
	TxHash       string                 `json:"tx_hash"`
	Timestamp    time.Time              `json:"timestamp"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	DeviceCount       int              `json:"device_count"`
	Devices          []GPUDevice      `json:"devices"`
	CUDAVersion      string           `json:"cuda_version"`
	DriverVersion    string           `json:"driver_version"`
	TotalMemoryGB    float64          `json:"total_memory_gb"`
	AvailableMemoryGB float64         `json:"available_memory_gb"`
}

// GPUDevice represents a single GPU device
type GPUDevice struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	ComputeCapability string `json:"compute_capability"`
	MemoryGB         float64 `json:"memory_gb"`
	Temperature      int     `json:"temperature"`
	Utilization      int     `json:"utilization"`
	PowerUsage       int     `json:"power_usage"`
}

// TrainingResult represents training results
type TrainingResult struct {
	ModelPath     string            `json:"model_path"`
	Accuracy      float64           `json:"accuracy"`
	Loss          float64           `json:"loss"`
	Epochs        int               `json:"epochs"`
	TrainingTime  time.Duration     `json:"training_time"`
	ValidationSet ValidationMetrics `json:"validation_set"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
}

// ValidationMetrics represents validation metrics
type ValidationMetrics struct {
	Accuracy    float64 `json:"accuracy"`
	Precision   float64 `json:"precision"`
	Recall      float64 `json:"recall"`
	F1Score     float64 `json:"f1_score"`
	ConfusionMatrix [][]int `json:"confusion_matrix"`
}

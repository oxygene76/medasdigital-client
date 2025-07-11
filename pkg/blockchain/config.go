package utils

import (
	"time"
)

// GPUConfig configuration for GPU management
type GPUConfig struct {
	Enabled         bool      `yaml:"enabled" json:"enabled"`
	DeviceCount     int       `yaml:"device_count" json:"device_count"`
	DeviceID        int       `yaml:"device_id" json:"device_id"`  // Primary device ID
	CUDADevices     []int     `yaml:"cuda_devices" json:"cuda_devices"`
	MaxMemoryGB     float64   `yaml:"max_memory_gb" json:"max_memory_gb"`
	BenchmarkOnInit bool      `yaml:"benchmark_on_init" json:"benchmark_on_init"`
	PowerLimit      float64   `yaml:"power_limit" json:"power_limit"`
	TempLimit       float64   `yaml:"temp_limit" json:"temp_limit"`
	UseAllDevices   bool      `yaml:"use_all_devices" json:"use_all_devices"`
}

// GetPrimaryDeviceID returns the primary device ID
func (g *GPUConfig) GetPrimaryDeviceID() int {
	if len(g.CUDADevices) > 0 {
		return g.CUDADevices[0]
	}
	return g.DeviceID
}

// GetDeviceIDs returns all configured device IDs
func (g *GPUConfig) GetDeviceIDs() []int {
	if len(g.CUDADevices) > 0 {
		return g.CUDADevices
	}
	if g.UseAllDevices {
		// Return all available devices up to DeviceCount
		devices := make([]int, g.DeviceCount)
		for i := 0; i < g.DeviceCount; i++ {
			devices[i] = i
		}
		return devices
	}
	return []int{g.DeviceID}
}

// DefaultGPUConfig returns default GPU configuration
func DefaultGPUConfig() *GPUConfig {
	return &GPUConfig{
		Enabled:         true,
		DeviceCount:     1,
		DeviceID:        0,
		CUDADevices:     []int{0},
		MaxMemoryGB:     24.0,
		BenchmarkOnInit: false,
		PowerLimit:      400.0,
		TempLimit:       85.0,
		UseAllDevices:   false,
	}
}

// ClientConfig configuration for the client
type ClientConfig struct {
	Chain     ChainConfig     `yaml:"chain" json:"chain"`
	Client    ClientSettings  `yaml:"client" json:"client"`
	GPU       GPUConfig       `yaml:"gpu" json:"gpu"`
	Analysis  AnalysisConfig  `yaml:"analysis" json:"analysis"`
}

// ChainConfig blockchain configuration
type ChainConfig struct {
	ID           string `yaml:"id" json:"id"`
	RPCEndpoint  string `yaml:"rpc_endpoint" json:"rpc_endpoint"`
	Bech32Prefix string `yaml:"bech32_prefix" json:"bech32_prefix"`
	GasPrice     string `yaml:"gas_price" json:"gas_price"`
}

// ClientSettings client-specific settings
type ClientSettings struct {
	KeyringDir     string        `yaml:"keyring_dir" json:"keyring_dir"`
	KeyringBackend string        `yaml:"keyring_backend" json:"keyring_backend"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	RetryAttempts  int           `yaml:"retry_attempts" json:"retry_attempts"`
}

// AnalysisConfig analysis configuration
type AnalysisConfig struct {
	MaxConcurrent     int           `yaml:"max_concurrent" json:"max_concurrent"`
	DefaultTimeout    time.Duration `yaml:"default_timeout" json:"default_timeout"`
	ResultsDir        string        `yaml:"results_dir" json:"results_dir"`
	EnablePlanet9     bool          `yaml:"enable_planet9" json:"enable_planet9"`
	EnablePhotometric bool          `yaml:"enable_photometric" json:"enable_photometric"`
	EnableClustering  bool          `yaml:"enable_clustering" json:"enable_clustering"`
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Chain: ChainConfig{
			ID:           "medasdigital-2",
			RPCEndpoint:  "https://rpc.medas-digital.io:26657",
			Bech32Prefix: "medas",
			GasPrice:     "0.025umedas",
		},
		Client: ClientSettings{
			KeyringDir:     "",
			KeyringBackend: "os",
			Timeout:        30 * time.Second,
			RetryAttempts:  3,
		},
		GPU: *DefaultGPUConfig(),
		Analysis: AnalysisConfig{
			MaxConcurrent:     4,
			DefaultTimeout:    10 * time.Minute,
			ResultsDir:        "./results",
			EnablePlanet9:     true,
			EnablePhotometric: true,
			EnableClustering:  true,
		},
	}
}

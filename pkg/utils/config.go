package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the client configuration
type Config struct {
	Chain  ChainConfig  `yaml:"chain" mapstructure:"chain"`
	Client ClientConfig `yaml:"client" mapstructure:"client"`
	GPU    GPUConfig    `yaml:"gpu" mapstructure:"gpu"`
	Analysis AnalysisConfig `yaml:"analysis" mapstructure:"analysis"`
	Resources ResourcesConfig `yaml:"resources" mapstructure:"resources"`
}

// ChainConfig contains blockchain-specific configuration
type ChainConfig struct {
	ID          string `yaml:"id" mapstructure:"id"`
	RPCEndpoint string `yaml:"rpc_endpoint" mapstructure:"rpc_endpoint"`
	GRPCEndpoint string `yaml:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	RESTEndpoint string `yaml:"rest_endpoint" mapstructure:"rest_endpoint"`
}

// ClientConfig contains client-specific configuration
type ClientConfig struct {
	Capabilities []string `yaml:"capabilities" mapstructure:"capabilities"`
	KeyringDir   string   `yaml:"keyring_dir" mapstructure:"keyring_dir"`
	DataDir      string   `yaml:"data_dir" mapstructure:"data_dir"`
	LogLevel     string   `yaml:"log_level" mapstructure:"log_level"`
}

// GPUConfig contains GPU-specific configuration
type GPUConfig struct {
	Enabled           bool    `yaml:"enabled" mapstructure:"enabled"`
	CUDADevices       []int   `yaml:"cuda_devices" mapstructure:"cuda_devices"`
	MemoryLimitGB     int     `yaml:"memory_limit_gb" mapstructure:"memory_limit_gb"`
	ComputeCapability string  `yaml:"compute_capability" mapstructure:"compute_capability"`
	CUDAPath          string  `yaml:"cuda_path" mapstructure:"cuda_path"`
	CuDNNPath         string  `yaml:"cudnn_path" mapstructure:"cudnn_path"`
}

// AnalysisConfig contains analysis-specific configuration
type AnalysisConfig struct {
	DataSources    []string `yaml:"data_sources" mapstructure:"data_sources"`
	ModelsDir      string   `yaml:"models_dir" mapstructure:"models_dir"`
	CacheDir       string   `yaml:"cache_dir" mapstructure:"cache_dir"`
	BatchSize      int      `yaml:"batch_size" mapstructure:"batch_size"`
	MaxConcurrent  int      `yaml:"max_concurrent" mapstructure:"max_concurrent"`
}

// ResourcesConfig contains resource limits
type ResourcesConfig struct {
	MaxCPUCores  int    `yaml:"max_cpu_cores" mapstructure:"max_cpu_cores"`
	MaxMemoryGB  int    `yaml:"max_memory_gb" mapstructure:"max_memory_gb"`
	StoragePath  string `yaml:"storage_path" mapstructure:"storage_path"`
	TempDir      string `yaml:"temp_dir" mapstructure:"temp_dir"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	medasDir := filepath.Join(homeDir, ".medasdigital")

	return &Config{
		Chain: ChainConfig{
			ID:           "medasdigital-2",
			RPCEndpoint:  "https://rpc.medas-digital.io:26657",
			GRPCEndpoint: "https://grpc.medas-digital.io:9090",
			RESTEndpoint: "https://rest.medas-digital.io:1317",
		},
		Client: ClientConfig{
			Capabilities: []string{
				"orbital_dynamics",
				"photometric_analysis",
				"astrometric_validation",
			},
			KeyringDir: filepath.Join(medasDir, "keyring"),
			DataDir:    filepath.Join(medasDir, "data"),
			LogLevel:   "info",
		},
		GPU: GPUConfig{
			Enabled:           false,
			CUDADevices:       []int{0},
			MemoryLimitGB:     8,
			ComputeCapability: "8.0",
			CUDAPath:          "/usr/local/cuda",
			CuDNNPath:         "/usr/local/cuda",
		},
		Analysis: AnalysisConfig{
			DataSources: []string{
				"https://api.minorplanetcenter.net/data",
			},
			ModelsDir:     filepath.Join(medasDir, "models"),
			CacheDir:      filepath.Join(medasDir, "cache"),
			BatchSize:     32,
			MaxConcurrent: 4,
		},
		Resources: ResourcesConfig{
			MaxCPUCores: 4,
			MaxMemoryGB: 8,
			StoragePath: filepath.Join(medasDir, "storage"),
			TempDir:     filepath.Join(medasDir, "temp"),
		},
	}
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() (*Config, error) {
	// Set config paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add config paths
	homeDir, _ := os.UserHomeDir()
	viper.AddConfigPath(filepath.Join(homeDir, ".medasdigital"))
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set environment variable prefix
	viper.SetEnvPrefix("MEDASDIGITAL")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			return createDefaultConfig()
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".medasdigital")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create necessary directories
	if err := createDirectories(config); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration saved to: %s\n", configFile)
	return nil
}

// createDefaultConfig creates and saves a default configuration
func createDefaultConfig() (*Config, error) {
	config := DefaultConfig()
	
	if err := SaveConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Chain.ID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}

	if config.Chain.RPCEndpoint == "" {
		return fmt.Errorf("RPC endpoint cannot be empty")
	}

	if len(config.Client.Capabilities) == 0 {
		return fmt.Errorf("at least one capability must be specified")
	}

	// Validate capabilities
	validCapabilities := map[string]bool{
		"orbital_dynamics":      true,
		"photometric_analysis":  true,
		"astrometric_validation": true,
		"clustering_analysis":   true,
		"survey_processing":     true,
		"anomaly_detection":     true,
		"ai_training":          true,
		"gpu_compute":          true,
	}

	for _, cap := range config.Client.Capabilities {
		if !validCapabilities[cap] {
			return fmt.Errorf("invalid capability: %s", cap)
		}
	}

	// Validate GPU config
	if config.GPU.Enabled {
		if len(config.GPU.CUDADevices) == 0 {
			return fmt.Errorf("CUDA devices must be specified when GPU is enabled")
		}
		
		if config.GPU.MemoryLimitGB <= 0 {
			return fmt.Errorf("GPU memory limit must be positive")
		}
	}

	return nil
}

// createDirectories creates necessary directories based on config
func createDirectories(config *Config) error {
	dirs := []string{
		config.Client.KeyringDir,
		config.Client.DataDir,
		config.Analysis.ModelsDir,
		config.Analysis.CacheDir,
		config.Resources.StoragePath,
		config.Resources.TempDir,
	}

	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".medasdigital", "config.yaml"), nil
}

// UpdateConfig updates specific configuration values
func UpdateConfig(updates map[string]interface{}) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Apply updates using reflection or specific field updates
	for key, value := range updates {
		switch key {
		case "chain.id":
			config.Chain.ID = value.(string)
		case "chain.rpc_endpoint":
			config.Chain.RPCEndpoint = value.(string)
		case "gpu.enabled":
			config.GPU.Enabled = value.(bool)
		case "client.capabilities":
			config.Client.Capabilities = value.([]string)
		// Add more cases as needed
		}
	}

	return SaveConfig(config)
}

// GetCapabilities returns the list of enabled capabilities
func (c *Config) GetCapabilities() []string {
	return c.Client.Capabilities
}

// HasCapability checks if a specific capability is enabled
func (c *Config) HasCapability(capability string) bool {
	for _, cap := range c.Client.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// IsGPUEnabled returns whether GPU support is enabled
func (c *Config) IsGPUEnabled() bool {
	return c.GPU.Enabled
}

// GetGPUDevices returns the list of CUDA devices
func (c *Config) GetGPUDevices() []int {
	return c.GPU.CUDADevices
}

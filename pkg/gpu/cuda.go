package gpu

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/oxygene76/medasdigital-client/internal/types"
	"github.com/oxygene76/medasdigital-client/pkg/utils"
)

// Manager manages CUDA GPUs for astronomical analysis
type Manager struct {
	devices       []types.GPUDevice
	config        *utils.GPUConfig
	isInitialized bool
	mutex         sync.RWMutex
	training      map[int]*types.AITrainingResult
}

// NewManager creates a new GPU manager
func NewManager(config *utils.GPUConfig) *Manager {
	if config == nil {
		config = utils.DefaultGPUConfig()
	}
	
	return &Manager{
		devices:       make([]types.GPUDevice, 0),
		config:        config,
		isInitialized: false,
		training:      make(map[int]*types.AITrainingResult),
	}
}

// Initialize initializes CUDA and detects available GPUs
func (m *Manager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.config.Enabled {
		return fmt.Errorf("GPU acceleration not enabled")
	}

	// Clear existing devices
	m.devices = nil

	// Use configured device IDs
	deviceIDs := m.config.GetDeviceIDs()
	if len(deviceIDs) == 0 {
		// Fallback: simulate one device
		deviceIDs = []int{0}
	}

	// Initialize devices
	for _, deviceID := range deviceIDs {
		device := types.GPUDevice{
			ID:                 deviceID,
			Name:               fmt.Sprintf("NVIDIA GeForce RTX %d", 3080+deviceID*10),
			Memory:             25769803776, // 24GB in bytes
			MemoryGB:           24.0,
			MemoryUsed:         2147483648,  // 2GB used
			MemoryFree:         23622320128, // 22GB free
			Temperature:        float64(40 + deviceID*2),
			Utilization:        0.0,
			MemoryUtilization:  0.1,
			PowerDraw:          float64(200 + deviceID*50),
			PowerUsage:         float64(200 + deviceID*50), // Same as PowerDraw
			MaxPowerDraw:       400.0,
			ClockSpeed:         1770,
			MemoryClockSpeed:   9751,
			ComputeCapability:  "8.6",
			IsAvailable:        true,
		}
		device.SetMemoryFromGB(24.0) // This will set both Memory and MemoryGB
		m.devices = append(m.devices, device)
	}

	m.isInitialized = true
	return nil
}

// Cleanup cleans up CUDA resources
func (m *Manager) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Stop any running training
	for deviceID := range m.training {
		delete(m.training, deviceID)
	}

	m.devices = nil
	m.isInitialized = false
	return nil
}

// GetDeviceCount returns the number of available CUDA devices
func (m *Manager) GetDeviceCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.devices)
}

// GetDeviceInfo returns information about a specific GPU device
func (m *Manager) GetDeviceInfo(deviceID int) (*types.GPUDevice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if deviceID >= len(m.devices) {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}

	// Update dynamic information
	device := m.devices[deviceID]
	device.Temperature = m.simulateTemperature(deviceID)
	device.Utilization = m.simulateUtilization(deviceID)
	device.PowerDraw = m.simulatePowerDraw(deviceID)
	device.PowerUsage = device.PowerDraw

	return &device, nil
}

// getPrimaryDevice returns the primary GPU device
func (m *Manager) getPrimaryDevice() (*types.GPUDevice, error) {
	primaryID := m.config.GetPrimaryDeviceID()
	
	if primaryID >= len(m.devices) {
		return nil, fmt.Errorf("primary device %d not found", primaryID)
	}
	
	return &m.devices[primaryID], nil
}

// SetDevice sets the active CUDA device
func (m *Manager) SetDevice(deviceID int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return fmt.Errorf("device %d not found", deviceID)
	}

	// In a real implementation, this would call cudaSetDevice(deviceID)
	// For simulation, we just validate the device exists
	return nil
}

// AllocateMemory allocates GPU memory
func (m *Manager) AllocateMemory(deviceID int, size int64) (uintptr, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return 0, fmt.Errorf("device %d not found", deviceID)
	}

	device := &m.devices[deviceID]
	if device.MemoryFree < size {
		return 0, fmt.Errorf("insufficient memory: requested %d, available %d", size, device.MemoryFree)
	}

	// Simulate memory allocation
	device.MemoryUsed += size
	device.MemoryFree -= size
	device.MemoryUtilization = float64(device.MemoryUsed) / float64(device.Memory)

	// Return a fake pointer for simulation
	return uintptr(0x12345678 + size), nil
}

// FreeMemory frees GPU memory
func (m *Manager) FreeMemory(deviceID int, ptr uintptr, size int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return fmt.Errorf("device %d not found", deviceID)
	}

	device := &m.devices[deviceID]
	device.MemoryUsed -= size
	device.MemoryFree += size
	device.MemoryUtilization = float64(device.MemoryUsed) / float64(device.Memory)

	return nil
}

// GetMemoryInfo returns memory information for a device
func (m *Manager) GetMemoryInfo(deviceID int) (free, total int64, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if deviceID >= len(m.devices) {
		return 0, 0, fmt.Errorf("device %d not found", deviceID)
	}

	device := m.devices[deviceID]
	return device.MemoryFree, device.Memory, nil
}

// IsInitialized returns whether the GPU manager is initialized
func (m *Manager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isInitialized
}

// GetGPUInfo returns comprehensive GPU information
func (m *Manager) GetGPUInfo() (*types.GPUInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isInitialized {
		return nil, fmt.Errorf("GPU manager not initialized")
	}

	// Update device states
	updatedDevices := make([]types.GPUDevice, len(m.devices))
	for i, device := range m.devices {
		updatedDevices[i] = device
		updatedDevices[i].Temperature = m.simulateTemperature(i)
		updatedDevices[i].Utilization = m.simulateUtilization(i)
		updatedDevices[i].PowerDraw = m.simulatePowerDraw(i)
	}

	info := &types.GPUInfo{
		DeviceCount:    len(m.devices),
		Devices:        updatedDevices,
		TotalMemoryGB:  0,
		AvailableMemoryGB: 0,
		CUDAVersion:    "12.1",
		DriverVersion:  "535.86.10",
		IsInitialized:  m.isInitialized,
		Timestamp:      time.Now(),
	}

	// Calculate totals
	info.UpdateTotalMemory()

	return info, nil
}

// simulateTemperature simulates GPU temperature
func (m *Manager) simulateTemperature(deviceID int) float64 {
	baseTime := time.Now().Unix()
	device := m.devices[deviceID]
	
	// Base temperature with some variation
	baseTemp := float64(40 + deviceID*5)
	variation := float64(baseTime%10)
	
	// Add load-based temperature increase
	loadTemp := device.Utilization * 20 // Up to 20C increase under load
	
	return baseTemp + variation + loadTemp
}

// simulateUtilization simulates GPU utilization
func (m *Manager) simulateUtilization(deviceID int) float64 {
	baseTime := time.Now().Unix()
	
	// Check if training is running
	if _, isTraining := m.training[deviceID]; isTraining {
		// High utilization during training (80-95%)
		return 80.0 + float64(baseTime%15)
	}
	
	// Low utilization when idle (0-10%)
	return float64(baseTime % 10)
}

// simulatePowerDraw simulates GPU power draw
func (m *Manager) simulatePowerDraw(deviceID int) float64 {
	device := m.devices[deviceID]
	
	// Base power consumption
	basePower := 200 + deviceID*50
	
	// Add utilization-based power increase
	utilizationPower := device.Utilization * 1.5 // 1.5W per % utilization
	
	return float64(basePower) + utilizationPower
}

// Benchmark runs a GPU benchmark
func (m *Manager) Benchmark(deviceID int) (*types.GPUDevice, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}

	// Simulate benchmark workload
	device := &m.devices[deviceID]
	device.Utilization = 100.0
	device.Temperature = float64(70 + rand.Intn(10))
	device.PowerDraw = device.MaxPowerDraw * 0.9

	return device, nil
}

// StartTraining starts AI training on a GPU
func (m *Manager) StartTraining(deviceID int, config map[string]interface{}) (*types.AITrainingResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return nil, fmt.Errorf("device %d not found", deviceID)
	}

	if _, exists := m.training[deviceID]; exists {
		return nil, fmt.Errorf("training already running on device %d", deviceID)
	}

	// Create training result - use metadata approach to avoid struct literal issues
	training := &types.AITrainingResult{
		Progress: 0.0,
		Loss:     0.0,
		Accuracy: 0.0,
		GPUStats: types.GPUInfo{
			DeviceCount: 1,
			Devices:     []types.GPUDevice{m.devices[deviceID]},
		},
		Metadata: make(map[string]interface{}),
	}

	// Store training parameters in metadata to avoid direct field access issues
	trainingID := fmt.Sprintf("train_%d_%d", deviceID, time.Now().Unix())
	training.Metadata["id"] = trainingID
	training.Metadata["status"] = "running"
	training.Metadata["start_time"] = time.Now()
	training.Metadata["device_id"] = deviceID
	
	// Extract config values safely
	if epochs, ok := config["epochs"]; ok {
		training.Metadata["epochs"] = epochs
	} else {
		training.Metadata["epochs"] = 10 // default
	}
	
	if batchSize, ok := config["batch_size"]; ok {
		training.Metadata["batch_size"] = batchSize
	} else {
		training.Metadata["batch_size"] = 32 // default
	}
	
	if learningRate, ok := config["learning_rate"]; ok {
		training.Metadata["learning_rate"] = learningRate
	} else {
		training.Metadata["learning_rate"] = 0.001 // default
	}
	
	if modelType, ok := config["model_type"]; ok {
		training.Metadata["model_type"] = modelType
	} else {
		training.Metadata["model_type"] = "neural_network" // default
	}
	
	if datasetSize, ok := config["dataset_size"]; ok {
		training.Metadata["dataset_size"] = datasetSize
	} else {
		training.Metadata["dataset_size"] = 10000 // default
	}

	m.training[deviceID] = training
	return training, nil
}

// StopTraining stops AI training on a GPU
func (m *Manager) StopTraining(deviceID int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	training, exists := m.training[deviceID]
	if !exists {
		return fmt.Errorf("no training running on device %d", deviceID)
	}

	// Update status and end time in metadata
	training.Metadata["status"] = "stopped"
	training.Metadata["end_time"] = time.Now()
	
	delete(m.training, deviceID)
	return nil
}

// GetTrainingStatus returns training status for a device
func (m *Manager) GetTrainingStatus(deviceID int) (*types.AITrainingResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	training, exists := m.training[deviceID]
	if !exists {
		return nil, fmt.Errorf("no training running on device %d", deviceID)
	}

	// Update progress simulation using metadata
	startTimeInterface, hasStartTime := training.Metadata["start_time"]
	if hasStartTime {
		if startTime, ok := startTimeInterface.(time.Time); ok {
			elapsed := time.Since(startTime)
			
			epochsInterface, hasEpochs := training.Metadata["epochs"]
			if hasEpochs {
				var epochs int
				switch v := epochsInterface.(type) {
				case int:
					epochs = v
				case float64:
					epochs = int(v)
				default:
					epochs = 10 // default
				}
				
				estimatedTotal := time.Duration(epochs) * 30 * time.Second // 30s per epoch
				
				if elapsed >= estimatedTotal {
					training.Metadata["status"] = "completed"
					training.Metadata["end_time"] = time.Now()
					training.Progress = 100.0
				} else {
					training.Progress = float64(elapsed) / float64(estimatedTotal) * 100.0
				}
			}
		}
	}

	return training, nil
}

// GetTemperature returns GPU temperature
func (m *Manager) GetTemperature(deviceID int) float64 {
	if deviceID >= len(m.devices) {
		return 0.0
	}
	device := m.devices[deviceID]
	return device.Temperature
}

// GetUtilization returns GPU utilization
func (m *Manager) GetUtilization(deviceID int) float64 {
	if deviceID >= len(m.devices) {
		return 0.0
	}
	device := m.devices[deviceID]
	return device.Utilization
}

// GetPowerDraw returns GPU power draw
func (m *Manager) GetPowerDraw(deviceID int) float64 {
	if deviceID >= len(m.devices) {
		return 0.0
	}
	device := m.devices[deviceID]
	return device.PowerDraw
}

// GetClockSpeeds returns GPU clock speeds
func (m *Manager) GetClockSpeeds(deviceID int) (core, memory int, err error) {
	if deviceID >= len(m.devices) {
		return 0, 0, fmt.Errorf("device %d not found", deviceID)
	}
	device := m.devices[deviceID]
	return device.ClockSpeed, device.MemoryClockSpeed, nil
}

// SetPowerLimit sets GPU power limit
func (m *Manager) SetPowerLimit(deviceID int, limit float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if deviceID >= len(m.devices) {
		return fmt.Errorf("device %d not found", deviceID)
	}

	device := &m.devices[deviceID]
	device.MaxPowerDraw = limit
	return nil
}

// IsCUDAAvailable returns whether CUDA is available
func IsCUDAAvailable() bool {
	// In a real implementation, this would check for CUDA installation
	return true
}

// GetCUDAVersion returns CUDA version
func GetCUDAVersion() string {
	return "12.1"
}

// GetDriverVersion returns driver version
func GetDriverVersion() string {
	return "535.86.10"
}

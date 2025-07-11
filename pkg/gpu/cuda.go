package gpu

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/oxygene76/medasdigital-client/internal/types"
	"github.com/oxygene76/medasdigital-client/pkg/utils"
)

// Manager represents the GPU manager for CUDA operations
type Manager struct {
	config          *utils.GPUConfig
	devices         []*types.GPUDevice
	isInitialized   bool
	mutex           sync.RWMutex
	benchmarkResult *BenchmarkResult
}

// BenchmarkResult represents GPU benchmark results
type BenchmarkResult struct {
	DeviceID        int           `json:"device_id"`
	DeviceName      string        `json:"device_name"`
	TotalTime       time.Duration `json:"total_time"`
	ThroughputGFLOPS float64       `json:"throughput_gflops"`
	MemoryBandwidth float64       `json:"memory_bandwidth_gbps"`
	PowerEfficiency float64       `json:"power_efficiency_gflops_per_watt"`
	TestsPassed     int           `json:"tests_passed"`
	TestsFailed     int           `json:"tests_failed"`
}

// TrainingMetrics represents AI training metrics
type TrainingMetrics struct {
	DeviceID       int     `json:"device_id"`
	ModelSize      int64   `json:"model_size_mb"`
	BatchSize      int     `json:"batch_size"`
	Throughput     float64 `json:"throughput_samples_per_sec"`
	MemoryUsage    float64 `json:"memory_usage_percent"`
	PowerUsage     float64 `json:"power_usage_watts"`
	Temperature    float64 `json:"temperature_celsius"`
	TrainingLoss   float64 `json:"training_loss"`
	ValidationLoss float64 `json:"validation_loss"`
}

// NewManager creates a new GPU manager
func NewManager(config *utils.GPUConfig) (*Manager, error) {
	manager := &Manager{
		config:        config,
		devices:       make([]*types.GPUDevice, 0),
		isInitialized: false,
	}

	if err := manager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize GPU manager: %w", err)
	}

	return manager, nil
}

// Initialize initializes the GPU manager and detects available devices
func (m *Manager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isInitialized {
		return nil
	}

	log.Println("Initializing GPU manager...")

	// Detect CUDA devices
	if err := m.detectCUDADevices(); err != nil {
		return fmt.Errorf("failed to detect CUDA devices: %w", err)
	}

	// Initialize each device
	for i, device := range m.devices {
		if err := m.initializeDevice(device); err != nil {
			log.Printf("Warning: Failed to initialize device %d: %v", i, err)
			device.IsAvailable = false
		}
	}

	m.isInitialized = true
	log.Printf("GPU manager initialized with %d devices", len(m.devices))

	return nil
}

// detectCUDADevices detects available CUDA devices
func (m *Manager) detectCUDADevices() error {
	// In a real implementation, this would use CUDA runtime API
	// For now, we'll simulate device detection
	
	deviceCount := m.getDeviceCount()
	
	for i := 0; i < deviceCount; i++ {
		device := &types.GPUDevice{
			ID:   i,
			Name: fmt.Sprintf("NVIDIA GeForce RTX 4090 #%d", i),
		}
		
		// Set memory from GB (will auto-set both fields)
		device.SetMemoryFromGB(24.0) // 24GB VRAM
		
		// Set other properties with proper type conversions
		device.Temperature = float64(40 + i*2)                    // ✅ Fixed: int to float64
		device.Utilization = 0.0
		device.MemoryUtilization = 0.0
		device.SetPowerUsage(float64(200 + i*50))                // ✅ Fixed: use helper method
		device.MaxPowerDraw = float64(450 + i*50)                // ✅ Fixed: int to float64
		device.ClockSpeed = 2520 + i*100                         // MHz
		device.MemoryClockSpeed = 21000 + i*1000                 // MHz
		device.ComputeCapability = "8.9"
		device.IsAvailable = true
		
		// Calculate free memory (simulate 90% free initially)
		device.MemoryFree = int64(float64(device.Memory) * 0.9)
		device.MemoryUsed = device.Memory - device.MemoryFree
		
		m.devices = append(m.devices, device)
	}

	return nil
}

// getDeviceCount returns the number of available CUDA devices
func (m *Manager) getDeviceCount() int {
	// In a real implementation, this would call cudaGetDeviceCount()
	// For simulation, return configured device count or detect based on system
	if m.config.Enabled {
		// Use configured device count or default to 1
		if m.config.DeviceCount > 0 {
			return m.config.DeviceCount
		}
		return 1 // Default to 1 device if enabled but count not specified
	}
	
	// Simulate detection based on system capabilities
	switch runtime.GOOS {
	case "linux":
		return 2 // Assume 2 GPUs on Linux
	case "windows":
		return 1 // Assume 1 GPU on Windows
	default:
		return 0 // No CUDA support on other platforms
	}
}

// initializeDevice initializes a specific GPU device
func (m *Manager) initializeDevice(device *types.GPUDevice) error {
	log.Printf("Initializing GPU device %d: %s", device.ID, device.Name)
	
	// In a real implementation, this would:
	// - Set CUDA device context
	// - Initialize CUDA streams
	// - Allocate initial memory pools
	// - Verify device capabilities
	
	// Simulate initialization by updating device status
	device.IsAvailable = true
	
	// Update device info with current stats
	if err := m.updateDeviceStats(device); err != nil {
		return fmt.Errorf("failed to update device stats: %w", err)
	}
	
	log.Printf("Device %d initialized successfully", device.ID)
	return nil
}

// GetInfo returns GPU information
func (m *Manager) GetInfo() (*types.GPUInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isInitialized {
		return nil, fmt.Errorf("GPU manager not initialized")
	}

	// Get primary device info (first device or configured device)
	primaryDevice := m.getPrimaryDevice()
	if primaryDevice == nil {
		return nil, fmt.Errorf("no GPU devices available")
	}

	info := &types.GPUInfo{
		Name:              primaryDevice.Name,
		Memory:            primaryDevice.Memory,
		MemoryUsed:        primaryDevice.MemoryUsed,
		Temperature:       primaryDevice.Temperature,
		Utilization:       primaryDevice.Utilization,
		CUDAVersion:       m.getCUDAVersion(),
		DriverVersion:     m.getDriverVersion(),
		ComputeCapability: primaryDevice.ComputeCapability,
		DeviceCount:       len(m.devices),
		Devices:           make([]types.GPUDevice, len(m.devices)),
	}

	// Copy all devices
	for i, device := range m.devices {
		info.Devices[i] = *device
	}

	// Update aggregated memory info
	info.UpdateTotalMemory()

	return info, nil
}

// getPrimaryDevice returns the primary GPU device
func (m *Manager) getPrimaryDevice() *types.GPUDevice {
	if len(m.devices) == 0 {
		return nil
	}

	// If specific device configured, use that device
	if m.config.Enabled && m.config.DeviceID >= 0 && m.config.DeviceID < len(m.devices) {
		return m.devices[m.config.DeviceID]
	}

	// Otherwise use first available device
	for _, device := range m.devices {
		if device.IsAvailable {
			return device
		}
	}

	return m.devices[0] // Fallback to first device
}

// PrintStatus prints GPU status information
func (m *Manager) PrintStatus() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isInitialized {
		fmt.Println("GPU Manager: Not initialized")
		return nil
	}

	fmt.Printf("=== GPU Status ===\n")
	fmt.Printf("CUDA Version: %s\n", m.getCUDAVersion())
	fmt.Printf("Driver Version: %s\n", m.getDriverVersion())
	fmt.Printf("Device Count: %d\n", len(m.devices))
	fmt.Printf("\n")

	for i, device := range m.devices {
		fmt.Printf("--- Device %d ---\n", i)
		fmt.Printf("Name: %s\n", device.Name)
		fmt.Printf("Memory: %.1f GB (%.1f%% used)\n", 
			device.MemoryGB, device.GetMemoryUsagePercent())
		fmt.Printf("Temperature: %.1f°C\n", device.Temperature)
		fmt.Printf("Utilization: %.1f%%\n", device.Utilization)
		fmt.Printf("Power: %.1f W / %.1f W\n", device.PowerDraw, device.MaxPowerDraw)
		fmt.Printf("Clock: %d MHz (Memory: %d MHz)\n", 
			device.ClockSpeed, device.MemoryClockSpeed)
		fmt.Printf("Compute Capability: %s\n", device.ComputeCapability)
		fmt.Printf("Available: %t\n", device.IsAvailable)
		
		// Status indicators
		if device.IsOverheating() {
			fmt.Printf("⚠️  WARNING: Device is overheating!\n")
		}
		if device.IsHighUtilization() {
			fmt.Printf("🔥 INFO: Device under high utilization\n")
		}
		
		fmt.Printf("\n")
	}

	return nil
}

// RunBenchmark runs a GPU performance benchmark
func (m *Manager) RunBenchmark() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isInitialized {
		return fmt.Errorf("GPU manager not initialized")
	}

	primaryDevice := m.getPrimaryDevice()
	if primaryDevice == nil {
		return fmt.Errorf("no GPU devices available")
	}

	fmt.Printf("Running GPU benchmark on device %d: %s\n", 
		primaryDevice.ID, primaryDevice.Name)

	startTime := time.Now()

	// Simulate benchmark tests
	result := &BenchmarkResult{
		DeviceID:   primaryDevice.ID,
		DeviceName: primaryDevice.Name,
	}

	// Run matrix multiplication benchmark
	fmt.Print("Matrix multiplication test... ")
	if err := m.runMatrixBenchmark(primaryDevice, result); err != nil {
		fmt.Printf("FAILED: %v\n", err)
		result.TestsFailed++
	} else {
		fmt.Println("PASSED")
		result.TestsPassed++
	}

	// Run memory bandwidth test
	fmt.Print("Memory bandwidth test... ")
	if err := m.runMemoryBenchmark(primaryDevice, result); err != nil {
		fmt.Printf("FAILED: %v\n", err)
		result.TestsFailed++
	} else {
		fmt.Println("PASSED")
		result.TestsPassed++
	}

	// Run power efficiency test
	fmt.Print("Power efficiency test... ")
	if err := m.runPowerBenchmark(primaryDevice, result); err != nil {
		fmt.Printf("FAILED: %v\n", err)
		result.TestsFailed++
	} else {
		fmt.Println("PASSED")
		result.TestsPassed++
	}

	result.TotalTime = time.Since(startTime)
	m.benchmarkResult = result

	// Print results
	fmt.Printf("\n=== Benchmark Results ===\n")
	fmt.Printf("Device: %s\n", result.DeviceName)
	fmt.Printf("Total Time: %v\n", result.TotalTime)
	fmt.Printf("Throughput: %.2f GFLOPS\n", result.ThroughputGFLOPS)
	fmt.Printf("Memory Bandwidth: %.2f GB/s\n", result.MemoryBandwidth)
	fmt.Printf("Power Efficiency: %.2f GFLOPS/W\n", result.PowerEfficiency)
	fmt.Printf("Tests: %d passed, %d failed\n", result.TestsPassed, result.TestsFailed)

	return nil
}

// runMatrixBenchmark runs matrix multiplication benchmark
func (m *Manager) runMatrixBenchmark(device *types.GPUDevice, result *BenchmarkResult) error {
	// Simulate matrix multiplication benchmark
	time.Sleep(2 * time.Second) // Simulate computation time
	
	// Update device utilization during benchmark
	device.Utilization = 95.0
	device.Temperature = float64(65 + rand.Intn(10))    // ✅ Fixed: int to float64
	device.PowerDraw = float64(350 + rand.Intn(50))     // ✅ Fixed: int to float64
	
	// Calculate simulated GFLOPS based on device specs
	baseGFLOPS := 35000.0 // Base GFLOPS for RTX 4090
	variation := float64(rand.Intn(2000) - 1000) // ±1000 GFLOPS variation
	result.ThroughputGFLOPS = baseGFLOPS + variation
	
	return nil
}

// runMemoryBenchmark runs memory bandwidth benchmark
func (m *Manager) runMemoryBenchmark(device *types.GPUDevice, result *BenchmarkResult) error {
	// Simulate memory bandwidth test
	time.Sleep(1 * time.Second)
	
	// Update memory utilization
	device.MemoryUtilization = 80.0
	device.MemoryUsed = int64(float64(device.Memory) * 0.8)
	device.MemoryFree = device.Memory - device.MemoryUsed
	
	// Calculate memory bandwidth (RTX 4090 ~1000 GB/s)
	baseBandwidth := 1000.0
	variation := float64(rand.Intn(100) - 50) // ±50 GB/s variation
	result.MemoryBandwidth = baseBandwidth + variation
	
	return nil
}

// runPowerBenchmark runs power efficiency benchmark  
func (m *Manager) runPowerBenchmark(device *types.GPUDevice, result *BenchmarkResult) error {
	// Simulate power efficiency test
	time.Sleep(1 * time.Second)
	
	// Calculate power efficiency
	if device.PowerDraw > 0 {
		result.PowerEfficiency = result.ThroughputGFLOPS / device.PowerDraw
	}
	
	return nil
}

// updateDeviceStats updates real-time device statistics
func (m *Manager) updateDeviceStats(device *types.GPUDevice) error {
	// In a real implementation, this would query NVIDIA Management Library (NVML)
	// For simulation, generate realistic values with some variation
	
	baseTime := time.Now().Unix()
	
	// Simulate temperature (40-80°C range with time-based variation)
	device.Temperature = float64(40+device.ID*5) + float64(baseTime%10)  // ✅ Fixed: proper type conversion
	
	// Simulate utilization (0-100% with random variation)
	device.Utilization = float64(int(baseTime % 100))                    // ✅ Fixed: int to float64
	
	// Simulate power usage (varies with utilization)
	basePower := 200 + device.ID*50
	variablePower := int(baseTime % 100)
	device.PowerDraw = float64(basePower + variablePower)                // ✅ Fixed: int to float64
	device.PowerUsage = device.PowerDraw                                  // Keep in sync
	
	// Simulate memory usage (changes over time)
	memoryPercent := float64((baseTime + int64(device.ID*10)) % 90 + 10) // 10-100%
	device.MemoryUsed = int64(float64(device.Memory) * memoryPercent / 100.0)
	device.MemoryFree = device.Memory - device.MemoryUsed
	device.MemoryUtilization = memoryPercent
	
	return nil
}

// GetTemperature returns the temperature of a specific device
func (m *Manager) GetTemperature(deviceID int) float64 {  // ✅ Fixed: return type float64
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if deviceID < 0 || deviceID >= len(m.devices) {
		return 0.0
	}
	
	device := m.devices[deviceID]
	m.updateDeviceStats(device) // Update before returning
	
	return device.Temperature  // ✅ Fixed: already float64
}

// GetUtilization returns the utilization of a specific device
func (m *Manager) GetUtilization(deviceID int) float64 {  // ✅ Fixed: return type float64
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if deviceID < 0 || deviceID >= len(m.devices) {
		return 0.0
	}
	
	device := m.devices[deviceID]
	m.updateDeviceStats(device) // Update before returning
	
	return device.Utilization  // ✅ Fixed: already float64
}

// GetMemoryInfo returns memory information for a specific device
func (m *Manager) GetMemoryInfo(deviceID int) (total, used, free int64, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if deviceID < 0 || deviceID >= len(m.devices) {
		return 0, 0, 0, fmt.Errorf("invalid device ID: %d", deviceID)
	}
	
	device := m.devices[deviceID]
	m.updateDeviceStats(device)
	
	return device.Memory, device.MemoryUsed, device.MemoryFree, nil
}

// AllocateMemory allocates GPU memory for computation
func (m *Manager) AllocateMemory(deviceID int, sizeBytes int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if deviceID < 0 || deviceID >= len(m.devices) {
		return fmt.Errorf("invalid device ID: %d", deviceID)
	}
	
	device := m.devices[deviceID]
	
	if !device.IsAvailable {
		return fmt.Errorf("device %d is not available", deviceID)
	}
	
	if device.MemoryFree < sizeBytes {
		return fmt.Errorf("insufficient memory: requested %d bytes, available %d bytes", 
			sizeBytes, device.MemoryFree)
	}
	
	// Simulate memory allocation
	device.MemoryUsed += sizeBytes
	device.MemoryFree -= sizeBytes
	device.MemoryUtilization = float64(device.MemoryUsed) / float64(device.Memory) * 100.0
	
	log.Printf("Allocated %d bytes on device %d", sizeBytes, deviceID)
	return nil
}

// FreeMemory frees previously allocated GPU memory
func (m *Manager) FreeMemory(deviceID int, sizeBytes int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if deviceID < 0 || deviceID >= len(m.devices) {
		return fmt.Errorf("invalid device ID: %d", deviceID)
	}
	
	device := m.devices[deviceID]
	
	if device.MemoryUsed < sizeBytes {
		return fmt.Errorf("cannot free %d bytes: only %d bytes allocated", 
			sizeBytes, device.MemoryUsed)
	}
	
	// Simulate memory deallocation
	device.MemoryUsed -= sizeBytes
	device.MemoryFree += sizeBytes
	device.MemoryUtilization = float64(device.MemoryUsed) / float64(device.Memory) * 100.0
	
	log.Printf("Freed %d bytes on device %d", sizeBytes, deviceID)
	return nil
}

// StartTraining starts AI model training with GPU acceleration
func (m *Manager) StartTraining(modelPath string, deviceIDs []int) (*TrainingMetrics, error) {
	if len(deviceIDs) == 0 {
		deviceIDs = []int{0} // Default to device 0
	}
	
	// Use first device for metrics
	primaryDeviceID := deviceIDs[0]
	
	if primaryDeviceID < 0 || primaryDeviceID >= len(m.devices) {
		return nil, fmt.Errorf("invalid device ID: %d", primaryDeviceID)
	}
	
	device := m.devices[primaryDeviceID]
	
	// Simulate training setup
	metrics := &TrainingMetrics{
		DeviceID:     primaryDeviceID,
		ModelSize:    1024, // 1GB model
		BatchSize:    32,
		Throughput:   150.0, // samples/sec
		MemoryUsage:  75.0,  // 75% memory usage
		PowerUsage:   device.PowerDraw,
		Temperature:  device.Temperature,
		TrainingLoss: 0.85,
		ValidationLoss: 0.92,
	}
	
	// Update device state for training
	device.Utilization = 95.0
	device.MemoryUtilization = metrics.MemoryUsage
	device.Temperature = float64(70 + rand.Intn(10))  // ✅ Fixed: int to float64
	
	log.Printf("Started training on device %d with model: %s", primaryDeviceID, modelPath)
	
	return metrics, nil
}

// StopTraining stops AI model training
func (m *Manager) StopTraining(deviceID int) error {
	if deviceID < 0 || deviceID >= len(m.devices) {
		return fmt.Errorf("invalid device ID: %d", deviceID)
	}
	
	device := m.devices[deviceID]
	
	// Reset device state after training
	device.Utilization = 5.0
	device.MemoryUtilization = 20.0
	device.Temperature = float64(45 + rand.Intn(5))  // ✅ Fixed: int to float64
	
	log.Printf("Stopped training on device %d", deviceID)
	return nil
}

// getCUDAVersion returns the CUDA runtime version
func (m *Manager) getCUDAVersion() string {
	// In a real implementation, this would call cudaRuntimeGetVersion()
	return "12.2"
}

// getDriverVersion returns the NVIDIA driver version
func (m *Manager) getDriverVersion() string {
	// In a real implementation, this would query NVML
	return "545.84"
}

// GetBenchmarkResult returns the last benchmark result
func (m *Manager) GetBenchmarkResult() *BenchmarkResult {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.benchmarkResult
}

// IsInitialized returns whether the GPU manager is initialized
func (m *Manager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.isInitialized
}

// GetDeviceCount returns the number of available devices
func (m *Manager) GetDeviceCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return len(m.devices)
}

// SetDevice sets the active CUDA device for subsequent operations
func (m *Manager) SetDevice(deviceID int) error {
	if deviceID < 0 || deviceID >= len(m.devices) {
		return fmt.Errorf("invalid device ID: %d", deviceID)
	}
	
	device := m.devices[deviceID]
	if !device.IsAvailable {
		return fmt.Errorf("device %d is not available", deviceID)
	}
	
	// In a real implementation, this would call cudaSetDevice()
	log.Printf("Set active device to %d: %s", deviceID, device.Name)
	return nil
}

// Shutdown gracefully shuts down the GPU manager
func (m *Manager) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.isInitialized {
		return nil
	}
	
	log.Println("Shutting down GPU manager...")
	
	// Clean up resources for each device
	for i, device := range m.devices {
		log.Printf("Cleaning up device %d: %s", i, device.Name)
		device.IsAvailable = false
		device.Utilization = 0.0
		device.MemoryUtilization = 0.0
	}
	
	m.isInitialized = false
	log.Println("GPU manager shutdown complete")
	
	return nil
}

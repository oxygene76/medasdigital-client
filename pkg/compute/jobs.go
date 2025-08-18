package compute

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobStatus represents the status of a computation job
type JobStatus string

const (
	StatusSubmitted JobStatus = "submitted"
	StatusQueued    JobStatus = "queued"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
)

// JobType represents different types of computation jobs
type JobType string

const (
	JobTypePICalculation JobType = "pi_calculation"
	// Future job types can be added here
	// JobTypeMatrixMultiplication JobType = "matrix_multiplication"
	// JobTypeFourierTransform     JobType = "fourier_transform"
)

// ComputeJob represents a computation job
type ComputeJob struct {
	// Basic job information
	ID              string                 `json:"id"`
	Type            JobType                `json:"type"`
	Parameters      map[string]interface{} `json:"parameters"`
	Status          JobStatus              `json:"status"`
	Progress        int                    `json:"progress"`
	Result          interface{}            `json:"result,omitempty"`
	Error           string                 `json:"error,omitempty"`
	
	// Payment information
	PaymentTxHash   string                 `json:"payment_tx_hash"`
	PaymentVerified bool                   `json:"payment_verified"`
	PriceBreakdown  *PriceBreakdown        `json:"price_breakdown"`
	
	// Timing information
	SubmittedAt     time.Time              `json:"submitted_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Duration        string                 `json:"duration,omitempty"`
	
	// Client and service information
	ClientAddr      string                 `json:"client_addr"`
	Tier            ServiceTier            `json:"tier"`
	Priority        int                    `json:"priority"`
	
	// Resource tracking
	ResourceUsage   *ResourceUsage         `json:"resource_usage,omitempty"`
	
	// Internal context (not serialized)
	cancelFunc      context.CancelFunc     `json:"-"`
	ctx             context.Context        `json:"-"`
	progressChan    chan int               `json:"-"`
}

// ResourceUsage tracks actual resource consumption
type ResourceUsage struct {
	PeakCPUPercent   float64       `json:"peak_cpu_percent"`
	PeakMemoryMB     float64       `json:"peak_memory_mb"`
	ActualDuration   time.Duration `json:"actual_duration"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          *time.Time    `json:"end_time,omitempty"`
}

// JobManager manages computation jobs
type JobManager struct {
	jobs           map[string]*ComputeJob
	jobCounter     int64
	maxJobs        int
	mu             sync.RWMutex
	pricingManager *PricingManager
	
	// Job queues by priority
	basicQueue    []*ComputeJob
	standardQueue []*ComputeJob
	premiumQueue  []*ComputeJob
	queueMu       sync.Mutex
	
	// Worker management
	workers        int
	workerPool     chan struct{}
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
}

// NewJobManager creates a new job manager
func NewJobManager(maxJobs, workers int, pricingManager *PricingManager) *JobManager {
	jm := &JobManager{
		jobs:           make(map[string]*ComputeJob),
		maxJobs:        maxJobs,
		pricingManager: pricingManager,
		workers:        workers,
		workerPool:     make(chan struct{}, workers),
		shutdownChan:   make(chan struct{}),
	}
	
	// Start worker pool
	jm.startWorkers()
	
	return jm
}

// startWorkers initializes the worker pool
func (jm *JobManager) startWorkers() {
	for i := 0; i < jm.workers; i++ {
		jm.wg.Add(1)
		go jm.worker()
	}
}

// worker processes jobs from the queue
func (jm *JobManager) worker() {
	defer jm.wg.Done()
	
	for {
		select {
		case <-jm.shutdownChan:
			return
		case jm.workerPool <- struct{}{}: // Acquire worker slot
			job := jm.getNextJob()
			if job != nil {
				jm.processJob(job)
			}
			<-jm.workerPool // Release worker slot
		}
	}
}

// getNextJob gets the next job from priority queues
func (jm *JobManager) getNextJob() *ComputeJob {
	jm.queueMu.Lock()
	defer jm.queueMu.Unlock()
	
	// Premium jobs first
	if len(jm.premiumQueue) > 0 {
		job := jm.premiumQueue[0]
		jm.premiumQueue = jm.premiumQueue[1:]
		return job
	}
	
	// Then standard jobs
	if len(jm.standardQueue) > 0 {
		job := jm.standardQueue[0]
		jm.standardQueue = jm.standardQueue[1:]
		return job
	}
	
	// Finally basic jobs
	if len(jm.basicQueue) > 0 {
		job := jm.basicQueue[0]
		jm.basicQueue = jm.basicQueue[1:]
		return job
	}
	
	return nil
}

// SubmitJob submits a new computation job
func (jm *JobManager) SubmitJob(jobType JobType, parameters map[string]interface{}, clientAddr string, tier ServiceTier, paymentTxHash string) (*ComputeJob, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	
	// Check job limits
	if len(jm.jobs) >= jm.maxJobs {
		return nil, fmt.Errorf("maximum concurrent jobs reached (%d)", jm.maxJobs)
	}
	
	// Validate job type
	if !jm.isValidJobType(jobType) {
		return nil, fmt.Errorf("unsupported job type: %s", jobType)
	}
	
	// Validate parameters based on job type
	if err := jm.validateJobParameters(jobType, parameters); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Calculate pricing
	priceBreakdown, err := jm.calculateJobPrice(jobType, parameters, tier)
	if err != nil {
		return nil, fmt.Errorf("pricing calculation failed: %w", err)
	}
	
	// Create job ID
	jm.jobCounter++
	jobID := fmt.Sprintf("%s-%d", jobType, jm.jobCounter)
	
	// Create job context
	ctx, cancel := context.WithCancel(context.Background())
	progressChan := make(chan int, 10) // Buffered channel for progress updates
	
	// Determine priority based on tier
	priority := jm.getTierPriority(tier)
	
	// Create job
	job := &ComputeJob{
		ID:              jobID,
		Type:            jobType,
		Parameters:      parameters,
		Status:          StatusSubmitted,
		Progress:        0,
		PaymentTxHash:   paymentTxHash,
		PaymentVerified: false, // Will be verified separately
		PriceBreakdown:  priceBreakdown,
		SubmittedAt:     time.Now(),
		ClientAddr:      clientAddr,
		Tier:            tier,
		Priority:        priority,
		ctx:             ctx,
		cancelFunc:      cancel,
		progressChan:    progressChan,
	}
	
	// Store job
	jm.jobs[jobID] = job
	
	// Add to appropriate queue
	jm.enqueueJob(job)
	
	return job, nil
}

// enqueueJob adds a job to the appropriate priority queue
func (jm *JobManager) enqueueJob(job *ComputeJob) {
	jm.queueMu.Lock()
	defer jm.queueMu.Unlock()
	
	switch job.Tier {
	case TierPremium:
		jm.premiumQueue = append(jm.premiumQueue, job)
	case TierStandard:
		jm.standardQueue = append(jm.standardQueue, job)
	case TierBasic:
		jm.basicQueue = append(jm.basicQueue, job)
	}
	
	job.Status = StatusQueued
}

// processJob processes a computation job
func (jm *JobManager) processJob(job *ComputeJob) {
	defer func() {
		if r := recover(); r != nil {
			jm.failJob(job, fmt.Sprintf("job panicked: %v", r))
		}
		close(job.progressChan)
	}()
	
	// Check if job was cancelled before starting
	select {
	case <-job.ctx.Done():
		jm.cancelJob(job)
		return
	default:
	}
	
	// Update status to running
	jm.updateJobStatus(job, StatusRunning)
	now := time.Now()
	job.StartedAt = &now
	
	// Initialize resource tracking
	job.ResourceUsage = &ResourceUsage{
		StartTime: now,
	}
	
	// Process based on job type
	switch job.Type {
	case JobTypePICalculation:
		jm.processPICalculation(job)
	default:
		jm.failJob(job, fmt.Sprintf("unsupported job type: %s", job.Type))
		return
	}
	
	// Mark as completed if not already failed or cancelled
	if job.Status == StatusRunning {
		jm.completeJob(job)
	}
}

// processPICalculation processes a PI calculation job
func (jm *JobManager) processPICalculation(job *ComputeJob) {
	// Extract parameters
	digits, ok := job.Parameters["digits"].(float64)
	if !ok {
		jm.failJob(job, "invalid digits parameter")
		return
	}
	
	method, ok := job.Parameters["method"].(string)
	if !ok || method == "" {
		method = "chudnovsky" // Default method
	}
	
	// Create PI calculator
	calc := NewPICalculator(int(digits), method)
	
	// Start progress monitoring
	go jm.monitorProgress(job)
	
	// Calculate PI with progress updates
	result, err := calc.CalculateWithProgress(job.progressChan)
	if err != nil {
		jm.failJob(job, fmt.Sprintf("PI calculation failed: %v", err))
		return
	}
	
	// Store result
	job.Result = result
	job.Progress = 100
	
	// Update resource usage
	if job.ResourceUsage != nil {
		endTime := time.Now()
		job.ResourceUsage.EndTime = &endTime
		job.ResourceUsage.ActualDuration = endTime.Sub(job.ResourceUsage.StartTime)
		
		// Estimate resource usage (in production, this would be measured)
		estimate := jm.pricingManager.EstimateResourceUsage(int(digits), method)
		job.ResourceUsage.PeakCPUPercent = estimate.CPUPercent
		job.ResourceUsage.PeakMemoryMB = estimate.MemoryMB
	}
}

// monitorProgress monitors and updates job progress
func (jm *JobManager) monitorProgress(job *ComputeJob) {
	for {
		select {
		case progress, ok := <-job.progressChan:
			if !ok {
				return // Channel closed
			}
			job.Progress = progress
		case <-job.ctx.Done():
			return // Job cancelled
		}
	}
}

// completeJob marks a job as completed
func (jm *JobManager) completeJob(job *ComputeJob) {
	jm.updateJobStatus(job, StatusCompleted)
	now := time.Now()
	job.CompletedAt = &now
	
	if job.StartedAt != nil {
		job.Duration = now.Sub(*job.StartedAt).String()
	}
}

// failJob marks a job as failed
func (jm *JobManager) failJob(job *ComputeJob, errorMsg string) {
	jm.updateJobStatus(job, StatusFailed)
	job.Error = errorMsg
	now := time.Now()
	job.CompletedAt = &now
	
	if job.StartedAt != nil {
		job.Duration = now.Sub(*job.StartedAt).String()
	}
}

// cancelJob marks a job as cancelled
func (jm *JobManager) cancelJob(job *ComputeJob) {
	jm.updateJobStatus(job, StatusCancelled)
	now := time.Now()
	job.CompletedAt = &now
	
	if job.StartedAt != nil {
		job.Duration = now.Sub(*job.StartedAt).String()
	}
}

// updateJobStatus updates the status of a job
func (jm *JobManager) updateJobStatus(job *ComputeJob, status JobStatus) {
	job.Status = status
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(jobID string) (*ComputeJob, error) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	return job, nil
}

// ListJobs returns all jobs with optional filtering
func (jm *JobManager) ListJobs(clientAddr string, status JobStatus) []*ComputeJob {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	var filteredJobs []*ComputeJob
	
	for _, job := range jm.jobs {
		// Filter by client address if specified
		if clientAddr != "" && job.ClientAddr != clientAddr {
			continue
		}
		
		// Filter by status if specified
		if status != "" && job.Status != status {
			continue
		}
		
		filteredJobs = append(filteredJobs, job)
	}
	
	return filteredJobs
}

// CancelJob cancels a running job
func (jm *JobManager) CancelJob(jobID string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	
	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}
	
	if job.Status == StatusCompleted || job.Status == StatusFailed || job.Status == StatusCancelled {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}
	
	// Cancel the job context
	if job.cancelFunc != nil {
		job.cancelFunc()
	}
	
	return nil
}

// CleanupCompletedJobs removes old completed jobs
func (jm *JobManager) CleanupCompletedJobs(maxAge time.Duration) int {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	var removedCount int
	
	for jobID, job := range jm.jobs {
		if (job.Status == StatusCompleted || job.Status == StatusFailed || job.Status == StatusCancelled) &&
			job.SubmittedAt.Before(cutoff) {
			delete(jm.jobs, jobID)
			removedCount++
		}
	}
	
	return removedCount
}

// GetQueueStatus returns status of job queues
func (jm *JobManager) GetQueueStatus() QueueStatus {
	jm.queueMu.Lock()
	defer jm.queueMu.Unlock()
	
	return QueueStatus{
		BasicQueue:    len(jm.basicQueue),
		StandardQueue: len(jm.standardQueue),
		PremiumQueue:  len(jm.premiumQueue),
		TotalQueued:   len(jm.basicQueue) + len(jm.standardQueue) + len(jm.premiumQueue),
		ActiveWorkers: jm.workers - len(jm.workerPool),
		MaxWorkers:    jm.workers,
	}
}

// QueueStatus represents the status of job queues
type QueueStatus struct {
	BasicQueue    int `json:"basic_queue"`
	StandardQueue int `json:"standard_queue"`
	PremiumQueue  int `json:"premium_queue"`
	TotalQueued   int `json:"total_queued"`
	ActiveWorkers int `json:"active_workers"`
	MaxWorkers    int `json:"max_workers"`
}

// Helper methods

// isValidJobType validates job type
func (jm *JobManager) isValidJobType(jobType JobType) bool {
	switch jobType {
	case JobTypePICalculation:
		return true
	default:
		return false
	}
}

// validateJobParameters validates job parameters based on type
func (jm *JobManager) validateJobParameters(jobType JobType, parameters map[string]interface{}) error {
	switch jobType {
	case JobTypePICalculation:
		digits, ok := parameters["digits"].(float64)
		if !ok {
			return fmt.Errorf("missing or invalid 'digits' parameter")
		}
		if digits <= 0 {
			return fmt.Errorf("digits must be positive")
		}
		
		method, ok := parameters["method"].(string)
		if ok && method != "" {
			validMethods := GetAvailableMethods()
			methodValid := false
			for _, validMethod := range validMethods {
				if method == validMethod {
					methodValid = true
					break
				}
			}
			if !methodValid {
				return fmt.Errorf("invalid method: %s", method)
			}
		}
		
		return nil
	default:
		return fmt.Errorf("unknown job type: %s", jobType)
	}
}

// calculateJobPrice calculates price for a job
func (jm *JobManager) calculateJobPrice(jobType JobType, parameters map[string]interface{}, tier ServiceTier) (*PriceBreakdown, error) {
	switch jobType {
	case JobTypePICalculation:
		digits := int(parameters["digits"].(float64))
		method, ok := parameters["method"].(string)
		if !ok || method == "" {
			method = "chudnovsky"
		}
		
		return jm.pricingManager.CalculatePrice(digits, tier, method)
	default:
		return nil, fmt.Errorf("unsupported job type: %s", jobType)
	}
}

// getTierPriority returns priority value for a tier
func (jm *JobManager) getTierPriority(tier ServiceTier) int {
	switch tier {
	case TierPremium:
		return 3
	case TierStandard:
		return 2
	case TierBasic:
		return 1
	default:
		return 1
	}
}

// Shutdown gracefully shuts down the job manager
func (jm *JobManager) Shutdown(timeout time.Duration) error {
	// Signal shutdown
	close(jm.shutdownChan)
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		jm.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// GetStatistics returns job manager statistics
func (jm *JobManager) GetStatistics() JobStatistics {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	
	stats := JobStatistics{
		TotalJobs: len(jm.jobs),
	}
	
	for _, job := range jm.jobs {
		switch job.Status {
		case StatusSubmitted:
			stats.SubmittedJobs++
		case StatusQueued:
			stats.QueuedJobs++
		case StatusRunning:
			stats.RunningJobs++
		case StatusCompleted:
			stats.CompletedJobs++
		case StatusFailed:
			stats.FailedJobs++
		case StatusCancelled:
			stats.CancelledJobs++
		}
		
		switch job.Tier {
		case TierBasic:
			stats.BasicTierJobs++
		case TierStandard:
			stats.StandardTierJobs++
		case TierPremium:
			stats.PremiumTierJobs++
		}
	}
	
	return stats
}

// JobStatistics represents job manager statistics
type JobStatistics struct {
	TotalJobs        int `json:"total_jobs"`
	SubmittedJobs    int `json:"submitted_jobs"`
	QueuedJobs       int `json:"queued_jobs"`
	RunningJobs      int `json:"running_jobs"`
	CompletedJobs    int `json:"completed_jobs"`
	FailedJobs       int `json:"failed_jobs"`
	CancelledJobs    int `json:"cancelled_jobs"`
	BasicTierJobs    int `json:"basic_tier_jobs"`
	StandardTierJobs int `json:"standard_tier_jobs"`
	PremiumTierJobs  int `json:"premium_tier_jobs"`
}

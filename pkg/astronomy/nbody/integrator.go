package nbody

import (
    "math"
    
    astromath "github.com/oxygene76/medasdigital-client/pkg/astronomy/math"
)

// Body represents a celestial body in the N-body system
type Body struct {
    ID       string              // Identifier
    Mass     float64             // Mass in solar masses
    Position astromath.Vector3   // Position in AU
    Velocity astromath.Vector3   // Velocity in AU/day
}

type Snapshot struct {
    Time   float64
    Bodies []Body
}

// System represents the N-body system
type System struct {
    Bodies []Body
    Time   float64 // Current time in Julian days
    G      float64 // Gravitational constant in AU³/(M☉·day²)
}

// NewSystem creates a new N-body system
func NewSystem() *System {
    return &System{
        Bodies: make([]Body, 0),
        G:      2.959122e-4, // AU³/(M☉·day²) - correct for solar system units
        Time:   0,
    }
}

// Copy creates a deep copy of the system
func (s *System) Copy() *System {
    newSystem := &System{
        Bodies: make([]Body, len(s.Bodies)),
        Time:   s.Time,
        G:      s.G,
    }
    copy(newSystem.Bodies, s.Bodies)
    return newSystem
}

// Integrate runs the N-body simulation for the specified duration
// using the Leapfrog integration method (2nd order, symplectic)
func (s *System) Integrate(duration, timestep float64) [][]Body {
    numSteps := int(duration / timestep)
    history := make([][]Body, 0, numSteps)
    
    // Store initial state
    initialCopy := make([]Body, len(s.Bodies))
    copy(initialCopy, s.Bodies)
    history = append(history, initialCopy)
    
    // Integration loop
    for step := 0; step < numSteps; step++ {
        s.LeapfrogStep(timestep)
        
        // Store state periodically (every 100 steps to save memory)
        if (step+1)%100 == 0 || step == numSteps-1 {
            stateCopy := make([]Body, len(s.Bodies))
            copy(stateCopy, s.Bodies)
            history = append(history, stateCopy)
        }
    }
    
    return history
}

// LeapfrogStep performs one step of the Leapfrog integration
func (s *System) LeapfrogStep(dt float64) {
    // Calculate initial accelerations
    accelerations := s.calculateAccelerations()
    
    // Update velocities by half step (kick)
    for i := range s.Bodies {
        s.Bodies[i].Velocity = s.Bodies[i].Velocity.Add(
            accelerations[i].Scale(dt * 0.5),
        )
    }
    
    // Update positions by full step (drift)
    for i := range s.Bodies {
        s.Bodies[i].Position = s.Bodies[i].Position.Add(
            s.Bodies[i].Velocity.Scale(dt),
        )
    }
    
    // Recalculate accelerations with new positions
    accelerations = s.calculateAccelerations()
    
    // Update velocities by second half step (kick)
    for i := range s.Bodies {
        s.Bodies[i].Velocity = s.Bodies[i].Velocity.Add(
            accelerations[i].Scale(dt * 0.5),
        )
    }
    
    s.Time += dt
}

// calculateAccelerations computes gravitational accelerations for all bodies
func (s *System) calculateAccelerations() []astromath.Vector3 {
    n := len(s.Bodies)
    acc := make([]astromath.Vector3, n)
    
    // For each body
    for i := 0; i < n; i++ {
        // Skip if massless (test particle)
        if s.Bodies[i].Mass == 0 {
            // Test particles only feel gravity from massive bodies
            for j := 0; j < n; j++ {
                if i != j && s.Bodies[j].Mass > 0 {
                    acc[i] = acc[i].Add(s.gravitationalAcceleration(i, j))
                }
            }
        } else {
            // Massive bodies interact with all other massive bodies
            for j := 0; j < n; j++ {
                if i != j && s.Bodies[j].Mass > 0 {
                    acc[i] = acc[i].Add(s.gravitationalAcceleration(i, j))
                }
            }
        }
    }
    
    return acc
}

// gravitationalAcceleration calculates acceleration on body i due to body j
func (s *System) gravitationalAcceleration(i, j int) astromath.Vector3 {
    // Vector from body i to body j
    r := s.Bodies[j].Position.Sub(s.Bodies[i].Position)
    rMag := r.Magnitude()
    
    // Avoid singularity
    if rMag < 1e-10 {
        return astromath.Vector3{}
    }
    
    // Newton's law: a = G * M_j * r / |r|³
    return r.Scale(s.G * s.Bodies[j].Mass / (rMag * rMag * rMag))
}

// GetKineticEnergy calculates total kinetic energy of the system
func (s *System) GetKineticEnergy() float64 {
    energy := 0.0
    for _, body := range s.Bodies {
        if body.Mass > 0 {
            v2 := body.Velocity.Dot(body.Velocity)
            energy += 0.5 * body.Mass * v2
        }
    }
    return energy
}

// GetPotentialEnergy calculates total gravitational potential energy
func (s *System) GetPotentialEnergy() float64 {
    energy := 0.0
    n := len(s.Bodies)
    
    for i := 0; i < n-1; i++ {
        if s.Bodies[i].Mass == 0 {
            continue
        }
        for j := i + 1; j < n; j++ {
            if s.Bodies[j].Mass == 0 {
                continue
            }
            r := s.Bodies[i].Position.Distance(s.Bodies[j].Position)
            if r > 1e-10 {
                energy -= s.G * s.Bodies[i].Mass * s.Bodies[j].Mass / r
            }
        }
    }
    
    return energy
}

// GetTotalEnergy returns the total energy (should be conserved)
func (s *System) GetTotalEnergy() float64 {
    return s.GetKineticEnergy() + s.GetPotentialEnergy()
}

// GetAngularMomentum calculates total angular momentum (should be conserved)
func (s *System) GetAngularMomentum() astromath.Vector3 {
    totalL := astromath.Vector3{}
    
    for _, body := range s.Bodies {
        if body.Mass > 0 {
            L := body.Position.Cross(body.Velocity).Scale(body.Mass)
            totalL = totalL.Add(L)
        }
    }
    
    return totalL
}

// IntegrateAdaptive uses adaptive timestep for better accuracy
func (s *System) IntegrateAdaptive(duration, minStep, maxStep, tolerance float64) [][]Body {
    history := make([][]Body, 0)
    
    // Store initial state
    initialCopy := make([]Body, len(s.Bodies))
    copy(initialCopy, s.Bodies)
    history = append(history, initialCopy)
    
    totalTime := 0.0
    dt := maxStep
    
    for totalTime < duration {
        // Adjust timestep if needed
        if totalTime + dt > duration {
            dt = duration - totalTime
        }
        
        // Try a step
        systemCopy := s.Copy()
        systemCopy.LeapfrogStep(dt)
        
        // Estimate error (simplified - in practice would use higher order method)
        error := s.estimateError(systemCopy, dt)
        
        if error < tolerance {
            // Accept step
            s.Bodies = systemCopy.Bodies
            s.Time = systemCopy.Time
            totalTime += dt
            
            // Store state periodically
            if int(totalTime/maxStep)%100 == 0 {
                stateCopy := make([]Body, len(s.Bodies))
                copy(stateCopy, s.Bodies)
                history = append(history, stateCopy)
            }
            
            // Increase timestep if error is very small
            if error < tolerance*0.1 && dt < maxStep {
                dt = math.Min(dt*1.5, maxStep)
            }
        } else {
            // Reject step and decrease timestep
            dt = math.Max(dt*0.5, minStep)
        }
    }
    
    // Store final state
    finalCopy := make([]Body, len(s.Bodies))
    copy(finalCopy, s.Bodies)
    history = append(history, finalCopy)
    
    return history
}

// estimateError provides a simple error estimate for adaptive stepping
func (s *System) estimateError(other *System, dt float64) float64 {
    maxError := 0.0
    
    for i := range s.Bodies {
        // Compare positions
        posError := s.Bodies[i].Position.Distance(other.Bodies[i].Position)
        
        // Normalize by expected motion
        expectedMotion := s.Bodies[i].Velocity.Magnitude() * dt
        if expectedMotion > 0 {
            normalizedError := posError / expectedMotion
            if normalizedError > maxError {
                maxError = normalizedError
            }
        }
    }
    
    return maxError
}

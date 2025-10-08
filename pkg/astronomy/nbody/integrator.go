package nbody

import (
    astromath "github.com/oxygene76/medasdigital-client/pkg/astronomy/math"
)

type Body struct {
    ID       string
    Mass     float64 // Solar masses
    Position astromath.Vector3 // AU
    Velocity astromath.Vector3 // AU/day
}

type System struct {
    Bodies    []Body
    Time      float64 // Julian days
    G         float64 // Gravitational constant in AU³/(M☉·day²)
}

func NewSystem() *System {
    return &System{
        Bodies: make([]Body, 0),
        G:      2.959122e-4, // AU³/(M☉·day²)
    }
}

// Leapfrog integration step
func (s *System) Step(dt float64) {
    // Calculate accelerations
    accelerations := s.calculateAccelerations()
    
    // Update velocities by half step
    for i := range s.Bodies {
        s.Bodies[i].Velocity = s.Bodies[i].Velocity.Add(
            accelerations[i].Scale(dt * 0.5),
        )
    }
    
    // Update positions
    for i := range s.Bodies {
        s.Bodies[i].Position = s.Bodies[i].Position.Add(
            s.Bodies[i].Velocity.Scale(dt),
        )
    }
    
    // Recalculate accelerations with new positions
    accelerations = s.calculateAccelerations()
    
    // Update velocities by second half step
    for i := range s.Bodies {
        s.Bodies[i].Velocity = s.Bodies[i].Velocity.Add(
            accelerations[i].Scale(dt * 0.5),
        )
    }
    
    s.Time += dt
}

func (s *System) calculateAccelerations() []astromath.Vector3 {
    n := len(s.Bodies)
    acc := make([]astromath.Vector3, n)
    
    for i := 0; i < n; i++ {
        for j := 0; j < n; j++ {
            if i == j {
                continue
            }
            
            r := s.Bodies[j].Position.Sub(s.Bodies[i].Position)
            r3 := math.Pow(r.Magnitude(), 3)
            
            if r3 > 0 {
                acc[i] = acc[i].Add(
                    r.Scale(s.G * s.Bodies[j].Mass / r3),
                )
            }
        }
    }
    
    return acc
}

// Run integration for specified duration
func (s *System) Integrate(duration, timestep float64) [][]Body {
    steps := int(duration / timestep)
    history := make([][]Body, steps)
    
    for i := 0; i < steps; i++ {
        s.Step(timestep)
        
        // Deep copy current state
        snapshot := make([]Body, len(s.Bodies))
        copy(snapshot, s.Bodies)
        history[i] = snapshot
    }
    
    return history
}

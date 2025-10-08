package orbital

import (
    "math"
    astromath "github.com/oxygene76/medasdigital-client/pkg/astronomy/math"
)

type OrbitalElements struct {
    SemiMajorAxis          float64 // AU
    Eccentricity           float64 // 0-1
    Inclination            float64 // radians
    LongitudeAscendingNode float64 // radians
    ArgumentPerihelion     float64 // radians
    MeanAnomaly            float64 // radians
    Epoch                  float64 // JD
}

// Convert orbital elements to cartesian position and velocity
func (oe OrbitalElements) ToCartesian(mu float64) (pos, vel astromath.Vector3) {
    // Solve Kepler's equation for eccentric anomaly
    E := oe.solveKeplersEquation()
    
    // True anomaly
    nu := 2.0 * math.Atan2(
        math.Sqrt(1+oe.Eccentricity)*math.Sin(E/2),
        math.Sqrt(1-oe.Eccentricity)*math.Cos(E/2),
    )
    
    // Distance from focus
    r := oe.SemiMajorAxis * (1 - oe.Eccentricity*math.Cos(E))
    
    // Position in orbital plane
    x := r * math.Cos(nu)
    y := r * math.Sin(nu)
    
    // Rotate to inertial frame
    cosOmega := math.Cos(oe.LongitudeAscendingNode)
    sinOmega := math.Sin(oe.LongitudeAscendingNode)
    cosI := math.Cos(oe.Inclination)
    sinI := math.Sin(oe.Inclination)
    cosW := math.Cos(oe.ArgumentPerihelion)
    sinW := math.Sin(oe.ArgumentPerihelion)
    
    pos.X = (cosOmega*cosW - sinOmega*sinW*cosI)*x + 
            (-cosOmega*sinW - sinOmega*cosW*cosI)*y
    pos.Y = (sinOmega*cosW + cosOmega*sinW*cosI)*x + 
            (-sinOmega*sinW + cosOmega*cosW*cosI)*y
    pos.Z = sinW*sinI*x + cosW*sinI*y
    
    // Velocity calculation
    n := math.Sqrt(mu / math.Pow(oe.SemiMajorAxis, 3))
    vr := math.Sqrt(mu/oe.SemiMajorAxis) * oe.Eccentricity * math.Sin(nu) / 
          math.Sqrt(1 - oe.Eccentricity*oe.Eccentricity)
    vt := math.Sqrt(mu/oe.SemiMajorAxis) * (1 + oe.Eccentricity*math.Cos(nu)) / 
          math.Sqrt(1 - oe.Eccentricity*oe.Eccentricity)
    
    // Transform velocities
    vx := vr*math.Cos(nu) - vt*math.Sin(nu)
    vy := vr*math.Sin(nu) + vt*math.Cos(nu)
    
    vel.X = (cosOmega*cosW - sinOmega*sinW*cosI)*vx + 
            (-cosOmega*sinW - sinOmega*cosW*cosI)*vy
    vel.Y = (sinOmega*cosW + cosOmega*sinW*cosI)*vx + 
            (-sinOmega*sinW + cosOmega*cosW*cosI)*vy
    vel.Z = sinW*sinI*vx + cosW*sinI*vy
    
    return pos, vel
}

func (oe OrbitalElements) solveKeplersEquation() float64 {
    E := oe.MeanAnomaly
    for i := 0; i < 10; i++ {
        E = oe.MeanAnomaly + oe.Eccentricity*math.Sin(E)
    }
    return E
}

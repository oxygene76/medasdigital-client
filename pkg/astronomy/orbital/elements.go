package orbital

import (
    "math"
    astromath "github.com/oxygene76/medasdigital-client/pkg/astronomy/math"
)

// OrbitalElements represents Keplerian orbital elements
type OrbitalElements struct {
    SemiMajorAxis          float64 // a - Semi-major axis (AU)
    Eccentricity           float64 // e - Eccentricity (0-1)
    Inclination            float64 // i - Inclination (radians)
    LongitudeAscendingNode float64 // Ω - Longitude of ascending node (radians)
    ArgumentPerihelion     float64 // ω - Argument of perihelion (radians)
    MeanAnomaly            float64 // M - Mean anomaly at epoch (radians)
    Epoch                  float64 // JD - Julian date of epoch
}

// ToCartesian converts orbital elements to cartesian position and velocity
// mu is the gravitational parameter (G * M_sun) in AU³/day²
func (oe OrbitalElements) ToCartesian(mu float64) (pos, vel astromath.Vector3) {
    // Solve Kepler's equation for eccentric anomaly
    E := oe.solveKeplersEquation()
    
    // True anomaly from eccentric anomaly
    sinE := math.Sin(E)
    cosE := math.Cos(E)
    
    nu := 2.0 * math.Atan2(
        math.Sqrt(1+oe.Eccentricity)*math.Sin(E/2),
        math.Sqrt(1-oe.Eccentricity)*math.Cos(E/2),
    )
    
    // Distance from focus
    r := oe.SemiMajorAxis * (1 - oe.Eccentricity*cosE)
    
    // Position in orbital plane
    x := r * math.Cos(nu)
    y := r * math.Sin(nu)
    
    // Velocity in orbital plane
    factor := math.Sqrt(mu/oe.SemiMajorAxis) / math.Sqrt(1 - oe.Eccentricity*oe.Eccentricity)
    vx := -factor * oe.SemiMajorAxis * math.Sin(E)
    vy := factor * oe.SemiMajorAxis * math.Sqrt(1 - oe.Eccentricity*oe.Eccentricity) * cosE
    
    // Rotation matrices
    cosOmega := math.Cos(oe.LongitudeAscendingNode)
    sinOmega := math.Sin(oe.LongitudeAscendingNode)
    cosI := math.Cos(oe.Inclination)
    sinI := math.Sin(oe.Inclination)
    cosW := math.Cos(oe.ArgumentPerihelion)
    sinW := math.Sin(oe.ArgumentPerihelion)
    
    // Transform to inertial frame
    // Rotation matrix elements
    r11 := cosOmega*cosW - sinOmega*sinW*cosI
    r12 := -cosOmega*sinW - sinOmega*cosW*cosI
    r21 := sinOmega*cosW + cosOmega*sinW*cosI
    r22 := -sinOmega*sinW + cosOmega*cosW*cosI
    r31 := sinW * sinI
    r32 := cosW * sinI
    
    // Apply rotation to position
    pos.X = r11*x + r12*y
    pos.Y = r21*x + r22*y
    pos.Z = r31*x + r32*y
    
    // Apply rotation to velocity
    vel.X = r11*vx + r12*vy
    vel.Y = r21*vx + r22*vy
    vel.Z = r31*vx + r32*vy
    
    return pos, vel
}

// solveKeplersEquation solves Kepler's equation M = E - e*sin(E) for E
func (oe OrbitalElements) solveKeplersEquation() float64 {
    // Newton-Raphson iteration
    E := oe.MeanAnomaly
    if oe.Eccentricity > 0.8 {
        E = math.Pi // Better initial guess for high eccentricity
    }
    
    tolerance := 1e-10
    maxIterations := 50
    
    for i := 0; i < maxIterations; i++ {
        f := E - oe.Eccentricity*math.Sin(E) - oe.MeanAnomaly
        fp := 1 - oe.Eccentricity*math.Cos(E)
        
        deltaE := f / fp
        E = E - deltaE
        
        if math.Abs(deltaE) < tolerance {
            break
        }
    }
    
    return E
}

// GetPerihelion returns the perihelion distance
func (oe OrbitalElements) GetPerihelion() float64 {
    return oe.SemiMajorAxis * (1 - oe.Eccentricity)
}

// GetAphelion returns the aphelion distance
func (oe OrbitalElements) GetAphelion() float64 {
    return oe.SemiMajorAxis * (1 + oe.Eccentricity)
}

// GetOrbitalPeriod returns the orbital period in days
func (oe OrbitalElements) GetOrbitalPeriod(mu float64) float64 {
    return 2 * math.Pi * math.Sqrt(math.Pow(oe.SemiMajorAxis, 3) / mu)
}

// GetLongitudeOfPerihelion returns the longitude of perihelion
func (oe OrbitalElements) GetLongitudeOfPerihelion() float64 {
    return math.Mod(oe.LongitudeAscendingNode + oe.ArgumentPerihelion, 2*math.Pi)
}

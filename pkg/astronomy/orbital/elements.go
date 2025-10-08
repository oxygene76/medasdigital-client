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
// Add this function to pkg/astronomy/orbital/elements.go

// CartesianToOrbital converts position and velocity vectors to orbital elements
func CartesianToOrbital(pos, vel astromath.Vector3, mu float64) OrbitalElements {
    // Specific angular momentum
    h := pos.Cross(vel)
    
    // Eccentricity vector
    r := pos.Magnitude()
    v := vel.Magnitude()
    eVec := vel.Cross(h).Scale(1.0/mu).Sub(pos.Scale(1.0/r))
    e := eVec.Magnitude()
    
    // Semi-major axis
    a := 1.0 / (2.0/r - v*v/mu)
    
    // Inclination
    i := math.Acos(h.Z / h.Magnitude())
    
    // Longitude of ascending node
    n := astromath.Vector3{0, 0, 1}.Cross(h)
    Omega := 0.0
    if n.Magnitude() > 1e-10 {
        Omega = math.Atan2(n.Y, n.X)
        if Omega < 0 {
            Omega += 2 * math.Pi
        }
    }
    
    // Argument of perihelion
    omega := 0.0
    if n.Magnitude() > 1e-10 && e > 1e-10 {
        cosOmega := n.Dot(eVec) / (n.Magnitude() * e)
        if math.Abs(cosOmega) <= 1.0 {
            omega = math.Acos(cosOmega)
            if eVec.Z < 0 {
                omega = 2*math.Pi - omega
            }
        }
    }
    
    // Mean anomaly
    cosE := (1 - r/a) / e
    E := 0.0
    if math.Abs(cosE) <= 1.0 {
        E = math.Acos(cosE)
        if pos.Dot(vel) < 0 {
            E = 2*math.Pi - E
        }
    }
    M := E - e*math.Sin(E)
    
    return OrbitalElements{
        SemiMajorAxis:          a,
        Eccentricity:           e,
        Inclination:            i,
        LongitudeAscendingNode: Omega,
        ArgumentPerihelion:     omega,
        MeanAnomaly:            M,
    }
}

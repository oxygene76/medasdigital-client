package orbital

import (
    "math"
    "fmt"
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
// ToCartesian converts orbital elements to position and velocity
// Returns position in AU and velocity in AU/year when mu is in AU³/(M☉·year²)
func (o OrbitalElements) ToCartesian(mu float64) (astromath.Vector3, astromath.Vector3) {
    // Solve Kepler's equation for eccentric anomaly
    E := o.MeanAnomaly
    for i := 0; i < 10; i++ {
        E = o.MeanAnomaly + o.Eccentricity*math.Sin(E)
    }
    
    cosE := math.Cos(E)
    sinE := math.Sin(E)
    
    // Position in orbital plane
    a := o.SemiMajorAxis
    e := o.Eccentricity
    
    // Distance from focus
    r := a * (1 - e*cosE)
    
    // Position in orbital plane coordinates
    x := a * (cosE - e)
    y := a * math.Sqrt(1-e*e) * sinE
    
    // FIXED: Velocity in orbital plane (was wrong!)
    // Mean motion n = sqrt(mu/a³)
    n := math.Sqrt(mu / (a * a * a))
    
    // Velocity components in orbital plane
    vx := -a * n * sinE / (1 - e*cosE)
    vy := a * n * math.Sqrt(1-e*e) * cosE / (1 - e*cosE)
    
    // Create rotation matrices for orbital orientation
    cosOmega := math.Cos(o.LongitudeAscendingNode)
    sinOmega := math.Sin(o.LongitudeAscendingNode)
    cosI := math.Cos(o.Inclination)
    sinI := math.Sin(o.Inclination)
    cosW := math.Cos(o.ArgumentPerihelion)
    sinW := math.Sin(o.ArgumentPerihelion)
    
    // Transform to 3D space
    // Rotation matrix elements
    r11 := cosOmega*cosW - sinOmega*sinW*cosI
    r12 := -cosOmega*sinW - sinOmega*cosW*cosI
    r21 := sinOmega*cosW + cosOmega*sinW*cosI
    r22 := -sinOmega*sinW + cosOmega*cosW*cosI
    r31 := sinW*sinI
    r32 := cosW*sinI
    
    // Apply rotation to position
    pos := astromath.Vector3{
        X: r11*x + r12*y,
        Y: r21*x + r22*y,
        Z: r31*x + r32*y,
    }
    
    // Apply rotation to velocity
    vel := astromath.Vector3{
        X: r11*vx + r12*vy,
        Y: r21*vx + r22*vy,
        Z: r31*vx + r32*vy,
    }
    
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
// CartesianToOrbital converts position and velocity vectors to orbital elements
// Position in AU, velocity in AU/year, mu in AU³/(M☉·year²)
func CartesianToOrbital(pos, vel astromath.Vector3, mu float64) OrbitalElements {
    r := pos.Magnitude()
    v := vel.Magnitude()
    
    // Check for invalid inputs
    if r == 0 || v == 0 || mu == 0 {
        return OrbitalElements{}
    }
    
    // Angular momentum vector
    h := pos.Cross(vel)
    hMag := h.Magnitude()
    
    // Check if orbit is degenerate
    if hMag < 1e-10 {
        return OrbitalElements{}
    }
    
    // Specific orbital energy
    energy := (v*v)/2.0 - mu/r
    
    // Semi-major axis (from vis-viva equation)
    a := -mu / (2 * energy)
    
    // Eccentricity vector: e = ((v²-μ/r)r - (r·v)v) / μ
    rdotv := pos.Dot(vel)
    eVec := pos.Scale((v*v - mu/r) / mu).Sub(vel.Scale(rdotv / mu))
    e := eVec.Magnitude()
    
    // For near-circular orbits, eccentricity vector may be unreliable
    if e < 1e-10 {
        e = 0
    }
    
    // Validate eccentricity
    if e >= 1.0 {
        // This shouldn't happen for bound orbits
        fmt.Printf("DEBUG CartesianToOrbital: e=%.3f, a=%.1f, energy=%.6f\n", e, a, energy)
        // Clamp to just below parabolic
        e = 0.999
    }
    
    // Inclination (angle between h and z-axis)
    i := math.Acos(math.Min(1.0, math.Max(-1.0, h.Z/hMag)))
    
    // Node vector (points along line of nodes)
    n := astromath.Vector3{0, 0, 1}.Cross(h)
    nMag := n.Magnitude()
    
    // Longitude of ascending node (angle from x-axis to node)
    Omega := 0.0
    if nMag > 1e-10 {
        Omega = math.Atan2(n.Y, n.X)
        if Omega < 0 {
            Omega += 2 * math.Pi
        }
    }
    
    // Argument of perihelion (angle from node to perihelion)
    omega := 0.0
    if nMag > 1e-10 && e > 1e-10 {
        cosOmega := n.Dot(eVec) / (nMag * e)
        cosOmega = math.Min(1.0, math.Max(-1.0, cosOmega))
        omega = math.Acos(cosOmega)
        if eVec.Z < 0 {
            omega = 2*math.Pi - omega
        }
    } else if e > 1e-10 {
        // For zero inclination, use angle from x-axis
        omega = math.Atan2(eVec.Y, eVec.X)
        if omega < 0 {
            omega += 2 * math.Pi
        }
    }
    
    // True anomaly (angle from perihelion to current position)
    nu := 0.0
    if e > 1e-10 {
        cosNu := pos.Dot(eVec) / (r * e)
        cosNu = math.Min(1.0, math.Max(-1.0, cosNu))
        nu = math.Acos(cosNu)
        if rdotv < 0 {
            nu = 2*math.Pi - nu
        }
    } else {
        // For circular orbit, measure from node or x-axis
        if nMag > 1e-10 {
            cosNu := pos.Dot(n) / (r * nMag)
            cosNu = math.Min(1.0, math.Max(-1.0, cosNu))
            nu = math.Acos(cosNu)
        } else {
            nu = math.Atan2(pos.Y, pos.X)
        }
        if nu < 0 {
            nu += 2 * math.Pi
        }
    }
    
    // Eccentric anomaly and mean anomaly
    E := 0.0
    M := 0.0
    if e < 0.99 {
        // Elliptical orbit
        E = 2 * math.Atan(math.Tan(nu/2) / math.Sqrt((1+e)/(1-e)))
        if E < 0 {
            E += 2 * math.Pi
        }
        M = E - e*math.Sin(E)
    } else {
        // Near-parabolic, use true anomaly as approximation
        M = nu
    }
    
    return OrbitalElements{
        SemiMajorAxis:          a,
        Eccentricity:           e,
        Inclination:            i,
        LongitudeAscendingNode: Omega,
        ArgumentPerihelion:     omega,
        MeanAnomaly:            M,
    }
}

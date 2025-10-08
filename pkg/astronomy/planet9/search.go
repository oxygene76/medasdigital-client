package planet9

import (
    "fmt"
    "math"
    
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/nbody"
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/orbital"
    astromath "github.com/oxygene76/medasdigital-client/pkg/astronomy/math"
)

// SearchPreset represents predefined search parameters from published papers
type SearchPreset string

const (
    PresetBatyginBrown2016  SearchPreset = "batygin_brown_2016"
    PresetTrujilloSheppard  SearchPreset = "trujillo_sheppard_2014"
    PresetBrownBatygin2021  SearchPreset = "brown_batygin_2021"
    PresetCustom            SearchPreset = "custom"
)

type SearchParameters struct {
    Mass                   float64
    SemiMajorAxis          float64
    Eccentricity           float64
    Inclination            float64
    LongitudeAscendingNode float64
    ArgumentPerihelion     float64
}

type SearchResult struct {
    Parameters      SearchParameters
    ETNOEffects     []ETNOEffect
    ClusteringScore float64
}

type ETNOEffect struct {
    ObjectID          string
    InitialElements   orbital.OrbitalElements
    FinalElements     orbital.OrbitalElements
    PerihelionShift   float64
    InclinationChange float64
    LongPeriChange    float64  // Change in longitude of perihelion
}

// GetPresetParameters returns parameters for known presets
func GetPresetParameters(preset SearchPreset) SearchParameters {
    switch preset {
    case PresetBatyginBrown2016:
        // Original Planet Nine hypothesis
        return SearchParameters{
            Mass:                   10.0,  // Earth masses
            SemiMajorAxis:          700.0, // AU
            Eccentricity:           0.6,
            Inclination:            30.0,  // degrees
            LongitudeAscendingNode: 100.0,
            ArgumentPerihelion:     150.0,
        }
        
    case PresetTrujilloSheppard:
        // Based on 2012 VP113 discovery
        return SearchParameters{
            Mass:                   5.0,   // Smaller mass
            SemiMajorAxis:          300.0, // Closer
            Eccentricity:           0.3,
            Inclination:            20.0,
            LongitudeAscendingNode: 90.0,
            ArgumentPerihelion:     290.0,
        }
        
    case PresetBrownBatygin2021:
        // Updated parameters
        return SearchParameters{
            Mass:                   7.5,
            SemiMajorAxis:          500.0,
            Eccentricity:           0.4,
            Inclination:            20.0,
            LongitudeAscendingNode: 100.0,
            ArgumentPerihelion:     150.0,
        }
        
    default:
        // Default custom parameters
        return SearchParameters{
            Mass:                   7.5,
            SemiMajorAxis:          600.0,
            Eccentricity:           0.3,
            Inclination:            25.0,
            LongitudeAscendingNode: 100.0,
            ArgumentPerihelion:     145.0,
        }
    }
}

func RunSimulation(params SearchParameters, etnos []orbital.OrbitalElements, 
                   duration float64) SearchResult {
    // Initialize system with Sun + outer planets + Planet 9 + ETNOs
    system := nbody.NewSystem()
    
    // Add Sun
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:   "Sun",
        Mass: 1.0,
        Position: astromath.Vector3{0, 0, 0},
        Velocity: astromath.Vector3{0, 0, 0},
    })
    
    // Add outer planets (Jupiter, Saturn, Uranus, Neptune)
    addOuterPlanets(system)
    
    // Add Planet 9
    p9Elements := orbital.OrbitalElements{
        SemiMajorAxis:          params.SemiMajorAxis,
        Eccentricity:           params.Eccentricity,
        Inclination:            params.Inclination * math.Pi / 180,
        LongitudeAscendingNode: params.LongitudeAscendingNode * math.Pi / 180,
        ArgumentPerihelion:     params.ArgumentPerihelion * math.Pi / 180,
        MeanAnomaly:            0,
        Epoch:                  2460200.5, // J2000 + 23 years
    }
    
    p9Pos, p9Vel := p9Elements.ToCartesian(system.G)
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Planet9",
        Mass:     params.Mass * 3.003e-6, // Earth masses to solar masses
        Position: p9Pos,
        Velocity: p9Vel,
    })
    
    // Store initial ETNO elements for comparison
    initialETNOs := make([]orbital.OrbitalElements, len(etnos))
    copy(initialETNOs, etnos)
    
    // Add ETNOs as test particles
    for i, etno := range etnos {
        pos, vel := etno.ToCartesian(system.G)
        system.Bodies = append(system.Bodies, nbody.Body{
            ID:       fmt.Sprintf("ETNO_%d", i),
            Mass:     0, // Massless test particles
            Position: pos,
            Velocity: vel,
        })
    }
    
    // Run integration
    timestep := 10.0 // days
    history := system.Integrate(duration*365.25, timestep)
    
    // Analyze results
    result := SearchResult{
        Parameters: params,
    }
    
    // Calculate ETNO effects and clustering
    result.ETNOEffects = analyzeETNOChanges(history, initialETNOs, system)
    result.ClusteringScore = calculateClustering(result.ETNOEffects)
    
    return result
}

// addOuterPlanets adds Jupiter, Saturn, Uranus, Neptune
func addOuterPlanets(system *nbody.System) {
    planets := []struct {
        name string
        mass float64  // solar masses
        elem orbital.OrbitalElements
    }{
        {
            name: "Jupiter",
            mass: 0.0009545942,
            elem: orbital.OrbitalElements{
                SemiMajorAxis:          5.2038,
                Eccentricity:           0.0489,
                Inclination:            1.303 * math.Pi / 180,
                LongitudeAscendingNode: 100.464 * math.Pi / 180,
                ArgumentPerihelion:     273.867 * math.Pi / 180,
                MeanAnomaly:            20.020 * math.Pi / 180,
            },
        },
        {
            name: "Saturn",
            mass: 0.0002857214,
            elem: orbital.OrbitalElements{
                SemiMajorAxis:          9.5826,
                Eccentricity:           0.0565,
                Inclination:            2.485 * math.Pi / 180,
                LongitudeAscendingNode: 113.665 * math.Pi / 180,
                ArgumentPerihelion:     339.392 * math.Pi / 180,
                MeanAnomaly:            317.020 * math.Pi / 180,
            },
        },
        {
            name: "Uranus",
            mass: 0.00004365785,
            elem: orbital.OrbitalElements{
                SemiMajorAxis:          19.2012,
                Eccentricity:           0.0469,
                Inclination:            0.773 * math.Pi / 180,
                LongitudeAscendingNode: 74.006 * math.Pi / 180,
                ArgumentPerihelion:     96.998 * math.Pi / 180,
                MeanAnomaly:            142.238 * math.Pi / 180,
            },
        },
        {
            name: "Neptune",
            mass: 0.00005149497,
            elem: orbital.OrbitalElements{
                SemiMajorAxis:          30.0479,
                Eccentricity:           0.0087,
                Inclination:            1.767 * math.Pi / 180,
                LongitudeAscendingNode: 131.783 * math.Pi / 180,
                ArgumentPerihelion:     276.336 * math.Pi / 180,
                MeanAnomaly:            256.228 * math.Pi / 180,
            },
        },
    }
    
    for _, planet := range planets {
        pos, vel := planet.elem.ToCartesian(system.G)
        system.Bodies = append(system.Bodies, nbody.Body{
            ID:       planet.name,
            Mass:     planet.mass,
            Position: pos,
            Velocity: vel,
        })
    }
}

// Update this function in pkg/astronomy/planet9/search.go

func analyzeETNOChanges(history []nbody.Snapshot, initialETNOs []orbital.OrbitalElements) []ETNOEffect {
    if len(history) == 0 {
        return nil
    }
    
    effects := make([]ETNOEffect, 0)
    firstSnap := history[0]
    lastSnap := history[len(history)-1]
    
    // Skip Sun and Planet9, analyze ETNOs
    etnoStart := 2 // Index where ETNOs start (after Sun and Planet9)
    
    for i := 0; i < len(initialETNOs) && etnoStart+i < len(lastSnap.Bodies); i++ {
        initial := firstSnap.Bodies[etnoStart+i]
        final := lastSnap.Bodies[etnoStart+i]
        
        // Skip if positions are invalid
        if initial.Position.IsZero() || final.Position.IsZero() {
            continue
        }
        
        // Convert to orbital elements using solar gravitational parameter
        mu := 4 * math.Pi * math.Pi // In AU^3/year^2 units
        
        initialOrb := orbital.CartesianToOrbital(initial.Position, initial.Velocity, mu)
        finalOrb := orbital.CartesianToOrbital(final.Position, final.Velocity, mu)
        
        // Calculate changes
        perihelionInitial := initialOrb.SemiMajorAxis * (1 - initialOrb.Eccentricity)
        perihelionFinal := finalOrb.SemiMajorAxis * (1 - finalOrb.Eccentricity)
        
        effect := ETNOEffect{
            ObjectID:          fmt.Sprintf("ETNO_%d", i),
            InitialElements:   initialETNOs[i],
            FinalElements:     finalOrb,
            PerihelionShift:   perihelionFinal - perihelionInitial,
            InclinationChange: (finalOrb.Inclination - initialOrb.Inclination) * 180.0 / math.Pi,
        }
        
        effects = append(effects, effect)
    }
    
    return effects
}

// calculateClustering calculates the Rayleigh test statistic for longitude of perihelion
func calculateClustering(effects []ETNOEffect) float64 {
    if len(effects) == 0 {
        return 0.0
    }
    
    // Calculate mean vector for longitude of perihelion
    sumCos := 0.0
    sumSin := 0.0
    
    for _, effect := range effects {
        // Get final longitude of perihelion
        longPeri := effect.FinalElements.LongitudeAscendingNode + 
                   effect.FinalElements.ArgumentPerihelion
        
        sumCos += math.Cos(longPeri)
        sumSin += math.Sin(longPeri)
    }
    
    n := float64(len(effects))
    meanCos := sumCos / n
    meanSin := sumSin / n
    
    // Rayleigh statistic R
    R := math.Sqrt(meanCos*meanCos + meanSin*meanSin)
    
    // Normalize to 0-1 scale (R can be at most 1)
    return R
}

// cartesianToOrbital converts position and velocity to orbital elements
func cartesianToOrbital(pos, vel astromath.Vector3, mu float64) orbital.OrbitalElements {
    // This is a simplified conversion - full implementation would be more complex
    r := pos.Magnitude()
    v := vel.Magnitude()
    
    // Specific orbital energy
    energy := v*v/2 - mu/r
    
    // Semi-major axis
    a := -mu / (2 * energy)
    
    // Angular momentum vector
    h := pos.Cross(vel)
    hMag := h.Magnitude()
    
    // Eccentricity vector
    eVec := vel.Cross(h).Scale(1/mu).Sub(pos.Scale(1/r))
    e := eVec.Magnitude()
    
    // Inclination
    i := math.Acos(h.Z / hMag)
    
    // Node vector
    n := astromath.Vector3{0, 0, 1}.Cross(h)
    nMag := n.Magnitude()
    
    // Longitude of ascending node
    omega := 0.0
    if nMag > 0 {
        omega = math.Atan2(n.Y, n.X)
    }
    
    // Argument of perihelion
    w := 0.0
    if nMag > 0 && e > 0 {
        w = math.Acos(n.Dot(eVec) / (nMag * e))
        if eVec.Z < 0 {
            w = 2*math.Pi - w
        }
    }
    
    // True anomaly
    nu := 0.0
    if e > 0 {
        cosNu := eVec.Dot(pos) / (e * r)
        nu = math.Acos(math.Max(-1, math.Min(1, cosNu)))
        if pos.Dot(vel) < 0 {
            nu = 2*math.Pi - nu
        }
    }
    
    // Eccentric anomaly
    E := 2 * math.Atan(math.Sqrt((1-e)/(1+e)) * math.Tan(nu/2))
    
    // Mean anomaly
    M := E - e*math.Sin(E)
    
    return orbital.OrbitalElements{
        SemiMajorAxis:          a,
        Eccentricity:           e,
        Inclination:            i,
        LongitudeAscendingNode: omega,
        ArgumentPerihelion:     w,
        MeanAnomaly:            M,
    }
}

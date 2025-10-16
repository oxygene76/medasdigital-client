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
    // Initialize system with proper units
    system := nbody.NewSystem()
    // system.G = 2.959122e-4 (AU³/M☉·day²) is already set correctly
    
    // Add Sun at origin
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Sun",
        Mass:     1.0,  // Solar masses
        Position: astromath.Vector3{0, 0, 0},
        Velocity: astromath.Vector3{0, 0, 0},
    })
    
    
    p9Elements := orbital.OrbitalElements{
    SemiMajorAxis:          params.SemiMajorAxis,
    Eccentricity:           params.Eccentricity,
    Inclination:            params.Inclination * math.Pi / 180, // passt
    LongitudeAscendingNode: params.LongitudeAscendingNode * math.Pi / 180,
    ArgumentPerihelion:     params.ArgumentPerihelion * math.Pi / 180,
    MeanAnomaly:            0,
    }
    p9Elements.EnsureRadians() // <- schadet nie
    
    // Use mu in year units for ToCartesian (which expects year units)
    muYear := 4 * math.Pi * math.Pi  // AU³/(M☉·year²)
    
    p9Pos, p9Vel := p9Elements.ToCartesian(muYear)
    // CRITICAL: Convert velocity from AU/year to AU/day for integrator
    p9Vel = p9Vel.Scale(1.0 / 365.25)
    
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Planet9",
        Mass:     params.Mass * 3.003e-6,  // Earth masses to solar masses
        Position: p9Pos,
        Velocity: p9Vel,  // Now in AU/day
    })
    
    // Add outer planets (optional but recommended for realism)
    // Jupiter
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Jupiter",
        Mass:     0.000954,  // Solar masses
        Position: astromath.Vector3{X: 5.2, Y: 0, Z: 0},
        Velocity: astromath.Vector3{X: 0, Y: 13.07/365.25, Z: 0},  // Convert km/s to AU/day
    })
    
    // Saturn  
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Saturn",
        Mass:     0.000286,
        Position: astromath.Vector3{X: 9.5, Y: 0, Z: 0},
        Velocity: astromath.Vector3{X: 0, Y: 9.69/365.25, Z: 0},
    })
    
    // Neptune
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Neptune",
        Mass:     0.0000515,
        Position: astromath.Vector3{X: 30.1, Y: 0, Z: 0},
        Velocity: astromath.Vector3{X: 0, Y: 5.43/365.25, Z: 0},
    })

    // Add ETNOs as massless test particles
    for i, etno := range etnos {
        pos, vel := etno.ToCartesian(muYear)  // Returns AU and AU/year
        
        // CRITICAL: Convert velocity from AU/year to AU/day
        vel = vel.Scale(1.0 / 365.25)
        
        system.Bodies = append(system.Bodies, nbody.Body{
            ID:       fmt.Sprintf("ETNO_%d", i),
            Mass:     0,  // Massless test particles
            Position: pos,  // AU
            Velocity: vel,  // AU/day
        })
    }


    // Vor der Integration: Schwerpunkt/Impuls nullen (verhindert Drift)
    system.RecenterToBarycenter()

    // Zeitschritt wählen (siehe N-Body-Patches aus vorheriger Antwort)
    dtDays := system.ChooseStepForSystem(2000, 5.0, 30.0) // typ. 10–20 Tage gut
    durationDays := duration * 365.25

    history := system.Integrate(durationDays, dtDays /*, snapshotEveryDays e.g. 365.25 */)

    
    // Analyze results
    result := SearchResult{
        Parameters: params,
    }
    
    // Calculate ETNO effects and clustering
    result.ETNOEffects = analyzeETNOChanges(history, etnos)
    result.ClusteringScore = calculateClustering(result.ETNOEffects)
    
    return result
}

func analyzeETNOChanges(history []nbody.Snapshot, initialETNOs []orbital.OrbitalElements) []ETNOEffect {
    if len(history) < 2 {
        return nil
    }
    
    effects := make([]ETNOEffect, 0)
    firstSnap := history[0]
    lastSnap := history[len(history)-1]
    
    // Bodies order: Sun(0), Planet9(1), Jupiter(2), Saturn(3), Neptune(4), ETNOs(5+)
    etnoStart := 5  // Skip Sun, P9, and 3 giant planets
    
    // Gravitational parameter for conversions (year units)
    muYear := 4 * math.Pi * math.Pi  // AU³/(M☉·year²)
    
    for i := 0; i < len(initialETNOs) && etnoStart+i < len(lastSnap.Bodies); i++ {
        initial := firstSnap.Bodies[etnoStart+i]
        final := lastSnap.Bodies[etnoStart+i]
        
        // Skip if positions are invalid
        if initial.Position.IsZero() || final.Position.IsZero() {
            continue
        }
        
        // CRITICAL: Convert velocities from AU/day back to AU/year for orbital elements
        initialVelYear := initial.Velocity.Scale(365.25)
        finalVelYear := final.Velocity.Scale(365.25)
        
        // Convert Cartesian to orbital elements (using year units)
        initialOrb := orbital.CartesianToOrbital(initial.Position, initialVelYear, muYear)
        finalOrb := orbital.CartesianToOrbital(final.Position, finalVelYear, muYear)
        
        // Validate the conversion
        if finalOrb.Eccentricity >= 1.0 || finalOrb.Eccentricity < 0 {
            fmt.Printf("Warning: ETNO_%d bad eccentricity: %.3f\n", i, finalOrb.Eccentricity)
            continue
        }
        
        if finalOrb.SemiMajorAxis <= 0 || finalOrb.SemiMajorAxis > 10000 {
            fmt.Printf("Warning: ETNO_%d bad semi-major axis: %.1f\n", i, finalOrb.SemiMajorAxis)
            continue
        }
        
        // Calculate perihelion changes
        perihelionInitial := initialOrb.SemiMajorAxis * (1 - initialOrb.Eccentricity)
        perihelionFinal := finalOrb.SemiMajorAxis * (1 - finalOrb.Eccentricity)
        perihelionShift := perihelionFinal - perihelionInitial
        
        // Sanity check - perihelion shouldn't change by more than a few AU over thousands of years
        maxReasonableShift := 10.0  // AU
        if math.Abs(perihelionShift) > maxReasonableShift {
            fmt.Printf("Warning: ETNO_%d unrealistic perihelion shift: %.1f AU\n", i, perihelionShift)
            continue
        }
        
        // Calculate inclination change
        inclinationChange := (finalOrb.Inclination - initialOrb.Inclination) * 180.0 / math.Pi
        
        effect := ETNOEffect{
            ObjectID:          fmt.Sprintf("ETNO_%d", i),
            InitialElements:   initialETNOs[i],
            FinalElements:     finalOrb,
            PerihelionShift:   perihelionShift,
            InclinationChange: inclinationChange,
        }
        
        effects = append(effects, effect)
    }
    
    return effects
}

func calculateClustering(effects []ETNOEffect) float64 {
    if len(effects) < 2 {
        return 0.0
    }
    
    // Extract longitude of perihelion values
    longitudes := make([]float64, 0)
    for _, effect := range effects {
        // longitude of perihelion = Ω + ω
        longitude := effect.FinalElements.LongitudeAscendingNode + 
                    effect.FinalElements.ArgumentPerihelion
        
        // Normalize to [0, 2π]
        for longitude > 2*math.Pi {
            longitude -= 2*math.Pi
        }
        for longitude < 0 {
            longitude += 2*math.Pi
        }
        
        longitudes = append(longitudes, longitude)
    }
    
    // Calculate Rayleigh statistic for clustering
    sumCos := 0.0
    sumSin := 0.0
    for _, lon := range longitudes {
        sumCos += math.Cos(lon)
        sumSin += math.Sin(lon)
    }
    
    n := float64(len(longitudes))
    R := math.Sqrt(sumCos*sumCos + sumSin*sumSin) / n
    
    // R ranges from 0 (uniform) to 1 (perfectly clustered)
    return R
}

// addOuterPlanets adds Jupiter, Saturn, Uranus, Neptune
func addOuterPlanets(system *nbody.System) {
    // mu for ToCartesian in year units:
    muYear := 4 * math.Pi * math.Pi // AU^3 / yr^2

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

    for _, p := range planets {
        pos, velYr := p.elem.ToCartesian(muYear) // AU, AU/yr
        velDay := velYr.Scale(1.0 / 365.25)      // AU/day for integrator
        system.Bodies = append(system.Bodies, nbody.Body{
            ID:       p.name,
            Mass:     p.mass,
            Position: pos,
            Velocity: velDay,
        })
    }
}

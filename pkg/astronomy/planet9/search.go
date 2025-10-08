package planet9

import (
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/nbody"
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/orbital"
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
    Parameters    SearchParameters
    ETNOEffects   []ETNOEffect
    ClusteringScore float64
}

type ETNOEffect struct {
    ObjectID       string
    InitialElements orbital.OrbitalElements
    FinalElements   orbital.OrbitalElements
    PerihelionShift float64
    InclinationChange float64
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
    
    // Add Planet 9
    p9Elements := orbital.OrbitalElements{
        SemiMajorAxis:          params.SemiMajorAxis,
        Eccentricity:           params.Eccentricity,
        Inclination:            params.Inclination * math.Pi / 180,
        LongitudeAscendingNode: params.LongitudeAscendingNode * math.Pi / 180,
        ArgumentPerihelion:     params.ArgumentPerihelion * math.Pi / 180,
        MeanAnomaly:            0,
    }
    
    p9Pos, p9Vel := p9Elements.ToCartesian(system.G)
    system.Bodies = append(system.Bodies, nbody.Body{
        ID:       "Planet9",
        Mass:     params.Mass * 3.003e-6, // Earth masses to solar masses
        Position: p9Pos,
        Velocity: p9Vel,
    })
    
    // Add ETNOs
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
    result.ETNOEffects = analyzeETNOChanges(history, etnos)
    result.ClusteringScore = calculateClustering(result.ETNOEffects)
    
    return result
}

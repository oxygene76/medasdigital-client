package main

import (
    "encoding/json"
    "fmt"
    "math"
    "os"
    "os/exec"
    "strings"
    "time"
    
    "github.com/spf13/cobra"
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/planet9"
    "github.com/oxygene76/medasdigital-client/pkg/astronomy/orbital"
)

var planet9Cmd = &cobra.Command{
    Use:   "planet9",
    Short: "Planet 9 orbital search and analysis",
    Long:  `Search for Planet 9 parameters using N-body simulations and ETNO clustering analysis`,
}

var planet9SearchCmd = &cobra.Command{
    Use:   "search [preset]",
    Short: "Search for Planet 9 using presets or custom parameters",
    Long: `
Search for Planet 9 parameters using different strategies:

Presets available:
  batygin_brown_2016  - Original Planet Nine parameters (10 M⊕, 700 AU, e=0.6)
  trujillo_sheppard   - Based on 2012 VP113 discovery (broader search)
  brown_batygin_2021  - Updated parameters with recent discoveries
  custom              - Define your own parameter ranges
  akari2025           - Within IRAS/AKARI constraints (7–17 M⊕, 500–700 AU, q≈300 AU)

Observationally Constrained Ranges (IRAS/AKARI 2025 Analysis):
  --mass          7-17     (Earth masses) - IR detection limits
  --semi-major    500-700  (AU) - Angular motion constraints (3'/year)
  --temperature   28-53    (K) - Black-body emission range
  --eccentricity  0.3-0.6  - Maintains perihelion 200-350 AU
  --inclination   10-40    (degrees) - Non-detection in ecliptic
  --node          0-360    (degrees) - Longitude of ascending node
  --omega         100-200  (degrees) - ETNO clustering around 150°

Key Constraints from IR Observations:
  - Minimum mass: 7 M⊕ (IRAS/AKARI detection threshold)
  - Maximum mass: 17 M⊕ (Neptune mass, would be detected)
  - Distance: 500-700 AU (42-69.6 arcmin motion over 23 years)
  - Angular motion: ~3 arcmin/year at 700 AU
  - Temperature: 28-53 K (equilibrium temperature range)
  - Perihelion: >200 AU (non-detection constraint)

Examples:
  # Use Batygin & Brown 2016 preset
  medasdigital-client planet9 search batygin_brown_2016

  # Observationally constrained search
  medasdigital-client planet9 search custom --mass 7-17 --semi-major 500-700

  # Optimal search (best constraints)
  medasdigital-client planet9 search custom --mass 10-15 --semi-major 600-700 --eccentricity 0.4-0.6

  # Quick test with reduced grid
  medasdigital-client planet9 search --quick
`,
    Args: cobra.MaximumNArgs(1),
    RunE: runPlanet9Search,
}

var planet9JobCmd = &cobra.Command{
    Use:   "submit-job [preset]",
    Short: "Submit Planet 9 search job to blockchain",
    Long: `
Submit a Planet 9 search job to be processed by the distributed computing network.

The job will be assigned to a provider who will run the N-body simulations
and return the best Planet 9 parameters found.
`,
    Args: cobra.MaximumNArgs(1),
    RunE: submitPlanet9Job,
}

var planet9AnalyzeCmd = &cobra.Command{
    Use:   "analyze [result-file]",
    Short: "Analyze Planet 9 search results",
    Long:  `Analyze and visualize results from a Planet 9 search`,
    Args:  cobra.ExactArgs(1),
    RunE:  analyzePlanet9Results,
}

var planet9TestCmd = &cobra.Command{
    Use:   "test",
    Short: "Test Planet 9 calculations with a simple scenario",
    Long:  `Run a quick test of the N-body integrator with a simple Planet 9 scenario`,
    RunE:  runPlanet9Test,
}

var planet9RangesCmd = &cobra.Command{
    Use:   "ranges",
    Short: "Show scientifically plausible Planet 9 parameter ranges",
    Long:  `Display the current scientific understanding of where Planet 9 might exist`,
    RunE:  showPlanet9Ranges,
}

// Command-line flags for planet9
var (
    // Search parameters
    p9MassRange      string
    p9SemiMajorRange string
    p9EccRange       string
    p9IncRange       string
    p9NodeRange      string
    p9OmegaRange     string
    
    // Grid resolution
    p9GridPoints     int
    p9QuickSearch    bool
    p9FineSearch     bool
    
    // Simulation options
    p9SimYears       float64
    p9IncludeKozai   bool
    p9IncludeResonance bool
    
    // Output options
    p9OutputFile     string
    p9OutputFormat   string
    p9ShowProgress   bool
    
    // Job submission
    p9JobPayment     string
    p9JobPriority    string

    p9SnapshotEveryKyr float64
    p9SnapshotFile     string
)

func init() {
    // Add planet9 command to root (rootCmd should be defined in main.go)
    rootCmd.AddCommand(planet9Cmd)
    
    // Add subcommands
    planet9Cmd.AddCommand(planet9SearchCmd)
    planet9Cmd.AddCommand(planet9JobCmd)
    planet9Cmd.AddCommand(planet9AnalyzeCmd)
    planet9Cmd.AddCommand(planet9TestCmd)
    planet9Cmd.AddCommand(planet9RangesCmd)
    
    // Search command flags
    planet9SearchCmd.Flags().StringVar(&p9MassRange, "mass", "", "Mass range in Earth masses (e.g., 5-10)")
    planet9SearchCmd.Flags().StringVar(&p9SemiMajorRange, "semi-major", "", "Semi-major axis range in AU (e.g., 400-800)")
    planet9SearchCmd.Flags().StringVar(&p9EccRange, "eccentricity", "", "Eccentricity range (e.g., 0.2-0.5)")
    planet9SearchCmd.Flags().StringVar(&p9IncRange, "inclination", "", "Inclination range in degrees (e.g., 15-30)")
    planet9SearchCmd.Flags().StringVar(&p9NodeRange, "node", "", "Longitude of ascending node range (e.g., 85-115)")
    planet9SearchCmd.Flags().StringVar(&p9OmegaRange, "omega", "", "Argument of perihelion range (e.g., 140-160)")
    
    planet9SearchCmd.Flags().IntVar(&p9GridPoints, "grid-points", 0, "Total grid points (overrides resolution)")
    planet9SearchCmd.Flags().BoolVar(&p9QuickSearch, "quick", false, "Quick search with coarse grid")
    planet9SearchCmd.Flags().BoolVar(&p9FineSearch, "fine", false, "Fine search with dense grid")
    
    planet9SearchCmd.Flags().Float64Var(&p9SimYears, "sim-years", 1000, "Simulation duration in years")
    planet9SearchCmd.Flags().BoolVar(&p9IncludeKozai, "kozai", false, "Test for Kozai-Lidov oscillations")
    planet9SearchCmd.Flags().BoolVar(&p9IncludeResonance, "resonance", false, "Test for mean-motion resonances")
    
    planet9SearchCmd.Flags().StringVar(&p9OutputFile, "output", "", "Save results to file")
    planet9SearchCmd.Flags().StringVar(&p9OutputFormat, "format", "json", "Output format (json, csv, summary)")
    planet9SearchCmd.Flags().BoolVar(&p9ShowProgress, "progress", true, "Show progress bar")
    
    // Job submission flags
    planet9JobCmd.Flags().StringVar(&p9JobPayment, "payment", "10000000umedas", "Payment amount")
    planet9JobCmd.Flags().StringVar(&p9JobPriority, "priority", "normal", "Job priority (low, normal, high)")

    planet9SearchCmd.Flags().Float64Var(&p9SnapshotEveryKyr, "snapshot-every-kyr", 0.2, "Snapshot cadence in kyr (0 = disable)")
    planet9SearchCmd.Flags().StringVar(&p9SnapshotFile, "snapshot-file", "snapshots.jsonl", "Path for streamed JSONL snapshots")
}

func runPlanet9Search(cmd *cobra.Command, args []string) error {
    // Determine preset
    preset := planet9.PresetCustom
    if len(args) > 0 {
        switch args[0] {
        case "batygin_brown_2016":
            preset = planet9.PresetBatyginBrown2016
        case "trujillo_sheppard":
            preset = planet9.PresetTrujilloSheppard  
        case "brown_batygin_2021":
            preset = planet9.PresetBrownBatygin2021
        case "akari2025":
            preset = planet9.PresetAkari2025
        default:
            if args[0] != "custom" {
                return fmt.Errorf("unknown preset: %s", args[0])
            }
        }
    }
    
    // Get preset parameters or build custom
    var searchParams planet9.SearchParameters
    if preset != planet9.PresetCustom {
        searchParams = planet9.GetPresetParameters(preset)
    } else {
        searchParams = buildCustomParameters()
    }
    
    // Adjust simulation time for quick/fine search
    simDuration := p9SimYears
    if p9QuickSearch {
        simDuration = 100  // 100 years for quick test
    } else if p9FineSearch {
        simDuration = 10000 // 10,000 years for fine search
    }
    
    // Load TNO data
    dataFile := "data/solar_system_jpl.json"
    if _, err := os.Stat(dataFile); os.IsNotExist(err) {
        fmt.Println("\n⚠ TNO data file not found. Downloading from JPL...")
        if err := downloadJPLData(); err != nil {
            return fmt.Errorf("failed to download data: %w", err)
        }
    }
    
    // Load ETNOs from data file
    etnos, err := loadETNOData(dataFile)
    if err != nil {
        return fmt.Errorf("failed to load ETNO data: %w", err)
    }
    
    fmt.Println("========================================")
    fmt.Println("   PLANET 9 ORBITAL PARAMETER SEARCH")
    fmt.Println("========================================")
    fmt.Printf("\nPreset: %s\n", preset)
    fmt.Printf("Parameters:\n")
    fmt.Printf("  Mass: %.1f Earth masses\n", searchParams.Mass)
    fmt.Printf("  Semi-major axis: %.0f AU\n", searchParams.SemiMajorAxis)
    fmt.Printf("  Eccentricity: %.2f\n", searchParams.Eccentricity)
    fmt.Printf("  Inclination: %.1f°\n", searchParams.Inclination)
    fmt.Printf("  Simulation: %.0f years\n", simDuration)
    fmt.Printf("  ETNOs loaded: %d\n\n", len(etnos))
    
    // Run simulation
    startTime := time.Now()
    fmt.Println("Running N-body simulation...")
    
    result := planet9.RunSimulation(searchParams, etnos, simDuration)
    
    elapsed := time.Since(startTime)
    
    // Display results
    fmt.Printf("\n=== RESULTS ===\n")
    fmt.Printf("Clustering Score: %.3f\n", result.ClusteringScore)
    fmt.Printf("Compute Time: %v\n\n", elapsed)
    
    // Show ETNO effects
    if len(result.ETNOEffects) > 0 {
        fmt.Println("ETNO Orbital Changes:")
        fmt.Println("Object          Perihelion Shift  Inclination Change")
        fmt.Println("------------------------------------------------------")
        for i, effect := range result.ETNOEffects {
            if i >= 10 {
                break // Show only first 10
            }
            fmt.Printf("%-15s  %+6.2f AU         %+6.2f°\n",
                effect.ObjectID,
                effect.PerihelionShift,
                effect.InclinationChange)
        }
    }
    
    // Save results if requested
    if p9OutputFile != "" {
        if err := saveSearchResults(&result, p9OutputFile, p9OutputFormat); err != nil {
            return fmt.Errorf("failed to save results: %w", err)
        }
        fmt.Printf("\nResults saved to: %s\n", p9OutputFile)
    }
    
    return nil
}

func runPlanet9Test(cmd *cobra.Command, args []string) error {
    fmt.Println("Running Planet 9 test simulation...")
    
    // Simple test with one ETNO
    testParams := planet9.SearchParameters{
        Mass:                   10.0,
        SemiMajorAxis:          700.0,
        Eccentricity:           0.6,
        Inclination:            30.0,
        LongitudeAscendingNode: 100.0,
        ArgumentPerihelion:     150.0,
    }
    
    // Create a test ETNO (Sedna-like)
    testETNO := orbital.OrbitalElements{
        SemiMajorAxis:          483.3,
        Eccentricity:           0.8496,
        Inclination:            0.2082, // 11.93 degrees in radians
        LongitudeAscendingNode: 2.5166, // 144.26 degrees
        ArgumentPerihelion:     5.4280, // 311.02 degrees
        MeanAnomaly:            6.2657, // 359.46 degrees
    }
    
    result := planet9.RunSimulation(testParams, []orbital.OrbitalElements{testETNO}, 100)
    
    fmt.Printf("Test completed!\n")
    fmt.Printf("Clustering score: %.3f\n", result.ClusteringScore)
    
    return nil
}

func buildCustomParameters() planet9.SearchParameters {
    // Parse range strings and use middle value
    // Updated defaults based on IRAS/AKARI observational constraints
    mass := parseRangeMiddle(p9MassRange, 12.0)      // Middle of 7-17 M⊕
    semiMajor := parseRangeMiddle(p9SemiMajorRange, 600.0)  // Middle of 500-700 AU
    ecc := parseRangeMiddle(p9EccRange, 0.45)       // Middle of 0.3-0.6
    inc := parseRangeMiddle(p9IncRange, 25.0)       // Middle of 10-40°
    node := parseRangeMiddle(p9NodeRange, 100.0)    
    omega := parseRangeMiddle(p9OmegaRange, 150.0)  // ETNO clustering
    
    // Validate and warn about implausible values
    validateParameters(mass, semiMajor, ecc, inc)
    
    return planet9.SearchParameters{
        Mass:                   mass,
        SemiMajorAxis:          semiMajor,
        Eccentricity:           ecc,
        Inclination:            inc,
        LongitudeAscendingNode: node,
        ArgumentPerihelion:     omega,
    }
}

func validateParameters(mass, semiMajor, ecc, inc float64) {
    warnings := []string{}
    
    // Check mass (7-17 Earth masses from IRAS/AKARI observational constraints)
    if mass < 7 {
        warnings = append(warnings, fmt.Sprintf("Mass %.1f M⊕ below detection threshold - IRAS/AKARI analysis shows minimum 7 M⊕ needed", mass))
    } else if mass > 17 {
        warnings = append(warnings, fmt.Sprintf("Mass %.1f M⊕ exceeds Neptune mass - would have been detected in IR surveys", mass))
    }
    
    // Check semi-major axis (500-700 AU from observational analysis)
    if semiMajor < 500 {
        warnings = append(warnings, fmt.Sprintf("Semi-major axis %.0f AU too close - would produce >69.6' motion over 23 years", semiMajor))
    } else if semiMajor > 700 {
        warnings = append(warnings, fmt.Sprintf("Semi-major axis %.0f AU may be too far - produces <42' motion, harder to detect", semiMajor))
    }
    
    // Check perihelion (must be >200 AU to avoid past detection)
    perihelion := semiMajor * (1 - ecc)
    if perihelion < 200 {
        warnings = append(warnings, fmt.Sprintf("Perihelion %.0f AU too close - would have been detected by IR surveys", perihelion))
    }
    
    // Check eccentricity to maintain proper perihelion
    minEcc := 1.0 - (350.0 / semiMajor)  // Keep perihelion < 350 AU for ETNO influence
    maxEcc := 1.0 - (200.0 / semiMajor)  // Keep perihelion > 200 AU for non-detection
    if ecc < minEcc {
        warnings = append(warnings, fmt.Sprintf("Eccentricity %.2f too low - perihelion %.0f AU won't affect ETNOs", ecc, perihelion))
    } else if ecc > maxEcc {
        warnings = append(warnings, fmt.Sprintf("Eccentricity %.2f too high - perihelion %.0f AU would be detected", ecc, perihelion))
    }
    
    // Check inclination (10-40 degrees plausible)
    if inc < 10 {
        warnings = append(warnings, "Inclination <10° unlikely - would have been detected in ecliptic surveys")
    } else if inc > 40 {
        warnings = append(warnings, "Inclination >40° unlikely - difficult to explain formation")
    }
    
    // Temperature check (28-53 K from paper)
    expectedTemp := calculateTemperature(semiMajor)
    if expectedTemp < 28 || expectedTemp > 53 {
        warnings = append(warnings, fmt.Sprintf("Temperature %.0f K outside 28-53 K range from IR observations", expectedTemp))
    }
    
    // Display warnings if any
    if len(warnings) > 0 {
        fmt.Println("\n⚠️  Parameter Warnings:")
        for _, w := range warnings {
            fmt.Printf("   • %s\n", w)
        }
        fmt.Println("\n   Observationally constrained ranges (IRAS/AKARI 2025):")
        fmt.Println("   • Mass: 7-17 M⊕ (IR detection limits)")
        fmt.Println("   • Semi-major: 500-700 AU (angular motion constraints)")
        fmt.Println("   • Temperature: 28-53 K (black-body emission)")
        fmt.Println("   • Angular motion: ~3'/year at 700 AU")
        fmt.Println()
    }
}

func calculateTemperature(semiMajor float64) float64 {
    // Simple temperature estimate based on distance
    // T ∝ 1/√d for equilibrium temperature
    return 50.0 * math.Sqrt(600.0/semiMajor)  // Normalized to ~50K at 600 AU
}

func parseRangeMiddle(s string, defaultVal float64) float64 {
    if s == "" {
        return defaultVal
    }
    
    parts := strings.Split(s, "-")
    if len(parts) != 2 {
        // Try to parse as single value
        var val float64
        if _, err := fmt.Sscanf(s, "%f", &val); err == nil {
            return val
        }
        return defaultVal
    }
    
    var min, max float64
    fmt.Sscanf(parts[0], "%f", &min)
    fmt.Sscanf(parts[1], "%f", &max)
    
    if min == 0 && max == 0 {
        return defaultVal
    }
    
    return (min + max) / 2.0 // Return middle of range
}

func loadETNOData(dataFile string) ([]orbital.OrbitalElements, error) {
    data, err := os.ReadFile(dataFile)
    if err != nil {
        return nil, err
    }
    
    var solarSystem struct {
        ETNOs []struct {
            OrbitalElements struct {
                SemiMajorAxis          float64 `json:"semimajor_axis"`
                Eccentricity           float64 `json:"eccentricity"`
                Inclination            float64 `json:"inclination"`
                LongitudeAscendingNode float64 `json:"longitude_ascending_node"`
                ArgumentPerihelion     float64 `json:"argument_perihelion"`
                MeanAnomaly            float64 `json:"mean_anomaly"`
            } `json:"orbital_elements"`
        } `json:"etnos"`
    }
    
    if err := json.Unmarshal(data, &solarSystem); err != nil {
        return nil, err
    }
    
    // Convert to orbital.OrbitalElements with radians
    etnos := make([]orbital.OrbitalElements, 0)
    for _, e := range solarSystem.ETNOs {
        // Convert degrees to radians
        etnos = append(etnos, orbital.OrbitalElements{
            SemiMajorAxis:          e.OrbitalElements.SemiMajorAxis,
            Eccentricity:           e.OrbitalElements.Eccentricity,
            Inclination:            e.OrbitalElements.Inclination * 0.017453293,            // deg to rad
            LongitudeAscendingNode: e.OrbitalElements.LongitudeAscendingNode * 0.017453293,
            ArgumentPerihelion:     e.OrbitalElements.ArgumentPerihelion * 0.017453293,
            MeanAnomaly:            e.OrbitalElements.MeanAnomaly * 0.017453293,
        })
    }
    
    return etnos, nil
}

func submitPlanet9Job(cmd *cobra.Command, args []string) error {
    cfg := loadConfig()
    
    // Build job parameters
    preset := "custom"
    if len(args) > 0 {
        preset = args[0]
    }
    
    params := map[string]interface{}{
        "service_type": "planet9_search",
        "preset": preset,
        "sim_years": p9SimYears,
    }
    
    paramsJSON, _ := json.Marshal(params)
    
    // Use hardcoded contract address if not in config
    contractAddr := "medas1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3s3cca97"
    
    fmt.Println("Submitting Planet 9 search job to blockchain...")
    fmt.Printf("  Contract: %s\n", contractAddr)
    fmt.Printf("  Payment: %s\n", p9JobPayment)
    fmt.Printf("  Parameters: %s\n", string(paramsJSON))
    
    // Use provider key from config if available
    keyName := "test-client"
    if cfg.Provider.KeyName != "" {
        keyName = cfg.Provider.KeyName
    }
    
    // Build transaction command
    execCmd := exec.Command(
        "medasdigitald", "tx", "wasm", "execute",
        contractAddr,
        fmt.Sprintf(`{"submit_job":{"service_type":"planet9_search","parameters":"%s","max_price":"1000000","auto_accept":true}}`,
            strings.ReplaceAll(string(paramsJSON), `"`, `\"`)),
        "--from", keyName,
        "--amount", p9JobPayment,
        "--gas", "auto",
        "--gas-adjustment", "1.3",
        "--gas-prices", "0.025umedas",
        "--keyring-backend", cfg.Provider.KeyringBackend,
        "--node", cfg.Chain.RPCEndpoint,
        "--chain-id", cfg.Chain.ID,
        "-y",
    )
    
    output, err := execCmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("transaction failed: %w\n%s", err, output)
    }
    
    fmt.Println("\n✓ Job submitted successfully!")
    fmt.Printf("Transaction output:\n%s\n", output)
    
    return nil
}

func analyzePlanet9Results(cmd *cobra.Command, args []string) error {
    resultFile := args[0]
    
    data, err := os.ReadFile(resultFile)
    if err != nil {
        return fmt.Errorf("failed to read results: %w", err)
    }
    
    var result planet9.SearchResult
    if err := json.Unmarshal(data, &result); err != nil {
        return fmt.Errorf("failed to parse results: %w", err)
    }
    
    fmt.Println("Planet 9 Results Analysis")
    fmt.Println("========================")
    fmt.Printf("\nParameters:\n")
    fmt.Printf("  Mass: %.1f Earth masses\n", result.Parameters.Mass)
    fmt.Printf("  Semi-major axis: %.0f AU\n", result.Parameters.SemiMajorAxis)
    fmt.Printf("  Eccentricity: %.3f\n", result.Parameters.Eccentricity)
    fmt.Printf("  Inclination: %.1f°\n", result.Parameters.Inclination)
    fmt.Printf("\nScores:\n")
    fmt.Printf("  Clustering: %.3f\n", result.ClusteringScore)
    fmt.Printf("\nETNO Effects: %d objects analyzed\n", len(result.ETNOEffects))
    
    return nil
}

func saveSearchResults(result *planet9.SearchResult, filename, format string) error {
    switch format {
    case "json":
        data, err := json.MarshalIndent(result, "", "  ")
        if err != nil {
            return err
        }
        return os.WriteFile(filename, data, 0644)
        
    case "summary":
        summary := fmt.Sprintf(`Planet 9 Search Results
======================
Mass: %.1f Earth masses
Semi-major axis: %.0f AU
Eccentricity: %.3f
Inclination: %.1f°
Clustering Score: %.3f
ETNOs Analyzed: %d
`, 
            result.Parameters.Mass,
            result.Parameters.SemiMajorAxis,
            result.Parameters.Eccentricity,
            result.Parameters.Inclination,
            result.ClusteringScore,
            len(result.ETNOEffects))
        
        return os.WriteFile(filename, []byte(summary), 0644)
        
    default:
        return fmt.Errorf("unknown format: %s", format)
    }
}

func downloadJPLData() error {
    // Execute Python script to download data
    cmd := exec.Command("python3", "scripts/fetch_jpl_api.py")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func showPlanet9Ranges(cmd *cobra.Command, args []string) error {
    fmt.Println(`
╔════════════════════════════════════════════════════════════════════╗
║                  PLANET 9 PARAMETER RANGES                        ║
║         Observational Constraints from IR Surveys (2025)          ║
╚════════════════════════════════════════════════════════════════════╝

OBSERVATIONALLY CONSTRAINED (IRAS/AKARI Analysis):
────────────────────────────────────────────────────
  Mass:           7-17 Earth masses
    • Based on IR detection limits
    • <7 M⊕: Below IRAS/AKARI threshold
    • >17 M⊕: Neptune mass, would be detected
    
  Semi-major:     500-700 AU
    • Produces 42-69.6' motion over 23 years
    • ~3 arcmin/year angular motion
    • Optimal for detection vs influence balance
    
  Temperature:    28-53 K
    • Black-body equilibrium range
    • Consistent with far-IR observations
    
  Eccentricity:   0.3-0.6
    • Keeps perihelion 200-350 AU
    • Explains ETNO perihelion detachment
    
  Inclination:    10-40°
    • Explains non-detection in surveys
    • Consistent with ETNO orbital distribution

THEORETICAL PREDICTIONS (Dynamics-based):
────────────────────────────────────────────────────
  Batygin & Brown (2016):
    • 10 M⊕, 700 AU, e=0.6, i=30°
    • Based on ETNO clustering analysis
    
  Brown & Batygin (2021):
    • 6.2 M⊕, 380 AU, e=0.2, i=16°
    • Refined with more ETNO discoveries
    
  Optimal Search Region (Combined):
    • Mass: 10-15 M⊕
    • Semi-major: 600-700 AU
    • Eccentricity: 0.4-0.6
    • Temperature: 35-45 K

KEY OBSERVATIONAL FACTS:
────────────────────────────────────────────────────
  ✓ Angular motion: ~3'/year at 700 AU
  ✓ 23-year motion: 42-69.6 arcmin (500-700 AU)
  ✓ Perihelion must be >200 AU (non-detection)
  ✓ Perihelion should be <350 AU (ETNO influence)
  ✓ IR flux drops as d^-2 (vs d^-4 for optical)
  ✓ Current location: likely near aphelion

SEARCH STRATEGY:
────────────────────────────────────────────────────
  Conservative (highest probability):
    medasdigital-client planet9 search custom \
      --mass 10-15 --semi-major 600-700 --eccentricity 0.4-0.6
      
  Full observational range:
    medasdigital-client planet9 search custom \
      --mass 7-17 --semi-major 500-700 --eccentricity 0.3-0.6
      
  Quick test:
    medasdigital-client planet9 search batygin_brown_2016 --quick
`)
    return nil
}

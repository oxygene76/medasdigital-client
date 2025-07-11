package main

import (
	"fmt"
	"log"
	"os"

	"github.com/oxygene76/medasdigital-client/pkg/client"
	"github.com/oxygene76/medasdigital-client/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "medasdigital-client",
		Short: "MedasDigital Client for astronomical analysis",
		Long: `A distributed astronomical analysis client for Planet 9 and TNO research
using the MedasDigital blockchain infrastructure with GPU acceleration.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.medasdigital/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add commands
	rootCmd.AddCommand(
		initCmd(),
		registerCmd(),
		analyzeCmd(),
		trainCmd(),
		statusCmd(),
		resultsCmd(),
		queryCmd(),
		gpuCmd(),
	)

	// Initialize configuration
	cobra.OnInitialize(initConfig)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(home + "/.medasdigital")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize client configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID, _ := cmd.Flags().GetString("chain-id")
			rpcEndpoint, _ := cmd.Flags().GetString("rpc-endpoint")
			
			config := &utils.Config{
				Chain: utils.ChainConfig{
					ID:          chainID,
					RPCEndpoint: rpcEndpoint,
				},
				Client: utils.ClientConfig{
					Capabilities: []string{
						"orbital_dynamics",
						"photometric_analysis",
						"gpu_compute",
					},
				},
				GPU: utils.GPUConfig{
					Enabled:           true,
					CUDADevices:       []int{0},
					MemoryLimitGB:     24,
					ComputeCapability: "8.6",
				},
			}

			return utils.SaveConfig(config)
		},
	}

	cmd.Flags().String("chain-id", "medasdigital-2", "Chain ID")
	cmd.Flags().String("rpc-endpoint", "https://rpc.medas-digital.io:26657", "RPC endpoint")

	return cmd
}

func registerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register client on blockchain",
		RunE: func(cmd *cobra.Command, args []string) error {
			capabilities, _ := cmd.Flags().GetStringSlice("capabilities")
			metadata, _ := cmd.Flags().GetString("metadata")
			from, _ := cmd.Flags().GetString("from")

			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.Register(capabilities, metadata, from)
		},
	}

	cmd.Flags().StringSlice("capabilities", []string{}, "Client capabilities")
	cmd.Flags().String("metadata", "", "Client metadata")
	cmd.Flags().String("from", "", "Wallet key name")
	cmd.MarkFlagRequired("capabilities")
	cmd.MarkFlagRequired("from")

	return cmd
}

func analyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Run analysis workflows",
	}

	cmd.AddCommand(
		orbitalDynamicsCmd(),
		photometricCmd(),
		clusteringCmd(),
		aiDetectionCmd(),
	)

	return cmd
}

func orbitalDynamicsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "orbital-dynamics",
		Short: "Run orbital dynamics analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			output, _ := cmd.Flags().GetString("output")
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.AnalyzeOrbitalDynamics(input, output)
		},
	}
}

func photometricCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "photometric",
		Short: "Run photometric analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			surveyData, _ := cmd.Flags().GetString("survey-data")
			targetList, _ := cmd.Flags().GetString("target-list")
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.AnalyzePhotometric(surveyData, targetList)
		},
	}
}

func clusteringCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clustering",
		Short: "Run clustering analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.AnalyzeClustering()
		},
	}
}

func aiDetectionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ai-detection",
		Short: "Run AI-powered object detection",
		RunE: func(cmd *cobra.Command, args []string) error {
			model, _ := cmd.Flags().GetString("model")
			surveyImages, _ := cmd.Flags().GetString("survey-images")
			gpuAccel, _ := cmd.Flags().GetBool("gpu-acceleration")
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.AIDetection(model, surveyImages, gpuAccel)
		},
	}
}

func trainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "train",
		Short: "Train machine learning models",
	}

	cmd.AddCommand(
		trainDeepDetectorCmd(),
		trainAnomalyDetectorCmd(),
	)

	return cmd
}

func trainDeepDetectorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deep-detector",
		Short: "Train deep learning object detector",
		RunE: func(cmd *cobra.Command, args []string) error {
			trainingData, _ := cmd.Flags().GetString("training-data")
			architecture, _ := cmd.Flags().GetString("model-architecture")
			gpuDevices, _ := cmd.Flags().GetIntSlice("gpu-devices")
			batchSize, _ := cmd.Flags().GetInt("batch-size")
			epochs, _ := cmd.Flags().GetInt("epochs")
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.TrainDeepDetector(trainingData, architecture, gpuDevices, batchSize, epochs)
		},
	}
}

func trainAnomalyDetectorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anomaly-detector",
		Short: "Train anomaly detection model",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.TrainAnomalyDetector()
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check client status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.Status()
		},
	}
}

func resultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "results",
		Short: "View analysis results",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.Results(limit)
		},
	}
}

func queryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Query blockchain data",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			queryType := args[0]
			queryID := args[1]
			
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.Query(queryType, queryID)
		},
	}
}

func gpuCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gpu",
		Short: "GPU management commands",
	}

	cmd.AddCommand(
		gpuStatusCmd(),
		gpuBenchmarkCmd(),
	)

	return cmd
}

func gpuStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check GPU status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.GPUStatus()
		},
	}
}

func gpuBenchmarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "benchmark",
		Short: "Run GPU benchmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewMedasDigitalClient()
			if err != nil {
				return err
			}

			return client.GPUBenchmark()
		},
	}
}

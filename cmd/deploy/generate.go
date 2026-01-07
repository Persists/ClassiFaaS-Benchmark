package main

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/deployment"
	"ClassiFaaS/internal/globals"
	"ClassiFaaS/internal/utils"
	"fmt"
)

func runGenerate() {
	ep := utils.NewEventLogger()
	defer ep.Close()

	cfg, err := config.LoadDeployConfig(configPath)
	if err != nil {
		ep.SendEvent(utils.SeverityError, "load_config", fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	dplOrchClient, err := deployment.NewDeployOrchestratorClient(*cfg)
	if err != nil {
		ep.SendEvent(utils.SeverityError, "init_deployer", fmt.Sprintf("Failed to initialize deployer: %v", err))
		return
	}

	filterFunction := func(df deployment.DeployedFunction) bool {
		if cfg.MemorySizes == nil {
			return cfg.Benchmarks[df.Benchmark] > 0
		}

		for _, size := range cfg.MemorySizes {
			if df.Memory == size {
				return cfg.Benchmarks[df.Benchmark] > 0
			}
		}

		// Not in allowed memory sizes
		return false
	}

	// Transform the configuration so that it can be used for benchmarking
	transformFunction := func(df deployment.DeployedFunction) config.BenchmarkFunctionConfig {

		parameterizedUrl := fmt.Sprintf("%s?parameter=%d", df.URL, cfg.Benchmarks[df.Benchmark])

		return config.BenchmarkFunctionConfig{
			Name: fmt.Sprintf("%s-%s-%d", df.Provider, df.Benchmark, df.Memory),
			URL:  parameterizedUrl,
			Auth: config.AuthConfig{
				Key:   globals.AuthKeys[df.Provider],
				Value: df.Auth,
			},
			Provider: df.Provider,
			Region:   df.Region,
			MemSize:  df.Memory,
		}
	}

	// Generate benchmark functions based on deployed functions
	benchmarkFunctions, err := dplOrchClient.GenerateBenchmarkFunctions(ep, filterFunction, transformFunction)
	if err != nil {
		ep.SendEvent(utils.SeverityError, "generate_benchmark_functions", fmt.Sprintf("Failed to generate benchmark functions: %v", err))
		return
	}
	generatedBenchmarkConfig := config.BenchmarkConfig{
		WorkloadParameters: cfg.WorkloadParameters,
		Functions:          benchmarkFunctions,
	}

	generatedBenchmarkConfig.WriteToFile("configs/generated.yaml")

	ep.SendEvent(utils.SeverityInfo, "generate_benchmark_config", "Generated benchmark config at configs/generated.yaml")
}

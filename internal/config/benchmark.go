package config

import (
	"ClassiFaaS/internal/globals"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type BenchmarkConfig struct {
	WorkloadParameters WorkloadParameters        `yaml:"workload"`
	Functions          []BenchmarkFunctionConfig `yaml:"functions"`
}

type WorkloadParameters struct {
	ParallelRequests  int    `yaml:"parallelRequests"`
	TotalRequests     int    `yaml:"totalRequests"`
	RetriesPerRequest int    `yaml:"retriesPerRequest"`
	ResultFolder      string `yaml:"resultFolder"`
}

type BenchmarkFunctionConfig struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	Region   string `yaml:"region"`
	MemSize  int    `yaml:"memorySize"`
	URL      string `yaml:"URL"`

	Auth AuthConfig `yaml:"auth"`
}

type AuthConfig struct {
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// LoadBenchmarkConfig loads the config from a YAML file.
func LoadBenchmarkConfig(path string) (*BenchmarkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg BenchmarkConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}
	return &cfg, nil
}

// Validate checks if the config is valid.
func (c *BenchmarkConfig) validate() error {

	fmt.Println("Validating functions...")
	if len(c.Functions) == 0 {
		return fmt.Errorf("at least one function must be specified")
	}
	for i, fn := range c.Functions {
		if _, ok := allowedProviders[fn.Provider]; !ok {
			return fmt.Errorf("function[%d]: invalid provider '%s'", i, fn.Provider)
		}
		if fn.Region == "" {
			return fmt.Errorf("function[%d]: region must not be empty", i)
		}
		if fn.MemSize <= 0 {
			return fmt.Errorf("function[%d]: memorySize must be > 0", i)
		}
		if fn.URL == "" {
			return fmt.Errorf("function[%d]: URL must not be empty", i)
		}
		if fn.Auth.Key == "" && fn.Provider != "alibaba" {
			return fmt.Errorf("function[%d]: auth.key must not be empty", i)
		}
		if err := validateAuthKeys(fn.Auth.Key, fn.Provider); err != nil {
			return fmt.Errorf("function[%d]: %v", i, err)
		}
	}

	// all functions should have a unique name including alphabetic and numeric characters, dashes and underscores
	functionNames := make(map[string]bool)
	for i, fn := range c.Functions {
		if _, exists := functionNames[fn.Name]; exists {
			return fmt.Errorf("function[%d]: duplicate function name '%s' with region '%s'", i, fn.Name, fn.Region)
		}
		functionNames[fmt.Sprintf("%s:%s", fn.Name, fn.Region)] = true
	}
	return nil
}

func validateWorkloadParameters(param WorkloadParameters) error {
	if param.ParallelRequests <= 0 {
		return fmt.Errorf("workload.parallelRequests must be greater than 0")
	}
	if param.TotalRequests <= 0 {
		return fmt.Errorf("workload.totalRequests must be greater than 0")
	}
	if param.RetriesPerRequest <= 0 {
		return fmt.Errorf("workload.retriesPerRequest must be greater than 0")
	}
	if param.ResultFolder == "" {
		return fmt.Errorf("workload.resultFolder must not be empty")
	}

	return nil
}
func validateAuthKeys(key, provider string) error {
	expectedKey, ok := globals.AuthKeys[provider]
	if !ok {
		return fmt.Errorf("unsupported provider %q for auth key validation", provider)
	}

	if key != expectedKey {
		return fmt.Errorf("invalid auth.key for provider %s: expected '%s', got '%s'", provider, expectedKey, key)
	}
	return nil
}

func (c *BenchmarkConfig) WriteToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to file: %v", err)
	}
	return nil
}

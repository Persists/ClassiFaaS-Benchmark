package config

import (
	"ClassiFaaS/internal/globals"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DeploymentConfig struct {
	Provider string `yaml:"provider"`
	Region   string `yaml:"region"`
}

var allowedProviders = map[string]struct{}{
	"gcp":     {},
	"aws":     {},
	"azure":   {},
	"alibaba": {},
}

type DeployConfig struct {
	WorkloadParameters WorkloadParameters `yaml:"workload"`
	Benchmarks         map[string]int     `yaml:"benchmarks"`
	MemorySizes        []int              `yaml:"memorySizes"`
	Deployments        []struct {
		DeploymentConfig `yaml:",inline"`
	} `yaml:"deployments"`
}

// LoadDeployConfig loads and validates the deployment configuration from the specified YAML file.
func LoadDeployConfig(path string) (*DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg DeployConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := validateWorkloadParameters(cfg.WorkloadParameters); err != nil {
		return nil, fmt.Errorf("invalid workload parameters: %v", err)
	}

	if len(cfg.Benchmarks) == 0 {
		fmt.Println("⚠️  No benchmarks defined — defaulting to none.")
	}

	if len(cfg.MemorySizes) == 0 {
		fmt.Println("⚠️  No memory sizes configured — using all memory sizes.")
		cfg.MemorySizes = nil // nil means “all” later in your logic
	}

	fmt.Println("Validating deploy-only config...")
	for _, deploy := range cfg.Deployments {
		if _, ok := allowedProviders[deploy.Provider]; !ok {
			return nil, fmt.Errorf("invalid provider '%s'", deploy.Provider)
		}
		if err := deploy.DeploymentConfig.validate(deploy.Provider); err != nil {
			return nil, fmt.Errorf("invalid deployment config for provider %s: %w", deploy.Provider, err)
		}
	}

	return &cfg, nil
}

// validate checks if the DeploymentConfig is valid for the given provider.
func (c *DeploymentConfig) validate(provider string) error {

	if _, ok := allowedProviders[provider]; !ok {
		return fmt.Errorf("invalid provider '%s'", provider)
	}
	if c.Region == "" {
		return fmt.Errorf("region is required in deployment config")
	}

	switch provider {
	case "gcp":
		for _, r := range globals.ValidGCPRegions {
			if c.Region == r {
				goto RegionValid
			}
		}
		return fmt.Errorf("invalid GCP region: %s", c.Region)
	case "aws":
		for _, r := range globals.ValidAWSRegions {
			if c.Region == r {
				goto RegionValid
			}
		}
		return fmt.Errorf("invalid AWS region: %s", c.Region)
	case "azure":
		for _, r := range globals.ValidAzureRegions {
			if c.Region == r {
				goto RegionValid
			}
		}
		return fmt.Errorf("invalid Azure region: %s", c.Region)
	case "alibaba":
		for _, r := range globals.ValidAlibabaRegions {
			if c.Region == r {
				goto RegionValid
			}
		}
		return fmt.Errorf("invalid Alibaba region: %s", c.Region)
	}
RegionValid:
	if provider == "" {
		return fmt.Errorf("provider is required for deployment config validation")
	}
	return nil
}

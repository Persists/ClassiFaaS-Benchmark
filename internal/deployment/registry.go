package deployment

import (
	"ClassiFaaS/internal/config"
	"fmt"
)

type deployerFactory func(cfg config.DeploymentConfig) deployer

var deployers = make(map[string]deployerFactory)

func registerDeployer(name string, factory deployerFactory) {
	deployers[name] = factory
}

func getDeployer(cfg config.DeploymentConfig) (deployer, error) {
	if factory, exists := deployers[cfg.Provider]; exists {
		return factory(cfg), nil
	}
	return nil, fmt.Errorf("deployer %q not registered", cfg.Provider)
}

func init() {
	registerDeployer("gcp", deployerFactory(func(cfg config.DeploymentConfig) deployer {
		return newGCPDeployer(cfg)
	}))
	registerDeployer("aws", deployerFactory(func(cfg config.DeploymentConfig) deployer {
		return newAwsDeployer(cfg)
	}))
	registerDeployer("azure", deployerFactory(func(cfg config.DeploymentConfig) deployer {
		return newAzureDeployer(cfg)
	}))
	registerDeployer("alibaba", deployerFactory(func(cfg config.DeploymentConfig) deployer {
		return newAlibabaDeployer(cfg)
	}))
}

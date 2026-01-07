package deployment

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/globals"
)

type azureDeployer struct {
	scriptDefaultDeployer
}

func newAzureDeployer(cfg config.DeploymentConfig) deployer {
	return &azureDeployer{
		scriptDefaultDeployer: scriptDefaultDeployer{
			provider:  "azure",
			scriptDir: globals.AzureDeploymentScriptFolder,
			region:    cfg.Region,
		},
	}
}

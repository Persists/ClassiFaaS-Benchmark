package deployment

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/globals"
)

type AwsDeployer struct {
	scriptDefaultDeployer
}

func newAwsDeployer(cfg config.DeploymentConfig) deployer {
	return &AwsDeployer{
		scriptDefaultDeployer: scriptDefaultDeployer{
			provider:  "aws",
			scriptDir: globals.AWSDeploymentScriptFolder,
			region:    cfg.Region,
		},
	}
}

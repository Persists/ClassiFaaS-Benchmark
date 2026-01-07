package deployment

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/globals"
)

type alibabaDeployer struct {
	scriptDefaultDeployer
}

func newAlibabaDeployer(cfg config.DeploymentConfig) deployer {
	return &alibabaDeployer{
		scriptDefaultDeployer: scriptDefaultDeployer{
			provider:  "alibaba",
			scriptDir: globals.AlibabaDeploymentScriptFolder,
			region:    cfg.Region,
		},
	}
}

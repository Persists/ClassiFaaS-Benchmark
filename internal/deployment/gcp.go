package deployment

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/globals"
)

type GCPDeployer struct {
	scriptDefaultDeployer
}

func newGCPDeployer(cfg config.DeploymentConfig) deployer {
	return &GCPDeployer{
		scriptDefaultDeployer: scriptDefaultDeployer{
			provider:  "gcp",
			scriptDir: globals.GCPDeploymentScriptFolder,
			region:    cfg.Region,
		},
	}
}

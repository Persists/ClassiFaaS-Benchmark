package main

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/deployment"
	"ClassiFaaS/internal/utils"
)

func runDeploy() {
	cfg, err := config.LoadDeployConfig(configPath)
	if err != nil {
		panic(err)
	}

	dplOrchClient, err := deployment.NewDeployOrchestratorClient(*cfg)
	if err != nil {
		panic(err)
	}

	ep := utils.NewEventLogger()
	defer ep.Close()

	err = dplOrchClient.DeployAll(ep)
	if err != nil {
		panic(err)
	}

	ep.SendEvent(utils.SeverityInfo, "deploy", "✅ Deployments finished successfully.")
}

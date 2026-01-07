package main

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/deployment"
	"ClassiFaaS/internal/utils"
)

func runRemove() {
	cfg, err := config.LoadDeployConfig(configPath)
	if err != nil {
		panic(err)
	}

	ep := utils.NewEventLogger()
	defer ep.Close()

	dplOrchClient, err := deployment.NewDeployOrchestratorClient(*cfg)
	if err != nil {
		panic(err)
	}

	err = dplOrchClient.RemoveAll(ep)
	if err != nil {
		panic(err)
	}

	ep.SendEvent("info", "remove", "All deployments removed successfully")
}

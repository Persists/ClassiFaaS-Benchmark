package deployment

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/utils"
	"fmt"
)

type DeployOrechestratorClient struct {
	deployments map[string]DeployTargetClient
}

type DeployTargetClient struct {
	provider string
	region   string
	deployer deployer
}

func NewDeployTargetClient(cfg config.DeploymentConfig) (*DeployTargetClient, error) {
	deployer, err := getDeployer(cfg)
	if err != nil {
		return nil, err
	}

	return &DeployTargetClient{
		deployer: deployer,
		provider: cfg.Provider,
		region:   cfg.Region,
	}, nil
}

func NewDeployOrchestratorClient(cfg config.DeployConfig) (*DeployOrechestratorClient, error) {
	targets := make(map[string]DeployTargetClient)

	for _, deploymentTarget := range cfg.Deployments {

		targetClient, err := NewDeployTargetClient(deploymentTarget.DeploymentConfig)
		if err != nil {
			return nil, err
		}

		targets[fmt.Sprintf("%s:%s", deploymentTarget.Provider, deploymentTarget.Region)] = *targetClient
	}

	return &DeployOrechestratorClient{
		deployments: targets,
	}, nil
}

func (tc *DeployTargetClient) Deploy(ep utils.EventPublisher) error {
	ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("deploy_%s", tc.provider), fmt.Sprintf("Starting deployment in %s region", tc.region))

	if err := tc.deployer.Deploy(ep); err != nil {
		ep.SendEvent(utils.SeverityError, fmt.Sprintf("deploy_%s", tc.provider), fmt.Sprintf("Deployment failed in %s region: %v", tc.region, err))
		return err
	}

	ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("deploy_%s", tc.provider), fmt.Sprintf("Deployment succeeded in %s region", tc.region))
	return nil
}

func (oc *DeployOrechestratorClient) DeployAll(ep utils.EventPublisher) error {
	deployTimeLine := utils.NewTimeline("Deploy All Targets", utils.RunParallel)

	for _, deployer := range oc.deployments {
		deployTimeLine.Step(deployer.Deploy)
	}

	if err := deployTimeLine.Run(ep); err != nil {
		return err
	}

	return nil
}

func (oc *DeployOrechestratorClient) GenerateBenchmarkFunctions(ep utils.EventPublisher, filter func(DeployedFunction) bool, transformer func(DeployedFunction) config.BenchmarkFunctionConfig) ([]config.BenchmarkFunctionConfig, error) {

	var functions []config.BenchmarkFunctionConfig

	GetDeployedFunctionsTimeline := utils.NewTimeline("Get Deployed Functions", utils.RunParallel)
	for _, deployer := range oc.deployments {
		GetDeployedFunctionsTimeline.Step(func(ep utils.EventPublisher) error {
			funcs, err := deployer.deployer.LoadDeployedFunctions(ep)
			if err != nil {
				ep.SendEvent(utils.SeverityError, fmt.Sprintf("get_functions_%s", deployer.provider), fmt.Sprintf("Failed to load functions for %s in %s: %v", deployer.provider, deployer.region, err))
				return err
			}

			var filteredFuncs []DeployedFunction
			for _, f := range funcs {
				if filter(f) {
					filteredFuncs = append(filteredFuncs, f)
				}
			}

			ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("get_functions_%s", deployer.provider), fmt.Sprintf("Loaded %d functions for %s in %s, %d passed filtering", len(funcs), deployer.provider, deployer.region, len(filteredFuncs)))

			var transformedFuncs []config.BenchmarkFunctionConfig
			for _, f := range filteredFuncs {
				transformedFuncs = append(transformedFuncs, transformer(f))
			}

			functions = append(functions, transformedFuncs...)

			ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("get_functions_%s", deployer.provider), fmt.Sprintf("Successfully loaded %d functions for %s in %s", len(funcs), deployer.provider, deployer.region))
			return nil
		})
	}

	if err := GetDeployedFunctionsTimeline.Run(ep); err != nil {
		return functions, err
	}

	return functions, nil
}

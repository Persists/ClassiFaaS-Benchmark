package deployment

import (
	"ClassiFaaS/internal/utils"
	"fmt"
)

func (tc *DeployTargetClient) Remove(ep utils.EventPublisher) error {
	ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("remove_%s", tc.provider), fmt.Sprintf("Starting removal in %s region", tc.region))

	if err := tc.deployer.Remove(ep); err != nil {
		ep.SendEvent(utils.SeverityError, fmt.Sprintf("remove_%s", tc.provider), fmt.Sprintf("Removal failed in %s region: %v", tc.region, err))
		return err
	}

	ep.SendEvent(utils.SeverityInfo, fmt.Sprintf("remove_%s", tc.provider), fmt.Sprintf("Removal succeeded in %s region", tc.region))
	return nil
}

func (oc *DeployOrechestratorClient) RemoveAll(ep utils.EventPublisher) error {
	removeTimeLine := utils.NewTimeline("Remove All Targets", utils.RunParallel)

	for _, deployer := range oc.deployments {
		removeTimeLine.Step(deployer.Remove)
	}

	if err := removeTimeLine.Run(ep); err != nil {
		return err
	}

	return nil
}

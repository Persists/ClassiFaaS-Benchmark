package deployment

import (
	"ClassiFaaS/internal/utils"
)

type DeployedFunction struct {
	Provider  string `json:"provider"`
	URL       string `json:"url"`
	Auth      string `json:"auth,omitempty"`
	Memory    int    `json:"memory"`
	Benchmark string `json:"benchmark"`
	Region    string `json:"region"`
}
type deployer interface {
	// Deploy the functions to the target cloud provider
	Deploy(event utils.EventPublisher) error
	// Load the deployed functions from the target cloud provider (including the auth details)
	LoadDeployedFunctions(event utils.EventPublisher) ([]DeployedFunction, error)
	// Remove the deployed functions from the target cloud provider
	Remove(event utils.EventPublisher) error
	// GetProvider returns the name of the cloud provider
	GetProvider() string
}

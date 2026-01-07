package deployment

import (
	"ClassiFaaS/internal/utils"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type scriptDefaultDeployer struct {
	provider  string
	region    string
	scriptDir string
}

func (d *scriptDefaultDeployer) GetProvider() string {
	return d.provider
}

func (d *scriptDefaultDeployer) Deploy(ep utils.EventPublisher) error {
	ep.SendEvent(utils.SeverityInfo, "deploy_"+d.provider, "Starting deployment...")

	cmd := exec.Command("bash", "./manage-deployment.sh", "deploy", d.region)
	cmd.Dir = d.scriptDir

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		ep.SendEvent(utils.SeverityError, "deploy_"+d.provider, fmt.Sprintf("stdout pipe failed: %v", err))
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ep.SendEvent(utils.SeverityError, "deploy_"+d.provider, fmt.Sprintf("stderr pipe failed: %v", err))
		return err
	}

	go ep.StreamToEvents(stdOut, "deploy_"+d.provider+"_stdout")
	go ep.StreamToEvents(stderr, "deploy_"+d.provider+"_stderr")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy resources (script: %s, location: %s): %w",
			filepath.Join(d.scriptDir, "manage-deployment.sh"), d.region, err)
	}

	ep.SendEvent(utils.SeverityInfo, "deploy_"+d.provider, "deployment completed successfully")
	return nil
}

func (d *scriptDefaultDeployer) LoadDeployedFunctions(ep utils.EventPublisher) ([]DeployedFunction, error) {
	ep.SendEvent(utils.SeverityInfo, "load_functions_"+d.provider, "Fetching function URLs and keys for all function apps...")

	cmd := exec.Command("bash", "./manage-deployment.sh", "get-urls", d.region)
	cmd.Dir = d.scriptDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get %s function URLs: %v\nOutput: %s", d.provider, err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	deployedFunctions := []DeployedFunction{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var f DeployedFunction
		if err := json.Unmarshal([]byte(line), &f); err != nil {
			return nil, fmt.Errorf("failed to parse line: %s\nerror: %v", line, err)
		}

		f.Provider = d.provider

		deployedFunctions = append(deployedFunctions, f)
	}

	if len(deployedFunctions) == 0 {
		return nil, fmt.Errorf("no functions parsed from script output")
	}

	ep.SendEvent(utils.SeverityInfo, "deploy_"+d.provider, fmt.Sprintf("parsed %d functions", len(deployedFunctions)))
	return deployedFunctions, nil
}

func (d *scriptDefaultDeployer) Remove(ep utils.EventPublisher) error {
	ep.SendEvent(utils.SeverityInfo, "remove_"+d.provider, "Starting removal of deployment...")

	cmd := exec.Command("bash", "./manage-deployment.sh", "delete", d.region)
	cmd.Dir = d.scriptDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		ep.SendEvent(utils.SeverityError, "remove_"+d.provider, fmt.Sprintf("failed to remove deployment: %v, output: %s", err, string(output)))
		return err
	}

	ep.SendEvent(utils.SeverityInfo, "remove_"+d.provider, "removal completed successfully")
	return nil
}

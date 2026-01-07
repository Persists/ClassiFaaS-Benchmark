package main

import (
	"ClassiFaaS/internal/auth"
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/loadgenerator"
	"ClassiFaaS/internal/utils"
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func main() {
	configPath := flag.String("config", "configs/generated.yaml", "Path to the benchmark configuration YAML file")
	flag.Parse()

	cfg, err := config.LoadBenchmarkConfig(*configPath)
	if err != nil {
		panic(err)
	}

	ep := utils.NewEventLogger()
	defer ep.Close()

	// Group functions by provider and region
	loadGenerators := make(map[string]loadgenerator.LoadGenerator)
	WorkloadParameters := cfg.WorkloadParameters

	for _, fn := range cfg.Functions {

		if fn.Provider == "gcp" {
			gcpToken, err := auth.GetGoogleIdentityToken(fn.URL)
			fn.Auth.Value = fmt.Sprintf("Bearer %s", gcpToken)

			if err != nil {
				panic(err)
			}
		}

		name := fmt.Sprintf("%s-%s-%s-%d-%d", fn.Provider, fn.Region, fn.Name, fn.MemSize, rand.Intn(1000))

		ep.SendEvent(utils.SeverityInfo, "executor_setup", "Setting up executor for "+name)
		if strings.Contains(fn.Provider, "azure") && WorkloadParameters.ParallelRequests > 300 {
			ep.SendEvent(utils.SeverityWarning, "azure_limitations", "Azure can´t handle more than 300 parallel requests. Limiting to 300.")
			WorkloadParameters.ParallelRequests = 300
		}

		lgen, err := loadgenerator.NewLoadGenerator(&WorkloadParameters, fn)
		if err != nil {
			panic(err)
		}
		loadGenerators[name] = *lgen

	}

	// setup benchmark timeline
	benchTimeLine := utils.NewTimeline("Benchmark Timeline", utils.RunParallel)
	for _, e := range loadGenerators {
		benchTimeLine.Step(e.Run)
	}

	// periodic progress update
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for range ticker.C {
			var progressUpdate string
			for name, e := range loadGenerators {
				total, remaining := e.GetQueueState()
				progressUpdate += fmt.Sprintf("%s: %d/%d tasks started.", name, total-remaining, total)
			}

			ep.SendEvent(utils.SeverityInfo, "progress_update", progressUpdate)
		}
	}()

	// run the benchmark
	benchTimeLine.Run(ep)
}

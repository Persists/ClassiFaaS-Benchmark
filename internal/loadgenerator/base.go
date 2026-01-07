package loadgenerator

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// LoadGenerator manages the coordinated execution of benchmark jobs
// across multiple function configurations.
type LoadGenerator struct {
	workerPoolSize int
	queueLen       int
	workerSpec     workerSpec
	task           *task
}

// NewLoadGenerator constructs a new LoadGenerator instance using the provided
// benchmark parameters and function configurations.
//
// It prepares a task queue that populated with one task for each function configuration,
// repeated TotalRequests times. Each task is associated with a file archiver
// responsible for persisting benchmark results and metadata.
func NewLoadGenerator(
	WorkloadParameters *config.WorkloadParameters,
	fnCfg config.BenchmarkFunctionConfig,
) (*LoadGenerator, error) {
	taskQueueLen := WorkloadParameters.TotalRequests
	taskQueue := createTaskQueue(taskQueueLen)

	// group tasks by memory size

	metadata := map[string]string{
		"timestamp":              time.Now().Format(time.RFC3339),
		"url":                    fnCfg.URL,
		"function":               fnCfg.Name,
		"parallel-requests":      strconv.Itoa(WorkloadParameters.ParallelRequests),
		"iterationsPerBenchmark": strconv.Itoa(WorkloadParameters.TotalRequests),
		"retries":                strconv.Itoa(WorkloadParameters.RetriesPerRequest),
		"provider":               fnCfg.Provider,
		"region":                 fnCfg.Region,
		"memorySize":             strconv.Itoa(fnCfg.MemSize),
	}

	metaStr, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	resultFolder := fmt.Sprintf(
		"%s/%s/%s/%s/%s.log",
		WorkloadParameters.ResultFolder,
		time.Now().Format("2006-01-02_15-04"),
		fnCfg.Provider,
		fnCfg.Region,
		fnCfg.Name,
	)

	archiver, err := utils.NewFileArchiveClient(resultFolder, string(metaStr))
	if err != nil {
		log.Fatalf("Failed to create archive client for function %s: %v", fnCfg.Name, err)
	}

	archiver.Start()
	task := createTask(&fnCfg, archiver)

	// Populate the task queue with all tasks across TotalRequests iterations.
	for i := 0; i < WorkloadParameters.TotalRequests; i++ {
		taskQueue <- task
	}

	close(taskQueue)

	workerSpec := workerSpec{
		httpClient:     &http.Client{Timeout: 120 * time.Second},
		taskQueue:      taskQueue,
		requestRetries: WorkloadParameters.RetriesPerRequest,
	}

	return &LoadGenerator{
		workerPoolSize: WorkloadParameters.ParallelRequests,
		task:           task,
		queueLen:       len(taskQueue),
		workerSpec:     workerSpec,
	}, nil
}

// Run starts the load generator using the configured worker pool Size (ParallelRequests).
//
// Each worker consumes tasks from the queue until it is empty, executing
// benchmark jobs.
//
// Once all workers complete, the Run method closes all archive clients
// associated with the executed tasks.
func (l *LoadGenerator) Run(ep utils.EventPublisher) error {

	var workerWg sync.WaitGroup

	ep.SendEvent("info", "load_generator_start",
		fmt.Sprintf("(%s: %s) Starting load generator with %d workers", l.task.Function.Provider, l.task.Function.Region, l.workerPoolSize))

	for i := 0; i < l.workerPoolSize; i++ {
		workerWg.Add(1)
		go l.workerSpec.worker(&workerWg, ep)
	}

	workerWg.Wait()

	l.task.ArchiveClient.Stop()
	ep.SendEvent("info", "function_finished",
		fmt.Sprintf("Finished benchmarking function %s and closed archiver", l.task.Function.Name))
	time.Sleep(2 * time.Second) // wait for any last events to be sent

	return nil
}

// GetQueueState returns the initial and current length of the task queue.
func (l *LoadGenerator) GetQueueState() (int, int) {
	return l.queueLen, len(l.workerSpec.taskQueue)
}

package loadgenerator

import (
	"ClassiFaaS/internal/utils"
	"fmt"
	"net/http"
	"sync"
)

// WorkerSpec defines the configuration for a benchmark worker.
type workerSpec struct {
	// taskQueue provides the stream of tasks to execute.
	taskQueue chan *task

	requestRetries int
	httpClient     *http.Client
}

// worker consumes and executes tasks from the TaskQueue.
//
// A worker runs until the queue is closed and all tasks are processed. Each task is
// executed via Task.Execute using an HTTP client.
//
// The worker signals completion through the provided WaitGroup once all
// tasks are finished.
func (spec *workerSpec) worker(workerWg *sync.WaitGroup, ep utils.EventPublisher) {
	defer workerWg.Done()

	for task := range spec.taskQueue {
		err := task.execute(spec.httpClient, spec.requestRetries)

		if err != nil {
			ep.SendEvent(
				"error",
				"task_execution",
				fmt.Sprintf("Error executing task %s: %v", task.Function.Name, err),
			)
		}

	}
}

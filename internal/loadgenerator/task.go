package loadgenerator

import (
	"ClassiFaaS/internal/config"
	"ClassiFaaS/internal/utils"
	"fmt"
	"net/http"
	"time"
)

// Task represents a single benchmark job to be executed against
// a specific function configuration.
//
// Each Task holds a reference to its target function configuration
// and an ArchiveClient used to persist benchmark results.
type task struct {
	Function *config.BenchmarkFunctionConfig

	ArchiveClient *utils.ArchiveClient
}

// CreateTaskQueue initializes and returns a buffered channel that acts
// as the task queue for benchmark execution.
func createTaskQueue(bufferSize int) chan *task {
	return make(chan *task, bufferSize)
}

// CreateTask constructs a new task for the specified function configuration
// and archive client. The resulting task is ready to be scheduled in a
// LoadGenerator's task queue.
func createTask(function *config.BenchmarkFunctionConfig, archiveClient *utils.ArchiveClient) *task {
	return &task{
		Function:      function,
		ArchiveClient: archiveClient,
	}
}

// Execute performs the benchmark request associated with the Task.
//
// The function is invoked via an HTTP GET request to the configured URL,
// including any query parameters provided in the query map. If the request
// or decoding fails, it will be retried up to the specified number of retries.
func (t *task) execute(httpClient *http.Client, retries int) error {
	var err error

	for attempt := 0; attempt <= retries; attempt++ {
		req, err := http.NewRequest("GET", t.Function.URL, nil)
		if err != nil {
			return err
		}

		if t.Function.Auth.Key != "" && t.Function.Auth.Value != "" {
			req.Header.Set(t.Function.Auth.Key, t.Function.Auth.Value)
		}

		resp, err := httpClient.Do(req)

		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		result, decErr := utils.DecodeBenchmarkResponse(resp)
		if decErr != nil {
			fmt.Println("Error decoding benchmark response:", decErr)
			return decErr
		}

		// Persist result
		resultStr, err := result.ToString()
		if err != nil {
			return err
		}

		t.ArchiveClient.Write(resultStr)
		return nil
	}

	return fmt.Errorf("task %s failed after %d retries: %v", t.Function.Name, retries, err)
}

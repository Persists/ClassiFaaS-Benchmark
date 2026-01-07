package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// ArchiveClient is a struct that contains the configuration for the archive client
type ArchiveClient struct {
	debugMode  bool
	writer     *bufio.Writer
	writeChan  chan string
	allWritten sync.WaitGroup
}

// systemsBlockSize returns the block size of the system
// This is used to optimize the buffer size of the writer
func systemsBlockSize() (int, error) {
	os := runtime.GOOS

	var cmd *exec.Cmd
	switch os {
	case "linux":
		cmd = exec.Command("stat", "-fc", "%s", "/")
	case "darwin":
		cmd = exec.Command("stat", "-f", "%k", "/")
	default:
		fmt.Println("Unsupported OS, using default buffer size")
		return 4096, nil
	}

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting block size:", err)
		return 0, err
	}
	var blockSize int
	fmt.Sscanf(string(out), "%d", &blockSize)
	return blockSize, nil

}

// NewArchiveClient creates a new ArchiveClient
func NewFileArchiveClient(filePath string, metadata string) (*ArchiveClient, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Get the block size of the system, and use it to optimize the buffer size
	bs, err := systemsBlockSize()
	if err != nil {
		return nil, fmt.Errorf("failed to get block size: %w", err)
	}

	writer := bufio.NewWriterSize(file, bs)

	ac := &ArchiveClient{
		writer:    writer,
		writeChan: make(chan string),
	}

	_, err = ac.writer.WriteString(metadata + "\n")
	if err != nil {
		return nil, fmt.Errorf("failed to add metadata to file: %w", err)
	}
	err = ac.writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush metadata to file: %w", err)
	}

	return ac, nil

}

// Start starts the archive client
//
// Goroutines:
//   - Debug mode: If enabled, starts a goroutine that continuously reads from the
//     write channel, and discards the data.
//   - File writing: Starts a goroutine to handle writing data to a file.
func (ac *ArchiveClient) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		ac.Stop()
		os.Exit(0)
	}()

	if ac.debugMode {
		go func() {
			for range ac.writeChan {
			}
		}()
	}
	go ac.writeToFile()
}

// writeToFile writes the data from the write channel to the file using the writer
// from the bufio package.
func (ac *ArchiveClient) writeToFile() {
	ac.allWritten.Add(1)
	defer ac.allWritten.Done()
	for line := range ac.writeChan {
		_, err := ac.writer.WriteString(line)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}

func (ac *ArchiveClient) Write(line string) {
	ac.writeChan <- line + "\n"
}

func (ac *ArchiveClient) Stop() {
	close(ac.writeChan)
	ac.allWritten.Wait()
	ac.writer.Flush()

	time.Sleep(5 * time.Second) // wait for any last writes to finish
}

func (ac *ArchiveClient) Flush() {
	ac.writer.Flush()
}

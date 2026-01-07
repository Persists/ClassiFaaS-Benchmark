package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var logChannel = make(chan Event, 1000)
var wg sync.WaitGroup

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Event struct {
	Time     time.Time `json:"time,omitempty"`
	Severity Severity  `json:"severity,omitempty"`
	Type     string    `json:"type,omitempty"`
	Message  string    `json:"message,omitempty"`
}

// Publisher wraps a send-only channel
type EventPublisher chan<- Event

func NewEventLogger() EventPublisher {
	logger := log.New(os.Stdout, "", 0)
	wg.Add(1)

	go func() {
		defer wg.Done()
		for event := range logChannel {
			reset := "\033[0m"
			yellow := "\033[33m"
			orange := "\033[38;5;208m"
			green := "\033[32m"
			magenta := "\033[35m"

			// Set severity color
			var sevColor string
			switch event.Severity {
			case "info":
				sevColor = "\033[36m" // cyan
			case "warn":
				sevColor = "\033[33m" // yellow
			case "error":
				sevColor = "\033[31m" // red
			default:
				sevColor = "\033[0m" // default
			}

			typeStr := event.Type
			switch {
			case strings.Contains(strings.ToLower(event.Type), "gcp"):
				typeStr = yellow + event.Type + reset
			case strings.Contains(strings.ToLower(event.Type), "aws"):
				typeStr = orange + event.Type + reset
			case strings.Contains(strings.ToLower(event.Type), "azure"):
				typeStr = green + event.Type + reset
			case strings.Contains(strings.ToLower(event.Type), "alibaba"):
				typeStr = magenta + event.Type + reset
			}

			logger.Printf("[%s] [%s%s%s] [%s] %s",
				event.Time.Format(time.RFC3339),
				sevColor, event.Severity, reset,
				typeStr,
				event.Message,
			)
		}
	}()

	return EventPublisher(logChannel)
}

func (p EventPublisher) Close() {
	close(logChannel)
	wg.Wait()
}

// SendEvent sends an event through the send-only channel
func (p EventPublisher) SendEvent(severity Severity, eventType string, message string) {
	select {
	case p <- Event{
		Time:     time.Now(),
		Severity: severity,
		Type:     eventType,
		Message:  message,
	}:
	default:
		fmt.Printf("event channel is full, dropping event: %s\n", message)
	}
}

func (p EventPublisher) StreamToEvents(pipe io.ReadCloser, eventType string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		p.SendEvent(SeverityInfo, eventType, line)
	}
	if err := scanner.Err(); err != nil {
		p.SendEvent(SeverityError, eventType, fmt.Sprintf("error reading pipe: %v", err))
	}
}

package utils

import (
	"fmt"
	"sync"
)

type RunMode int

const (
	RunSequential RunMode = iota
	RunParallel
)

type Timeline struct {
	Description string
	Mode        RunMode
	Steps       []*Step
}

type StepFunc func(eventPublisher EventPublisher) error

type Step struct {
	Run StepFunc
}

func NewTimeline(description string, mode RunMode) *Timeline {
	return &Timeline{Description: description, Mode: mode}
}

// Run executes the timeline according to the provided RunMode.
// - events: channel to emit lifecycle events. Caller is responsible for closing it.
// - mode: RunSequential or RunParallel.
//
// Panics inside step functions are recovered and emitted as EventStepPanic.
func (tl *Timeline) Run(eventPublisher EventPublisher) error {
	// Emit timeline start
	eventPublisher.SendEvent(SeverityInfo, "start_timeline", tl.Description)
	var err error
	if tl.Mode == RunParallel {
		err = tl.runParallel(eventPublisher)
	} else {
		err = tl.runSequential(eventPublisher)
	}

	if err != nil {
		eventPublisher.SendEvent(SeverityError, "timeline_failed", err.Error())
		return err
	}

	eventPublisher.SendEvent(SeverityInfo, "finish_timeline", tl.Description)
	return nil
}

// runSequential runs steps one after another.
func (tl *Timeline) runSequential(eventPublisher EventPublisher) error {
	for _, step := range tl.Steps {
		// Call step safely
		err := step.RunStep(eventPublisher)
		if err != nil {
			return err
		}
	}
	return nil
}

// runParallel runs all steps concurrently.
func (tl *Timeline) runParallel(eventPublisher EventPublisher) error {
	var wg sync.WaitGroup
	errors := []error{}
	for _, step := range tl.Steps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := step.RunStep(eventPublisher); err != nil {
				eventPublisher.SendEvent(SeverityError, "parallel_timeline_step_failed", err.Error())
				errors = append(errors, err)
			}
		}()
	}

	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("one or more parallel steps failed: %v", errors)
	}
	return nil
}

func (tl *Timeline) Step(run func(events EventPublisher) error) *Timeline {
	step := &Step{
		Run: run,
	}
	tl.Steps = append(tl.Steps, step)
	return tl
}

func (s *Step) RunStep(eventPublisher EventPublisher) error {
	if s == nil || s.Run == nil {
		eventPublisher.SendEvent(SeverityError, "invalid_step", "step or step function is nil")
		return fmt.Errorf("step or step function is nil")
	}
	err := s.Run(eventPublisher)
	if err != nil {
		return err
	}

	return nil
}

package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fs1g17/LLM-Batch-Runner/internal/config"
)

// Today: make a single planner that, in a loop, process requests
// then tomorrow, we can split up each processing of request in goroutine

// this planner does the following:
// it accepts config
// it reads the Config.In directory for filenames
// then, in a loop, for each filename, it:
// - reads file contents
// - performs request to LLM
// - writes output to Config.Out directory

type Runner struct {
	RequestsComplete map[string]bool // maps if request file has corresponding response
	RequestFileNames []string
	RequestsSkipped  int // number of input folder files that have matching output files
	Provider         Provider
	Config           *config.Config
}

type Provider interface {
	Completion(request string) (string, error)
}

func NewRunner(provider Provider, config *config.Config) *Runner {
	return &Runner{
		RequestsComplete: make(map[string]bool),
		Provider:         provider,
		Config:           config,
	}
}

type ResultSummary struct {
	Skipped   int
	Succeeded int
	Failed    int
}

func (r *Runner) readFolder(path string, setComplete bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if setComplete {
			// first check if the request was in the "in" folder
			_, ok := r.RequestsComplete[e.Name()]
			if !ok {
				// skip files in the "out" folder that weren't in "in" folder
				continue
			}
		}
		r.RequestsComplete[e.Name()] = setComplete
	}

	return nil
}

func (r *Runner) plan() error {
	err := r.readFolder(r.Config.In, false)
	if err != nil {
		return fmt.Errorf("failed to read in dir: %w\n", err)
	}
	err = r.readFolder(r.Config.Out, true)
	if err != nil {
		return fmt.Errorf("failed to read out dir: %w\n", err)
	}

	r.RequestsSkipped = 0
	for requestFileName, complete := range r.RequestsComplete {
		if complete {
			// skip completed requests
			r.RequestsSkipped++
			continue
		}
		r.RequestFileNames = append(r.RequestFileNames, requestFileName)
	}

	return nil
}

type Event int

const (
	Start Event = iota
	Success
	Fail
)

func (r *Runner) worker(id int, ctx context.Context, jobs <-chan string, events chan<- Event) {
	// jobs receive filepath - this allows us to skip files
	for {
		select {
		case <-ctx.Done():
			ctx.Err()
			return
		case requestFileName, ok := <-jobs:
			if !ok {
				return
			}
			events <- Start
			time.Sleep(10 * time.Second)
			dat, err := os.ReadFile(filepath.Join(r.Config.In, requestFileName))
			if err != nil {
				// log and skip
				fmt.Printf("worker %d failed to read file %s/%s, %v\n", id, r.Config.In, requestFileName, err)
				events <- Fail
				continue
			}

			var content string = string(dat)
			fmt.Printf("worker %d got contents of %s/%s are: '%s'\n", id, r.Config.In, requestFileName, content)

			response, err := r.Provider.Completion(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "worker %d failed to get completion for file: %s, %v\n", id, requestFileName, err)
				events <- Fail
				continue
			}

			writePath := filepath.Join(r.Config.Out, requestFileName+".tmp")
			err = os.WriteFile(writePath, []byte(response), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "worker %d failed to write temp output file: %s, %v\n", id, requestFileName, err)
				events <- Fail
				continue
			}

			err = os.Rename(writePath, filepath.Join(r.Config.Out, requestFileName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "worker %d failed to rename output file: %v\n", id, err)
				events <- Fail
				continue
			}

			events <- Success
		}
	}
}

/*
in directory contains requests, out directory will contain responses with the same filename
in order to resume, we can simply see which output files don't exist for the input files
*/
func (r *Runner) Run(ctx context.Context) (ResultSummary, error) {
	// must make it thread safe later - because concurrent access will have race conditions
	resultSummary := ResultSummary{
		Skipped:   0,
		Succeeded: 0,
		Failed:    0,
	}

	err := r.plan()
	if err != nil {
		return resultSummary, err
	}

	resultSummary.Skipped = r.RequestsSkipped

	numJobs := len(r.RequestFileNames)
	jobs := make(chan string, numJobs)
	events := make(chan Event, numJobs)

	var wg sync.WaitGroup
	for i := range r.Config.Workers {
		wg.Go(func() {
			r.worker(i, ctx, jobs, events)
		})
	}

	go func() {
		wg.Wait()
		close(events)
	}()

	var wg2 sync.WaitGroup
	wg2.Go(func() {
		inFlightCounter := 0
		for event := range events {
			if event == Start {
				inFlightCounter++
			} else {
				inFlightCounter--
			}

			if event == Success {
				resultSummary.Succeeded++
			}
			if event == Fail {
				resultSummary.Failed++
			}

			fmt.Printf("in-flight: %d, done: %d, failed: %d\n", inFlightCounter, resultSummary.Succeeded, resultSummary.Failed)
		}
	})

	// now we can start/resume
outer:
	for _, requestFileName := range r.RequestFileNames {
		select {
		case jobs <- requestFileName:
		case <-ctx.Done():
			break outer
		}
	}
	close(jobs)

	wg2.Wait()

	return resultSummary, nil
}

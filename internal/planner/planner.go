package planner

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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

type Planner struct {
	RequestsComplete map[string]bool // maps if request file has corresponding response
	Provider         Provider
	Config           *config.Config
}

type Provider interface {
	Completion(request string) (string, error)
}

func NewPlanner(provider Provider, config *config.Config) *Planner {
	return &Planner{
		RequestsComplete: make(map[string]bool),
		Provider:         provider,
		Config:           config,
	}
}

type PlanSummary struct {
	Skipped int
	Done    int
	Failed  int
}

/*
in directory contains requests, out directory will contain responses with the same filename
in order to resume, we can simply see which output files don't exist for the input files
*/
func (p *Planner) Plan() (PlanSummary, error) {
	// must make it thread safe later - because concurrent access will have race conditions
	planSummary := PlanSummary{
		Skipped: 0,
		Done:    0,
		Failed:  0,
	}

	err := p.readFolder(p.Config.In, false)
	if err != nil {
		return planSummary, fmt.Errorf("failed to read in dir: %w\n", err)
	}
	err = p.readFolder(p.Config.Out, true)
	if err != nil {
		return planSummary, fmt.Errorf("failed to read out dir: %w\n", err)
	}

	// now we can start/resume
	for requestFileName, completed := range p.RequestsComplete {
		if completed {
			planSummary.Skipped++
			continue
		}

		dat, err := os.ReadFile(path.Join(p.Config.In, requestFileName))
		if err != nil {
			// log and skip
			fmt.Printf("failed to read file %s/%s, %v\n", p.Config.In, requestFileName, err)
			continue
		}

		var content string = string(dat)
		fmt.Printf("contents of %s/%s are: '%s'\n", p.Config.In, requestFileName, content)

		response, err := p.Provider.Completion(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get completion for file: %s, %v\n", requestFileName, err)
			planSummary.Failed++
			continue
		}

		writePath := filepath.Join(p.Config.Out, requestFileName+".tmp")
		err = os.WriteFile(writePath, []byte(response), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write temp output file: %s, %v\n", requestFileName, err)
			planSummary.Failed++
			continue
		}

		err = os.Rename(writePath, filepath.Join(p.Config.Out, requestFileName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to rename output file: %v\n", err)
			continue
		}

		planSummary.Done++
	}

	return planSummary, nil
}

func (p *Planner) readFolder(path string, setComplete bool) error {
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
			_, ok := p.RequestsComplete[e.Name()]
			if !ok {
				// skip files in the "out" folder that weren't in "in" folder
				continue
			}
		}
		p.RequestsComplete[e.Name()] = setComplete
	}

	return nil
}

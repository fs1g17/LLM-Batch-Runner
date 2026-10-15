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

/*
in directory contains requests, out directory will contain responses with the same filename
in order to resume, we can simply see which output files don't exist for the input files
*/
func (p *Planner) Plan() {
	p.readFolder(p.Config.In, false)
	p.readFolder(p.Config.Out, true)

	// now we can start/resume
	for requestFileName, completed := range p.RequestsComplete {
		if completed {
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
			fmt.Printf("failed to get completion for file: %s", requestFileName)
			continue
		}

		writePath := filepath.Join(p.Config.Out, requestFileName)
		err = os.WriteFile(writePath, []byte(response), 0644)
		if err != nil {
			fmt.Printf("failed to write completion for file: %s", requestFileName)
			continue
		}
	}
}

func (p *Planner) readFolder(path string, setComplete bool) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
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
}

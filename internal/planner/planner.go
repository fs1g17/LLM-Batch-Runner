package planner

import (
	"fmt"
	"os"

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
}

/*
in directory contains requests, out directory will contain responses with the same filename
in order to resume, we can simply see which output files don't exist for the input files
*/
func (p *Planner) Plan(config config.Config) {
	p.RequestsComplete = make(map[string]bool)

	p.readFolder(config.In, false)
	p.readFolder(config.Out, true)
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

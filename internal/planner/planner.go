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
	entries, err := os.ReadDir(config.In)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	p.RequestsComplete = make(map[string]bool)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p.RequestsComplete[e.Name()] = false
	}

	entries, err = os.ReadDir(config.Out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p.RequestsComplete[e.Name()] = true
	}
}

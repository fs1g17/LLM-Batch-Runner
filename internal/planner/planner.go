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
}

func (p *Planner) Plan(config config.Config) {
	entries, err := os.ReadDir(config.In)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fmt.Println(e.Name())
	}
}

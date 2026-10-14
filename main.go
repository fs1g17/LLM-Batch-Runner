package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fs1g17/LLM-Batch-Runner/internal/config"
	"github.com/fs1g17/LLM-Batch-Runner/internal/planner"
)

func main() {
	configLimits, err := config.LoadConfigLimits()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	in := flag.String("in", "", "path to prompts")
	out := flag.String("out", "", "path to results")
	workers := flag.Int("workers", 1, "number of workers")
	rps := flag.Float64("rps", 1, "request limiting per second")
	fake := flag.Bool("fake", false, "fake requests")

	flag.Parse()

	fmt.Printf("in: '%s', out: '%s', workers: %d, rps: %f, fake: %t\n", *in, *out, *workers, *rps, *fake)

	config := config.Config{In: *in, Out: *out, Workers: *workers, Rps: *rps, Fake: *fake}
	err = config.Validate(configLimits)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	planner := planner.Planner{}
	planner.Plan(config)

	for request, complete := range planner.RequestsComplete {
		fmt.Println(request, "is complete:", complete)
	}
}

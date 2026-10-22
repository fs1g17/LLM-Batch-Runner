package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/fs1g17/LLM-Batch-Runner/internal/config"
	"github.com/fs1g17/LLM-Batch-Runner/internal/llm"
	"github.com/fs1g17/LLM-Batch-Runner/internal/runner"
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

	llm, err := llm.NewLLM()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	runner := runner.NewRunner(llm, &config)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
	}()
	summary, err := runner.Run(ctx)

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(summary)

	if ctx.Err() != nil {
		fmt.Println("Run was stopped by user")
		os.Exit(130)
	}

	if summary.Failed != 0 {
		os.Exit(1)
	}
}

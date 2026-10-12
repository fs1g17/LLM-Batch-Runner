package main

import (
	"flag"
	"fmt"
)

func main() {
	in := flag.String("in", "", "path to prompts")
	out := flag.String("out", "", "path to results")

	workers := flag.Int("workers", 1, "number of workers")
	rps := flag.Int("rps", 1, "request limiting per second")

	fake := flag.Bool("fake", false, "fake requests")

	flag.Parse()

	fmt.Printf("in: '%s', out: '%s', workers: %d, rps: %d, fake: %t\n", *in, *out, *workers, *rps, *fake)
}

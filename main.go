package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var invalidPath = errors.New("path is invalid")
var undefinedVal = errors.New("value is required")
var incorrectVal = errors.New("value is incorrect")

type Config struct {
	In      *string
	Out     *string
	Workers *int
	Rps     *float64
	Fake    *bool
}

type ConfigLimits struct {
	MinWorkers int
	MaxWorkers int
	MinRps     float64
	MaxRps     float64
}

func loadConfigLimits() ConfigLimits {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}

	configLimits := ConfigLimits{
		MinWorkers: 1,
		MaxWorkers: 100,
		MinRps:     0.5,
		MaxRps:     100,
	}

	minWorkers, exists := os.LookupEnv("MIN_WORKERS")
	if exists {
		minWorkersInt, err := strconv.Atoi(minWorkers)
		if err == nil {
			configLimits.MinWorkers = minWorkersInt
		}
	}

	maxWorkers, exists := os.LookupEnv("MAX_WORKERS")
	if exists {
		maxWorkersInt, err := strconv.Atoi(maxWorkers)
		if err == nil {
			configLimits.MaxWorkers = maxWorkersInt
		}
	}

	minRps, exists := os.LookupEnv("MIN_RPS")
	if exists {
		minRpsFloat, err := strconv.ParseFloat(minRps, 64)
		if err == nil {
			configLimits.MinRps = minRpsFloat
		}
	}

	maxRps, exists := os.LookupEnv("MAX_RPS")
	if exists {
		maxRpsFloat, err := strconv.ParseFloat(maxRps, 64)
		if err == nil {
			configLimits.MaxRps = maxRpsFloat
		}
	}

	return configLimits
}

func main() {
	configLimits := loadConfigLimits()

	in := flag.String("in", "", "path to prompts")
	out := flag.String("out", "", "path to results")
	workers := flag.Int("workers", 1, "number of workers")
	rps := flag.Float64("rps", 1, "request limiting per second")
	fake := flag.Bool("fake", false, "fake requests")

	flag.Parse()

	fmt.Printf("in: '%s', out: '%s', workers: %d, rps: %f, fake: %t\n", *in, *out, *workers, *rps, *fake)

	config := Config{in, out, workers, rps, fake}
	err := config.Validate(configLimits)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (c *Config) Validate(cl ConfigLimits) error {
	// first, make sure all values have been passed
	if c.In == nil {
		return fmt.Errorf("-in %w", undefinedVal)
	}
	if c.Out == nil {
		return fmt.Errorf("-out %w", undefinedVal)
	}
	if c.Workers == nil {
		return fmt.Errorf("-workers %w", undefinedVal)
	}
	if c.Rps == nil {
		return fmt.Errorf("-rps %w", undefinedVal)
	}

	// second, validate the values are sensible
	if err := validateDir(*c.In); err != nil {
		return fmt.Errorf("-in %w", err)
	}
	if err := validateDir(*c.Out); err != nil {
		return fmt.Errorf("-out %w", err)
	}

	if *c.Workers < cl.MinWorkers || *c.Workers > cl.MaxWorkers {
		return fmt.Errorf("-workers must be between %d and %d, %w", cl.MinWorkers, cl.MaxWorkers, incorrectVal)
	}

	if *c.Rps < cl.MinRps || *c.Rps > cl.MaxRps {
		return fmt.Errorf("-rps must be between %f and %f, %w", cl.MinRps, cl.MaxRps, incorrectVal)
	}

	return nil
}

func validateDir(path string) error {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("doesn't exist %w", invalidPath)
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory %w", invalidPath)
	}
	return nil
}

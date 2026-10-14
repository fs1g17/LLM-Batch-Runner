package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var invalidPath = errors.New("path is invalid")
var undefinedVal = errors.New("value is required")
var incorrectVal = errors.New("value is incorrect")
var incorrectEnv = errors.New("env is incorrect")

type Config struct {
	In      string
	Out     string
	Workers int
	Rps     float64
	Fake    bool
}

func (c *Config) Validate(cl ConfigLimits) error {
	// first, make sure all values have been passed
	if c.In == "" {
		return fmt.Errorf("-in %w", undefinedVal)
	}
	if c.Out == "" {
		return fmt.Errorf("-out %w", undefinedVal)
	}

	// second, validate the values are sensible
	if err := validateDir(c.In); err != nil {
		return fmt.Errorf("-in %w", err)
	}
	if err := validateDir(c.Out); err != nil {
		return fmt.Errorf("-out %w", err)
	}

	if c.Workers < cl.MinWorkers || c.Workers > cl.MaxWorkers {
		return fmt.Errorf("-workers must be between %d and %d, %w", cl.MinWorkers, cl.MaxWorkers, incorrectVal)
	}

	if c.Rps < cl.MinRps || c.Rps > cl.MaxRps {
		return fmt.Errorf("-rps must be between %f and %f, %w", cl.MinRps, cl.MaxRps, incorrectVal)
	}

	return nil
}

func validateDir(path string) error {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("doesn't exist %w", invalidPath)
	}
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory %w", invalidPath)
	}
	return nil
}

type ConfigLimits struct {
	MinWorkers int
	MaxWorkers int
	MinRps     float64
	MaxRps     float64
}

func LoadConfigLimits() (ConfigLimits, error) {
	configLimits := ConfigLimits{
		MinWorkers: 1,
		MaxWorkers: 100,
		MinRps:     0.5,
		MaxRps:     100,
	}

	err := godotenv.Load(".env")
	if err != nil {
		// if .env file doesn't exist, use defaaults or check env vars
		// set in other ways
		if !errors.Is(err, fs.ErrNotExist) {
			return configLimits, err
		}
	}

	minWorkers, exists := os.LookupEnv("MIN_WORKERS")
	if exists {
		minWorkersInt, err := strconv.Atoi(minWorkers)
		if err != nil {
			return configLimits, fmt.Errorf("MIN_WORKERS env var is not a int, %w", incorrectEnv)
		}
		configLimits.MinWorkers = minWorkersInt
	}

	maxWorkers, exists := os.LookupEnv("MAX_WORKERS")
	if exists {
		maxWorkersInt, err := strconv.Atoi(maxWorkers)
		if err != nil {
			return configLimits, fmt.Errorf("MAX_WORKERS env var is not a int, %w", incorrectEnv)
		}
		configLimits.MaxWorkers = maxWorkersInt
	}

	minRps, exists := os.LookupEnv("MIN_RPS")
	if exists {
		minRpsFloat, err := strconv.ParseFloat(minRps, 64)
		if err != nil {
			return configLimits, fmt.Errorf("MIN_RPS env var is not a float, %w", incorrectEnv)
		}
		configLimits.MinRps = minRpsFloat
	}

	maxRps, exists := os.LookupEnv("MAX_RPS")
	if exists {
		maxRpsFloat, err := strconv.ParseFloat(maxRps, 64)
		if err != nil {
			return configLimits, fmt.Errorf("MAX_RPS env var is not a float, %w", incorrectEnv)
		}
		configLimits.MaxRps = maxRpsFloat
	}

	return configLimits, nil
}

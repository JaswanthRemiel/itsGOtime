package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// loadConfig reads and parses the monitoring configuration from the specified YAML file.
// It loads the file from the given path, unmarshals its contents into a Config struct,
// and sets a default interval of 60 seconds if none is specified.
// Returns the parsed Config and any error encountered during file reading or parsing.
func loadConfig(path string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	if cfg.IntervalSeconds == 0 {
		cfg.IntervalSeconds = 60
	}
	return cfg, nil
}

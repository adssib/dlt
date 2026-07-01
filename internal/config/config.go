// Package config loads dlt's YAML configuration files into typed structs.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration so it can be parsed from YAML strings like
// "5s", "10ms", "1m30s". Use .Std() to get the underlying time.Duration.
type Duration time.Duration

// Std returns the standard library time.Duration.
func (d Duration) Std() time.Duration { return time.Duration(d) }

// UnmarshalYAML parses a duration string via time.ParseDuration.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// TargetConfig is the on-disk shape of configs/target.yaml — the deliberately
// breakable system under test (the "defender").
type TargetConfig struct {
	Target struct {
		Listen        string `yaml:"listen"`
		MetricsListen string `yaml:"metrics_listen"`
		Seed          int64  `yaml:"seed"` // 0 = random per run

		Latency struct {
			Base         Duration `yaml:"base"`
			Jitter       Duration `yaml:"jitter"`
			Distribution string   `yaml:"distribution"` // normal | exponential
		} `yaml:"latency"`

		Concurrency struct {
			Capacity    int    `yaml:"capacity"`
			Slowdown    string `yaml:"slowdown"` // none | linear | quadratic
			MaxInflight int    `yaml:"max_inflight"`
		} `yaml:"concurrency"`

		Tail struct {
			Probability float64  `yaml:"probability"`
			Extra       Duration `yaml:"extra"`
		} `yaml:"tail"`

		Faults struct {
			ErrorRate float64 `yaml:"error_rate"`
			Spike     struct {
				Every     Duration `yaml:"every"`
				Duration  Duration `yaml:"duration"`
				ErrorRate float64  `yaml:"error_rate"`
			} `yaml:"spike"`
		} `yaml:"faults"`

		RateLimit struct {
			Enabled   bool    `yaml:"enabled"`
			Algorithm string  `yaml:"algorithm"` // token_bucket | sliding_window
			Rate      float64 `yaml:"rate"`
			Burst     int     `yaml:"burst"`
		} `yaml:"ratelimit"`
	} `yaml:"target"`
}

// LoadTarget reads and parses a target config file.
func LoadTarget(path string) (*TargetConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read target config: %w", err)
	}
	var cfg TargetConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse target config: %w", err)
	}
	return &cfg, nil
}

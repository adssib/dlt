package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CoordinatorConfig is the on-disk shape of configs/coordinator.yaml: the test
// definition plus coordination settings. Note min_workers is a *readiness
// barrier*, not a spawn count — the orchestrator owns the worker count (ADR-0003).
type CoordinatorConfig struct {
	Coordinator struct {
		Listen         string   `yaml:"listen"`
		MinWorkers     int      `yaml:"min_workers"`
		WaitForWorkers Duration `yaml:"wait_for_workers"`
	} `yaml:"coordinator"`

	Test struct {
		TargetURL        string   `yaml:"target_url"`
		TotalRequests    int      `yaml:"total_requests"`
		TotalConcurrency int      `yaml:"total_concurrency"`
		Timeout          Duration `yaml:"timeout"`
		RampUp           Duration `yaml:"ramp_up"`
	} `yaml:"test"`

	Report struct {
		Format string `yaml:"format"` // text | json
	} `yaml:"report"`
}

// WorkerConfig is the on-disk shape of configs/worker.yaml — deliberately simple:
// where the coordinator is, and this worker's capacity.
type WorkerConfig struct {
	Worker struct {
		Coordinator    string `yaml:"coordinator"`
		MaxConcurrency int    `yaml:"max_concurrency"`
		MetricsListen  string `yaml:"metrics_listen"`
		ID             string `yaml:"id"` // optional; defaults to hostname
	} `yaml:"worker"`
}

// LoadCoordinator reads and parses a coordinator config file.
func LoadCoordinator(path string) (*CoordinatorConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read coordinator config: %w", err)
	}
	var cfg CoordinatorConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse coordinator config: %w", err)
	}
	return &cfg, nil
}

// LoadWorker reads and parses a worker config file.
func LoadWorker(path string) (*WorkerConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read worker config: %w", err)
	}
	var cfg WorkerConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse worker config: %w", err)
	}
	return &cfg, nil
}

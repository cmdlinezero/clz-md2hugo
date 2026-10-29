package generator

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Version int   `json:"version"`
	Jobs    []Job `json:"jobs"`
}

type Job struct {
	Name     string  `json:"name"`
	Source   string  `json:"source"`
	Template string  `json:"template"`
	Output   string  `json:"output"`
	Items    string  `json:"items"`
	Path     string  `json:"path"`
	Merge    []Merge `json:"merge"`
}

type Merge struct {
	Source     string `json:"source"`
	Items      string `json:"items"`
	LocalKey   string `json:"localKey"`
	ForeignKey string `json:"foreignKey"`
	As         string `json:"as"`
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read generator config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse generator config: %w", err)
	}
	if cfg.Version != 1 {
		return Config{}, fmt.Errorf("unsupported generator config version %d", cfg.Version)
	}
	if len(cfg.Jobs) == 0 {
		return Config{}, fmt.Errorf("generator config contains no jobs")
	}
	for i, job := range cfg.Jobs {
		if job.Name == "" || job.Source == "" || job.Template == "" || job.Output == "" || job.Path == "" {
			return Config{}, fmt.Errorf("job %d requires name, source, template, output, and path", i+1)
		}
		seenAliases := map[string]bool{}
		for j, merge := range job.Merge {
			if merge.Source == "" || merge.LocalKey == "" || merge.ForeignKey == "" || merge.As == "" {
				return Config{}, fmt.Errorf("job %d merge %d requires source, localKey, foreignKey, and as", i+1, j+1)
			}
			if seenAliases[merge.As] {
				return Config{}, fmt.Errorf("job %d has duplicate merge alias %q", i+1, merge.As)
			}
			seenAliases[merge.As] = true
		}
	}
	return cfg, nil
}

package agent

import (
	"fmt"
	"time"
)

// Config represents the agent configuration
type Config struct {
	NomadAddr  string        `json:"nomad_addr"`
	NomadToken string        `json:"nomad_token"`
	Sinks      []string      `json:"sinks"`
	EventTypes []string      `json:"event_types"`
	RateLimit  time.Duration `json:"rate_limit"`
	FileConfig FileConfig    `json:"file_config"`
}

// FileConfig holds configuration for file sink
type FileConfig struct {
	Path string `json:"path"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.NomadAddr == "" {
		return fmt.Errorf("nomad address is required")
	}

	if len(c.Sinks) == 0 {
		return fmt.Errorf("at least one sink must be specified")
	}

	for _, sink := range c.Sinks {
		switch sink {
		case "stdout":
			// No additional validation needed
		case "file":
			if c.FileConfig.Path == "" {
				return fmt.Errorf("file path is required when using file sink")
			}
		default:
			return fmt.Errorf("unknown sink type: %s", sink)
		}
	}

	// Validate event types if specified
	if len(c.EventTypes) > 0 {
		validEventTypes := map[string]bool{
			EventTypeAllocation: true,
			EventTypeEvaluation: true,
			EventTypeNode:       true,
			EventTypeJob:        true,
			EventTypeDeployment: true,
			EventTypeTask:       true,
		}

		for _, eventType := range c.EventTypes {
			if !validEventTypes[eventType] {
				return fmt.Errorf("unknown event type: %s", eventType)
			}
		}
	}

	return nil
}

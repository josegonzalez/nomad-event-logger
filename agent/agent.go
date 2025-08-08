package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Agent represents the main event collection agent
type Agent struct {
	config   *Config
	managers []EventManager
	sinks    []Sink
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	logger   *slog.Logger
}

// New creates a new agent with the given configuration
func New(config *Config) (*Agent, error) {
	// Create sinks based on configuration
	var sinks []Sink

	for _, sinkType := range config.Sinks {
		var sink Sink
		var err error

		switch sinkType {
		case "stdout":
			sink = NewStdoutSink()
		case "file":
			sink, err = NewFileSink(config.FileConfig.Path)
		default:
			return nil, fmt.Errorf("unknown sink type: %s", sinkType)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create %s sink: %w", sinkType, err)
		}

		sinks = append(sinks, sink)
	}

	// Determine which event types to monitor
	eventTypes := config.EventTypes
	if len(eventTypes) == 0 {
		// Default to all event types if none specified
		eventTypes = []string{
			EventTypeAllocation,
			EventTypeEvaluation,
			EventTypeNode,
			EventTypeJob,
			EventTypeDeployment,
			EventTypeTask,
		}
	}

	// Create event managers for specified event types
	var managers []EventManager

	for _, eventType := range eventTypes {
		var manager EventManager
		var err error

		switch eventType {
		case EventTypeAllocation:
			manager, err = NewAllocationManager(config.NomadAddr, config.NomadToken, sinks, config.RateLimit)
		case EventTypeEvaluation:
			manager, err = NewEvaluationManager(config.NomadAddr, config.NomadToken, sinks)
		case EventTypeNode:
			manager, err = NewNodeManager(config.NomadAddr, config.NomadToken, sinks)
		case EventTypeJob:
			manager, err = NewJobManager(config.NomadAddr, config.NomadToken, sinks)
		case EventTypeDeployment:
			manager, err = NewDeploymentManager(config.NomadAddr, config.NomadToken, sinks)
		case EventTypeTask:
			// Task events are handled by the AllocationManager
			manager, err = NewAllocationManager(config.NomadAddr, config.NomadToken, sinks, config.RateLimit)
		default:
			return nil, fmt.Errorf("unknown event type: %s", eventType)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create %s manager: %w", eventType, err)
		}

		managers = append(managers, manager)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:   config,
		managers: managers,
		sinks:    sinks,
		ctx:      ctx,
		cancel:   cancel,
		logger:   GetLogger(),
	}, nil
}

// createSinks creates sink instances based on configuration
func createSinks(config *Config) ([]Sink, error) {
	var sinks []Sink

	for _, sinkType := range config.Sinks {
		switch sinkType {
		case "stdout":
			sinks = append(sinks, NewStdoutSink())
		case "file":
			fileSink, err := NewFileSink(config.FileConfig.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to create file sink: %w", err)
			}
			sinks = append(sinks, fileSink)
		default:
			return nil, fmt.Errorf("unknown sink type: %s", sinkType)
		}
	}

	return sinks, nil
}

// Start starts the agent and all event managers
func (a *Agent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Starting Nomad event collection agent")

	// Start all event managers
	for _, manager := range a.managers {
		a.wg.Add(1)
		go func(m EventManager) {
			defer a.wg.Done()
			if err := m.Start(a.ctx); err != nil {
				a.logger.Error("Manager failed to start",
					"event_type", m.GetEventType(),
					"error", err.Error(),
				)
			}
		}(manager)
	}

	a.logger.Info("Agent started successfully",
		"event_manager_count", len(a.managers),
		"rate_limit_seconds", a.config.RateLimit.Seconds(),
		"sinks", len(a.sinks),
	)
	return nil
}

// Stop stops the agent and all event managers
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Stopping Nomad event collection agent")

	// Cancel context to stop all managers
	a.cancel()

	// Wait for all managers to stop
	a.wg.Wait()

	// Close all sinks
	for _, sink := range a.sinks {
		if err := sink.Close(); err != nil {
			a.logger.Error("Failed to close sink",
				"error", err.Error(),
			)
		}
	}

	a.logger.Info("Agent stopped successfully")
	return nil
}

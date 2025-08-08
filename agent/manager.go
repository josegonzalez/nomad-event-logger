package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// EventManager defines the interface for managing different types of Nomad events
type EventManager interface {
	Start(ctx context.Context) error
	Stop() error
	GetEventType() string
}

// BaseManager provides common functionality for all event managers
type BaseManager struct {
	nomadClient  *nomadapi.Client
	sinks        []Sink
	eventType    string
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	lastIndex    uint64
	logger       *slog.Logger
	firstRun     bool
	lastCallTime time.Time
	rateLimit    time.Duration
}

// NewBaseManager creates a new base manager
func NewBaseManager(nomadAddr, nomadToken string, sinks []Sink, eventType string, rateLimit time.Duration) (*BaseManager, error) {
	config := nomadapi.DefaultConfig()
	config.Address = nomadAddr
	if nomadToken != "" {
		config.SecretID = nomadToken
	}

	client, err := nomadapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nomad client: %w", err)
	}

	return &BaseManager{
		nomadClient: client,
		sinks:       sinks,
		eventType:   eventType,
		stopChan:    make(chan struct{}),
		lastIndex:   0,
		logger:      GetLogger(),
		firstRun:    true,
		rateLimit:   rateLimit,
	}, nil
}

// GetEventType returns the event type this manager handles
func (m *BaseManager) GetEventType() string {
	return m.eventType
}

// WriteEvent writes an event to all configured sinks
func (m *BaseManager) WriteEvent(event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sink := range m.sinks {
		if err := sink.Write(event); err != nil {
			m.logger.Error("Failed to write event to sink",
				"event_type", m.eventType,
				"error", err.Error(),
			)
		}
	}
	return nil
}

// Stop stops the manager
func (m *BaseManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.stopChan)
	m.wg.Wait()
	return nil
}

// runWatcherWithRateLimit provides a common rate limiting wrapper for all managers
func (m *BaseManager) runWatcherWithRateLimit(ctx context.Context, watchFunc func(context.Context) error) {
	for {
		select {
		case <-m.stopChan:
			return
		default:
		}

		// Check if enough time has passed since the last call
		now := time.Now()
		if now.Sub(m.lastCallTime) < m.rateLimit {
			// Wait for the remaining time
			sleepDuration := m.rateLimit - now.Sub(m.lastCallTime)
			time.Sleep(sleepDuration)
			continue
		}

		// Update last call time before making the call
		m.lastCallTime = now

		if err := watchFunc(ctx); err != nil {
			m.logger.Error("Watcher error, retrying in 5 seconds",
				"event_type", m.eventType,
				"error", err.Error(),
			)
			time.Sleep(5 * time.Second)
		}
	}
}

// markFirstRunComplete marks the first run as complete
func (m *BaseManager) markFirstRunComplete() {
	if m.firstRun {
		m.logger.Info("First run completed, events will be processed on subsequent runs",
			"event_type", m.eventType)
		m.firstRun = false
	}
}

// isFirstRun returns whether this is the first run
func (m *BaseManager) isFirstRun() bool {
	return m.firstRun
}

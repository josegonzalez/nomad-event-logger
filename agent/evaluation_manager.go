package agent

import (
	"context"
	"fmt"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// EvaluationManager manages evaluation events
type EvaluationManager struct {
	*BaseManager
}

func NewEvaluationManager(nomadAddr, nomadToken string, sinks []Sink) (*EvaluationManager, error) {
	baseManager, err := NewBaseManager(nomadAddr, nomadToken, sinks, EventTypeEvaluation, 0) // No rate limit for evaluations
	if err != nil {
		return nil, err
	}
	return &EvaluationManager{
		BaseManager: baseManager,
	}, nil
}

func (m *EvaluationManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runEvaluationWatcher(ctx)
	}()
	return nil
}

func (m *EvaluationManager) runEvaluationWatcher(ctx context.Context) {
	m.runWatcherWithRateLimit(ctx, func(ctx context.Context) error {
		return m.watchEvaluations(ctx)
	})
}

func (m *EvaluationManager) watchEvaluations(ctx context.Context) error {
	// Use the Nomad API to poll evaluations
	opts := &nomadapi.QueryOptions{
		AllowStale: true,
		WaitIndex:  m.lastIndex,
		WaitTime:   30 * time.Second,
	}

	// Get evaluations with blocking query
	evaluations, meta, err := m.nomadClient.Evaluations().List(opts)
	if err != nil {
		return fmt.Errorf("failed to get evaluations: %w", err)
	}

	// Process evaluations
	for _, eval := range evaluations {
		// Skip event output on first run
		if m.isFirstRun() {
			continue
		}

		// Skip evaluations that have not changed since last query
		if eval.ModifyIndex <= m.lastIndex {
			continue
		}

		// Skip any evaluations that do not have a SnapshotIndex
		// as these are not yet persisted to the nomad raft state
		if eval.SnapshotIndex == 0 {
			continue
		}

		nomadEvent := NewEvent(EventTypeEvaluation, eval)
		if err := m.WriteEvent(nomadEvent); err != nil {
			m.logger.Error("Failed to write evaluation event",
				"error", err.Error(),
			)
		}
	}

	// Update last index
	if meta != nil && meta.LastIndex > m.lastIndex {
		m.lastIndex = meta.LastIndex
	}

	// Mark first run as complete after processing
	m.markFirstRunComplete()

	return nil
}

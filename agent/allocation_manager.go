package agent

import (
	"context"
	"fmt"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// AllocationManager manages allocation events
type AllocationManager struct {
	*BaseManager
	lastSeenTime int64
}

func NewAllocationManager(nomadAddr, nomadToken string, sinks []Sink, rateLimit time.Duration) (*AllocationManager, error) {
	baseManager, err := NewBaseManager(nomadAddr, nomadToken, sinks, EventTypeAllocation, rateLimit)
	if err != nil {
		return nil, err
	}
	return &AllocationManager{
		BaseManager: baseManager,
	}, nil
}

func (m *AllocationManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runAllocationWatcher(ctx)
	}()
	return nil
}

func (m *AllocationManager) runAllocationWatcher(ctx context.Context) {
	m.runWatcherWithRateLimit(ctx, func(ctx context.Context) error {
		return m.watchAllocations(ctx)
	})
}

func (m *AllocationManager) watchAllocations(ctx context.Context) error {
	// Use the Nomad API to poll allocations
	opts := &nomadapi.QueryOptions{
		AllowStale: true,
		WaitIndex:  m.lastIndex,
		WaitTime:   30 * time.Second,
	}

	// Get allocations with blocking query
	allocations, meta, err := m.nomadClient.Allocations().List(opts)
	if err != nil {
		return fmt.Errorf("failed to get allocations: %w", err)
	}

	// Update last index
	if meta != nil && meta.LastIndex > m.lastIndex {
		m.lastIndex = meta.LastIndex
	}

	// Track the maximum task event time across all allocations in this query
	maxEventTime := m.lastSeenTime

	// Process allocations for task events
	for _, allocStub := range allocations {
		allocationMaxTime, err := m.processAllocationTaskEvents(allocStub)
		if err != nil {
			m.logger.Error("Failed to process allocation task events",
				"allocation_id", allocStub.ID,
				"error", err.Error(),
				"job_id", allocStub.JobID,
			)
		}

		// Update max event time if this allocation had newer events
		if allocationMaxTime > maxEventTime {
			maxEventTime = allocationMaxTime
		}
	}

	// Update the global last seen time to the maximum found in this query
	if maxEventTime > m.lastSeenTime {
		m.lastSeenTime = maxEventTime
	}

	// Mark first run as complete after processing
	m.markFirstRunComplete()

	return nil
}

func (m *AllocationManager) processAllocationTaskEvents(alloc *nomadapi.AllocationListStub) (int64, error) {
	if alloc.TaskStates == nil {
		return m.lastSeenTime, nil
	}

	maxEventTime := m.lastSeenTime

	for taskName, taskState := range alloc.TaskStates {
		if taskState.Events == nil {
			continue
		}

		// Process unseen task events
		for _, event := range taskState.Events {
			// Convert Unix timestamp to time.Time
			eventTime := event.Time

			// Skip events we've already seen
			if eventTime < m.lastSeenTime {
				continue
			}

			// Update local max event time for this allocation
			if eventTime > maxEventTime {
				maxEventTime = eventTime
			}

			// Skip event output on first run, only track time
			if m.isFirstRun() {
				continue
			}

			taskStateMap := map[string]any{
				"State":       taskState.State,
				"Failed":      taskState.Failed,
				"Restarts":    taskState.Restarts,
				"LastRestart": taskState.LastRestart,
				"StartedAt":   taskState.StartedAt,
				"FinishedAt":  taskState.FinishedAt,
			}

			// Create task event
			taskEvent := NewTaskEvent(alloc, taskName, event, taskStateMap)
			nomadEvent := NewEvent(EventTypeTask, taskEvent)

			// Write event
			if err := m.WriteEvent(nomadEvent); err != nil {
				m.logger.Error("Failed to write task event",
					"allocation_id", alloc.ID,
					"job_id", alloc.JobID,
					"task_name", taskName,
					"error", err.Error(),
				)
			}
		}
	}

	return maxEventTime, nil
}

package agent

import (
	"context"
	"fmt"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// JobManager manages job events
type JobManager struct {
	*BaseManager
}

func NewJobManager(nomadAddr, nomadToken string, sinks []Sink) (*JobManager, error) {
	baseManager, err := NewBaseManager(nomadAddr, nomadToken, sinks, EventTypeJob, 0) // No rate limit for jobs
	if err != nil {
		return nil, err
	}
	return &JobManager{
		BaseManager: baseManager,
	}, nil
}

func (m *JobManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runJobWatcher(ctx)
	}()
	return nil
}

func (m *JobManager) runJobWatcher(ctx context.Context) {
	m.runWatcherWithRateLimit(ctx, func(ctx context.Context) error {
		return m.watchJobs(ctx)
	})
}

func (m *JobManager) watchJobs(ctx context.Context) error {
	// Use the Nomad API to poll jobs
	opts := &nomadapi.QueryOptions{
		WaitIndex: m.lastIndex,
		WaitTime:  30 * time.Second,
	}

	// Get jobs with blocking query
	jobs, _, err := m.nomadClient.Jobs().List(opts)
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	// Track the maximum ModifyIndex from all jobs in this query
	maxModifyIndex := m.lastIndex

	// Process jobs
	for _, job := range jobs {
		// Only process jobs whose ModifyIndex is greater than our lastIndex
		if job.ModifyIndex <= m.lastIndex {
			continue
		}

		// Update max ModifyIndex if this job has a newer index
		if job.ModifyIndex > maxModifyIndex {
			maxModifyIndex = job.ModifyIndex
		}

		// Skip event output on first run
		if m.isFirstRun() {
			continue
		}

		nomadEvent := NewEvent(EventTypeJob, job)
		if err := m.WriteEvent(nomadEvent); err != nil {
			m.logger.Error("Failed to write job event",
				"error", err.Error(),
			)
		}
	}

	// Update the manager's lastIndex to the maximum ModifyIndex found
	if maxModifyIndex > m.lastIndex {
		m.lastIndex = maxModifyIndex
	}

	// Mark first run as complete after processing
	m.markFirstRunComplete()

	return nil
}

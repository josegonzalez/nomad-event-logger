package agent

import (
	"context"
	"fmt"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// DeploymentManager manages deployment events
type DeploymentManager struct {
	*BaseManager
}

func NewDeploymentManager(nomadAddr, nomadToken string, sinks []Sink) (*DeploymentManager, error) {
	baseManager, err := NewBaseManager(nomadAddr, nomadToken, sinks, EventTypeDeployment, 0) // No rate limit for deployments
	if err != nil {
		return nil, err
	}
	return &DeploymentManager{
		BaseManager: baseManager,
	}, nil
}

func (m *DeploymentManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runDeploymentWatcher(ctx)
	}()
	return nil
}

func (m *DeploymentManager) runDeploymentWatcher(ctx context.Context) {
	m.runWatcherWithRateLimit(ctx, func(ctx context.Context) error {
		return m.watchDeployments(ctx)
	})
}

func (m *DeploymentManager) watchDeployments(ctx context.Context) error {
	// Use the Nomad API to poll deployments
	opts := &nomadapi.QueryOptions{
		WaitIndex: m.lastIndex,
		WaitTime:  30 * time.Second,
	}

	// Get deployments with blocking query
	deployments, meta, err := m.nomadClient.Deployments().List(opts)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}

	// Process deployments
	for _, deployment := range deployments {
		// Skip event output on first run
		if m.isFirstRun() {
			continue
		}

		// Skip deployments that have not changed since last query
		if deployment.ModifyIndex <= m.lastIndex {
			continue
		}

		nomadEvent := NewEvent(EventTypeDeployment, deployment)
		if err := m.WriteEvent(nomadEvent); err != nil {
			m.logger.Error("Failed to write deployment event",
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

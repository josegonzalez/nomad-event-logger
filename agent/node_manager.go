package agent

import (
	"context"
	"fmt"
	"time"

	nomadapi "github.com/hashicorp/nomad/api"
)

// NodeManager manages node events
type NodeManager struct {
	*BaseManager
}

func NewNodeManager(nomadAddr, nomadToken string, sinks []Sink) (*NodeManager, error) {
	baseManager, err := NewBaseManager(nomadAddr, nomadToken, sinks, EventTypeNode, 0) // No rate limit for nodes
	if err != nil {
		return nil, err
	}
	return &NodeManager{
		BaseManager: baseManager,
	}, nil
}

func (m *NodeManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.runNodeWatcher(ctx)
	}()
	return nil
}

func (m *NodeManager) runNodeWatcher(ctx context.Context) {
	m.runWatcherWithRateLimit(ctx, func(ctx context.Context) error {
		return m.watchNodes(ctx)
	})
}

func (m *NodeManager) watchNodes(ctx context.Context) error {
	// Use the Nomad API to poll nodes
	opts := &nomadapi.QueryOptions{
		WaitIndex: m.lastIndex,
		WaitTime:  30 * time.Second,
	}

	// Get nodes with blocking query
	nodes, meta, err := m.nomadClient.Nodes().List(opts)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	// Process nodes
	for _, node := range nodes {
		// Skip event output on first run
		if m.isFirstRun() {
			continue
		}

		// Skip nodes that have not changed since last query
		if node.ModifyIndex <= m.lastIndex {
			continue
		}

		nomadEvent := NewEvent(EventTypeNode, node)
		if err := m.WriteEvent(nomadEvent); err != nil {
			m.logger.Error("Failed to write node event",
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

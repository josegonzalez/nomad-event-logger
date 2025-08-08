package agent

import (
	"encoding/json"
	"time"

	"github.com/hashicorp/nomad/api"
)

// Event represents a Nomad event with metadata
type Event struct {
	Time time.Time `json:"time"`
	Type string    `json:"type"`
	Data any       `json:"data"`
}

// TaskEvent represents a task event with allocation and task information
type TaskEvent struct {
	// Allocation information
	AllocationName     string `json:"AllocationName"`
	AllocationID       string `json:"AllocationID"`
	NodeID             string `json:"NodeID"`
	EvalID             string `json:"EvalID"`
	DesiredStatus      string `json:"DesiredStatus"`
	DesiredDescription string `json:"DesiredDescription"`
	ClientStatus       string `json:"ClientStatus"`
	ClientDescription  string `json:"ClientDescription"`
	JobID              string `json:"JobID"`
	TaskGroup          string `json:"TaskGroup"`

	// Task information
	TaskName  string         `json:"TaskName"`
	TaskEvent *api.TaskEvent `json:"TaskEvent"`
	TaskInfo  map[string]any `json:"TaskInfo"`
}

// NewEvent creates a new event with the current time
func NewEvent(eventType string, data any) *Event {
	return &Event{
		Time: time.Now(),
		Type: eventType,
		Data: data,
	}
}

// NewTaskEvent creates a new task event
func NewTaskEvent(allocation *api.AllocationListStub, taskName string, taskEvent *api.TaskEvent, taskInfo map[string]any) *TaskEvent {
	return &TaskEvent{
		AllocationName:     allocation.Name,
		AllocationID:       allocation.ID,
		NodeID:             allocation.NodeID,
		EvalID:             allocation.EvalID,
		DesiredStatus:      allocation.DesiredStatus,
		DesiredDescription: allocation.DesiredDescription,
		ClientStatus:       allocation.ClientStatus,
		ClientDescription:  allocation.ClientDescription,
		JobID:              allocation.JobID,
		TaskGroup:          allocation.TaskGroup,
		TaskName:           taskName,
		TaskEvent:          taskEvent,
		TaskInfo:           taskInfo,
	}
}

// ToJSON converts the event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Event types
const (
	EventTypeAllocation = "allocation"
	EventTypeEvaluation = "evaluation"
	EventTypeNode       = "node"
	EventTypeJob        = "job"
	EventTypeDeployment = "deployment"
	EventTypeTask       = "task"
)

# Nomad Event Logger

A Go-based event collection agent that processes Hashicorp Nomad events and dumps them to external sink providers. The agent supports multiple event types and can output to various destinations.

## Features

- **Nomad SDK Integration**: Uses the official Nomad Go API for reliable event collection
- **JSON Logging**: Structured JSON logging with custom timestamp format
- **Multiple Sink Support**: Output to stdout or files
- **Event Type Filtering**: Monitor specific event types or all events
- **Rate Limiting**: Configurable rate limiting for allocation queries to prevent API overload
- **Task Event Tracking**: Comprehensive task event monitoring with allocation context

## Event Types

The agent collects the following Nomad event types:

- **Allocation**: Job allocation events (start, stop, update, etc.)
- **Evaluation**: Job evaluation events (triggered, completed, failed, etc.)
- **Node**: Node lifecycle events (registration, deregistration, etc.)
- **Job**: Job lifecycle events (submission, update, deletion, etc.)
- **Deployment**: Deployment events (start, success, failure, etc.)
- **Task**: Task events from allocation TaskStates (start, stop, restart, etc.)

## Sink Providers

### Stdout Sink

Outputs events to standard output in JSON format.

### File Sink

Writes events to a specified file, one event per line in JSON format.

## Installation

```bash
go install github.com/josegonzalez/nomad-event-logger@latest
```

## Usage

### Basic Usage

Start the agent with default settings (stdout sink):

```bash
nomad-event-logger start
```

### With Custom Configuration

```bash
nomad-event-logger start \
  --nomad-addr http://localhost:4646 \
  --nomad-token your-token \
  --sinks stdout,file \
  --file-path /tmp/nomad-events.json
```

### Multiple Sinks

```bash
nomad-event-logger start \
  --nomad-addr http://localhost:4646 \
  --sinks stdout,file \
  --file-path /tmp/nomad-events.json \
```

### Monitor Specific Event Types

Monitor only allocation and job events:

```bash
nomad-event-logger start \
  --nomad-addr http://localhost:4646 \
  --event-types allocation,job
```

Monitor only deployment events:

```bash
nomad-event-logger start \
  --nomad-addr http://localhost:4646 \
  --event-types deployment
```

Monitor all event types (default behavior):

```bash
nomad-event-logger start \
  --nomad-addr http://localhost:4646
```

## Configuration

### Command Line Flags

- `--nomad-addr`: Nomad server address (default: <http://localhost:4646>)
- `--nomad-token`: Nomad ACL token (optional)
- `--sinks`: Comma-separated list of sink providers (default: stdout)
- `--event-types`: Comma-separated list of event types to monitor (allocation, evaluation, node, job, deployment, task). Defaults to all if not specified.
- `--rate-limit`: Rate limit for allocation queries (e.g., 5s, 1m). Defaults to 5 seconds.
- `--file-path`: File path for file sink (default: /tmp/nomad-events.json)

### Configuration File

You can also use a configuration file. The default location is `$HOME/.nomad-event-logger.yaml`:

```yaml
nomad:
  addr: http://localhost:4646
  token: your-token

sinks:
  - stdout
  - file

event_types:
  - allocation
  - job
  - deployment

file:
  path: /tmp/nomad-events.json

```

## Usage Examples

### Basic Usage

```bash
# Start with default settings
./nomad-event-logger start

# Specify Nomad server address
./nomad-event-logger start --nomad-addr http://nomad.example.com:4646

# Use custom sinks
./nomad-event-logger start --sinks stdout,file --file-path /var/log/nomad-events.json
```

### Monitor Specific Event Types

```bash
# Monitor only allocation events
./nomad-event-logger start --event-types allocation

# Monitor multiple event types
./nomad-event-logger start --event-types allocation,evaluation,node

# Monitor task events with rate limiting
./nomad-event-logger start --event-types task --rate-limit 10s
```

### Rate Limiting

```bash
# Use default 5-second rate limit
./nomad-event-logger start --event-types task

# Custom rate limit (10 seconds)
./nomad-event-logger start --event-types task --rate-limit 10s

# Very conservative rate limit (1 minute)
./nomad-event-logger start --event-types task --rate-limit 1m
```

## Event Format

All events are output in JSON format with the following structure:

```json
{
  "time": "2022-01-01T00:00:00Z",
  "type": "allocation",
  "data": {
    // Raw Nomad event data
  }
}
```

- `time`: RFC3339 formatted timestamp when the event was received
- `type`: Event type (allocation, evaluation, node, job, deployment, task)
- `data`: Raw event data from Nomad

### Task Event Format

Task events include comprehensive allocation and task information:

```json
{
  "time": "2022-01-01T00:00:00Z",
  "type": "task",
  "data": {
    "AllocationName": "example.abc123",
    "AllocationID": "abc123",
    "NodeID": "node-1",
    "EvalID": "eval-456",
    "DesiredStatus": "run",
    "DesiredDescription": "Task is running",
    "ClientStatus": "running",
    "ClientDescription": "Task is running",
    "JobID": "example",
    "TaskGroup": "web",
    "TaskName": "web",
    "TaskEvent": {
      "Type": "Started",
      "Time": 1640995200,
      "Message": "Task started by client",
      "ExitCode": 0
    },
    "TaskInfo": {
      "State": "running",
      "Failed": false,
      "Restarts": 0,
      "StartedAt": "2022-01-01T00:00:00Z"
    }
  }
}
```

## Rate Limiting

The agent implements configurable rate limiting for allocation queries to prevent API overload:

- **Default**: 5 seconds between allocation queries
- **Configurable**: Use `--rate-limit` flag to set custom intervals
- **Sliding Window**: Ensures maximum of 1 query per time window
- **Prevents Flooding**: Avoids overwhelming the Nomad API with rapid requests

### Rate Limit Examples

```bash
# Conservative (1 query per minute)
./nomad-event-logger start --rate-limit 1m

# Moderate (1 query per 10 seconds)
./nomad-event-logger start --rate-limit 10s

# Aggressive (1 query per second)
./nomad-event-logger start --rate-limit 1s
```

## Logging

The agent uses structured JSON logging for better observability and log aggregation. All log messages are output in JSON format with the following structure:

```json
{
  "time": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "msg": "Agent started successfully",
  "event_manager_count": 5,
  "rate_limit_seconds": 5,
  "sinks": 2
}
```

### Log Levels

- **INFO**: General operational messages (startup, shutdown, successful operations)
- **ERROR**: Error conditions that don't stop the application (connection failures, sink write errors)

### Log Fields

Common log fields include:

- `event_type`: The type of event being processed
- `error`: Error message when operations fail
- `event_managers`: Number of active event managers
- `sinks`: Number of configured sinks

This structured logging format makes it easy to:

- Parse logs with log aggregation tools (ELK, Splunk, etc.)
- Filter and search logs by specific fields
- Monitor application health and performance
- Debug issues with specific event types or operations

## Architecture

The agent uses the official HashiCorp Nomad API to collect events through blocking queries. This approach provides:

- **Reliability**: Uses the official Nomad SDK for stable API interactions
- **Efficiency**: Blocking queries reduce unnecessary polling while ensuring real-time updates
- **Compatibility**: Works with all Nomad versions that support the API
- **Authentication**: Supports Nomad ACL tokens for secure access

Each event type has its own manager that:

- Establishes connections to Nomad using the official API
- Uses blocking queries to efficiently monitor for changes
- Handles reconnections automatically
- Writes events to all configured sinks
- Maintains state using index-based polling for optimal performance

## Development

### Building from Source

```bash
git clone https://github.com/josegonzalez/nomad-event-logger.git
cd nomad-event-logger
go build -o nomad-event-logger .
```

### Running Tests

```bash
go test ./...
```

## Requirements

- Go 1.21 or later
- Access to a Nomad server
- Optional: Nomad ACL token for authenticated access

## License

This project is licensed under the MIT License.

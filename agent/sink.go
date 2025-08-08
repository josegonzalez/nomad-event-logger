package agent

import (
	"fmt"
	"os"
	"sync"
)

// Sink defines the interface for event output providers
type Sink interface {
	Write(event *Event) error
	Close() error
}

// StdoutSink writes events to stdout
type StdoutSink struct {
	mu sync.Mutex
}

func NewStdoutSink() *StdoutSink {
	return &StdoutSink{}
}

func (s *StdoutSink) Write(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func (s *StdoutSink) Close() error {
	return nil
}

// FileSink writes events to a file
type FileSink struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileSink(path string) (*FileSink, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	return &FileSink{file: file}, nil
}

func (s *FileSink) Write(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return s.file.Sync()
}

func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

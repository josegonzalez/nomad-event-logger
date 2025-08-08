package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josegonzalez/nomad-event-logger/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Nomad event collection agent",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Add flags
	startCmd.Flags().String("nomad-addr", "http://localhost:4646", "Nomad server address")
	startCmd.Flags().String("nomad-token", "", "Nomad ACL token")
	startCmd.Flags().StringSlice("sinks", []string{"stdout"}, "Sink providers (stdout, file)")
	startCmd.Flags().StringSlice("event-types", []string{}, "Event types to monitor (allocation, evaluation, node, job, deployment, task). Defaults to all if not specified.")
	startCmd.Flags().String("file-path", "/tmp/nomad-events.json", "File path for file sink")
	startCmd.Flags().Duration("rate-limit", 5*time.Second, "Rate limit for allocation queries (e.g., 5s, 1m)")

	// Bind flags to viper
	viper.BindPFlag("nomad_addr", startCmd.Flags().Lookup("nomad-addr"))
	viper.BindPFlag("nomad_token", startCmd.Flags().Lookup("nomad-token"))
	viper.BindPFlag("sinks", startCmd.Flags().Lookup("sinks"))
	viper.BindPFlag("event_types", startCmd.Flags().Lookup("event-types"))
	viper.BindPFlag("file_config.path", startCmd.Flags().Lookup("file-path"))
	viper.BindPFlag("rate_limit", startCmd.Flags().Lookup("rate-limit"))
}

func runStart(cmd *cobra.Command, args []string) error {
	config := &agent.Config{
		NomadAddr:  viper.GetString("nomad_addr"),
		NomadToken: viper.GetString("nomad_token"),
		Sinks:      viper.GetStringSlice("sinks"),
		EventTypes: viper.GetStringSlice("event_types"),
		RateLimit:  viper.GetDuration("rate_limit"),
		FileConfig: agent.FileConfig{
			Path: viper.GetString("file_config.path"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create and start agent
	eventAgent, err := agent.New(config)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Start the agent
	if err := eventAgent.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown gracefully
	slog.Info("Shutting down agent")
	return eventAgent.Stop()
}

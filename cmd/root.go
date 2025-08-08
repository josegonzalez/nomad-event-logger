package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	nomadAddr  string
	nomadToken string
)

var rootCmd = &cobra.Command{
	Use:   "nomad-event-logger",
	Short: "A Nomad event collection agent",
	Long: `A tool that processes Hashicorp Nomad Events and dumps them to external sink providers.
Supports stdout and file sinks with JSON formatted output.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nomad-event-logger.yaml)")
	rootCmd.PersistentFlags().StringVar(&nomadAddr, "nomad-addr", "", "Nomad server address")
	rootCmd.PersistentFlags().StringVar(&nomadToken, "nomad-token", "", "Nomad ACL token")

	if err := viper.BindPFlag("nomad.addr", rootCmd.PersistentFlags().Lookup("nomad-addr")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("nomad.token", rootCmd.PersistentFlags().Lookup("nomad-token")); err != nil {
		cobra.CheckErr(err)
	}

	// Bind environment variables
	if err := viper.BindEnv("nomad.addr", "NOMAD_ADDR"); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindEnv("nomad.token", "NOMAD_TOKEN"); err != nil {
		cobra.CheckErr(err)
	}

	// Set environment variable defaults
	viper.SetDefault("nomad.addr", "http://localhost:4646")
	viper.SetDefault("nomad.token", "")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nomad-event-logger")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

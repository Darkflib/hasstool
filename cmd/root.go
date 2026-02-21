package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"hasstool/ha"
)

var (
	haURL   string
	haToken string
)

var rootCmd = &cobra.Command{
	Use:   "hasstool",
	Short: "CLI tool for Home Assistant",
	Long:  "hasstool interacts with a Home Assistant instance via its REST and WebSocket APIs.",
}

// NewClient builds an ha.Client from the global flags, exiting on missing config.
func NewClient() *ha.Client {
	if haURL == "" {
		fmt.Fprintln(os.Stderr, "error: --url is required (or set HA_URL)")
		os.Exit(1)
	}
	// Fall back to env var since we don't set it as the flag default.
	token := haToken
	if token == "" {
		token = os.Getenv("HA_TOKEN")
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: --token is required (or set HA_TOKEN)")
		os.Exit(1)
	}
	return ha.NewClient(haURL, token)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&haURL, "url", os.Getenv("HA_URL"),
		"Home Assistant base URL (e.g. http://homeassistant.local:8123)")

	// Don't pass the env value as a default — that causes cobra to print the
	// full token in --help output. Read it manually in NewClient instead.
	rootCmd.PersistentFlags().StringVar(&haToken, "token", "",
		"Long-lived access token (or set HA_TOKEN)")
}

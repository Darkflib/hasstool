package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"hasstool/ha"
)

var watchCmd = &cobra.Command{
	Use:   "watch [entity_id]",
	Short: "Watch state changes via WebSocket",
	Long: `Stream state_changed events. Press Ctrl-C to stop.

  hasstool watch                      # all entities
  hasstool watch light.living_room    # single entity`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient()
		outputJSON, _ := cmd.Flags().GetBool("json")

		entityFilter := ""
		if len(args) == 1 {
			entityFilter = args[0]
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			cancel()
		}()

		fmt.Fprintln(os.Stderr, "Watching state changes (Ctrl-C to stop)...")

		err := client.WatchStateChanges(ctx, entityFilter, func(data ha.StateChangedData) {
			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.Encode(data)
				return
			}
			oldState := "<none>"
			if data.OldState != nil {
				oldState = data.OldState.State
			}
			newState := "<none>"
			if data.NewState != nil {
				newState = data.NewState.State
			}
			fmt.Printf("%-40s %s -> %s\n", data.EntityID, oldState, newState)
		})

		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	watchCmd.Flags().Bool("json", false, "Output events as JSON")
	rootCmd.AddCommand(watchCmd)
}

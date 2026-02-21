package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"hasstool/ha"
)

var callCmd = &cobra.Command{
	Use:   "call <domain.service> [entity_id...]",
	Short: "Call a Home Assistant service",
	Long: `Call a service, optionally targeting specific entities.

  hasstool call light.turn_on light.living_room
  hasstool call light.turn_on light.living_room --data '{"brightness": 128}'
  hasstool call input_boolean.toggle input_boolean.vacation_mode`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := NewClient()

		parts := strings.SplitN(args[0], ".", 2)
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "error: service must be in the form domain.service (e.g. light.turn_on)")
			os.Exit(1)
		}
		domain, service := parts[0], parts[1]

		var target *ha.Target
		if len(args) > 1 {
			target = &ha.Target{EntityID: args[1:]}
		}

		var serviceData map[string]any
		if rawData, _ := cmd.Flags().GetString("data"); rawData != "" {
			if err := json.Unmarshal([]byte(rawData), &serviceData); err != nil {
				fmt.Fprintln(os.Stderr, "error: --data is not valid JSON:", err)
				os.Exit(1)
			}
		}

		result, err := client.CallService(ctx, domain, service, target, serviceData)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		if !result.Success {
			fmt.Fprintf(os.Stderr, "error: service call failed: %s: %s\n",
				result.Error.Code, result.Error.Message)
			os.Exit(1)
		}

		fmt.Println("OK")
	},
}

func init() {
	callCmd.Flags().String("data", "", "Extra service data as JSON (e.g. '{\"brightness\": 128}')")
	rootCmd.AddCommand(callCmd)
}

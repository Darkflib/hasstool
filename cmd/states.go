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

var statesCmd = &cobra.Command{
	Use:   "states [entity_id]",
	Short: "Get entity state(s)",
	Long: `Retrieve one or all entity states.

  hasstool states                              # all states
  hasstool states light.living_room            # single entity
  hasstool states light.living_room --attr brightness
  hasstool states --domain light --attr brightness`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := NewClient()

		outputJSON, _ := cmd.Flags().GetBool("json")
		filterDomain, _ := cmd.Flags().GetString("domain")
		attrKey, _ := cmd.Flags().GetString("attr")

		printState := func(s *ha.State) {
			if outputJSON {
				printJSON(s)
				return
			}
			if attrKey != "" {
				val, ok := s.Attributes[attrKey]
				if !ok {
					fmt.Fprintf(os.Stderr, "%s: attribute %q not found\n", s.EntityID, attrKey)
					return
				}
				fmt.Printf("%-40s %v\n", s.EntityID, val)
				return
			}
			fmt.Printf("%-40s %s\n", s.EntityID, s.State)
		}

		if len(args) == 1 {
			state, err := client.GetState(ctx, args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			printState(state)
			return
		}

		states, err := client.GetStates(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		for i := range states {
			s := &states[i]
			if filterDomain != "" && !strings.HasPrefix(s.EntityID, filterDomain+".") {
				continue
			}
			if attrKey != "" {
				if _, ok := s.Attributes[attrKey]; !ok {
					continue // skip entities that don't have the attribute
				}
			}
			printState(s)
		}
	},
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func init() {
	statesCmd.Flags().Bool("json", false, "Output as JSON")
	statesCmd.Flags().String("domain", "", "Filter by domain (e.g. light, switch)")
	statesCmd.Flags().String("attr", "", "Show a specific attribute value instead of state")
	rootCmd.AddCommand(statesCmd)
}

// Package cmd implements the different commands.
package cmd

import (
	"context"
	"os"
	"time"

	"github.com/leonhfr/cete/pkg/game"
	"github.com/spf13/cobra"
)

// options represents the global options
type options struct {
	broadcast bool
	noPGN     bool
	port      int
}

const (
	black     = "black"
	broadcast = "broadcast"
	noPGN     = "no-pgn"
	port      = "port"
	white     = "white"
)

var version = "0.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cete",
	Short: "Pit UCI-compliant chess engines against each other",
	Long: `cete pits UCI-compliant chess engines against each other.

Cete was originally developed to easily test the honey badger chess engine.
A cete is a group of honey badgers.`,
	Args:              cobra.MatchAll(cobra.NoArgs),
	Example:           "  cete -b --white stockfish --black stockfish",
	Version:           version,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGame(
			cmd.Context(),
			getInput(cmd),
			getOptions(cmd),
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().BoolP(broadcast, "b", false, "live broadcast in a web view")
	rootCmd.PersistentFlags().Bool(noPGN, false, "do not print game in PGN format")
	rootCmd.PersistentFlags().IntP(port, "p", 6061, "port used for lived broadcast")

	// Local flags
	rootCmd.Flags().String(white, "stockfish", "path or command to the white engine")
	rootCmd.Flags().String(black, "stockfish", "path or command to the black engine")
	_ = rootCmd.MarkFlagFilename(white)
	_ = rootCmd.MarkFlagFilename(black)
}

// getInput returns a game.Input from the root command local flags
func getInput(cmd *cobra.Command) game.Input {
	white, _ := cmd.Flags().GetString(white)
	black, _ := cmd.Flags().GetString(black)

	return game.Input{
		WhiteEngine: white,
		BlackEngine: black,
		Time:        500 * time.Millisecond,
	}
}

// getOptions returns the options from the root command persistent flags
func getOptions(cmd *cobra.Command) options {
	broadcast, _ := cmd.Flags().GetBool(broadcast)
	noPGN, _ := cmd.Flags().GetBool(noPGN)
	port, _ := cmd.Flags().GetInt(port)

	return options{
		broadcast: broadcast,
		noPGN:     noPGN,
		port:      port,
	}
}

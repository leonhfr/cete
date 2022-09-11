package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/structs"
	"github.com/leonhfr/cete/game"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type (
	yamlPlayer struct {
		Engine  string            `yaml:"engine"`
		Options map[string]string `yaml:"options"`
	}

	yamlInput struct {
		White yamlPlayer `yaml:"white"`
		Black yamlPlayer `yaml:"black"`
		Time  int        `yaml:"time"`
	}
)

// gameCmd represents the game command
var gameCmd = &cobra.Command{
	Use:   "game <yaml file>",
	Short: "play a game using a yaml template file",
	Long: `The game command allows you to play a game
using a yaml template file instead of passing all the
options every call.`,
	Args:    cobra.MatchAll(cobra.ExactArgs(1)),
	Example: "  cete game ./game.yaml -b",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := parseYaml(args[0])
		if err != nil {
			return err
		}
		return runGame(
			cmd.Context(),
			game.Input{
				WhiteEngine:  input.White.Engine,
				BlackEngine:  input.Black.Engine,
				WhiteOptions: input.White.Options,
				BlackOptions: input.Black.Options,
				Time:         time.Duration(input.Time * 10e5),
			},
			getOptions(cmd),
		)
	},
}

func init() {
	rootCmd.AddCommand(gameCmd)
}

func runGame(ctx context.Context, input game.Input, options options) error {
	g, err := game.Run(ctx, input)
	if err != nil {
		return err
	}

	if options.pgn {
		fmt.Printf("PGN:%s\n", g.String())
	}

	return nil
}

func parseYaml(filename string) (*yamlInput, error) {
	var input *yamlInput
	contents, err := os.ReadFile(filename)
	if err != nil {
		return input, err
	}

	err = yaml.Unmarshal(contents, input)
	if err != nil {
		return input, err
	}

	if invalid := structs.HasZero(input); invalid {
		return input, errors.New("yaml file is missing some inputs")
	}

	return input, nil
}

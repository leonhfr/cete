// Package game pits two engines against each other.
package game

import (
	"context"
	"errors"
	"time"

	"github.com/leonhfr/cete/uci"
	"github.com/notnil/chess"
)

// Input is a game play input.
type Input struct {
	WhiteEngine  string
	BlackEngine  string
	WhiteOptions map[string]string
	BlackOptions map[string]string
	Time         time.Duration
}

// Run plays a game.
func Run(ctx context.Context, input Input) (*chess.Game, error) {
	len := nameLength(input.WhiteEngine, input.BlackEngine)

	white, err := new(input.WhiteEngine, len, chess.White, input.WhiteOptions)
	if err != nil {
		return nil, err
	}

	black, err := new(input.BlackEngine, len, chess.Black, input.BlackOptions)
	if err != nil {
		return nil, err
	}

	defer white.Close()
	defer black.Close()

	return runGame(ctx, input, white, black)
}

func runGame(ctx context.Context, input Input, white, black *uci.Engine) (*chess.Game, error) {
	game := chess.NewGame()
	for game.Outcome() == chess.NoOutcome {
		select {
		case <-ctx.Done():
			return game, nil
		default:
		}

		var move *chess.Move
		var err error

		switch game.Position().Turn() {
		case chess.White:
			move, err = search(white, game.Position(), input.Time)
		case chess.Black:
			move, err = search(black, game.Position(), input.Time)
		case chess.NoColor:
			err = errors.New("expected valid color")
		}

		if err != nil {
			return nil, err
		}

		if err := game.Move(move); err != nil {
			return nil, err
		}
	}

	return game, nil
}

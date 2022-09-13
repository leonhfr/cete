// Package game pits two engines against each other.
package game

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/leonhfr/cete/internal/uci"
	"github.com/leonhfr/cete/pkg/engine"
	"github.com/leonhfr/cete/pkg/live"
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
	white, black, err := startEngines(input)
	if err != nil {
		return nil, err
	}

	defer engine.Close(white)
	defer engine.Close(black)

	game := chess.NewGame()
	for game.Outcome() == chess.NoOutcome {
		select {
		case <-ctx.Done():
			return game, nil
		default:
		}

		_, err := playMove(game, input.Time, white, black)
		if err != nil {
			return game, err
		}
	}

	return game, nil
}

// RunWithLive plays a game and broadcast it to a live view.
func RunWithLive(ctx context.Context, input Input, static fs.FS, port int) (game *chess.Game, err error) {
	view, errc, err := live.New(static, port, log.New(os.Stdout, "cete: ", 0))
	if err != nil {
		return nil, err
	}
	defer func() {
		if tErr := view.Shutdown(); tErr != nil {
			err = tErr
		}
	}()

	white, black, err := startEngines(input)
	if err != nil {
		return nil, err
	}
	defer engine.Close(white)
	defer engine.Close(black)

	view.Wait()

	game = chess.NewGame()
	for game.Outcome() == chess.NoOutcome {
		select {
		case <-ctx.Done():
			return game, nil
		case err := <-errc:
			return game, err
		default:
		}

		move, err := playMove(game, input.Time, white, black)
		if err != nil {
			return game, err
		}

		view.Update(move, game.Position())
	}

	return game, err
}

// playMove plays a single move.
func playMove(game *chess.Game, t time.Duration, white, black *uci.Engine) (*chess.Move, error) {
	var move *chess.Move
	var err error

	switch game.Position().Turn() {
	case chess.White:
		move, err = engine.Search(white, game.Position(), t)
	case chess.Black:
		move, err = engine.Search(black, game.Position(), t)
	case chess.NoColor:
		err = errors.New("expected valid color")
	}

	if err != nil {
		return move, err
	}

	if err := game.Move(move); err != nil {
		return move, err
	}

	return move, nil
}

// startEngines starts up both white and black engines
func startEngines(input Input) (*uci.Engine, *uci.Engine, error) {
	len := engine.NameLength(input.WhiteEngine, input.BlackEngine)

	white, err := engine.Start(input.WhiteEngine, len, chess.White, input.WhiteOptions)
	if err != nil {
		return nil, nil, err
	}

	black, err := engine.Start(input.BlackEngine, len, chess.Black, input.BlackOptions)
	if err != nil {
		return nil, nil, err
	}

	return white, black, nil
}

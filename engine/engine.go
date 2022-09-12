// Package engine implements helper functions to easily manage UCI engines.
package engine

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"
	"time"

	ansi "github.com/fatih/color"
	"github.com/leonhfr/cete/uci"
	"github.com/notnil/chess"
)

// Start starts a UCI engine and sets it up to run searches.
func Start(exec string, len int, color chess.Color, options map[string]string) (*uci.Engine, error) {
	name := path.Base(exec)
	ew := newEngineWriter(name, len, color)
	logFn := uci.Logger(log.New(ew, "", 0))

	e, err := uci.New(exec, logFn)
	if err != nil {
		return nil, err
	}
	uci.Debug(e)

	commands := []uci.Cmd{uci.CmdUCI}
	for name, value := range options {
		commands = append(commands, uci.CmdSetOption{Name: name, Value: value})
	}
	commands = append(commands, uci.CmdIsReady, uci.CmdUCINewGame)

	if err = e.Run(commands...); err != nil {
		return e, err
	}

	return e, nil
}

// Search runs a single search.
func Search(e *uci.Engine, p *chess.Position, moveTime time.Duration) (*chess.Move, error) {
	err := e.Run(
		uci.CmdPosition{Position: p},
		uci.CmdGo{MoveTime: moveTime},
	)
	if err != nil {
		return nil, err
	}
	return e.SearchResults().BestMove, nil
}

// Close gracefully shuts down an engine.
func Close(e *uci.Engine) {
	e.Close()
}

// NameLength returns the length of the longest od two base names.
func NameLength(exec1, exec2 string) int {
	name1 := path.Base(exec1)
	name2 := path.Base(exec2)
	return int(math.Max(float64(len(name1)), float64(len(name2))))
}

type engineWriter struct {
	name string
	len  int
	ansi *ansi.Color
}

func newEngineWriter(name string, len int, c chess.Color) engineWriter {
	switch c { //nolint
	case chess.White:
		return engineWriter{name, len, ansi.New(ansi.FgHiBlack, ansi.BgHiWhite)}
	case chess.Black:
		return engineWriter{name, len, ansi.New(ansi.FgHiWhite, ansi.BgHiBlack)}
	default:
		panic("expected valid color")
	}
}

// Write implements the io.Writer interface.
func (ew engineWriter) Write(p []byte) (n int, err error) {
	prefix := fmt.Sprintf("%*s:", ew.len-len(ew.name), ew.name)
	n, err = ew.ansi.Fprint(os.Stdout, prefix)
	if err != nil {
		return
	}

	m, err := fmt.Fprintf(os.Stdout, " %s %s", arrow(string(p)), string(p))
	n += m

	return
}

func arrow(message string) string {
	for _, prefix := range guiToEnginePrefixes {
		if strings.HasPrefix(message, prefix) {
			return "-->"
		}
	}
	return "<--"
}

var guiToEnginePrefixes = []string{
	"uci", "debug", "isready", "setoption", "ucinewgame",
	"position", "go", "stop", "quit",
}

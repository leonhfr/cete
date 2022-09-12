// Package live broadcasts the moves in a web view.
package live

import (
	"fmt"
	"io/fs"

	"github.com/notnil/chess"
)

// View represents a live web view.
type View struct{}

// New creates a new live view.
func New(static fs.FS, port int) (*View, error) {
	return &View{}, nil
}

// Wait awaits that the user confirms the view is live.
func (v *View) Wait() {}

// Update updates the live view with the latest move and position.
func (v *View) Update(move *chess.Move, position *chess.Position) {
	fmt.Printf("LIVE VIEW: Move %s FEN: %s\n", move.String(), position.String())
}

// Shutdown shuts down the live view gracefully.
func (v *View) Shutdown() {}

// Package live broadcasts the moves in a web view.
package live

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/notnil/chess"
)

// View represents a live web view.
type View struct {
	logger   *log.Logger
	serveMux http.ServeMux
	shutdown func() error
}

// New creates a new live view.
func New(static fs.FS, port int, logger *log.Logger) (*View, chan error, error) {
	errc := make(chan error, 1)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errc, err
	}
	logger.Printf("listening on http://%v", l.Addr())

	view := &View{
		logger: logger,
	}

	view.serveMux.Handle("/", http.FileServer(http.FS(static)))

	server := &http.Server{
		Handler:      view,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		defer close(errc)
		errc <- server.Serve(l)
	}()

	view.shutdown = func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

	return view, errc, nil
}

// Wait awaits that the user confirms the view is live.
func (v *View) Wait() {}

// Update updates the live view with the latest move and position.
func (v *View) Update(move *chess.Move, position *chess.Position) {
	v.logger.Printf("move %s FEN: %s\n", move.String(), position.String())
}

// Shutdown shuts down the live view gracefully.
func (v *View) Shutdown() error {
	return v.shutdown()
}

// ServeHTTP implements the http.Handler interface.
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.serveMux.ServeHTTP(w, r)
}

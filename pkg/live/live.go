// Package live broadcasts the moves in a web view.
package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/leonhfr/cete/static"
	"github.com/notnil/chess"
	"nhooyr.io/websocket"
)

// View represents a live web view.
type View struct {
	logger                  *log.Logger
	mu                      sync.Mutex
	port                    int
	position                *chess.Position
	serveMux                http.ServeMux
	shutdown                func() error
	subscribeMessageBuffer  int
	subscribeMessageLimiter time.Duration
	subscribers             map[*subscriber]struct{}
	wait                    chan struct{}
}

type subscriber struct {
	msgs chan []byte
	kick func()
}

// New creates a new live view.
func New(port int, logger *log.Logger) (*View, chan error, error) {
	errc := make(chan error, 1)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, errc, err
	}

	view := &View{
		logger:                  logger,
		port:                    port,
		position:                chess.StartingPosition(),
		subscribeMessageBuffer:  16,
		subscribeMessageLimiter: 200 * time.Millisecond,
		subscribers:             make(map[*subscriber]struct{}),
		wait:                    make(chan struct{}),
	}

	view.serveMux.Handle("/", http.FileServer(http.FS(static.FileSystem)))
	view.serveMux.HandleFunc("/start", view.startHandler)
	view.serveMux.HandleFunc("/subscribe", view.subscribeHandler)

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
func (v *View) Wait(ctx context.Context) {
	v.logger.Printf("live view on http://localhost:%d\n", v.port)
	v.logger.Printf("press start to continue\n")
	select {
	case <-v.wait:
	case <-ctx.Done():
	}
}

// Update updates the live view with the latest move and position.
func (v *View) Update(move *chess.Move, position *chess.Position) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.position = position
	msg, err := json.Marshal(&message{Move: move, Position: position})
	if err != nil {
		return err
	}

	for s := range v.subscribers {
		select {
		case s.msgs <- msg:
		default:
			go s.kick()
		}
	}

	return nil
}

// Shutdown shuts down the live view gracefully.
func (v *View) Shutdown() error {
	return v.shutdown()
}

// ServeHTTP implements the http.Handler interface.
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.serveMux.ServeHTTP(w, r)
}

// startHandler receives the start request that confirms the view is live
// and unblocks the wait function
func (v *View) startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer w.WriteHeader(http.StatusAccepted)

	select {
	case <-v.wait:
		return
	default:
		close(v.wait)
	}
}

// subscribeHandler accepts the WebSocket connection
// sends the current game state and subscribes it to all future messages
func (v *View) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		v.logger.Printf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	err = v.subscribe(r.Context(), c)
	if errors.Is(err, context.Canceled) {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		v.logger.Printf("%v", err)
		return
	}
}

// subscribe subscribes the given WebSocket to all broadcasted messages
func (v *View) subscribe(ctx context.Context, c *websocket.Conn) error {
	ctx = c.CloseRead(ctx)

	s := &subscriber{
		msgs: make(chan []byte, v.subscribeMessageBuffer),
		kick: func() {
			c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}
	v.addSubscriber(s)
	defer v.deleteSubscriber(s)

	// first message sets the position
	msg, err := json.Marshal(&message{Position: v.position})
	if err != nil {
		return err
	}
	s.msgs <- msg

	limiter := time.NewTicker(v.subscribeMessageLimiter)
	for {
		select {
		case <-limiter.C:
			select {
			case msg := <-s.msgs:
				err := writeTimeout(ctx, time.Second, c, msg)
				if err != nil {
					return err
				}
			default:
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// addSubscriber adds a subscriber
func (v *View) addSubscriber(s *subscriber) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.subscribers[s] = struct{}{}
}

// deleteSubscriber deletes a subscriber
func (v *View) deleteSubscriber(s *subscriber) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.subscribers, s)
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

// message is broadcast between cete and the live view.
type message struct {
	Move     *chess.Move
	Position *chess.Position
}

// MarshalJSON implements the encoding/json.Marshaler interface.
//
// If no move is passed, then the position is encoded.
// This is used to set the position in the live view.
// If a move is passed, the position is only encoded in case of a promotion.
func (m *message) MarshalJSON() ([]byte, error) {
	kingCastles := map[chess.Rank]string{
		chess.Rank1: "h1-f1",
		chess.Rank8: "h8-f8",
	}
	queenCastles := map[chess.Rank]string{
		chess.Rank1: "a1-d1",
		chess.Rank8: "a8-d8",
	}

	var move, castlingMove, position string

	if m.Move != nil {
		move = fmt.Sprintf("%v-%v", m.Move.S1().String(), m.Move.S2().String())

		if m.Move.HasTag(chess.KingSideCastle) {
			castlingMove = kingCastles[m.Move.S1().Rank()]
		}

		if m.Move.HasTag(chess.QueenSideCastle) {
			castlingMove = queenCastles[m.Move.S1().Rank()]
		}
	}

	if (m.Move == nil || m.Move.Promo() != chess.NoPieceType) && m.Position != nil {
		position = m.Position.String()
	}

	return json.Marshal(&struct {
		Move         string `json:"move,omitempty"`
		CastlingMove string `json:"castlingMove,omitempty"`
		Position     string `json:"position,omitempty"`
	}{
		Move:         move,
		CastlingMove: castlingMove,
		Position:     position,
	})
}

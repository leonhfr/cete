package live

import (
	"encoding/json"
	"testing"

	"github.com/notnil/chess"
	"github.com/stretchr/testify/assert"
)

func TestMessageMarshalJSON(t *testing.T) {
	fen0 := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	fen1 := "r4r2/1b2bppk/ppq1p3/2pp3n/5P2/1P2P3/PBPPQ1PP/R4RK1 w - - 0 2"
	fen2 := "8/8/8/8/8/5K2/4p2R/5k2 b - - 0 1"
	fen3 := "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"
	fen4 := "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1"

	tests := []struct {
		name string
		args message
		want string
	}{
		{
			"single move",
			message{move("b1a3", fen0), chess.StartingPosition()},
			`{"move":"b1-a3"}`,
		},
		{
			"set position",
			message{Position: position(fen1)},
			`{"position":"` + fen1 + `"}`,
		},
		{
			"promotion",
			message{Move: move("e2e1n", fen2), Position: position(fen2)},
			`{"move":"e2-e1","position":"` + fen2 + `"}`,
		},
		{
			"white king side castle",
			message{Move: move("e1g1", fen3), Position: position(fen3)},
			`{"move":"e1-g1","castlingMove":"h1-f1"}`,
		},
		{
			"black king side castle",
			message{Move: move("e8g8", fen4), Position: position(fen4)},
			`{"move":"e8-g8","castlingMove":"h8-f8"}`,
		},
		{
			"white queen side castle",
			message{Move: move("e1c1", fen3), Position: position(fen3)},
			`{"move":"e1-c1","castlingMove":"a1-d1"}`,
		},
		{
			"black queen side castle",
			message{Move: move("e8c8", fen4), Position: position(fen4)},
			`{"move":"e8-c8","castlingMove":"a8-d8"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := json.Marshal(&tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(msg))
		})
	}
}

func position(fen string) *chess.Position {
	fn, _ := chess.FEN(fen)
	game := chess.NewGame(fn)
	return game.Position()
}

func move(uci, fen string) *chess.Move {
	p := position(fen)
	m, _ := chess.UCINotation{}.Decode(p, uci)
	return m
}

package uchess

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corentings/chess/v2"
)

// newTestGameState returns a minimal GameState suitable for testing ProcessCmd.
func newTestGameState() *GameState {
	return &GameState{
		Game: chess.NewGame(),
	}
}

func TestProcessCmdNormalizesCastlingNotation(t *testing.T) {
	gs := newTestGameState()

	msg, game := ProcessCmd("0-0", gs)
	// We don't assert legality (the underlying library may reject castling
	// from the default position for various reasons), but we *do* assert that
	// the algebraic notation in the error was normalized to O-O.
	if !strings.Contains(msg, "O-O") {
		t.Fatalf("expected normalized castling notation in message, got %q", msg)
	}

	// Ensure the game pointer is still non-nil and unchanged in error case.
	if game == nil {
		t.Fatalf("expected non-nil game after ProcessCmd")
	}
}

func TestProcessCmdReturnsDetailedIllegalMessage(t *testing.T) {
	gs := newTestGameState()

	msg, _ := ProcessCmd("e9", gs) // clearly invalid algebraic notation

	if !strings.HasPrefix(msg, "\u26A0 Illegal: ") {
		t.Fatalf("expected detailed illegal move message, got %q", msg)
	}
}

func TestProcessCmdRecognizesOnlineCommand(t *testing.T) {
	gs := newTestGameState()

	msg, game := ProcessCmd("online", gs)

	if msg != "online" {
		t.Fatalf("ProcessCmd(online) message = %q, want %q", msg, "online")
	}
	if game != gs.Game {
		t.Fatal("ProcessCmd(online) should preserve the current game")
	}
}

func TestProcessMoveAppliesAlgebraicMove(t *testing.T) {
	gs := newTestGameState()

	msg, game := ProcessMove("e4", gs)

	if strings.TrimSpace(msg) != "" {
		t.Fatalf("expected empty/cleared message after a legal move, got %q", msg)
	}
	if len(game.Moves()) != 1 {
		t.Fatalf("expected one move after ProcessMove, got %d", len(game.Moves()))
	}
}

func TestProcessCmdBlocksOnlineMovesUntilLocalTurn(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*GameState)
		want    string
	}{
		{
			name: "while connecting",
			prepare: func(gs *GameState) {
				gs.Online.Enabled = true
			},
			want: "Waiting for opponent to connect",
		},
		{
			name: "opponent turn",
			prepare: func(gs *GameState) {
				gs.Online.Enabled = true
				gs.Online.Connected = true
				gs.Online.LocalSide = chess.Black
			},
			want: "Waiting for opponent's move",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := newTestGameState()
			tt.prepare(gs)
			msg, game := ProcessCmd("e4", gs)
			if msg != tt.want {
				t.Fatalf("ProcessCmd(e4) message = %q, want %q", msg, tt.want)
			}
			if game != gs.Game || len(game.Moves()) != 0 {
				t.Fatal("move should not be applied while waiting")
			}
		})
	}
}

// TestProcessCmdCastlingUserReport reproduces the position reported in
// GitHub issue #3 where the user attempted to castle with "0-0".
// This ensures we handle the input safely and surface a clear message.
func TestProcessCmdCastlingUserReport(t *testing.T) {
	gs := newTestGameState()

	// Reproduce the reported move sequence:
	// 1.d4 d5 2.Bf4 c5 3.c3 Nf6 4.Nf3 Nc6 5.e3 Nh5 6.g3 Qb6
	// 7.Qc2 h6 8.Bg2 g5 9.Be5 Ng7 10.Bxg7 Bxg7 *
	moves := []string{
		"d4", "d5",
		"Bf4", "c5",
		"c3", "Nf6",
		"Nf3", "Nc6",
		"e3", "Nh5",
		"g3", "Qb6",
		"Qc2", "h6",
		"Bg2", "g5",
		"Be5", "Ng7",
		"Bxg7", "Bxg7",
	}

	for _, mv := range moves {
		pos := gs.Game.Position()
		move, err := chess.AlgebraicNotation{}.Decode(pos, mv)
		if err != nil {
			t.Fatalf("failed to decode move %q: %v", mv, err)
		}
		if err := gs.Game.Move(move, nil); err != nil {
			t.Fatalf("failed to apply move %q: %v", mv, err)
		}
	}

	// Ask the underlying chess library whether a kingside castle for white
	// is actually legal in this position.
	hasWhiteKingSideCastle := false
	for _, m := range gs.Game.ValidMoves() {
		if m.HasTag(chess.KingSideCastle) && m.S1().Rank() == chess.Rank1 {
			hasWhiteKingSideCastle = true
			break
		}
	}

	if !hasWhiteKingSideCastle {
		t.Fatalf("expected white kingside castling to be a legal move in this position")
	}

	msg, game := ProcessCmd("0-0", gs)

	// On success, ProcessCmd should clear the message label (spaces only).
	if strings.TrimSpace(msg) != "" {
		t.Fatalf("expected empty/cleared message after successful castling, got %q", msg)
	}

	// Ensure the last move played is a white kingside castle.
	gameMoves := game.Moves()
	if len(gameMoves) == 0 {
		t.Fatalf("expected at least one move after castling")
	}
	last := gameMoves[len(gameMoves)-1]
	if !last.HasTag(chess.KingSideCastle) {
		t.Fatalf("expected last move to be kingside castle, got %v", last)
	}
}

// TestSaveGameWritesPgnFile verifies that SaveGame writes the game as a .pgn
// file (so it opens in third-party analysis tools) containing the PGN moves.
func TestSaveGameWritesPgnFile(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(wd)

	game := chess.NewGame()
	for _, mv := range []string{"e4", "e5"} {
		move, err := chess.AlgebraicNotation{}.Decode(game.Position(), mv)
		if err != nil {
			t.Fatalf("decode %q: %v", mv, err)
		}
		if err := game.Move(move, nil); err != nil {
			t.Fatalf("apply %q: %v", mv, err)
		}
	}

	msg := SaveGame(game)
	if !strings.HasPrefix(msg, "Saved ") {
		t.Fatalf("expected a save confirmation, got %q", msg)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var pgnFile string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".pgn") {
			pgnFile = e.Name()
		}
	}
	if pgnFile == "" {
		t.Fatalf("expected a .pgn file to be created, got entries %v", entries)
	}

	data, err := os.ReadFile(filepath.Join(dir, pgnFile))
	if err != nil {
		t.Fatalf("read pgn: %v", err)
	}
	if !strings.Contains(string(data), "e4") {
		t.Fatalf("expected PGN move text in saved file, got %q", string(data))
	}
}

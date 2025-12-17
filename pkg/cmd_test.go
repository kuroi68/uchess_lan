package uchess

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
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
		if err := gs.Game.MoveStr(mv); err != nil {
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




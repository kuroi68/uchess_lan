package uchess

import (
	"testing"

	"github.com/corentings/chess/v2"
)

func TestIdxToRankAndFileAndSquare(t *testing.T) {
	if got := idxToRank(0); got != "1" {
		t.Fatalf("idxToRank(0) = %q, want %q", got, "1")
	}
	if got := idxToRank(7); got != "8" {
		t.Fatalf("idxToRank(7) = %q, want %q", got, "8")
	}
	if got := idxToFile(0); got != "a" {
		t.Fatalf("idxToFile(0) = %q, want %q", got, "a")
	}
	if got := idxToFile(7); got != "h" {
		t.Fatalf("idxToFile(7) = %q, want %q", got, "h")
	}

	if got := idxToSquare(0, 0); got != "a1" {
		t.Fatalf("idxToSquare(0,0) = %q, want %q", got, "a1")
	}
	if got := idxToSquare(7, 7); got != "h8" {
		t.Fatalf("idxToSquare(7,7) = %q, want %q", got, "h8")
	}
}

func TestMoveIdxPaginatesLastFivePairs(t *testing.T) {
	game := chess.NewGame()

	// Create 6 full move pairs (12 ply) so that pagination logic kicks in.
	sequence := []string{
		"e4", "e5",
		"Nf3", "Nc6",
		"Bb5", "a6",
		"Ba4", "Nf6",
		"O-O", "Be7",
		"Re1", "b5",
	}

	for _, mv := range sequence {
		pos := game.Position()
		move, err := chess.AlgebraicNotation{}.Decode(pos, mv)
		if err != nil {
			t.Fatalf("failed to decode move %q: %v", mv, err)
		}
		if err := game.Move(move, nil); err != nil {
			t.Fatalf("failed to apply move %q: %v", mv, err)
		}
	}

	// We expect moveIdx to show only the last 5 move pairs.
	// idx 0 should correspond to "2." (second pair), since the first pair is paged off.
	idx, _, _ := moveIdx(game, 0)
	if idx != "2." {
		t.Fatalf("moveIdx(game,0) index = %q, want %q", idx, "2.")
	}

	// The last visible index (idx 4) should correspond to "6."
	idx, _, _ = moveIdx(game, 4)
	if idx != "6." {
		t.Fatalf("moveIdx(game,4) index = %q, want %q", idx, "6.")
	}
}

func TestLastMoveUsesFromAndToSquares(t *testing.T) {
	game := chess.NewGame()

	pos := game.Position()
	move, err := chess.AlgebraicNotation{}.Decode(pos, "e4")
	if err != nil {
		t.Fatalf("failed to decode move: %v", err)
	}
	if err := game.Move(move, nil); err != nil {
		t.Fatalf("failed to apply move e4: %v", err)
	}

	// Last move e2 -> e4 should mark both e2 and e4 as part of the last move.
	if !lastMove(game, "e4") {
		t.Fatalf("expected lastMove(game, \"e4\") to be true")
	}
	if !lastMove(game, "e2") {
		t.Fatalf("expected lastMove(game, \"e2\") to be true")
	}
	if lastMove(game, "a1") {
		t.Fatalf("expected lastMove(game, \"a1\") to be false")
	}
}

func TestHintSqMatchesMoveEndpoints(t *testing.T) {
	game := chess.NewGame()
	// Take an arbitrary valid move from the starting position.
	validMoves := game.ValidMoves()
	if len(validMoves) == 0 {
		t.Fatalf("expected at least one valid move from starting position")
	}
	mv := &validMoves[0]

	sqFrom := mv.S1().String()
	sqTo := mv.S2().String()

	if !hintSq(mv, sqFrom) {
		t.Fatalf("expected hintSq to be true for from square %q", sqFrom)
	}
	if !hintSq(mv, sqTo) {
		t.Fatalf("expected hintSq to be true for to square %q", sqTo)
	}
	if hintSq(mv, "a1") && sqFrom != "a1" && sqTo != "a1" {
		t.Fatalf("expected hintSq to be false for unrelated square \"a1\"")
	}
}



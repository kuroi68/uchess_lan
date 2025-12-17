package uchess

import (
	"math"
	"testing"

	"github.com/notnil/chess"
)

func TestIsInteractive(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "human_vs_cpu",
			cfg: Config{
				WhitePiece: "human",
				BlackPiece: "cpu",
			},
			want: true,
		},
		{
			name: "cpu_vs_human",
			cfg: Config{
				WhitePiece: "cpu",
				BlackPiece: "human",
			},
			want: true,
		},
		{
			name: "cpu_vs_cpu",
			cfg: Config{
				WhitePiece: "cpu",
				BlackPiece: "cpu",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInteractive(tt.cfg)
			if got != tt.want {
				t.Fatalf("IsInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCPU(t *testing.T) {
	cfg := Config{
		WhitePiece: "cpu",
		BlackPiece: "human",
	}

	if !IsCPU(chess.White, cfg) {
		t.Fatalf("IsCPU(White) = false, want true")
	}

	if IsCPU(chess.Black, cfg) {
		t.Fatalf("IsCPU(Black) = true, want false")
	}
}

func TestWinProbSymmetryAndBounds(t *testing.T) {
	if got := WinProb(0); got != 0.5 {
		t.Fatalf("WinProb(0) = %v, want 0.5", got)
	}

	whiteFav := WinProb(400)  // white up about a piece
	blackFav := WinProb(-400) // black up about a piece

	if !(whiteFav > 0.5 && whiteFav < 1.0) {
		t.Fatalf("WinProb(400) = %v, want between 0.5 and 1.0", whiteFav)
	}
	if !(blackFav < 0.5 && blackFav > 0.0) {
		t.Fatalf("WinProb(-400) = %v, want between 0.0 and 0.5", blackFav)
	}

	// Check approximate symmetry around 0.5.
	if diff := math.Abs(whiteFav-(1-blackFav)); diff > 1e-6 {
		t.Fatalf("WinProb symmetry broken: WinProb(400) = %v, 1-WinProb(-400) = %v (diff=%v)", whiteFav, 1-blackFav, diff)
	}
}

func TestRoundNearest(t *testing.T) {
	if got := RoundNearest(47, 5); got != 45 {
		t.Fatalf("RoundNearest(47,5) = %v, want 45", got)
	}
	if got := RoundNearest(52, 5); got != 50 {
		t.Fatalf("RoundNearest(52,5) = %v, want 50", got)
	}
}

func TestScoreStrInitialPositionIsEqual(t *testing.T) {
	positions := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	white, black := ScoreStr(positions)
	if white != "" || black != "" {
		t.Fatalf("ScoreStr(initial) = (%q,%q), want both empty", white, black)
	}
}

func TestScoreStrWhiteUpPawn(t *testing.T) {
	// Position where black is missing one pawn (white has captured a pawn).
	positions := "rnbqkbnr/ppppppp1/8/8/8/8/PPPPPPPP/RNBQKBNR"
	white, black := ScoreStr(positions)

	if white != "+1" || black != "" {
		t.Fatalf("ScoreStr(white up pawn) = (%q,%q), want (+1, \"\")", white, black)
	}
}



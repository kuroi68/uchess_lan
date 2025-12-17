package uchess

import (
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



package uchess

import (
	"github.com/corentings/chess/v2"
	"github.com/gdamore/tcell/v2"
)

// GameState encapsulates everything needed to run the game
type GameState struct {
	S          tcell.Screen // Screen
	Input      *Input       // Input
	Game       *chess.Game  // Chess Board
	UCI        UCIState     // UCI State
	Config     Config       // Global Config
	Theme      Theme        // Theme
	Score      int          // Score in centipawns
	CheckWhite bool         // White is in check
	CheckBlack bool         // Black is in check
	Hint       *chess.Move  // Hint when available
}

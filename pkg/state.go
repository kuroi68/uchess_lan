package uchess

import (
	onlinepkg "github.com/tmountain/uchess/pkg/online"

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
	Online     OnlineState  // Online
}

// OnlineState encapsulates online game settings
// TODO: переписать пояснения к полям
type OnlineState struct {
	Enabled   bool               // Online is active
	Connected bool               // User is connected
	IsHost    bool               // Local pc is the host
	PeerAddr  string             // Address of opponent
	LocalSide chess.Color        // Local color of the chess pieces
	Session   *onlinepkg.Session // Session connection
}

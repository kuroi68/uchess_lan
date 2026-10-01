package uchess

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/corentings/chess/v2/image"
	"github.com/corentings/chess/v2/uci"
)

// selectEngine returns the UCI engine and its corresponding config based upon the current turn
func selectEngine(game *chess.Game, us UCIState) (*uci.Engine, *UCIEngine) {
	var eng *uci.Engine
	var cfg *UCIEngine

	if game.Position().Turn() == chess.White {
		eng = us.UciWhite
		cfg = us.CfgWhite
	} else {
		eng = us.UciBlack
		cfg = us.CfgBlack
	}
	return eng, cfg
}

// EngScore provides the current board score in centipawns for whomever the current
// game position identifies as active
func EngScore(game *chess.Game, us UCIState, config Config) int {
	eng, _ := selectEngine(game, us)
	// Update the engine on the game position
	cmdPos := uci.CmdPosition{Position: game.Position()}
	// Do a quick analysis of the current board
	cmdGo := uci.CmdGo{Depth: 10}
	// If the command fails for some reason, exit
	if err := eng.Run(cmdPos, cmdGo); err != nil {
		panic(err)
	}
	info := eng.SearchResults().Info
	// Return the score in centipawns
	return info.Score.CP
}

// EngMove executes a move using the UCI engine
func EngMove(game *chess.Game, us UCIState, config Config) string {
	eng, engCfg := selectEngine(game, us)

	// Update the engine on the game position
	cmdPos := uci.CmdPosition{Position: game.Position()}
	// Params to send to the engine
	cmdGo := &uci.CmdGo{Depth: engCfg.Depth}
	cmdGo.MoveTime = engCfg.MoveTime * time.Millisecond
	// If SearchMoves is specified, include it
	if engCfg.SearchMoves != "" {
		cmdGo.SearchMoves = searchMoves(engCfg.SearchMoves)
	}
	// Run the actual commands
	if err := eng.Run(cmdPos, *cmdGo); err != nil {
		return "\u26A0 Error. Engine command."
	}
	// Fetch the results
	move := eng.SearchResults().BestMove
	// Validate the move
	if err := game.Move(move, nil); err != nil {
		return "\u26A0 Error. Engine move."
	}
	// If this is a CPU vs CPU match and the game has just finished,
	// automatically save the PGN so it can be analyzed later.
	if !IsInteractive(config) && game.Outcome() != "*" {
		return SaveGame(game)
	}
	// Clear the label for normal interactive play
	return strings.Repeat(" ", 32)
}

func undoMove(game *chess.Game) *chess.Game {
	newGame := chess.NewGame()
	moves := game.Moves()
	for i := 0; i < len(moves)-2; i++ {
		move := moves[i]
		newGame.Move(move, nil)
	}
	return newGame
}

func resetGame(game *chess.Game) *chess.Game {
	newGame := chess.NewGame()
	return newGame
}

// SaveGame writes the game's PGN to a timestamped .pgn file so it can be
// opened directly in third-party analysis tools. Returns a status label.
func SaveGame(game *chess.Game) string {
	ts := Timestamp()
	file := fmt.Sprintf("uchess_%v.pgn", ts)
	f, err := os.Create(file)
	if err != nil {
		return err.Error()
	}
	defer f.Close()

	f.WriteString(game.String())
	return fmt.Sprintf("Saved %v", file)
}

func saveImage(game *chess.Game) string {
	ts := Timestamp()
	file := fmt.Sprintf("uchess_%v.svg", ts)
	f, err := os.Create(file)

	if err != nil {
		return err.Error()
	}

	defer f.Close()

	board := game.Position().Board()
	if err := image.SVG(f, board); err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Saved %v", file)
}

func resign(game *chess.Game) *chess.Game {
	game.Resign(game.Position().Turn())
	return game
}

func hint(gs *GameState) string {
	DrawMsgLabel(gs.S, "Thinking...", gs.Theme)
	Render(gs)
	eng := gs.UCI.UciHint
	engCfg := gs.UCI.CfgHint
	// Update the engine on the game position
	cmdPos := uci.CmdPosition{Position: gs.Game.Position()}
	// Params to send to the engine
	cmdGo := &uci.CmdGo{Depth: engCfg.Depth}
	cmdGo.MoveTime = engCfg.MoveTime * time.Millisecond
	// If SearchMoves is specified, include it
	if engCfg.SearchMoves != "" {
		cmdGo.SearchMoves = searchMoves(engCfg.SearchMoves)
	}
	// Run the actual commands
	if err := eng.Run(cmdPos, *cmdGo); err != nil {
		return "\u26A0 Error. Engine command."
	}
	// Take the best move
	bestMove := eng.SearchResults().BestMove
	// Success, set the move in the game state
	gs.Hint = bestMove
	return strings.Repeat(" ", 80)
}

func quit(gs *GameState) string {
	return "quit"
}

func online(gs *GameState) string {
	return "online"
}

// ProcessCmd processes a move request or command
func ProcessCmd(cmd string, gs *GameState) (string, *chess.Game) {
	cmd = strings.TrimSpace(cmd)

	switch cmd {
	// Back one turn
	case "back":
		return strings.Repeat(" ", 80), undoMove(gs.Game)
		// Save the PGN string
	case "save":
		return SaveGame(gs.Game), gs.Game
		// SVG snapshot of the current board
	case "image":
		return saveImage(gs.Game), gs.Game
		// Output the FEN string
	case "fen":
		return gs.Game.Position().String(), gs.Game
		// Reset the game
	case "reset":
		return strings.Repeat(" ", 80), resetGame(gs.Game)
		// Current player resigns
	case "resign":
		return strings.Repeat(" ", 80), resign(gs.Game)
		// Process a move string
	case "hint":
		if !gs.Online.Enabled {
			return hint(gs), gs.Game
		}
		return "\u26A0 Illegal: Hints are not available online", gs.Game
	case "quit": // TODO: грамотно выходить из игры в онлайне
		return quit(gs), gs.Game
		// Online game
	case "online":
		return online(gs), gs.Game
	default:
		if gs.Online.Enabled && !gs.Online.Connected {
			return "Waiting for opponent to connect", gs.Game
		}
		if gs.Online.Connected && gs.Game.Position().Turn() != gs.Online.LocalSide {
			return "Waiting for opponent's move", gs.Game
		}
		return ProcessMove(cmd, gs)
	}
}

// ProcessMove applies a single algebraic chess move to the current game.
func ProcessMove(cmd string, gs *GameState) (string, *chess.Game) {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.ReplaceAll(cmd, "0", "O")
	cmd = strings.ReplaceAll(cmd, "o", "O")

	pos := gs.Game.Position()
	move, err := chess.AlgebraicNotation{}.Decode(pos, cmd)
	if err != nil {
		return "\u26A0 Illegal: " + err.Error(), gs.Game
	}
	if err := gs.Game.Move(move, nil); err != nil {
		return "\u26A0 Illegal: " + err.Error(), gs.Game
	}
	fmt.Println("Ход совершен")
	return strings.Repeat(" ", 80), gs.Game
}

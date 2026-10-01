package uchess

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/corentings/chess/v2"
	"github.com/gdamore/tcell/v2"
)

const (
	leftMargin          = 4
	topMargin           = 4
	numofSquaresInBoard = 64
	numOfSquaresInRow   = 8
)

// drawText places text at the specified coordinates with the provided style
func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, r := range []rune(text) {
		s.SetContent(x, y, r, nil, style)
		x++
	}
}

func clearTextLine(s tcell.Screen, x, y int, style tcell.Style) {
	width, _ := s.Size()
	for col := x; col < width; col++ {
		s.SetContent(col, y, ' ', nil, style)
	}
}

func drawTextLine(s tcell.Screen, x, y int, style tcell.Style, text string) {
	width, _ := s.Size()
	if x >= width {
		return
	}
	clearTextLine(s, x, y, style)
	maxRunes := width - x
	runes := []rune(text)
	if len(runes) > maxRunes {
		if maxRunes == 1 {
			text = string(runes[:1])
		} else {
			text = string(runes[:maxRunes-1]) + "…"
		}
	}
	drawText(s, x, y, style, text)
}

// drawRune places a rune at the specified coordinates with the provided style
func drawRune(s tcell.Screen, x, y int, style tcell.Style, r rune) {
	s.SetContent(x, y, r, nil, style)
}

// DefStyle is the default style for tcell rendering
var DefStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

// stylePiece applies the theme's style to a piece based upon its color
func stylePiece(p chess.Piece, sqBg tcell.Color, t Theme) tcell.Style {
	pieceStyle := tcell.StyleDefault.Background(sqBg)

	if p.Color() == chess.White {
		return pieceStyle.Foreground(t.White)
	}
	return pieceStyle.Foreground(t.Black)

}

// squareBg returns the theme's color corresponding to the square
func squareBg(sq chess.Square, t Theme) tcell.Color {
	// The color is used to draw the square
	squareColor := squareColor(sq)
	if squareColor == chess.Black {
		return t.SquareDark
	}
	return t.SquareLight
}

// drawSquare draws a board square and its corresponding piece
func drawSquare(s tcell.Screen, col, row int, p chess.Piece, sqBg tcell.Color, t Theme) {
	// Empty square
	if p == chess.NoPiece {
		// Fill two columns wide to make it square
		s.SetContent(col, row, ' ', nil, tcell.StyleDefault.Background(sqBg))
		s.SetContent(col+1, row, ' ', nil, tcell.StyleDefault.Background(sqBg))
		// Square contains a piece
	} else {
		piece, _ := utf8.DecodeRuneInString(p.String())
		pieceStyle := stylePiece(p, sqBg, t)
		// Fill with the piece and then pad the rest with blank
		s.SetContent(col, row, piece, nil, pieceStyle)
		s.SetContent(col+1, row, ' ', nil, tcell.StyleDefault.Background(sqBg))
	}
}

// drawRank draws the rank (row indicator)
func drawRank(s tcell.Screen, col, row int, r chess.Rank, t Theme) {
	// drawRune wants a rune
	rank, _ := utf8.DecodeRuneInString(r.String())
	// Display the rank (row)
	rankStyle := tcell.StyleDefault.Foreground(t.Rank)
	drawRune(s, col, row, rankStyle, rank)
}

// drawMoveLabel displays the current move above the board
func drawMoveLabel(s tcell.Screen, game *chess.Game, t Theme) {
	var nextPlayer string
	playerTurn := game.Position().Turn()

	if playerTurn == chess.Black {
		nextPlayer = " Black to Move "
	} else {
		nextPlayer = " White to Move "
	}
	labelStyle := tcell.StyleDefault.Background(t.MoveLabelBg).Foreground(t.MoveLabelFg)
	drawTextLine(s, leftMargin+2, topMargin-2, labelStyle, nextPlayer)
}

// DrawMsgLabel displays the current message from the command
func DrawMsgLabel(s tcell.Screen, msg string, t Theme) {
	topMargin := topMargin + 10
	labelStyle := tcell.StyleDefault.Foreground(t.Msg)
	drawTextLine(s, leftMargin, topMargin, labelStyle, msg)
}

// drawPlayers displays the names of the players and their scores
func drawPlayers(s tcell.Screen, config Config, game *chess.Game, t Theme) {
	leftMargin := leftMargin + 22
	nameStyle := tcell.StyleDefault.Foreground(t.Emoji)
	drawPlayerName(s, leftMargin, topMargin-2, nameStyle, "Black", config.BlackName)
	drawPlayerName(s, leftMargin, topMargin+8, nameStyle, "White", config.WhiteName)
	fen := game.Position().String()
	pos := strings.Split(fen, " ")
	whiteAdv, blackAdv := Advantages(pos[0])
	whiteScore, blackScore := ScoreStr(pos[0])
	blackRes := fmt.Sprintf("%v %-10v", blackAdv, blackScore)
	advStyle := tcell.StyleDefault.Foreground(t.Advantage)
	drawTextLine(s, leftMargin, topMargin-1, advStyle, blackRes)
	whiteRes := fmt.Sprintf("%v %-10v", whiteAdv, whiteScore)
	drawTextLine(s, leftMargin, topMargin+7, advStyle, whiteRes)
}

func drawPlayerName(s tcell.Screen, x, y int, style tcell.Style, color, name string) {
	label := playerLabel(color, name, screenWidthFrom(s, x))
	if label != "" {
		drawText(s, x, y, style, label)
	}
}

func screenWidthFrom(s tcell.Screen, x int) int {
	width, _ := s.Size()
	return width - x
}

func playerLabel(color, name string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" {
		name = "Player"
	}
	label := fmt.Sprintf("%s: %s", color, name)
	runes := []rune(label)
	if len(runes) > maxRunes {
		if maxRunes == 1 {
			label = string(runes[:1])
		} else {
			label = string(runes[:maxRunes-1]) + "…"
		}
	}
	labelRunes := []rune(label)
	if len(labelRunes) < maxRunes {
		label += strings.Repeat(" ", maxRunes-len(labelRunes))
	}
	return label
}

// drawScore displays the current game score
func drawScore(s tcell.Screen, cp int, game *chess.Game, t Theme) {
	topMargin := topMargin + 13
	leftMargin := leftMargin
	prob := WinProb(cp) * 100
	scoreStyle := tcell.StyleDefault.Foreground(t.Score)
	score := fmt.Sprintf("cp=%v, pct=%-10.2f", cp, prob)
	status := ""

	// Outcome "*" means the game is in progress
	if game.Outcome() == "*" {
		status = strings.Repeat(" ", 80)
		// Otherwise, the game has ended
	} else {
		status = fmt.Sprintf("%v (%v)", game.Outcome(), game.Method())
	}
	drawTextLine(s, leftMargin, topMargin, scoreStyle, score)
	drawTextLine(s, leftMargin, topMargin+1, scoreStyle, status)
}

// Render draws the screen
func Render(gs *GameState) {
	drawMoveLabel(gs.S, gs.Game, gs.Theme)
	drawBoard(gs.S, gs.Game, gs.Theme, gs.CheckWhite, gs.CheckBlack, gs.Hint)
	drawPrompt(gs.S, gs.Input, gs.Theme)
	drawScore(gs.S, gs.Score, gs.Game, gs.Theme)
	drawScoreMeter(gs.S, gs.Score, gs.Theme)
	drawPlayers(gs.S, gs.Config, gs.Game, gs.Theme)
	drawMoves(gs.S, gs.Game, gs.Theme)
	// Update screen
	gs.S.Show()
}

// drawScoreCell draws a cell of the score meter
func drawScoreCell(s tcell.Screen, cp, idx int, t Theme) {
	block := '█'
	// At 8 start at top margin moving down as idx decreases
	ypos := topMargin - idx + 8
	// Round this by 5 because the meter is low resolution and we
	// don't want values like 49.25 showing lower than 50%
	winProb := RoundNearest(WinProb(cp)*100, 5.0)
	baseColor := t.MeterBase
	neutralColor := t.MeterNeutral
	winColor := t.MeterWin
	loseColor := t.MeterLose
	blockStyle := DefStyle

	if winProb == 50 {
		blockStyle = tcell.StyleDefault.Foreground(neutralColor)
	} else if winProb < 50 {
		blockStyle = tcell.StyleDefault.Foreground(loseColor)
	} else if winProb > 50 {
		blockStyle = tcell.StyleDefault.Foreground(winColor)
	}

	midStyle := tcell.StyleDefault.Foreground(t.MeterMid)
	drawRune(s, leftMargin+19, topMargin+3, midStyle, '_')
	if AtScale(idx, 8, winProb) {
		drawRune(s, leftMargin+20, ypos, blockStyle, block)
	} else {
		baseStyle := tcell.StyleDefault.Foreground(baseColor)
		drawRune(s, leftMargin+20, ypos, baseStyle, block)
	}
}

// drawScoreMeter displays a graphical representation of the score
func drawScoreMeter(s tcell.Screen, cp int, t Theme) {
	for i := 8; i > 0; i-- {
		drawScoreCell(s, cp, i, t)
	}
}

// drawPrompt draws the prompt
func drawPrompt(s tcell.Screen, i *Input, t Theme) {
	topMargin := topMargin + 11
	clearTextLine(s, leftMargin, topMargin, DefStyle)
	promptStyle := tcell.StyleDefault.Foreground(t.Prompt)
	drawRune(s, leftMargin, topMargin, promptStyle, '❯')
	inputStyle := tcell.StyleDefault.Foreground(t.Input)
	inputX := leftMargin + 2
	width, _ := s.Size()
	input := i.Current()
	maxRunes := width - inputX
	if maxRunes < 0 {
		maxRunes = 0
	}
	if len([]rune(input)) > maxRunes {
		runes := []rune(input)
		input = string(runes[:maxRunes])
	}
	drawText(s, inputX, topMargin, inputStyle, input)
	cursorX := inputX + len([]rune(input))
	if cursorX >= width {
		cursorX = width - 1
	}
	s.ShowCursor(cursorX, topMargin)
}

// idxToRank converts an index to its corresponding rank string
func idxToRank(idx chess.Rank) string {
	ranks := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	return ranks[idx]
}

// idxToFile converts an index to its corresponding file string
func idxToFile(idx int) string {
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	return files[idx]
}

// idxToSquare returns a string representing the algebraic notation
// for a square given a rank index and a file index
func idxToSquare(rIdx chess.Rank, fIdx int) string {
	return fmt.Sprintf("%v%v", idxToFile(fIdx), idxToRank(rIdx))
}

// drawMoves displays recent moves
func drawMoves(s tcell.Screen, game *chess.Game, t Theme) {
	leftMargin := leftMargin + 22
	boxStyle := tcell.StyleDefault.Foreground(t.MoveBox)
	drawText(s, leftMargin, topMargin, boxStyle, "┏━━━━━━━━━━━━━━━━━━━━━┓")
	for i := 0; i < 5; i++ {
		idx, white, black := moveIdx(game, i)
		row := fmt.Sprintf("┃ %-3v %-7v %-7v ┃", idx, white, black)
		drawText(s, leftMargin, topMargin+i+1, boxStyle, row)
	}
	drawText(s, leftMargin, topMargin+6, boxStyle, "┗━━━━━━━━━━━━━━━━━━━━━┛")
}

// gameMove is used to store intermediate data in moveIdx
type gameMove = struct {
	index string
	white string
	black string
}

// moveIdx gets the move at the requested index windowed by the
// number of moves capable of being displayed at once (5 at this moment)
func moveIdx(game *chess.Game, idx int) (string, string, string) {
	positions := game.Positions()
	gameMoves := make([]gameMove, 0)
	var gm gameMove

	// First, pull all of the moves into a slice for further analysis
	for i, move := range game.Moves() {
		pos := positions[i]
		txt := chess.AlgebraicNotation.Encode(chess.AlgebraicNotation{}, pos, move)
		// On even indicies, write the white move / index
		if i%2 == 0 {
			gm = gameMove{}
			gm.index = fmt.Sprintf("%v.", ((i / 2) + 1))
			gm.white = txt
			// Every odd index, terminates a move pair
		} else {
			gm.black = txt
			gameMoves = append(gameMoves, gm)
		}
	}

	// We can only display five move pairs, so the moveOffset
	// is used to paginate to the most recent
	moveOffset := 0
	if len(gameMoves) > 5 {
		moveOffset = len(gameMoves) - 5
	}

	// Make sure the idx+offset is not out of bounds
	if idx+moveOffset <= len(gameMoves)-1 {
		move := gameMoves[idx+moveOffset]
		return move.index, move.white, move.black
	}

	return "", "", ""
}

// lastMove returns a boolean representing whether sq was part of the
// last move made
func lastMove(game *chess.Game, sq string) bool {
	moves := game.Moves()
	if len(moves) > 0 {
		lastMove := moves[len(moves)-1]
		if lastMove.S1().String() == sq || lastMove.S2().String() == sq {
			return true
		}
	}
	return false
}

// hintSq returns a boolean representing whether the square passed matches
// any hint that may be set in the game state
func hintSq(move *chess.Move, sq string) bool {
	if move != nil && (move.S1().String() == sq || move.S2().String() == sq) {
		return true
	}
	return false
}

// drawBoard draws the board on the screen
func drawBoard(s tcell.Screen, game *chess.Game, t Theme, checkWhite, checkBlack bool, hint *chess.Move) {
	pos := game.Position()
	board := pos.Board()
	row := topMargin

	var r chess.Rank
	// Step through the ranks starting with the top row
	for r = 7; r >= 0; r-- {
		// Add some space on the left-hand side of the screen
		col := leftMargin
		// Draw the rank indicator to the left of the squares
		drawRank(s, col, row, r, t)
		// Add some space between the rank and squares
		col += 2

		// Walk the board
		for f := 0; f < numOfSquaresInRow; f++ {
			sq := getSquare(chess.File(f), chess.Rank(r))
			// This may contain a piece
			p := board.Piece(sq)
			// Square background color (from theme)
			sqBg := squareBg(sq, t)

			// Use the highlight color to mark last moves unless there is a hint
			if lastMove(game, idxToSquare(r, f)) {
				sqBg = t.SquareHigh
			}

			// Show hints
			if hintSq(hint, idxToSquare(r, f)) {
				sqBg = t.SquareHint
			}

			if (p == chess.BlackKing && checkBlack) ||
				(p == chess.WhiteKing && checkWhite) {
				sqBg = t.SquareCheck
			}

			// Draw the square
			drawSquare(s, col, row, p, sqBg, t)
			// Increment to next square
			col += 2
		}
		// Go to the next row
		row++
	}
	// Display the file (column)
	fileStyle := tcell.StyleDefault.Foreground(t.File)
	drawText(s, leftMargin+2, row, fileStyle, "a b c d e f g h")
}

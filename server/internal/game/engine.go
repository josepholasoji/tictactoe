// Package game implements the Tic-Tac-Toe rules engine: board state,
// move validation, and win/draw detection. It holds no session or
// networking concerns — those live in internal/session and internal/ws.
package game

import "errors"

const (
	Empty = ""
	MarkX = "X"
	MarkO = "O"
)

var ErrCellTaken = errors.New("cell already occupied")
var ErrPositionOutOfRange = errors.New("position must be between 0 and 8")

// Board is a 3x3 grid flattened row-major:
//
//	0 1 2
//	3 4 5
//	6 7 8
type Board [9]string

func NewBoard() Board {
	return Board{}
}

// ApplyMove returns a new board with mark placed at position.
func ApplyMove(b Board, position int, mark string) (Board, error) {
	if position < 0 || position > 8 {
		return b, ErrPositionOutOfRange
	}
	if b[position] != Empty {
		return b, ErrCellTaken
	}
	next := b
	next[position] = mark
	return next, nil
}

// IsFull reports whether every cell is occupied.
func IsFull(b Board) bool {
	for _, cell := range b {
		if cell == Empty {
			return false
		}
	}
	return true
}

// Opponent returns the other player's mark.
func Opponent(mark string) string {
	if mark == MarkX {
		return MarkO
	}
	return MarkX
}

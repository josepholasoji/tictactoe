package game

import (
	"errors"
	"testing"
)

func TestApplyMove(t *testing.T) {
	b := NewBoard()
	b, err := ApplyMove(b, 4, MarkX)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b[4] != MarkX {
		t.Fatalf("expected center to be X, got %q", b[4])
	}

	if _, err := ApplyMove(b, 4, MarkO); !errors.Is(err, ErrCellTaken) {
		t.Fatalf("expected ErrCellTaken, got %v", err)
	}

	if _, err := ApplyMove(b, 9, MarkO); !errors.Is(err, ErrPositionOutOfRange) {
		t.Fatalf("expected ErrPositionOutOfRange, got %v", err)
	}
}

func TestWinnerRow(t *testing.T) {
	b := Board{MarkX, MarkX, MarkX, Empty, Empty, Empty, Empty, Empty, Empty}
	winner, ok := Winner(b)
	if !ok || winner != MarkX {
		t.Fatalf("expected X to win, got %q ok=%v", winner, ok)
	}
}

func TestWinnerDiagonal(t *testing.T) {
	b := Board{MarkO, Empty, Empty, Empty, MarkO, Empty, Empty, Empty, MarkO}
	winner, ok := Winner(b)
	if !ok || winner != MarkO {
		t.Fatalf("expected O to win, got %q ok=%v", winner, ok)
	}
}

func TestNoWinner(t *testing.T) {
	b := Board{MarkX, MarkO, MarkX, MarkX, MarkO, MarkO, MarkO, MarkX, MarkX}
	if _, ok := Winner(b); ok {
		t.Fatalf("expected no winner on full drawn board")
	}
	if !IsFull(b) {
		t.Fatalf("expected board to be full")
	}
}

func TestOpponent(t *testing.T) {
	if Opponent(MarkX) != MarkO {
		t.Fatalf("expected opponent of X to be O")
	}
	if Opponent(MarkO) != MarkX {
		t.Fatalf("expected opponent of O to be X")
	}
}

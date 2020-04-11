package main

import (
	term "github.com/wvanlint/twf/terminal"
)

type StatusView struct {
	state *AppState
}

func (v *StatusView) Position(totalRows int, totalCols int) term.Position {
	return term.Position{
		Top:  totalRows,
		Left: 1,
		Rows: 1,
		Cols: totalCols,
	}
}

func (v *StatusView) HasBorder() bool {
	return false
}

func (v *StatusView) ShouldRender() bool {
	return true
}

func (v *StatusView) Render(p term.Position) []term.Line {
	line := term.NewLine(&term.Graphics{Reverse: true}, p.Cols)
	line.Append("", &term.Graphics{})
	return []term.Line{line}
}

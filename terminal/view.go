package terminal

type View interface {
	Position(int, int) Position
	HasBorder() bool
	ShouldRender() bool
	Render(Position) []Line
}

type Position struct {
	Top  int
	Left int

	Rows int
	Cols int
}

func (p *Position) Shrink(i int) Position {
	if p.Rows <= 0 || p.Cols <= 0 {
		return *p
	}
	return Position{
		Top:  p.Top + i,
		Left: p.Left + i,
		Rows: p.Rows - 2*i,
		Cols: p.Cols - 2*i,
	}
}

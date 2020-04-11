package terminal

type View interface {
	Dims() (float32, float32)
	HasBorder() bool
	Render() ([]Line, bool)
}

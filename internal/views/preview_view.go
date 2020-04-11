package views

import (
	"math"
	"os/exec"
	"strings"

	"github.com/wvanlint/twf/internal/config"
	"github.com/wvanlint/twf/internal/state"
	term "github.com/wvanlint/twf/internal/terminal"
	"go.uber.org/zap"
)

type previewView struct {
	config      *config.TwfConfig
	state       *state.State
	lastPreview string
	scroll      int
	noLines     int
}

func NewPreviewView(config *config.TwfConfig, state *state.State) term.View {
	return &previewView{
		config: config,
		state:  state,
	}
}

func (v *previewView) Position(totalRows int, totalCols int) term.Position {
	return term.Position{
		Top:  1,
		Left: int(math.Ceil(float64(totalCols)/2.0)) + 1,
		Rows: totalRows - 1,
		Cols: int(math.Floor(float64(totalCols) / 2.0)),
	}
}

func (v *previewView) HasBorder() bool {
	return true
}

func (v *previewView) ShouldRender() bool {
	node, _ := v.state.Root.FindPath(v.state.Cursor)
	return v.config.Preview.Enabled && !node.IsDir()
}

func (v *previewView) Render(p term.Position) []term.Line {
	if v.lastPreview != v.state.Cursor {
		v.lastPreview = v.state.Cursor
		v.scroll = 0
	}
	output, err := getPreview(v.config.Preview.PreviewCommand, v.state.Cursor)
	output = strings.ReplaceAll(output, "\t", "    ")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(output, "\n")
	if v.scroll > len(lines)-p.Rows {
		if len(lines) < p.Rows {
			v.scroll = 0
		} else {
			v.scroll = len(lines) - p.Rows
		}
	}

	termLines := []term.Line{}
	for i := v.scroll; i-v.scroll < p.Rows && i < len(lines); i++ {
		termLine := term.NewLine(&term.Graphics{}, p.Cols)
		termLine.AppendRaw(lines[i])
		zap.L().Sugar().Info(p.Cols)
		zap.L().Sugar().Info(termLine.Length())
		zap.L().Sugar().Info(termLine.Text())
		termLines = append(termLines, termLine)
	}
	return termLines
}

func getPreview(cmdTemplate string, path string) (string, error) {
	cmd := strings.ReplaceAll(cmdTemplate, "{}", path)
	var output strings.Builder
	preview := exec.Command("bash", "-c", cmd)
	preview.Stdout = &output
	err := preview.Run()
	return output.String(), err
}

func (v *previewView) GetCommands() map[string]term.Command {
	return map[string]term.Command{
		"preview:down": v.down,
		"preview:up":   v.up,
	}
}

func (v *previewView) up(helper term.TerminalHelper, args ...interface{}) {
	if v.scroll > 0 {
		v.scroll -= 1
	}
}

func (v *previewView) down(helper term.TerminalHelper, args ...interface{}) {
	v.scroll += 1
}

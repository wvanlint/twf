package main

import (
	"math"
	"os/exec"
	"strings"

	term "github.com/wvanlint/twf/terminal"
	"go.uber.org/zap"
)

type PreviewView struct {
	state *AppState
}

func (v *PreviewView) Position(totalRows int, totalCols int) term.Position {
	return term.Position{
		Top:  1,
		Left: int(math.Ceil(float64(totalCols)/2.0)) + 1,
		Rows: totalRows - 1,
		Cols: int(math.Floor(float64(totalCols) / 2.0)),
	}
}

func (v *PreviewView) HasBorder() bool {
	return true
}

func (v *PreviewView) ShouldRender() bool {
	return !v.state.Root.FindPath(v.state.Cursor).IsDir()
}

func (v *PreviewView) Render(p term.Position) []term.Line {
	output, err := getPreview("bat --color always {}", v.state.Cursor)
	output = strings.ReplaceAll(output, "\t", "    ")
	if err != nil {
		panic(err)
	}
	termLines := []term.Line{}
	for _, outputLine := range strings.Split(output, "\n") {
		zap.L().Sugar().Info("Output line:", outputLine)
		line := term.NewLine(&term.Graphics{}, p.Cols)
		zap.L().Sugar().Info("Cols:", p.Cols)
		line.AppendRaw(outputLine)
		zap.L().Sugar().Info("Line length:", line.Length())
		termLines = append(termLines, line)
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

func (v *PreviewView) GetCommands() map[string]term.Command {
	return map[string]term.Command{}
}

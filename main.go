package main

import (
	sys "golang.org/x/sys/unix"
	"os"
	"os/signal"
)

func main() {
	PrintWd()

	t, err := InitTerm()
	if err != nil {
		panic(err)
	}

	tree, err := InitTreeFromWd()
	if err != nil {
		panic(err)
	}
	state := &AppState{
		Root: tree,
	}
	t.Render(state)

	go func() {
		for {
			input := make([]byte, 1)
			_, err := os.Stdout.Read(input)
			if err == nil {
				if input[0] == 'j' {
					state.CursorLine += 1
					t.Render(state)
				} else if input[0] == 'k' {
					state.CursorLine -= 1
					t.Render(state)
				}
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		t.Close()
		done <- true
	}()
	signal.Notify(sigs, sys.SIGINT, sys.SIGTERM)
	<-done
}

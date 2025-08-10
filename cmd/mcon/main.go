package main

import (
	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	model, err := initialModel(cli)
	ctx.FatalIfErrorf(err)

	prog := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = prog.Run()
	ctx.FatalIfErrorf(err)
}

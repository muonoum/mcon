package main

import (
	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Stdin         string
	StdoutCommand []string `arg:""`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	stdout, err := streamCommand(cli.StdoutCommand[0], cli.StdoutCommand[1:]...)
	ctx.FatalIfErrorf(err)

	model, err := initialModel(100, cli.Stdin, stdout)
	ctx.FatalIfErrorf(err)

	prog := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = prog.Run()
	ctx.FatalIfErrorf(err)
}

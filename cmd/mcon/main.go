package main

import (
	"bufio"
	"os/exec"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct {
	Stdin         string
	StdoutCommand []string `arg:""`
}

func main() {
	var cli CLI
	kong.Parse(&cli)

	stdout, err := stream(cli.StdoutCommand[0], cli.StdoutCommand[1:]...)
	if err != nil {
		panic(err)
	}

	prog := tea.NewProgram(initialModel(100, cli.Stdin, stdout),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := prog.Run(); err != nil {
		panic(err)
	}
}

func stream(arg string, args ...string) (chan string, error) {
	cmd := exec.Command(arg, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	channel := make(chan string)

	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				panic(err)
			}

			channel <- scanner.Text()
		}

		cmd.Process.Kill()
	}()

	return channel, nil
}

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CLI struct {
	Stdin         string
	StdoutCommand []string `arg:""`

	StdoutBackgroundColor string `default:""`
	StdoutTextColor       string `default:"#bcb5b3"`
	PromptBackgroundColor string `default:"#bcb5b3"`
	PromptTextColor       string `default:"#2b252b"`
	CursorColor           string `default:"#6c6563"`
}

type Model struct {
	cli         CLI
	initialised bool

	buffer   []string
	limit    int
	receiver chan string

	prompt textinput.Model

	stdin  io.Writer
	stdout viewport.Model
}

type LogMsg string

func receive(receiver chan string) tea.Cmd {
	return func() tea.Msg {
		return LogMsg(<-receiver)
	}
}

func initialModel(limit int, cli CLI, receiver chan string) (m Model, err error) {
	m.cli = cli
	m.limit = limit
	m.receiver = receiver
	m.prompt = textinput.New()
	m.prompt.Prompt = ""
	m.prompt.Focus()
	m.prompt.TextStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(m.cli.PromptBackgroundColor)).
		Foreground(lipgloss.Color(m.cli.PromptTextColor))
	m.prompt.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.cli.CursorColor))
	m.prompt.Cursor.TextStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(cli.PromptBackgroundColor))

	if cli.Stdin != "" {
		if m.stdin, err = os.OpenFile(cli.Stdin, os.O_WRONLY, 0644); err != nil {
			return m, err
		}
	}

	return m, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(receive(m.receiver), textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			value := m.prompt.Value()
			m.prompt.Reset()

			if m.stdin != nil {
				fmt.Fprintln(m.stdin, value)
			}
		}

	case tea.WindowSizeMsg:
		height := msg.Height - lipgloss.Height(m.prompt.View())
		m.prompt.Width = msg.Width

		if !m.initialised {
			m.initialised = true
			m.stdout = viewport.New(msg.Width, height)
			m.stdout.Style = lipgloss.NewStyle().
				Background(lipgloss.Color(m.cli.StdoutBackgroundColor)).
				Foreground(lipgloss.Color(m.cli.StdoutTextColor))

			m.stdout.SetContent(strings.Join(m.buffer, "\n"))
			m.stdout.GotoBottom()
		} else {
			m.stdout.Width = msg.Width
			m.stdout.Height = height
			m.stdout.GotoBottom()
		}

	case LogMsg:
		m.buffer = append(m.buffer, string(msg))
		if len(m.buffer) > m.limit {
			m.buffer = m.buffer[len(m.buffer)-m.limit:]
		}

		m.stdout.SetContent(strings.Join(m.buffer, "\n"))
		m.stdout.GotoBottom()
		return m, receive(m.receiver)
	}

	var inputCmd, viewportCmd tea.Cmd
	m.prompt, inputCmd = m.prompt.Update(msg)
	m.stdout, viewportCmd = m.stdout.Update(msg)
	return m, tea.Batch(inputCmd, viewportCmd)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.stdout.View(),
		m.prompt.View(),
	)
}

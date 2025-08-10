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
	Scrollback    int      `default:"1000"`
	StdoutCommand []string `arg:""`
	Theme         Theme    `embed:""`
}

type Theme struct {
	StdoutBackgroundColor string `default:""`
	StdoutTextColor       string `default:"#bcb5b3"`
	PromptBackgroundColor string `default:"#bcb5b3"`
	PromptTextColor       string `default:"#2b252b"`
	CursorColor           string `default:"#6c6563"`
}

type Model struct {
	initialised bool

	theme Theme

	buffer     []string
	scrollBack int
	receiver   chan string

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

func initialModel(cli CLI) (m Model, _ error) {
	receiver, err := stdoutChannel(cli.StdoutCommand[0], cli.StdoutCommand[1:]...)
	if err != nil {
		return m, err
	}

	m.theme = cli.Theme
	m.receiver = receiver
	m.scrollBack = cli.Scrollback

	m.prompt = textinput.New()
	m.prompt.Prompt = ""
	m.prompt.TextStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(m.theme.PromptBackgroundColor)).
		Foreground(lipgloss.Color(m.theme.PromptTextColor))
	m.prompt.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.CursorColor))
	m.prompt.Cursor.TextStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(m.theme.PromptBackgroundColor))

	m.prompt.Focus()

	if cli.Stdin != "" {
		m.stdin, err = os.OpenFile(cli.Stdin, os.O_WRONLY, 0644)
	}

	return m, err
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
				Background(lipgloss.Color(m.theme.StdoutBackgroundColor)).
				Foreground(lipgloss.Color(m.theme.StdoutTextColor))

			m.stdout.SetContent(strings.Join(m.buffer, "\n"))
			m.stdout.GotoBottom()
		} else {
			m.stdout.Width = msg.Width
			m.stdout.Height = height
			m.stdout.GotoBottom()
		}

	case LogMsg:
		m.buffer = append(m.buffer, string(msg))
		if len(m.buffer) > m.scrollBack {
			m.buffer = m.buffer[len(m.buffer)-m.scrollBack:]
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

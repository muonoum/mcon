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

type Model struct {
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

func initialModel(limit int, fifo string, receiver chan string) (Model, error) {
	var m Model

	m.limit = limit
	m.receiver = receiver
	m.prompt = textinput.New()
	m.prompt.Prompt = ""
	m.prompt.Focus()

	if fifo != "" {
		var err error
		if m.stdin, err = os.OpenFile(fifo, os.O_WRONLY, 0644); err != nil {
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

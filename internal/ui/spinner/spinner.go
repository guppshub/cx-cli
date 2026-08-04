package spinner

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	message  string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case quitMsg:
		m.quitting = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

type quitMsg struct{}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	spinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Electric blue spinner
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))  // Dim gray text
	return fmt.Sprintf("\r%s %s", spinStyle.Render(m.spinner.View()), msgStyle.Render(m.message))
}

type Spinner struct {
	program *tea.Program
	done    chan struct{}
	quitted bool
}

// Start launches the spinner and returns a stop function.
func Start(message string) *Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	m := model{
		spinner: s,
		message: message,
	}

	done := make(chan struct{})
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	go func() {
		_, _ = p.Run()
		close(done)
	}()

	return &Spinner{program: p, done: done}
}

// Stop cleanly terminates the spinner and clears the line.
func (s *Spinner) Stop() {
	if s.quitted {
		return
	}
	s.quitted = true
	s.program.Send(quitMsg{})
	<-s.done // Wait for the Bubble Tea program to exit and restore the terminal cooked state
	fmt.Fprint(os.Stderr, "\r\033[K")
}

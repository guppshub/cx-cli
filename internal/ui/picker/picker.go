package picker

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Row represents a single structured entry in the picker menu.
type Row struct {
	ID     string   // Unique ID returned on selection
	Fields []string // List of column values
}

type model struct {
	title       string
	headers     []string
	colWidths   []int
	items       []Row
	filtered    []Row
	cursor      int
	windowStart int
	windowSize  int
	searchInput textinput.Model
	selectedID  string
	cancelled   bool
}

// SingleSelect displays a scrollable, fuzzy-filterable menu of items and blocks until selection or cancellation.
func SingleSelect(title string, headers []string, items []Row) (string, error) {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	colWidths := calculateColWidths(headers, items)

	m := model{
		title:       title,
		headers:     headers,
		colWidths:   colWidths,
		items:       items,
		filtered:    items,
		windowStart: 0,
		windowSize:  12,
		searchInput: ti,
		selectedID:  "",
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	fm := finalModel.(model)
	if fm.cancelled {
		return "", nil
	}
	return fm.selectedID, nil
}

func calculateColWidths(headers []string, items []Row) []int {
	numCols := len(headers)
	for _, item := range items {
		if len(item.Fields) > numCols {
			numCols = len(item.Fields)
		}
	}

	widths := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		maxLen := 0
		if i < len(headers) {
			maxLen = len(headers[i])
		}
		for _, item := range items {
			if i < len(item.Fields) {
				if len(item.Fields[i]) > maxLen {
					maxLen = len(item.Fields[i])
				}
			}
		}
		widths[i] = maxLen
	}
	return widths
}

func formatRow(fields []string, widths []int, spacing int) string {
	var parts []string
	for i, f := range fields {
		width := 0
		if i < len(widths) {
			width = widths[i]
		}
		padLen := width - len(f)
		if padLen < 0 {
			padLen = 0
		}
		parts = append(parts, f+strings.Repeat(" ", padLen))
	}
	return strings.Join(parts, strings.Repeat(" ", spacing))
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.filtered) > 0 && m.cursor >= 0 && m.cursor < len(m.filtered) {
				m.selectedID = m.filtered[m.cursor].ID
				return m, tea.Quit
			}
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyUp:
			if len(m.filtered) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.filtered) - 1
				}
				m.adjustScroll()
			}
			return m, nil

		case tea.KeyDown:
			if len(m.filtered) > 0 {
				m.cursor++
				if m.cursor >= len(m.filtered) {
					m.cursor = 0
				}
				m.adjustScroll()
			}
			return m, nil
		}
	}

	// Update text input
	oldQuery := m.searchInput.Value()
	m.searchInput, cmd = m.searchInput.Update(msg)
	newQuery := m.searchInput.Value()

	if oldQuery != newQuery {
		m.applyFilter(newQuery)
	}

	return m, cmd
}

func (m *model) adjustScroll() {
	if m.cursor < m.windowStart {
		m.windowStart = m.cursor
	} else if m.cursor >= m.windowStart+m.windowSize {
		m.windowStart = m.cursor - m.windowSize + 1
	}
}

func (m *model) applyFilter(query string) {
	if query == "" {
		m.filtered = m.items
	} else {
		var res []Row
		words := strings.Fields(strings.ToLower(query))
		for _, item := range m.items {
			matched := true
			for _, word := range words {
				wordFound := false
				for _, field := range item.Fields {
					if strings.Contains(strings.ToLower(field), word) {
						wordFound = true
						break
					}
				}
				if !wordFound {
					matched = false
					break
				}
			}
			if matched {
				res = append(res, item)
			}
		}
		m.filtered = res
	}

	m.cursor = 0
	m.windowStart = 0
}

func (m model) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	s.WriteString(titleStyle.Render(m.title) + "\n\n")
	s.WriteString("Search: " + m.searchInput.View() + "\n")
	s.WriteString(strings.Repeat("─", 50) + "\n")

	// Render headers
	if len(m.headers) > 0 {
		formattedHeader := formatRow(m.headers, m.colWidths, 3)
		s.WriteString("  " + headerStyle.Render(formattedHeader) + "\n")
		s.WriteString(strings.Repeat("─", 50) + "\n")
	}

	if len(m.filtered) == 0 {
		s.WriteString("  No matches found\n")
	} else {
		end := m.windowStart + m.windowSize
		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		for i := m.windowStart; i < end; i++ {
			item := m.filtered[i]
			formatted := formatRow(item.Fields, m.colWidths, 3)

			if i == m.cursor {
				s.WriteString("❯ " + selectedStyle.Render(formatted) + "\n")
			} else {
				s.WriteString("  " + formatted + "\n")
			}
		}
	}

	s.WriteString(strings.Repeat("─", 50) + "\n")
	s.WriteString(helpStyle.Render("↑/↓ Navigate • Enter Select • Esc Cancel") + "\n")

	return s.String()
}

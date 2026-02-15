package prompter

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InteractivePrompter provides interactive prompts via bubbletea.
type InteractivePrompter struct{}

var _ Prompter = InteractivePrompter{}

// Styles
var (
	questionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	checkedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// --- Confirm ---

type confirmModel struct {
	question     string
	defaultValue bool
	value        bool
	done         bool
	cancelled    bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "Y":
			m.value = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.value = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.value = m.defaultValue
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		result := "No"
		if m.value {
			result = "Yes"
		}
		return fmt.Sprintf("%s %s\n", questionStyle.Render(m.question), result)
	}
	hint := "[y/N]"
	if m.defaultValue {
		hint = "[Y/n]"
	}
	return fmt.Sprintf("%s %s ", questionStyle.Render(m.question), helpStyle.Render(hint))
}

func (InteractivePrompter) Confirm(_ context.Context, message string, defaultValue bool) (bool, error) {
	m := confirmModel{question: message, defaultValue: defaultValue}
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return defaultValue, fmt.Errorf("bubbletea: %w", err)
	}
	cm := result.(confirmModel)
	if cm.cancelled {
		return defaultValue, context.Canceled
	}
	return cm.value, nil
}

// --- Select ---

type selectModel struct {
	question string
	options  []string
	cursor   int
	done     bool
	cancelled bool
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	if m.done {
		return fmt.Sprintf("%s %s\n", questionStyle.Render(m.question), m.options[m.cursor])
	}
	var b strings.Builder
	b.WriteString(questionStyle.Render(m.question) + "\n")
	for i, opt := range m.options {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, opt))
	}
	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • q quit"))
	return b.String()
}

func (InteractivePrompter) Select(_ context.Context, message string, options []string, defaultIndex int) (int, error) {
	m := selectModel{question: message, options: options, cursor: defaultIndex}
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return defaultIndex, fmt.Errorf("bubbletea: %w", err)
	}
	sm := result.(selectModel)
	if sm.cancelled {
		return defaultIndex, context.Canceled
	}
	return sm.cursor, nil
}

// --- MultiSelect ---

type multiSelectModel struct {
	question  string
	options   []string
	cursor    int
	selected  map[int]bool
	done      bool
	cancelled bool
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "a":
			allSelected := true
			for i := range m.options {
				if !m.selected[i] {
					allSelected = false
					break
				}
			}
			for i := range m.options {
				m.selected[i] = !allSelected
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	if m.done {
		var selected []string
		for i, opt := range m.options {
			if m.selected[i] {
				selected = append(selected, opt)
			}
		}
		return fmt.Sprintf("%s %s\n", questionStyle.Render(m.question), strings.Join(selected, ", "))
	}
	var b strings.Builder
	b.WriteString(questionStyle.Render(m.question) + "\n")
	for i, opt := range m.options {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}
		check := "[ ] "
		if m.selected[i] {
			check = checkedStyle.Render("[x] ")
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, check, opt))
	}
	b.WriteString(helpStyle.Render("↑/↓ navigate • space toggle • a all • enter confirm • q quit"))
	return b.String()
}

func (InteractivePrompter) MultiSelect(_ context.Context, message string, options []string, defaults []bool) ([]int, error) {
	selected := make(map[int]bool)
	for i, d := range defaults {
		selected[i] = d
	}
	m := multiSelectModel{question: message, options: options, selected: selected}
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		var indices []int
		for i, d := range defaults {
			if d {
				indices = append(indices, i)
			}
		}
		return indices, fmt.Errorf("bubbletea: %w", err)
	}
	msm := result.(multiSelectModel)
	if msm.cancelled {
		var indices []int
		for i, d := range defaults {
			if d {
				indices = append(indices, i)
			}
		}
		return indices, context.Canceled
	}
	var indices []int
	for i := range options {
		if msm.selected[i] {
			indices = append(indices, i)
		}
	}
	return indices, nil
}

// --- Input ---

type inputModel struct {
	question     string
	defaultValue string
	value        string
	done         bool
	cancelled    bool
}

func (m inputModel) Init() tea.Cmd { return nil }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			if m.value == "" {
				m.value = m.defaultValue
			}
			m.done = true
			return m, tea.Quit
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.value += msg.String()
			}
		}
	}
	return m, nil
}

func (m inputModel) View() string {
	if m.done {
		return fmt.Sprintf("%s %s\n", questionStyle.Render(m.question), m.value)
	}
	defaultHint := ""
	if m.defaultValue != "" {
		defaultHint = helpStyle.Render(fmt.Sprintf(" (%s)", m.defaultValue))
	}
	return fmt.Sprintf("%s%s %s", questionStyle.Render(m.question), defaultHint, m.value)
}

func (InteractivePrompter) Input(_ context.Context, message string, defaultValue string) (string, error) {
	m := inputModel{question: message, defaultValue: defaultValue}
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return defaultValue, fmt.Errorf("bubbletea: %w", err)
	}
	im := result.(inputModel)
	if im.cancelled {
		return defaultValue, context.Canceled
	}
	return im.value, nil
}

// Package tui provides an interactive terminal user interface for selecting
// target directories and entering page titles when running til without flags.
package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/src"
	"github.com/senorprogrammer/til/src/commands"
	"github.com/senorprogrammer/til/src/logging"
)

const (
	stateSelectingTarget = iota
	stateEnteringTitle
	stateDone
	stateError
)

const (
	defaultTitleInputWidth = 50
	defaultTitleCharLimit  = 200
	listHeightOffset       = 2
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type targetItem struct {
	key  string
	path string
}

func (i targetItem) FilterValue() string { return i.key }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(targetItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%6s\t%s", item.key, item.path)

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("> "+str))
	} else {
		fmt.Fprint(w, itemStyle.Render("  "+str))
	}
}

type model struct {
	targets        map[string]string
	targetList     list.Model
	titleInput     textinput.Model
	selectedTarget string
	state          int
	config         *config.Config
	logger         logging.Logger
	err            error
}

// NewInteractiveModel creates a new interactive TUI model
func NewInteractiveModel(cfg *config.Config, logger logging.Logger) (*model, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	dirMap, err := cfg.Map("targetDirectories")
	if err != nil {
		return nil, fmt.Errorf("failed to get target directories: %w", err)
	}

	targets := make(map[string]string)
	items := []list.Item{}

	for key, dir := range dirMap {
		str, ok := dir.(string)
		if !ok {
			continue
		}
		targets[key] = str
		items = append(items, targetItem{key: key, path: str})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no target directories configured")
	}

	m := &model{
		targets: targets,
		config:  cfg,
		logger:  logger,
		state:   stateSelectingTarget,
	}

	// If only one target, skip selection
	if len(targets) == 1 {
		for key := range targets {
			m.selectedTarget = key
		}
		m.state = stateEnteringTitle
	} else {
		// Setup list for target selection
		delegate := itemDelegate{}
		m.targetList = list.New(items, delegate, 0, 0)
		m.targetList.Title = "Select target directory"
		m.targetList.SetShowStatusBar(false)
		m.targetList.SetFilteringEnabled(false)
		m.targetList.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	}

	// Setup text input for title
	ti := textinput.New()
	ti.Placeholder = "Enter page title..."
	ti.Focus()
	ti.CharLimit = defaultTitleCharLimit
	ti.Width = defaultTitleInputWidth
	m.titleInput = ti

	return m, nil
}

func (m *model) Init() tea.Cmd {
	if m.state == stateSelectingTarget {
		return nil
	}
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.state == stateSelectingTarget {
			m.targetList.SetWidth(msg.Width)
			m.targetList.SetHeight(msg.Height - listHeightOffset)
		} else if m.state == stateEnteringTitle {
			// Update title input width to be responsive to window size
			// Leave some margin for padding
			if msg.Width > 20 {
				m.titleInput.Width = msg.Width - 10
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch m.state {
		case stateSelectingTarget:
			switch msg.String() {
			case "enter":
				selectedItem := m.targetList.SelectedItem()
				if selectedItem != nil {
					item, ok := selectedItem.(targetItem)
					if ok {
						m.selectedTarget = item.key
						m.state = stateEnteringTitle
						return m, textinput.Blink
					}
				}
			default:
				var cmd tea.Cmd
				m.targetList, cmd = m.targetList.Update(msg)
				return m, cmd
			}

		case stateEnteringTitle:
			switch msg.String() {
			case "enter":
				title := m.titleInput.Value()
				if title == "" {
					return m, nil
				}
				return m, m.executeCreateCommand(title)
			default:
				var cmd tea.Cmd
				m.titleInput, cmd = m.titleInput.Update(msg)
				return m, cmd
			}
		}

	case error:
		m.err = msg
		m.state = stateError
		return m, nil

	case createSuccessMsg:
		m.state = stateDone
		return m, tea.Quit
	}

	return m, nil
}

func (m *model) View() string {
	switch m.state {
	case stateSelectingTarget:
		return m.targetList.View()

	case stateEnteringTitle:
		return titleStyle.Render(fmt.Sprintf(
			"\n📝 Enter page title:\n\n%s\n\n%s",
			m.titleInput.View(),
			"(Press Enter to create, Ctrl+C to quit)",
		))

	case stateError:
		return fmt.Sprintf("\n✘ Error: %v\n\nPress Ctrl+C to quit.\n", m.err)

	case stateDone:
		return ""
	}

	return ""
}

// RunInteractive launches the interactive TUI
func RunInteractive(cfg *config.Config, logger logging.Logger) error {
	m, err := NewInteractiveModel(cfg, logger)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

type createSuccessMsg struct{}

// executeCreateCommand executes the create command synchronously.
// This is intentional - the command will exit the TUI after execution.
func (m *model) executeCreateCommand(title string) tea.Cmd {
	return func() tea.Msg {
		tDir, err := src.GetTargetDir(m.config, m.selectedTarget, true)
		if err != nil {
			return err
		}

		editor := m.config.UString("editor", "open")
		if editor == "" {
			editor = "open"
		}

		createCmd := commands.NewCreateCommand(tDir, title, editor, m.logger)
		if err := createCmd.Execute(); err != nil {
			return err
		}

		return createSuccessMsg{}
	}
}

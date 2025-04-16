package initprogram

import (
	"embed"
	"fmt"
	"strings"

	"github.com/Kong/konnect-orchestrator/internal/config"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/platform"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Go! ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Go!"))

	programReceiver func(tea.Msg)
)

type responseMsg string

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type model struct {
	focusIndex    int
	inputs        []textinput.Model
	cursorMode    cursor.Mode
	err           error
	responses     []string
	resourceFiles embed.FS
	lastMsg       tea.Msg
}

func initialModel(files embed.FS, c config.Config) model {
	m := model{
		inputs:        make([]textinput.Model, 2),
		resourceFiles: files,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 150
		t.Width = 100

		switch i {
		case 0:
			t.Placeholder = "GitHub URL"
			t.SetValue(c.PlatformRepoURL)
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "GitHub Token"
			t.SetValue(c.PlatformRepoGHToken)
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case responseMsg:
		m.lastMsg = msg
		m.responses = append(m.responses, string(msg))
		return m, nil
	case errMsg:
		m.lastMsg = msg
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Change cursor mode
		case tea.KeyCtrlR:
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				run(m.inputs[0].Value(), m.inputs[1].Value(), m.resourceFiles)
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(m.err.Error())
		b.WriteString("\n")
		return b.String()
	}

	if m.responses != nil {
		for _, res := range m.responses {
			b.WriteString(res)
		}
		return b.String()
	}

	b.WriteString("\n")
	b.WriteString("The `koctl init` command initializes a\n")
	b.WriteString("Platform Team GitHub repository to prepare\n")
	b.WriteString("it to participate in the Konnect Reference Platform.\n\n")
	b.WriteString("In order to initialize the repository, we must collect\n")
	b.WriteString("two pieces of information. The GitHub URL and a GitHub token\n")
	b.WriteString("that has access to the repository.\n\n")
	b.WriteString("See the FAQ page for details on the token permission requirements:\n")
	b.WriteString("https://deploy-preview-783--kongdeveloper.netlify.app/konnect-reference-platform/faq/\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func run(url, token string, files embed.FS) {
	statusChan := make(chan string)
	retChan := make(chan error)

	go func() {
		for {
			select {
			case status := <-statusChan:
				programReceiver(responseMsg(status))
			case e := <-retChan:
				programReceiver(errMsg{e})
				return
			}
		}
	}()

	go func() {
		gitCfg := manifest.LoadGitConfigFromGhValues(url, token, "", "")
		retChan <- platform.Init(gitCfg, files, statusChan)
	}()
}

func Execute(files embed.FS, cfg config.Config) error {
	p := tea.NewProgram(initialModel(files, cfg))
	programReceiver = p.Send

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

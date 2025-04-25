package runprogram

import (
	"fmt"
	"strings"

	"github.com/Kong/konnect-orchestrator/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	currentView    runView
	welcome        *welcomeView
	setupGitHubApp *setupGitHubAppView
}

type runView interface {
	view(*model) string
	update(model, tea.Msg) (tea.Model, tea.Cmd)
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
	exitButton   = "Exit"
	nextButton   = "Next"
	prevButton   = "Back"
)

type inputsProvider interface {
	getInputs() []textinput.Model
}

type welcomeView struct {
	focusIndex int
}

type setupGitHubAppView struct {
	focusIndex int
	inputs     []textinput.Model
}

func focusButton(btn string) string {
	return fmt.Sprintf("[ %s ]", focusedStyle.Render(btn))
}

func blurButton(btn string) string {
	return fmt.Sprintf("[ %s ]", blurredStyle.Render(btn))
}

func (v *welcomeView) getInputs() []textinput.Model {
	return nil
}
func (v *welcomeView) view(_ *model) string {
	var b strings.Builder
	b.WriteString("How to run the Konnect Reference Platform Self Service App\n")
	if v.focusIndex == 0 {
		fmt.Fprintf(&b, "\n%s\t%s\n\n", focusButton(exitButton), blurButton(nextButton))
	} else {
		fmt.Fprintf(&b, "\n%s\t%s\n\n", blurButton(exitButton), focusButton(nextButton))
	}
	return b.String()
}

func (v *welcomeView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyTab:
			if v.focusIndex == 0 {
				v.focusIndex = 1
			} else {
				v.focusIndex = 0
			}
		case tea.KeyEnter:
			if v.focusIndex == 0 {
				return m, tea.Quit
			} // else
			m.currentView = m.setupGitHubApp
		}
	}
	return m, nil
}

func (v *setupGitHubAppView) getInputs() []textinput.Model {
	return v.inputs
}

func (v *setupGitHubAppView) view(_ *model) string {
	var b strings.Builder
	b.WriteString("Setup GitHub App View\n\n")
	for i := 0; i < len(v.inputs); i++ {
		b.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	fmt.Fprintf(&b, "\n\n%s\t%s\n\n", blurButton(prevButton), blurButton(nextButton))
	return b.String()
}

func (v *setupGitHubAppView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	updateInputs := func() []tea.Cmd {
		rv := make([]tea.Cmd, len(v.inputs))
		for i := range v.inputs {
			if i == v.focusIndex {
				rv[i] = v.inputs[i].Focus()
				v.inputs[i].PromptStyle = focusedStyle
				v.inputs[i].TextStyle = focusedStyle
			} else {
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = noStyle
				v.inputs[i].TextStyle = noStyle
			}
		}
		return rv
	}

	//var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyTab, tea.KeyDown, tea.KeyShiftTab, tea.KeyUp, tea.KeyEnter:
			s := msg.String()

			if s == "tab" || s == "down" {
				v.focusIndex++
				if v.focusIndex > len(v.inputs) {
					v.focusIndex = 0
				}
				//cmds = updateInputs()
			}

			//v.focusIndex--
			//if v.focusIndex < 0 {
			//	v.focusIndex = len(v.inputs)
			//}
			//cmds = updateInputs()

			//// if Enter is pressed on Next button, move to next view
			//// if Enter is pressed on Previous button, move to previous view
			//// otherwise, advance focus
			////return m, nil

			//return m, tea.Batch(cmds...)
		}
	}

	updateInputs()
	return m, nil
}

func updateInputs(inputs []textinput.Model, msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range inputs {
		inputs[i], cmds[i] = inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func initialModel(cfg config.Config) model {
	m := model{}
	m.welcome = &welcomeView{
		focusIndex: 1,
	}
	m.currentView = m.welcome

	m.setupGitHubApp = &setupGitHubAppView{
		focusIndex: 0,
	}
	m.setupGitHubApp.inputs = make([]textinput.Model, 2)
	m.setupGitHubApp.inputs[0] = textinput.New()
	m.setupGitHubApp.inputs[0].Prompt = "GitHub Client ID: "
	m.setupGitHubApp.inputs[0].SetValue(cfg.GitHubClientID)
	m.setupGitHubApp.inputs[0].PromptStyle = focusedStyle
	m.setupGitHubApp.inputs[0].TextStyle = focusedStyle

	m.setupGitHubApp.inputs[1] = textinput.New()
	m.setupGitHubApp.inputs[1].Prompt = "GitHub Client Secret: "
	m.setupGitHubApp.inputs[1].SetValue(cfg.GitHubClientSecret)
	m.setupGitHubApp.inputs[1].EchoCharacter = '•'

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m.currentView.update(m, msg)
}

func (m model) View() string {
	return m.currentView.view(&m)
}

func Execute(cfg config.Config) error {
	p := tea.NewProgram(initialModel(cfg))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

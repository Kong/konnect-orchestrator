package runprogram

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kong/konnect-orchestrator/internal/config"
	"github.com/Kong/konnect-orchestrator/internal/docker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	currentView       runView
	welcome           *welcomeView
	setupGitHubApp    *setupGitHubAppView
	setupPlatformRepo *setupPlatformRepoView
	focusedIndex      int
}

type runView interface {
	view(*model) string
	update(model, tea.Msg) (tea.Model, tea.Cmd)
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#286FEB"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
	exitButton   = "Exit"
	nextButton   = "Next"
	backButton   = "Back"
	runButton    = "Run"
)

type welcomeView struct{}

type setupGitHubAppView struct {
	inputs []textinput.Model
}

type setupPlatformRepoView struct {
	inputs []textinput.Model
}

func focusButton(btn string) string {
	return fmt.Sprintf("[ %s ]", focusedStyle.Render(btn))
}

func blurButton(btn string) string {
	return fmt.Sprintf("[ %s ]", blurredStyle.Render(btn))
}

func (v *welcomeView) view(m *model) string {
	var b strings.Builder
	b.WriteString("This command will guide you through running the Konnect Reference Platform\n")
	b.WriteString("self-service UI application. The self-service app allows teams to onboard\n")
	b.WriteString("themselves and their services to the Konnect Reference Platform.\n\n")
	b.WriteString("For more information on the reference platform:\n")
	b.WriteString("  https://developer.konghq.com/konnect-reference-platform\n\n")
	b.WriteString("On the following screens, press Tab to navigate between the buttons and inputs,\n")
	b.WriteString("and Enter to select the focused button.\n\n")

	if m.focusedIndex == 0 {
		fmt.Fprintf(&b, "\n%s     %s\n\n", focusButton(exitButton), blurButton(nextButton))
	} else {
		fmt.Fprintf(&b, "\n%s     %s\n\n", blurButton(exitButton), focusButton(nextButton))
	}

	return b.String()
}

func (v *welcomeView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyTab:
			if m.focusedIndex == 0 {
				m.focusedIndex = 1
			} else {
				m.focusedIndex = 0
			}
		case tea.KeyEnter:
			if m.focusedIndex == 0 {
				return m, tea.Quit
			}
			m.focusedIndex = 0
			m.currentView = m.setupGitHubApp
			m.setupGitHubApp.inputs[0].PromptStyle = focusedStyle
			m.setupGitHubApp.inputs[0].TextStyle = focusedStyle
			return m, m.setupGitHubApp.inputs[0].Focus()
		}
	}
	return m, nil
}

func (v *setupGitHubAppView) view(m *model) string {
	var b strings.Builder
	b.WriteString("The Reference Platform self-service application identifies\nitself to GitHub using the OAuth Apps integration.\n\n")
	b.WriteString("Before proceeding, you will need an OAuth Client ID and secret\nwhich are created within GitHub.\n\n")
	b.WriteString("For details on properly creating and configuring an OAuth App, see:\n")
	b.WriteString("  https://developer.konghq.com/konnect-reference-platform/self-service\n\n")
	b.WriteString("Once you have created the OAuth App, enter the Client ID and secret below.\n\n")

	for i := 0; i < len(v.inputs); i++ {
		b.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	if m.focusedIndex == len(v.inputs) {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", blurButton(backButton), focusButton(nextButton))
	} else if m.focusedIndex == len(v.inputs)+1 {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", focusButton(backButton), blurButton(nextButton))
	} else {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", blurButton(backButton), blurButton(nextButton))
	}

	return b.String()
}

func (v *setupGitHubAppView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyTab, tea.KeyDown, tea.KeyShiftTab, tea.KeyUp, tea.KeyEnter:
			s := msg.String()

			// if Enter is pressed on Next button, move to next view
			// if Enter is pressed on Previous button, move to previous view
			// otherwise, advance focus
			if s == "enter" {
				if m.focusedIndex == len(v.inputs) {
					// Next button
					m.focusedIndex = 0
					m.currentView = m.setupPlatformRepo
					m.setupPlatformRepo.inputs[0].PromptStyle = focusedStyle
					m.setupPlatformRepo.inputs[0].TextStyle = focusedStyle
					return m, m.setupPlatformRepo.inputs[0].Focus()
				} else if m.focusedIndex == len(v.inputs)+1 {
					// Previous button
					m.focusedIndex = 1
					m.currentView = m.welcome
					return m, textinput.Blink
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focusedIndex--
			} else {
				m.focusedIndex++
			}
			if m.focusedIndex > len(v.inputs)+1 {
				m.focusedIndex = 0
			} else if m.focusedIndex < 0 {
				m.focusedIndex = len(v.inputs) + 1
			}

			cmds := make([]tea.Cmd, len(v.inputs))
			for i := 0; i <= len(v.inputs)-1; i++ {
				if i == m.focusedIndex {
					// Set focused state
					cmds[i] = v.inputs[i].Focus()
					v.inputs[i].PromptStyle = focusedStyle
					v.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = noStyle
				v.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := updateInputs(v.inputs, msg)

	return m, cmd
}

func (v *setupPlatformRepoView) view(m *model) string {
	var b strings.Builder
	b.WriteString("When a development team uses the self-service app to onboard an API they\n")
	b.WriteString("build, the self-service application files a PR to the central platform\n")
	b.WriteString("repository introducing the necessary changes.\n\n")
	b.WriteString("The self-service app needs to be able to authenticate to that\n")
	b.WriteString("repository using a GitHub token.\n\n")
	b.WriteString("For details on properly creating and configuring a GitHub repository and token, see:\n")
	b.WriteString("  https://developer.konghq.com/konnect-reference-platform/self-service\n\n")
	b.WriteString("Once you have created the GitHub token, enter the Platform Repo URL and\n")
	b.WriteString("GitHub token below.\n\n")

	for i := 0; i < len(v.inputs); i++ {
		b.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	if m.focusedIndex == len(v.inputs) {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", blurButton(backButton), focusButton(runButton))
	} else if m.focusedIndex == len(v.inputs)+1 {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", focusButton(backButton), blurButton(runButton))
	} else {
		fmt.Fprintf(&b, "\n\n%s     %s\n\n", blurButton(backButton), blurButton(runButton))
	}
	return b.String()
}

func (v *setupPlatformRepoView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyTab, tea.KeyDown, tea.KeyShiftTab, tea.KeyUp, tea.KeyEnter:
			s := msg.String()

			// if Enter is pressed on Next button, move to next view
			// if Enter is pressed on Previous button, move to previous view
			// otherwise, advance focus
			if s == "enter" {
				if m.focusedIndex == len(v.inputs) {
					// Next button
					run(m.setupGitHubApp.inputs[0].Value(), m.setupGitHubApp.inputs[1].Value(),
						m.setupPlatformRepo.inputs[0].Value(), m.setupPlatformRepo.inputs[1].Value())
					return m, tea.Quit
				} else if m.focusedIndex == len(v.inputs)+1 {
					// Previous button
					m.focusedIndex = 0
					m.currentView = m.setupGitHubApp
					m.setupGitHubApp.inputs[0].PromptStyle = focusedStyle
					m.setupGitHubApp.inputs[0].TextStyle = focusedStyle
					return m, m.setupGitHubApp.inputs[0].Focus()
				}
			}

			if s == "up" || s == "shift+tab" {
				m.focusedIndex--
			} else {
				m.focusedIndex++
			}
			if m.focusedIndex > len(v.inputs)+1 {
				m.focusedIndex = 0
			} else if m.focusedIndex < 0 {
				m.focusedIndex = len(v.inputs) + 1
			}

			cmds := make([]tea.Cmd, len(v.inputs))
			for i := 0; i <= len(v.inputs)-1; i++ {
				if i == m.focusedIndex {
					// Set focused state
					cmds[i] = v.inputs[i].Focus()
					v.inputs[i].PromptStyle = focusedStyle
					v.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = noStyle
				v.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := updateInputs(v.inputs, msg)

	return m, cmd
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
	m := model{
		focusedIndex: 1,
	}
	m.welcome = &welcomeView{}
	m.currentView = m.welcome

	m.setupGitHubApp = &setupGitHubAppView{}
	m.setupGitHubApp.inputs = make([]textinput.Model, 2)
	m.setupGitHubApp.inputs[0] = textinput.New()
	m.setupGitHubApp.inputs[0].Prompt = "GitHub Client ID: "
	m.setupGitHubApp.inputs[0].SetValue(cfg.GitHubClientID)
	m.setupGitHubApp.inputs[0].PromptStyle = focusedStyle
	m.setupGitHubApp.inputs[0].TextStyle = focusedStyle
	m.setupGitHubApp.inputs[0].Cursor.Style = cursorStyle

	m.setupGitHubApp.inputs[1] = textinput.New()
	m.setupGitHubApp.inputs[1].Prompt = "GitHub Client Secret: "
	m.setupGitHubApp.inputs[1].SetValue(cfg.GitHubClientSecret)
	m.setupGitHubApp.inputs[1].EchoMode = textinput.EchoPassword
	m.setupGitHubApp.inputs[1].EchoCharacter = '•'
	m.setupGitHubApp.inputs[1].Cursor.Style = cursorStyle

	m.setupPlatformRepo = &setupPlatformRepoView{}
	m.setupPlatformRepo.inputs = make([]textinput.Model, 2)
	m.setupPlatformRepo.inputs[0] = textinput.New()
	m.setupPlatformRepo.inputs[0].Prompt = "Platform Repo URL: "
	m.setupPlatformRepo.inputs[0].SetValue(cfg.PlatformRepoURL)
	m.setupPlatformRepo.inputs[0].PromptStyle = focusedStyle
	m.setupPlatformRepo.inputs[0].TextStyle = focusedStyle
	m.setupPlatformRepo.inputs[0].Cursor.Style = cursorStyle

	m.setupPlatformRepo.inputs[1] = textinput.New()
	m.setupPlatformRepo.inputs[1].Prompt = "Platform Repo GitHub Token: "
	m.setupPlatformRepo.inputs[1].SetValue(cfg.PlatformRepoGHToken)
	m.setupPlatformRepo.inputs[1].EchoMode = textinput.EchoPassword
	m.setupPlatformRepo.inputs[1].EchoCharacter = '•'
	m.setupPlatformRepo.inputs[1].Cursor.Style = cursorStyle

	return m
}

func run(githubClientID, githubClientSecret, platformRepoURL, platformRepoGHToken string) {
	// expose the arguments as env variables to the docker containers
	envVars := map[string]string{
		"GITHUB_CLIENT_ID":           githubClientID,
		"GITHUB_CLIENT_SECRET":       githubClientSecret,
		"PLATFORM_REPO_URL":          platformRepoURL,
		"PLATFORM_REPO_GITHUB_TOKEN": platformRepoGHToken,
	}
	_, _ = docker.ComposeUp(context.Background(), docker.KoctlRunComposeFile, "koctl-run", envVars)
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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

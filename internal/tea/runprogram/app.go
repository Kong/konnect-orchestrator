package runprogram

import (
	"context"
	"fmt"
	"os"
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
	inputs     []textinput.Model
	saveConfig bool
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

	// … your introductory text …

	// 1) render the two text‐inputs
	for i := range v.inputs {
		b.WriteString(v.inputs[i].View())
		b.WriteRune('\n')
	}

	// 2) render the checkbox
	check := "[ ] Save config to .env file"
	if v.saveConfig {
		check = "[X] Save config to .env file"
	}
	// if it’s focused, give it the “focus” styling
	if m.focusedIndex == len(v.inputs) {
		check = focusedStyle.Render(check)
	}
	b.WriteString("\n")
	b.WriteString(check)
	b.WriteString("\n\n")

	// 3) render Back / Run buttons
	backIdx := len(v.inputs) + 1
	runIdx := len(v.inputs) + 2

	switch m.focusedIndex {
	case backIdx:
		fmt.Fprintf(&b, "%s     %s",
			focusButton(backButton),
			blurButton(runButton),
		)
	case runIdx:
		fmt.Fprintf(&b, "%s     %s",
			blurButton(backButton),
			focusButton(runButton),
		)
	default:
		fmt.Fprintf(&b, "%s     %s",
			blurButton(backButton),
			blurButton(runButton),
		)
	}

	return b.String()
}

func (v *setupPlatformRepoView) update(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	maxIdx := len(v.inputs) + 2 // inputs + checkbox + Back + Run

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		// ─── Move focus backwards ───────────────────────────────────────
		case "up", "shift+tab":
			if m.focusedIndex == len(v.inputs) {
				// from checkbox back to last text input
				m.focusedIndex = len(v.inputs) - 1
			} else {
				m.focusedIndex--
			}

		// ─── Move focus forwards ────────────────────────────────────────
		case "down", "tab":
			if m.focusedIndex == len(v.inputs) {
				// skip Back, go straight to Run
				m.focusedIndex = len(v.inputs) + 2
			} else {
				m.focusedIndex++
			}

		// ─── Enter: advance focus or activate Back/Run ──────────────────
		case "enter":
			switch m.focusedIndex {
			case len(v.inputs):
				// on checkbox → skip to Run
				m.focusedIndex = len(v.inputs) + 2
			case len(v.inputs) + 1:
				// Back button
				m.focusedIndex = 0
				m.currentView = m.setupGitHubApp
				return m, m.setupGitHubApp.inputs[0].Focus()
			case len(v.inputs) + 2:
				// Run button
				run(
					m.setupGitHubApp.inputs[0].Value(),
					m.setupGitHubApp.inputs[1].Value(),
					m.setupPlatformRepo.inputs[0].Value(),
					m.setupPlatformRepo.inputs[1].Value(),
					v.saveConfig,
				)
				return m, tea.Quit
			default:
				// on any text input → move next
				m.focusedIndex++
			}

		// ─── Space: only toggle the checkbox ─────────────────────────────
		case " ":
			if m.focusedIndex == len(v.inputs) {
				v.saveConfig = !v.saveConfig
				return m, nil
			}
			// otherwise fall through to default so text inputs get the space

		// ─── All other keys ─────────────────────────────────────────────
		default:
			if m.focusedIndex < len(v.inputs) {
				// forward typing, backspace, delete, etc.
				return m, updateInputs(v.inputs, msg)
			}
			return m, nil
		}

		// ─── Wrap‐around focus index ────────────────────────────────────
		if m.focusedIndex < 0 {
			m.focusedIndex = maxIdx
		} else if m.focusedIndex > maxIdx {
			m.focusedIndex = 0
		}

		// ─── Update text‐input focus/blur styling ───────────────────────
		var cmds []tea.Cmd
		for i := range v.inputs {
			if i == m.focusedIndex {
				v.inputs[i].PromptStyle = focusedStyle
				v.inputs[i].TextStyle = focusedStyle
				cmds = append(cmds, v.inputs[i].Focus())
			} else {
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = noStyle
				v.inputs[i].TextStyle = noStyle
			}
		}
		return m, tea.Batch(cmds...)
	}

	// non-KeyMsg: no change
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
	m.setupPlatformRepo.saveConfig = false
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

func run(
	githubClientID,
	githubClientSecret,
	platformRepoURL,
	platformRepoGHToken string,
	saveConfig bool,
) {
	// expose the arguments as env variables to the docker containers
	envVars := map[string]string{
		"GITHUB_CLIENT_ID":           githubClientID,
		"GITHUB_CLIENT_SECRET":       githubClientSecret,
		"PLATFORM_REPO_URL":          platformRepoURL,
		"PLATFORM_REPO_GITHUB_TOKEN": platformRepoGHToken,
	}

	// buffer all post‐ComposeUp messages here
	var postMsgs []string

	if saveConfig {
		const (
			envFile    = ".env"
			backupFile = ".env.bak"
		)

		// 1) back up existing .env (if any), but don't print yet
		if _, err := os.Stat(envFile); err == nil {
			if err := os.Rename(envFile, backupFile); err != nil {
				postMsgs = append(postMsgs, fmt.Sprintf(
					"warning: could not backup %s: %v", envFile, err,
				))
			} else {
				postMsgs = append(postMsgs, fmt.Sprintf(
					"backed up existing %s → %s", envFile, backupFile,
				))
			}
		}

		// 2) write new .env
		f, err := os.Create(envFile)
		if err != nil {
			postMsgs = append(postMsgs, fmt.Sprintf(
				"error: could not create %s: %v", envFile, err,
			))
		} else {
			defer f.Close()
			for k, v := range envVars {
				fmt.Fprintf(f, "%s=%s\n", k, v)
			}
			postMsgs = append(postMsgs, fmt.Sprintf("wrote new %s", envFile))
		}
	}

	// 3) actually start Docker
	_, _ = docker.ComposeUp(
		context.Background(),
		docker.KoctlRunComposeFile,
		"koctl-run",
		envVars,
	)

	// **then** push the cursor to a new line…
	fmt.Fprint(os.Stdout, "\r\n")

	// 4) now print all the buffered messages
	for _, msg := range postMsgs {
		fmt.Fprintln(os.Stdout, msg)
	}
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

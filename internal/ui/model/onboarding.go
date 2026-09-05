package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/dialog"
	"github.com/richavery/bvr-cli/internal/ui/util"
)

type onboardingStep int

const (
	onboardingStepWelcome onboardingStep = iota
	onboardingStepHome
	onboardingStepModels
	onboardingStepFileFinder
	onboardingStepNotifications
	onboardingStepThemes
	onboardingStepComplete
)

type onboardingModel struct {
	com   *common.Common
	step  onboardingStep
	done  bool
	error string
	info  string

	hasGit    bool
	hasOllama bool

	providers []string

	skillsSynced bool
	skillsCount  int
}

var _ dialog.Dialog = (*onboardingModel)(nil)

func newOnboardingModel(com *common.Common) *onboardingModel {
	return &onboardingModel{
		com:  com,
		step: onboardingStepWelcome,
	}
}

func (o *onboardingModel) ID() string {
	return "onboarding"
}

func (o *onboardingModel) Init() tea.Cmd {
	return tea.Batch(
		o.checkDependencies(),
		o.detectProviders(),
	)
}

func (o *onboardingModel) checkDependencies() tea.Cmd {
	return func() tea.Msg {
		gitPath, gitErr := lookPath("git")
		ollamaPath, ollamaErr := lookPath("ollama")
		return dependencyCheckMsg{
			hasGit:    gitErr == nil && gitPath != "",
			hasOllama: ollamaErr == nil && ollamaPath != "",
		}
	}
}

func (o *onboardingModel) detectProviders() tea.Cmd {
	return func() tea.Msg {
		providers, err := config.Providers(o.com.Config())
		ids := make([]string, 0)
		if err == nil {
			for _, p := range providers {
				ids = append(ids, string(p.ID))
			}
		}
		return providerDetectMsg{providers: ids, err: err}
	}
}

func (o *onboardingModel) syncSkills() tea.Cmd {
	return func() tea.Msg {
		count := 0
		for _, p := range config.GlobalSkillsDirs() {
			entries, err := os.ReadDir(p)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !strings.HasPrefix(e.Name(), ".") {
					count++
				}
			}
		}
		return skillsSyncedMsg{count: count}
	}
}

type dependencyCheckMsg struct {
	hasGit    bool
	hasOllama bool
}

type providerDetectMsg struct {
	providers []string
	err       error
}

type skillsSyncedMsg struct {
	count int
}

func (o *onboardingModel) HandleMsg(msg tea.Msg) dialog.Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return o.handleKey(msg)
	case dependencyCheckMsg:
		o.hasGit = msg.hasGit
		o.hasOllama = msg.hasOllama
		return nil
	case providerDetectMsg:
		o.providers = msg.providers
		return nil
	case skillsSyncedMsg:
		o.skillsCount = msg.count
		o.skillsSynced = true
		return nil
	case util.InfoMsg:
		if msg.Type == util.InfoTypeError {
			o.error = msg.Msg
		} else {
			o.info = msg.Msg
		}
		return nil
	}
	return nil
}

func (o *onboardingModel) handleKey(msg tea.KeyPressMsg) dialog.Action {
	switch msg.String() {
	case "esc":
		return dialog.ActionSkipOnboarding{}
	case "enter":
		switch o.step {
		case onboardingStepComplete:
			return dialog.ActionClose{}
		default:
			o.nextStep()
			return nil
		}
	case "o", "O":
		return dialog.ActionSkipOnboarding{}
	}
	return nil
}

func (o *onboardingModel) nextStep() {
	switch o.step {
	case onboardingStepWelcome:
		o.step = onboardingStepHome
	case onboardingStepHome:
		o.step = onboardingStepModels
	case onboardingStepModels:
		o.step = onboardingStepFileFinder
	case onboardingStepFileFinder:
		o.step = onboardingStepNotifications
	case onboardingStepNotifications:
		o.step = onboardingStepThemes
	case onboardingStepThemes:
		o.step = onboardingStepComplete
	case onboardingStepComplete:
		o.done = true
	}
}

func (o *onboardingModel) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	view := o.render()
	cur := &tea.Cursor{}
	styled := uv.NewStyledString(view)
	styled.Draw(scr, area)
	return cur
}

func (o *onboardingModel) render() string {
	var b strings.Builder
	b.WriteString(o.com.Styles.Initialize.Header.Render("Welcome to BVR"))
	b.WriteString("\n\n")

	switch o.step {
	case onboardingStepWelcome:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Let's take a quick tour."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("You'll see the home screen, models, file finder, notifications, and themes."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepHome:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Home"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("The home view is your cockpit."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Pick a project, start a session, and choose a model from one place."))
		b.WriteString("\n\n")
		b.WriteString(o.renderPreview("resources/screenshots/home-menu.jpg"))
		b.WriteString("\n")
		b.WriteString(o.renderASCIIPreview("home-menu"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepModels:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Models"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Open the models menu to switch providers and models."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Use /models or the models shortcut to browse local and cloud models."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Screenshot: resources/screenshots/menu-models.jpg"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepFileFinder:
		b.WriteString(o.com.Styles.Initialize.Content.Render("File Finder"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Browse project files, previews, metadata, and hidden files."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Use the finder to pull files into context without leaving the TUI."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Screenshot: resources/screenshots/file-finder.jpg"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepNotifications:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Notifications"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("BVR shows status updates, model warm-up, and task progress in the notification area."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Watch for provider, model, and workflow updates here."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Screenshot: resources/screenshots/notification-select.jpg"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepThemes:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Themes"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("BVR includes multiple themes."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Available theme screenshots:"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("- Pink: resources/screenshots/theme-pink.jpg"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("- Purple: resources/screenshots/theme-purple.jpg"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("- Default green home: resources/screenshots/home-green.jpg"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepComplete:
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Setup complete!"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("You're ready to use BVR."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Open /models to choose a model, then start chatting."))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to launch BVR"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to skip and explore later"))
	}

	if o.error != "" {
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Error: " + o.error))
	}
	if o.info != "" {
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render(o.info))
	}

	return b.String()
}

// lookPath is a tiny local wrapper around exec.LookPath to avoid importing
// os/exec in the UI layer if that package is unavailable.
func lookPath(name string) (string, error) {
	for _, dir := range []string{"/usr/local/bin", "/usr/bin", "/bin", "/opt/homebrew/bin"} {
		p := dir + "/" + name
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("%s not found", name)
}

func (o *onboardingModel) renderPreview(path string) string {
	label := filepath.Base(path)
	exists, err := pathExists(path)
	if err != nil {
		return o.com.Styles.Initialize.Content.Render("[preview unavailable: " + label + "]")
	}
	if !exists {
		return o.com.Styles.Initialize.Content.Render("[preview missing: " + label + "]")
	}
	switch terminalPreviews() {
	case "inline":
		return o.com.Styles.Initialize.Accent.Render("[preview " + label + "]")
	default:
		return o.com.Styles.Initialize.Content.Render("[preview " + label + "]")
	}
}

func (o *onboardingModel) renderASCIIPreview(name string) string {
	switch name {
	case "home-menu":
		return o.com.Styles.Initialize.Content.Render(strings.TrimSpace(`
┌─────────────────────┐
│ ▶ BVR              │
│ Project: my-app     │
│ Model:  mistral     │
│ Status: ready       │
└─────────────────────┘`))
	}
	return ""
}

func terminalPreviews() string {
	if previewEnv() != "" {
		return "inline"
	}
	return "placeholder"
}

func previewEnv() string {
	switch strings.ToLower(os.Getenv("TERM_PROGRAM")) {
	case "ghostty", "kitty", "wezterm", "alacritty", "iterm2":
		return strings.ToLower(os.Getenv("TERM_PROGRAM"))
	}
	return ""
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

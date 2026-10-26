package model

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/dialog"
	fimage "github.com/richavery/bvr-cli/internal/ui/image"
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
	hasRg     bool

	// Image rendering for the tour screenshots. caps is injected by the UI
	// because the terminal's graphics support is discovered by probing, not by
	// reading the environment.
	caps     *common.Capabilities
	shot     tourShot
	shotEnc  fimage.Encoding
	shotCols int
	shotRows int

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
		o.sendShot(),
	)
}

func (o *onboardingModel) checkDependencies() tea.Cmd {
	return func() tea.Msg {
		gitPath, gitErr := lookPath("git")
		ollamaPath, ollamaErr := lookPath("ollama")
		rgPath, rgErr := lookPath("rg")
		return dependencyCheckMsg{
			hasGit:    gitErr == nil && gitPath != "",
			hasOllama: ollamaErr == nil && ollamaPath != "",
			hasRg:     rgErr == nil && rgPath != "",
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
	hasRg     bool
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
		o.hasRg = msg.hasRg
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
			// Each step has its own picture, so each step asks for it. The
			// decode and encode happen on the command goroutine.
			if cmd := o.sendShot(); cmd != nil {
				return dialog.ActionCmd{Cmd: cmd}
			}
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
		b.WriteString(o.renderShot(o.renderASCIIPreview("home")))
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
		b.WriteString(o.renderShot(""))
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
		b.WriteString(o.renderShot(""))
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
		b.WriteString(o.com.Styles.Initialize.Accent.Render("Press Enter to continue"))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("Press O or Esc to opt out and start using BVR now"))

	case onboardingStepThemes:
		b.WriteString(o.com.Styles.Initialize.Content.Render("Themes"))
		b.WriteString("\n\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("BVR ships eight themes; ctrl+shift+t switches them."))
		b.WriteString("\n")
		b.WriteString(o.com.Styles.Initialize.Content.Render("The tour picture below is the pink one."))
		b.WriteString("\n\n")
		b.WriteString(o.renderShot(""))
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
		if deps := o.renderDependencies(); deps != "" {
			b.WriteString("\n\n")
			b.WriteString(deps)
		}
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

// renderDependencies names the binaries onboarding could not find and how to
// install them. The agent is blocked from running package managers, so this is
// the only place a missing tool can be surfaced to the user.
func (o *onboardingModel) renderDependencies() string {
	var missing []string
	for _, dep := range []struct {
		pkg       string
		installed bool
		why       string
	}{
		{"git", o.hasGit, "project detection and diffs"},
		{"ripgrep", o.hasRg, "grep and glob search is much slower without it"},
		{"ollama", o.hasOllama, "local models"},
	} {
		if dep.installed {
			continue
		}
		line := "  " + dep.pkg + " - " + dep.why
		if cmd := installCommand(dep.pkg); cmd != "" {
			line += "\n    " + cmd
		}
		missing = append(missing, line)
	}
	if len(missing) == 0 {
		return ""
	}
	content := o.com.Styles.Initialize.Content
	return content.Render("Not detected:") + "\n" + content.Render(strings.Join(missing, "\n"))
}

// installCommand is the command that gets pkg onto this host, or "" when there
// is no single obvious one.
func installCommand(pkg string) string {
	return installCommandFor(runtime.GOOS, distroIDs(), pkg)
}

// installCommandFor is the mapping on its own so the Linux branches can be
// tested from any host.
func installCommandFor(goos string, ids []string, pkg string) string {
	switch goos {
	case "darwin":
		return "brew install " + pkg
	case "linux":
		switch {
		// Omarchy and the other Arch-derived setups match on ID_LIKE, which is
		// why the ids are a list rather than a single value.
		case slices.Contains(ids, "arch"):
			if pkg == "ollama" {
				// Arch keeps ollama in the AUR, where pacman has nothing to do.
				return "yay -S ollama"
			}
			return "sudo pacman -S " + pkg
		case slices.Contains(ids, "debian"):
			return "sudo apt install -y " + pkg
		case slices.Contains(ids, "fedora"), slices.Contains(ids, "rhel"):
			return "sudo dnf install " + pkg
		}
	}
	return ""
}

// distroIDs caches /etc/os-release: the onboarding view re-renders per keypress.
var distroIDs = sync.OnceValue(func() []string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil
	}
	return osReleaseIDs(data)
})

// osReleaseIDs collects ID and ID_LIKE from an /etc/os-release body.
func osReleaseIDs(data []byte) []string {
	var ids []string
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || (key != "ID" && key != "ID_LIKE") {
			continue
		}
		ids = append(ids, strings.Fields(strings.Trim(value, `"`))...)
	}
	return ids
}

// lookPath resolves an executable using $PATH (plus a few well-known
// install dirs), so onboarding checks work on Arch/Omarchy installs where
// tools live outside the macOS Homebrew layout.
func lookPath(name string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	home, _ := os.UserHomeDir()
	for _, dir := range []string{
		"/usr/local/bin", "/usr/bin", "/bin",
		"/opt/homebrew/bin", "/opt/homebrew/sbin",
		"/home/linuxbrew/.linuxbrew/bin",
		filepath.Join(home, ".local", "bin"),
	} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("%s not found", name)
}

// renderASCIIPreview is the Home step's stand-in on terminals that cannot paint
// pixels, so the step never renders as an empty gap.
func (o *onboardingModel) renderASCIIPreview(name string) string {
	switch name {
	case "home":
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

package model

import (
	"strings"
	"testing"

	"github.com/richavery/bvr-cli/internal/ui/common"
	uistyles "github.com/richavery/bvr-cli/internal/ui/styles"
	"github.com/stretchr/testify/require"
)

func TestInstallCommandFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		goos string
		ids  []string
		pkg  string
		want string
	}{
		{"macos", "darwin", nil, "ripgrep", "brew install ripgrep"},
		{"arch", "linux", []string{"arch"}, "ripgrep", "sudo pacman -S ripgrep"},
		// Omarchy sets ID=omarchy and only claims Arch through ID_LIKE, so the
		// Arch suggestion has to survive that indirection.
		{"omarchy", "linux", []string{"omarchy", "arch"}, "ripgrep", "sudo pacman -S ripgrep"},
		{"arch ollama is AUR only", "linux", []string{"arch"}, "ollama", "yay -S ollama"},
		{"debian", "linux", []string{"ubuntu", "debian"}, "ripgrep", "sudo apt install -y ripgrep"},
		{"fedora", "linux", []string{"fedora"}, "ripgrep", "sudo dnf install ripgrep"},
		{"unknown distro", "linux", []string{"void"}, "ripgrep", ""},
		{"unknown os", "windows", nil, "ripgrep", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, installCommandFor(tc.goos, tc.ids, tc.pkg))
		})
	}
}

func TestOSReleaseIDs(t *testing.T) {
	t.Parallel()

	ids := osReleaseIDs([]byte(strings.Join([]string{
		`NAME="Omarchy"`,
		`ID=omarchy`,
		`ID_LIKE="arch Linux"`,
		`BUILD_ID="2026.10"`,
		`HOME_URL="https://omarchy.org/"`,
	}, "\n")))
	require.Equal(t, []string{"omarchy", "arch", "Linux"}, ids)

	require.Empty(t, osReleaseIDs([]byte("PRETTY_NAME=\"no ids here\"\n")))
}

func TestRenderDependenciesListsOnlyWhatIsMissing(t *testing.T) {
	t.Parallel()

	st := uistyles.CharmtonePantera()
	o := &onboardingModel{
		com:       &common.Common{Styles: &st},
		hasGit:    true,
		hasOllama: true,
	}

	view := o.renderDependencies()
	require.Contains(t, view, "ripgrep")
	require.NotContains(t, view, "git")

	o.hasRg = true
	require.Empty(t, o.renderDependencies())
}

// TestCheckDependenciesReportsRg: the onboarding hint is only as good as the
// probe behind it, and the probe must keep resolving rg off $PATH.
func TestCheckDependenciesReportsRg(t *testing.T) {
	t.Parallel()

	msg, ok := newOnboardingModel(&common.Common{Styles: &uistyles.Styles{}}).checkDependencies()().(dependencyCheckMsg)
	require.True(t, ok)
	require.Equal(t, lookPathOK("rg"), msg.hasRg)
	require.Equal(t, lookPathOK("git"), msg.hasGit)
}

func lookPathOK(name string) bool {
	path, err := lookPath(name)
	return err == nil && path != ""
}

// TestUpdateRoutesOnboardingProbeResults: the dependency probe runs in a
// command, and only key presses were routed to the onboarding dialog, so every
// result was dropped and onboarding claimed all three tools were missing.
func TestUpdateRoutesOnboardingProbeResults(t *testing.T) {
	m := newBusyUI(&countingWorkspace{ready: true})
	o := newOnboardingModel(m.com)
	m.onboardingDialog = o
	m.setState(uiOnboarding, uiFocusEditor)

	_, cmd := m.Update(dependencyCheckMsg{hasGit: true, hasRg: true})
	runCmds(m, cmd)

	require.True(t, o.hasRg, "the probe result must reach the onboarding dialog")
	require.True(t, o.hasGit)
	require.False(t, o.hasOllama)
}

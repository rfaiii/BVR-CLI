package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

const DefaultTheme = "rich-aizen-green"

type ThemeDefinition struct {
	ID, Name                   string
	Primary, Secondary         color.Color
	Gradient                   color.Color
	Surface                    color.Color
	SurfaceSubtle              color.Color
	SurfaceMuted               color.Color
	OnSurface                  color.Color
	Muted                      color.Color
	Subtle                     color.Color
	Border                     color.Color
	StatusSuccess, StatusError color.Color
	StatusWarning, StatusInfo  color.Color
	CodeBackground             color.Color
}

func (t ThemeDefinition) Palette() quickStyleOpts {
	return quickStyleOpts{
		primary:   t.Primary,
		secondary: t.Secondary,
		accent:    t.OnSurface,
		keyword:   t.Primary,

		fgBase:       t.OnSurface,
		fgMoreSubtle: t.Muted,
		fgSubtle:     t.Subtle,
		fgMostSubtle: t.SurfaceMuted,

		onPrimary: t.Surface,

		bgBase:         t.Surface,
		bgLeastVisible: t.SurfaceSubtle,
		bgLessVisible:  t.SurfaceMuted,
		bgMostVisible:  t.Border,

		separator: t.Border,

		destructive:       t.StatusError,
		error:             t.StatusError,
		warningSubtle:     t.StatusWarning,
		warning:           t.StatusWarning,
		denied:            t.StatusWarning,
		busy:              t.StatusSuccess,
		info:              t.StatusInfo,
		infoMoreSubtle:    t.StatusInfo,
		infoMostSubtle:    t.StatusInfo,
		success:           t.StatusSuccess,
		successMoreSubtle: t.StatusSuccess,
		successMostSubtle: t.StatusSuccess,

		ansiBlack:   t.SurfaceSubtle,
		ansiRed:     t.StatusError,
		ansiGreen:   t.StatusSuccess,
		ansiYellow:  t.StatusWarning,
		ansiBlue:    t.StatusInfo,
		ansiMagenta: t.Primary,
		ansiCyan:    t.StatusInfo,
		ansiWhite:   t.OnSurface,

		ansiBrightBlack:   t.SurfaceMuted,
		ansiBrightRed:     t.StatusError,
		ansiBrightGreen:   t.StatusSuccess,
		ansiBrightYellow:  t.StatusWarning,
		ansiBrightBlue:    t.StatusInfo,
		ansiBrightMagenta: t.Primary,
		ansiBrightCyan:    t.StatusInfo,
		ansiBrightWhite:   t.OnSurface,
	}
}

var themeDefinitions = []ThemeDefinition{
	{
		DefaultTheme,
		"Rich Aizen Green",
		lipgloss.Color("#3BF66B"), lipgloss.Color("#6BFF91"), lipgloss.Color("#3BF66B"),
		lipgloss.Color("#0C0E0D"), lipgloss.Color("#141716"), lipgloss.Color("#222825"),
		lipgloss.Color("#FFFFFF"), lipgloss.Color("#8C9691"), lipgloss.Color("#59645E"),
		lipgloss.Color("#222825"), lipgloss.Color("#3BF66B"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#F2C14E"), lipgloss.Color("#8CDED0"), lipgloss.Color("#141716"),
	},
	{
		"crazy-jeff-pink",
		"Crazy Jeff Pink",
		lipgloss.Color("#FF4FA3"), lipgloss.Color("#FF86C8"), lipgloss.Color("#FF4FA3"),
		lipgloss.Color("#14060C"), lipgloss.Color("#1E0814"), lipgloss.Color("#2E1020"),
		lipgloss.Color("#FFF0F5"), lipgloss.Color("#C27A8C"), lipgloss.Color("#7A4050"),
		lipgloss.Color("#2E1020"), lipgloss.Color("#FF4FA3"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFB86C"), lipgloss.Color("#FF86C8"), lipgloss.Color("#1E0814"),
	},
	{
		"kobe-yang-purple",
		"Kobe Yang Purple",
		lipgloss.Color("#B56CFF"), lipgloss.Color("#E0B3FF"), lipgloss.Color("#B56CFF"),
		lipgloss.Color("#0E0816"), lipgloss.Color("#18101F"), lipgloss.Color("#261633"),
		lipgloss.Color("#F5F3FF"), lipgloss.Color("#A090B8"), lipgloss.Color("#5E4F73"),
		lipgloss.Color("#261633"), lipgloss.Color("#B56CFF"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFD866"), lipgloss.Color("#C4A8FF"), lipgloss.Color("#18101F"),
	},
	{
		"steve-dabeav-blue",
		"Steve DaBeav Blue",
		lipgloss.Color("#5CC8FF"), lipgloss.Color("#A9E7FF"), lipgloss.Color("#5CC8FF"),
		lipgloss.Color("#081218"), lipgloss.Color("#0F1C24"), lipgloss.Color("#162C38"),
		lipgloss.Color("#F0F9FF"), lipgloss.Color("#82A8BF"), lipgloss.Color("#4C6878"),
		lipgloss.Color("#162C38"), lipgloss.Color("#5CC8FF"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFD866"), lipgloss.Color("#A9E7FF"), lipgloss.Color("#0F1C24"),
	},
	{
		"jenny-ann-orange",
		"Jenny Ann Orange",
		lipgloss.Color("#FF8A3D"), lipgloss.Color("#FFC078"), lipgloss.Color("#FF8A3D"),
		lipgloss.Color("#16100A"), lipgloss.Color("#241C12"), lipgloss.Color("#3A2A18"),
		lipgloss.Color("#FFF7ED"), lipgloss.Color("#B08C6E"), lipgloss.Color("#7C5E40"),
		lipgloss.Color("#3A2A18"), lipgloss.Color("#FF8A3D"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFD866"), lipgloss.Color("#FFC078"), lipgloss.Color("#241C12"),
	},
	{
		"felix-tornado-white",
		"Felix Tornado White",
		lipgloss.Color("#FFFFFF"), lipgloss.Color("#E8EEF2"), lipgloss.Color("#FFFFFF"),
		lipgloss.Color("#111316"), lipgloss.Color("#1C1F24"), lipgloss.Color("#2C3038"),
		lipgloss.Color("#F8FAFC"), lipgloss.Color("#C7CDD4"), lipgloss.Color("#8A929C"),
		lipgloss.Color("#2C3038"), lipgloss.Color("#E8EEF2"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#F2C14E"), lipgloss.Color("#A9E7FF"), lipgloss.Color("#1C1F24"),
	},
	{
		"luis-mellow-yellow",
		"Luis Mellow Yellow",
		lipgloss.Color("#D6C84A"), lipgloss.Color("#F2E98B"), lipgloss.Color("#D6C84A"),
		lipgloss.Color("#14120A"), lipgloss.Color("#1F1D12"), lipgloss.Color("#302B18"),
		lipgloss.Color("#FFFDF0"), lipgloss.Color("#B5AD70"), lipgloss.Color("#6E6940"),
		lipgloss.Color("#302B18"), lipgloss.Color("#D6C84A"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFD866"), lipgloss.Color("#F2E98B"), lipgloss.Color("#1F1D12"),
	},
	{
		"bobur-blood-red",
		"Bobur Blood Red",
		lipgloss.Color("#FF1F1F"), lipgloss.Color("#FF6B6B"), lipgloss.Color("#FF1F1F"),
		lipgloss.Color("#140606"), lipgloss.Color("#1E0A0A"), lipgloss.Color("#2E1414"),
		lipgloss.Color("#FFF0F0"), lipgloss.Color("#C28A8A"), lipgloss.Color("#7A4A4A"),
		lipgloss.Color("#2E1414"), lipgloss.Color("#FF1F1F"), lipgloss.Color("#FF5F56"),
		lipgloss.Color("#FFB86C"), lipgloss.Color("#FF8A8A"), lipgloss.Color("#1E0A0A"),
	},
}

func Themes() []ThemeDefinition { return append([]ThemeDefinition(nil), themeDefinitions...) }

// themeAccentAlt returns the (Accent, Alt) color pair for a theme, per the
// BVR color-mapping table:
//
//	Primary Theme | Accent | Alt
//	Green         | Pink   | Purple
//	Pink          | Purple | Blue
//	Purple        | Blue   | Orange
//	Blue          | Orange | White
//	Orange        | White  | Yellow
//	White         | Yellow | Red
//	Yellow        | Red    | Green
//	Red           | Green  | Pink
//
// The values intentionally match the theme primaries so highlights stay
// on-brand across every theme.
func themeAccentAlt(id string) (color.Color, color.Color) {
	green, pink, purple := lipgloss.Color("#3BF66B"), lipgloss.Color("#FF4FA3"), lipgloss.Color("#B56CFF")
	blue, orange, white := lipgloss.Color("#5CC8FF"), lipgloss.Color("#FF8A3D"), lipgloss.Color("#FFFFFF")
	yellow, red := lipgloss.Color("#D6C84A"), lipgloss.Color("#FF1F1F")
	switch id {
	case "crazy-jeff-pink":
		return purple, blue
	case "kobe-yang-purple":
		return blue, orange
	case "steve-dabeav-blue":
		return orange, white
	case "jenny-ann-orange":
		return white, yellow
	case "felix-tornado-white":
		return yellow, red
	case "luis-mellow-yellow":
		return red, green
	case "bobur-blood-red":
		return green, pink
	default: // rich-aizen-green
		return pink, purple
	}
}

func ThemeByID(id string) ThemeDefinition {
	for _, theme := range themeDefinitions {
		if theme.ID == id {
			return theme
		}
	}
	return themeDefinitions[0]
}
func ThemeForName(id string) Styles {
	theme := ThemeByID(id)
	s := quickStyle(theme.Palette())
	return applyTheme(s, theme)
}

func applyTheme(s Styles, theme ThemeDefinition) Styles {
	p := theme.Palette()
	primary, secondary, accent := theme.Primary, theme.Secondary, p.accent
	onPrimary := p.onPrimary
	borderColor := theme.Border
	selectedFill := primary
	selectedText := onPrimary
	if !compat.HasDarkBackground {
		selectedFill = onPrimary
		selectedText = primary
	}

	// Store the derived accent/Alt colors for the active theme so the rest of
	// the UI (boot banner, buttons, metadata, separators, help text) can use
	// them consistently. "accent" stays meaningful for selected-state fills;
	// ThemeColor.Accent is the highlight color and ThemeColor.Alt the
	// alternate (metadata) color per the mapping table.
	s.ThemeColor.Accent, s.ThemeColor.Alt = themeAccentAlt(theme.ID)

	s.Header.LogoGradFromColor, s.Header.LogoGradToColor = secondary, primary
	s.Header.HypercreditIcon = s.Header.HypercreditIcon.Foreground(secondary)
	s.Header.Percentage = s.Header.Percentage.Foreground(primary)

	s.Editor.PromptNormalFocused = s.Editor.PromptNormalFocused.Foreground(primary)
	s.Editor.PromptBeastmodeIconFocused = s.Editor.PromptBeastmodeIconFocused.Foreground(primary)
	s.Editor.PromptBangIconFocused = s.Editor.PromptBangIconFocused.Foreground(primary).
		Background(primary)

	s.ModelInfo.HypercreditIcon = s.ModelInfo.HypercreditIcon.Foreground(secondary)

	s.Dialog.Title = s.Dialog.Title.Foreground(primary)
	s.Dialog.TitleText = s.Dialog.TitleText.Foreground(primary)
	s.Dialog.TitleGradFromColor = primary
	s.Dialog.TitleGradToColor = secondary
	s.Dialog.SelectedItem = s.Dialog.SelectedItem.
		BorderForegroundBlend(primary, secondary).
		BorderForegroundBlendOffset(1).
		Background(selectedFill).
		Foreground(selectedText)
	s.Dialog.View = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(primary, secondary).
		BorderForegroundBlendOffset(1).
		Foreground(p.fgBase).
		Background(p.bgLessVisible)
	s.Dialog.Quit.Frame = s.Dialog.Quit.Frame.
		BorderForegroundBlend(primary, secondary).
		BorderForegroundBlendOffset(1)
	s.Dialog.PrimaryText = s.Dialog.PrimaryText.Foreground(primary)
	s.Dialog.ContentPanel = s.Dialog.ContentPanel.
		BorderForegroundBlend(primary, secondary).
		BorderForegroundBlendOffset(1).
		Background(p.bgLessVisible)

	s.Messages.UserFocused = s.Messages.UserFocused.BorderForeground(primary)
	s.Messages.AssistantFocused = s.Messages.AssistantFocused.BorderForeground(secondary)
	s.Messages.ShellBarFocused = s.Messages.ShellBarFocused.BorderForeground(primary)
	s.Messages.ShellPrompt = s.Messages.ShellPrompt.Foreground(primary)
	s.Messages.ToolCallFocused = s.Messages.ToolCallFocused.BorderForeground(secondary)

	s.Status.ResourceFilled = lipgloss.NewStyle().Foreground(s.ThemeColor.Accent)
	s.Status.ResourceEmpty = lipgloss.NewStyle().Foreground(borderColor)
	s.Status.ResourceLabel = s.Status.ResourceLabel.Foreground(p.fgMoreSubtle)
	s.Status.ResourceValue = s.Status.ResourceValue.Foreground(s.ThemeColor.Accent)
	s.Status.SuccessIndicator = s.Status.SuccessIndicator.Background(p.success)
	s.Status.WarnIndicator = s.Status.WarnIndicator.Background(p.warning)
	s.Status.ErrorIndicator = s.Status.ErrorIndicator.Background(p.error)
	s.Status.WarnMessage = s.Status.WarnMessage.Background(p.warningSubtle)

	s.Pills.Focused = s.Pills.Focused.
		BorderForegroundBlend(primary, secondary).
		BorderForegroundBlendOffset(1)
	s.Pills.QueueGradFromColor = primary
	s.Pills.QueueGradToColor = secondary

	s.Logo.TitleColorA, s.Logo.TitleColorB = secondary, primary
	s.Logo.FieldColor = primary

	s.WorkingGradFromColor = primary
	s.WorkingGradToColor = secondary
	s.WorkingLabelColor = p.fgMostSubtle
	s.WorkingTimerColor = p.fgMostSubtle

	s.Dialog.Sessions.DeletingTitleGradientFromColor = p.error
	s.Dialog.Sessions.DeletingTitleGradientToColor = primary
	s.Dialog.Sessions.RenamingTitleGradientFromColor = p.warningSubtle
	s.Dialog.Sessions.RenamingTitleGradientToColor = accent

	s.CompactDetails.View = s.CompactDetails.View.BorderForegroundBlend(primary, secondary).BorderForegroundBlendOffset(1)
	s.Completions.Focused = s.Completions.Focused.Background(primary).Foreground(onPrimary)
	s.Completions.Match = s.Completions.Match.Underline(true).Foreground(secondary)
	s.Attachments.Image = s.Attachments.Image.Background(p.success)
	s.Attachments.Text = s.Attachments.Text.Background(p.info)
	s.Attachments.Skill = s.Attachments.Skill.Background(p.primary)
	s.Attachments.Remove = s.Attachments.Remove.Background(p.bgLessVisible)
	s.Attachments.Deleting = s.Attachments.Deleting.Background(p.destructive)
	s.Pills.Base = s.Pills.Base.Foreground(p.fgBase)
	s.Pills.TodoLabel = s.Pills.TodoLabel.Foreground(p.fgBase)
	s.Pills.TodoProgress = s.Pills.TodoProgress.Foreground(primary)
	s.Pills.TodoSpinner = s.Pills.TodoSpinner.Foreground(secondary)

	// Lighten the section heading / help text (LOCAL DEVICE, SKILLS, MCP, etc.)
	// so it reads against the dark surface. These previously used the darkest
	// muted tone; bumping one step brighter keeps it subtle but legible.
	s.Resource.Heading = s.Resource.Heading.Foreground(p.fgMoreSubtle)
	s.Resource.AdditionalText = s.Resource.AdditionalText.Foreground(p.fgMoreSubtle)

	// Command list metadata (shortcut column) uses the ALT COLOR when focused
	// so command buttons read as ACCENT-highlighted with ALT metadata/help.
	s.Dialog.ListItem.InfoFocused = s.Dialog.ListItem.InfoFocused.Foreground(s.ThemeColor.Alt)

	// File Finder uses the theme's ALT COLOR for metadata, the rule
	// separators, and the close button (instead of the default red/ muted
	// tones) so it stays on-brand per the theme color-mapping table.
	s.Dialog.FileBrowser.Close = s.Dialog.FileBrowser.Close.Foreground(s.ThemeColor.Alt)
	s.Dialog.FileBrowser.Rule = s.Dialog.FileBrowser.Rule.Foreground(s.ThemeColor.Alt)
	s.Dialog.FileBrowser.Preview = s.Dialog.FileBrowser.Preview.Foreground(s.ThemeColor.Alt)

	return s
}

// ThemeKeyForProvider returns a stable identifier for the theme
// associated with the given provider ID. Providers that share a theme
// yield the same key, so callers can cheaply detect when switching
// providers would not actually change the active theme and skip the
// expensive style rebuild. This is the single source of truth for the
// provider-to-theme mapping; [ThemeForProvider] builds on it.
func ThemeKeyForProvider(providerID string) string {
	switch providerID {
	case "hyper":
		return "dark-bvr-theme"
	default:
		return "dark-bvr-theme"
	}
}

// ThemeForProvider returns the Dark BVR theme for every provider. Provider
// identity is retained for compatibility, but BVR has one visual identity.
func ThemeForProvider(providerID string) Styles {
	return DarkBvrTheme()
}

// DarkBvrTheme returns BVR's dark green theme. The palette is shared with
// the cross-language references in R&D/BVR-DARK-COLOR.
func DarkBvrTheme() Styles {
	theme := ThemeByID(DefaultTheme)
	s := quickStyle(theme.Palette())
	return applyTheme(s, theme)
}

// CharmtonePantera is retained as a source-compatible alias for existing UI
// tests and callers. New code should use DarkBvrTheme.
func CharmtonePantera() Styles { return DarkBvrTheme() }

// HyperbvrObsidiana returns the Hyperbvr dark theme.
func HyperbvrObsidiana() Styles {
	return DarkBvrTheme()
}

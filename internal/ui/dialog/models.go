package dialog

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/localmodel"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/util"
)

// ModelType represents the type of model to select.
type ModelType int

const (
	ModelTypeLarge ModelType = iota
	ModelTypeSmall
)

// String returns the string representation of the [ModelType].
func (mt ModelType) String() string {
	switch mt {
	case ModelTypeLarge:
		return "Large Task"
	case ModelTypeSmall:
		return "Small Task"
	default:
		return "Unknown"
	}
}

// Config returns the corresponding config model type.
func (mt ModelType) Config() config.SelectedModelType {
	switch mt {
	case ModelTypeLarge:
		return config.SelectedModelTypeLarge
	case ModelTypeSmall:
		return config.SelectedModelTypeSmall
	default:
		return ""
	}
}

// Placeholder returns the input placeholder for the model type.
func (mt ModelType) Placeholder() string {
	switch mt {
	case ModelTypeLarge:
		return largeModelInputPlaceholder
	case ModelTypeSmall:
		return smallModelInputPlaceholder
	default:
		return ""
	}
}

const (
	onboardingModelInputPlaceholder = "Find your fave"
	largeModelInputPlaceholder      = "Choose a model for large, complex tasks"
	smallModelInputPlaceholder      = "Choose a model for small, simple tasks"
)

// ModelsID is the identifier for the model selection dialog.
const ModelsID = "models"

const defaultModelsDialogMaxWidth = 73

// Models represents a model selection dialog.
type Models struct {
	com          *common.Common
	isOnboarding bool

	modelType   ModelType
	providers   []catwalk.Provider
	localModels []localmodel.Model

	keyMap struct {
		Tab      key.Binding
		UpDown   key.Binding
		Select   key.Binding
		Edit     key.Binding
		Next     key.Binding
		Previous key.Binding
		Close    key.Binding
		Refresh  key.Binding
		Pull     key.Binding
		Start    key.Binding
	}
	list       *ModelsList
	input      textinput.Model
	help       help.Model
	pulling    bool
	pullStatus localmodel.PullProgress
	pullCancel context.CancelFunc
}

var _ Dialog = (*Models)(nil)

// NewModels creates a new Models dialog.
func NewModels(com *common.Common, isOnboarding bool) (*Models, error) {
	t := com.Styles
	m := &Models{}
	m.com = com
	m.isOnboarding = isOnboarding

	help := help.New()
	help.Styles = t.DialogHelpStyles()

	m.help = help
	m.list = NewModelsList(t)
	m.list.Focus()
	m.list.SetSelected(0)

	m.input = textinput.New()
	m.input.SetVirtualCursor(false)
	m.input.Placeholder = onboardingModelInputPlaceholder
	m.input.SetStyles(com.Styles.TextInput)
	m.input.Focus()

	m.keyMap.Tab = key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "toggle type"),
	)
	m.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "confirm"),
	)
	m.keyMap.Edit = key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit"),
	)
	m.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "choose"),
	)
	m.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	m.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	m.keyMap.Close = CloseKey
	m.keyMap.Refresh = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh local models"))
	m.keyMap.Pull = key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pull selected model"))
	m.keyMap.Start = key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start Ollama"))

	var err error
	m.providers, err = config.Providers(m.com.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	if err := m.setProviderItems(); err != nil {
		return nil, fmt.Errorf("failed to set provider items: %w", err)
	}

	return m, nil
}

// ID implements Dialog.
func (m *Models) ID() string {
	return ModelsID
}

// HandleMsg implements Dialog.
func (m *Models) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case LocalModelsMsg:
		m.localModels = append(msg.Models, msg.Additional...)
		if err := m.setProviderItems(); err != nil {
			return util.ReportError(err)
		}
		return nil
	case OllamaPullMsg:
		m.pullStatus = msg.Progress
		if msg.Done {
			m.pulling = false
			if m.pullCancel != nil {
				m.pullCancel()
				m.pullCancel = nil
			}
			if msg.Err != nil {
				return ActionCmd{Cmd: util.ReportError(msg.Err)}
			}
			return ActionCmd{Cmd: tea.Batch(m.LocalModelsCmd(), util.ReportInfo("Ollama model pulled: "+msg.Model))}
		}
		return ActionCmd{Cmd: msg.Next}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.Close):
			if m.isOnboarding {
				return ActionSkipOnboarding{}
			}
			return ActionClose{}
		case key.Matches(msg, m.keyMap.Refresh):
			return ActionCmd{Cmd: m.LocalModelsCmd()}
		case key.Matches(msg, m.keyMap.Start):
			return ActionCmd{Cmd: startOllamaCmd()}
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			if m.pulling && m.pullCancel != nil {
				m.pullCancel()
				m.pullCancel = nil
				m.pulling = false
				return ActionCmd{Cmd: util.ReportInfo("Ollama pull cancelled")}
			}
		case key.Matches(msg, m.keyMap.Pull):
			if item, ok := m.list.SelectedItem().(*ModelItem); ok && item.prov.ID == "ollama-local" {
				if m.pulling {
					return nil
				}
				ctx, cancel := context.WithCancel(context.Background())
				m.pullCancel, m.pulling, m.pullStatus = cancel, true, localmodel.PullProgress{Status: "starting"}
				return ActionCmd{Cmd: pullOllamaCmd(ctx, item.model.ID)}
			}
		case key.Matches(msg, m.keyMap.Previous):
			m.list.Focus()
			if m.list.IsSelectedFirst() {
				m.list.SelectLast()
			} else {
				m.list.SelectPrev()
			}
			m.list.ScrollToSelected()
		case key.Matches(msg, m.keyMap.Next):
			m.list.Focus()
			if m.list.IsSelectedLast() {
				m.list.SelectFirst()
			} else {
				m.list.SelectNext()
			}
			m.list.ScrollToSelected()
		case key.Matches(msg, m.keyMap.Select, m.keyMap.Edit):
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				break
			}

			modelItem, ok := selectedItem.(*ModelItem)
			if !ok {
				break
			}

			isEdit := key.Matches(msg, m.keyMap.Edit)

			return ActionSelectModel{
				Provider:       modelItem.prov,
				Model:          modelItem.SelectedModel(),
				ModelType:      modelItem.SelectedModelType(),
				ReAuthenticate: isEdit,
				LocalModel:     modelItem.prov.ID == "ollama-local",
			}
		case key.Matches(msg, m.keyMap.Tab):
			if m.isOnboarding {
				break
			}
			if m.modelType == ModelTypeLarge {
				m.modelType = ModelTypeSmall
			} else {
				m.modelType = ModelTypeLarge
			}
			if err := m.setProviderItems(); err != nil {
				return util.ReportError(err)
			}
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			value := m.input.Value()
			m.list.Focus()
			m.list.SetFilter(value)
			m.list.SelectFirst()
			m.list.ScrollToTop()
			return ActionCmd{cmd}
		}
	}
	return nil
}

type LocalModelsMsg struct {
	Models     []localmodel.Model
	Additional []localmodel.Model
	Status     localmodel.RuntimeStatus
	Err        error
}

type OllamaPullMsg struct {
	Model    string
	Progress localmodel.PullProgress
	Done     bool
	Err      error
	Next     tea.Cmd
}

func (m *Models) LocalModelsCmd() tea.Cmd {
	return func() tea.Msg {
		runtime := localmodel.NewOllama("")
		status := runtime.Status(context.Background())
		if status.Status != localmodel.StatusOnline {
			if status.Installed {
				if startErr := runtime.Start(context.Background()); startErr != nil {
					return LocalModelsMsg{Status: status, Err: startErr}
				}
				status = runtime.Status(context.Background())
			}
			if status.Status != localmodel.StatusOnline {
				return LocalModelsMsg{Status: status, Err: status.Error}
			}
		}
		models, err := runtime.Models(context.Background())
		if err != nil {
			return LocalModelsMsg{Status: status, Err: err}
		}
		additional := make([]localmodel.Model, 0)
		for _, compatible := range []*localmodel.CompatibleRuntime{localmodel.NewLMStudio(""), localmodel.NewLlamaCPP("")} {
			if discovered, discoverErr := compatible.ListModels(context.Background()); discoverErr == nil {
				additional = append(additional, discovered...)
			}
		}
		return LocalModelsMsg{Models: models, Additional: additional, Status: status}
	}
}

// Cursor returns the cursor for the dialog.
func (m *Models) Cursor() *tea.Cursor {
	return InputCursor(m.com.Styles, m.input.Cursor())
}

// modelTypeRadioView returns the radio view for model type selection.
func (m *Models) modelTypeRadioView() string {
	t := m.com.Styles
	textStyle := t.Radio.Label
	largeRadioStyle := t.Radio.Off
	smallRadioStyle := t.Radio.Off
	if m.modelType == ModelTypeLarge {
		largeRadioStyle = t.Radio.On
	} else {
		smallRadioStyle = t.Radio.On
	}

	largeRadio := largeRadioStyle.Padding(0, 1).Render()
	smallRadio := smallRadioStyle.Padding(0, 1).Render()

	return fmt.Sprintf("%s%s  %s%s",
		largeRadio, textStyle.Render(ModelTypeLarge.String()),
		smallRadio, textStyle.Render(ModelTypeSmall.String()))
}

// Draw implements [Dialog].
func (m *Models) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := m.com.Styles
	width := max(0, min(defaultModelsDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	m.input.SetWidth(dialogInputTextWidth(t, m.input, innerWidth))

	listHeight, listTotalHeight, _ := sizeDialogList(t, m.list, innerWidth, height)

	rc := NewRenderContext(t, width)
	rc.Title = "Switch Model"
	rc.TitleInfo = m.modelTypeRadioView()
	if m.pulling {
		progress := "Pulling: " + m.pullStatus.Status
		if m.pullStatus.Total > 0 {
			progress += " " + strconv.Itoa(int(m.pullStatus.Completed*100/m.pullStatus.Total)) + "%"
		}
		if m.pullStatus.Digest != "" {
			progress += " " + ansi.Truncate(m.pullStatus.Digest, 18, "…")
		}
		rc.AddPart(t.Dialog.PrimaryText.Render(progress + "  (c cancel)"))
	}

	if m.isOnboarding {
		titleText := t.Dialog.PrimaryText.Render("To start, let's choose a provider and model.")
		skipHint := t.Dialog.SecondaryText.Render("or press esc to skip and explore later.")
		rc.AddPart(titleText)
		rc.AddPart(skipHint)
	}

	inputView := t.Dialog.InputPrompt.Render(m.input.View())
	rc.AddPart(inputView)

	listView := t.Dialog.List.Height(m.list.Height()).Render(m.list.Render())
	listView = joinScrollbar(t, listView, listHeight, listTotalHeight, listHeight, m.list.Offset())
	rc.AddPart(listView)

	rc.Help = renderDialogHelp(t, &m.help, m, innerWidth)

	cur := m.Cursor()

	if m.isOnboarding {
		rc.Title = ""
		rc.TitleInfo = ""
		rc.IsOnboarding = true
		view := rc.Render()
		cur = adjustOnboardingInputCursor(t, cur)
		DrawOnboardingCursor(scr, area, view, cur)
	} else {
		view := rc.Render()
		DrawCenterCursor(scr, area, view, cur)
	}
	return cur
}

// ShortHelp returns the short help view.
func (m *Models) ShortHelp() []key.Binding {
	if m.isOnboarding {
		return []key.Binding{
			m.keyMap.UpDown,
			m.keyMap.Select,
			m.keyMap.Refresh,
			m.keyMap.Start,
		}
	}
	h := []key.Binding{
		m.keyMap.UpDown,
		m.keyMap.Tab,
		m.keyMap.Select,
	}
	if m.isSelectedConfigured() {
		h = append(h, m.keyMap.Edit)
	}
	h = append(h, m.keyMap.Close)
	h = append(h, m.keyMap.Refresh, m.keyMap.Pull, m.keyMap.Start)
	if m.pulling {
		h = append(h, key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel pull")))
	}
	return h
}

func startOllamaCmd() tea.Cmd {
	return func() tea.Msg {
		runtime := localmodel.NewOllama("")
		if err := runtime.Start(context.Background()); err != nil {
			return util.NewErrorMsg(err)
		}
		return util.NewInfoMsg("Ollama started")
	}
}
func pullOllamaCmd(ctx context.Context, name string) tea.Cmd {
	progressCh := make(chan localmodel.PullProgress, 32)
	doneCh := make(chan error, 1)
	go func() {
		err := localmodel.NewOllama("").Pull(ctx, name, func(progress localmodel.PullProgress) {
			select {
			case progressCh <- progress:
			case <-ctx.Done():
			}
		})
		doneCh <- err
		close(progressCh)
	}()
	var read tea.Cmd
	read = func() tea.Msg {
		select {
		case progress, ok := <-progressCh:
			if !ok {
				return OllamaPullMsg{Model: name, Done: true, Err: <-doneCh}
			}
			return OllamaPullMsg{Model: name, Progress: progress, Next: read}
		case <-ctx.Done():
			return OllamaPullMsg{Model: name, Done: true, Err: context.Canceled}
		}
	}
	return read
}

// FullHelp returns the full help view.
func (m *Models) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m *Models) isSelectedConfigured() bool {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return false
	}
	modelItem, ok := selectedItem.(*ModelItem)
	if !ok {
		return false
	}
	providerID := string(modelItem.prov.ID)
	_, isConfigured := m.com.Config().Providers.Get(providerID)
	return isConfigured
}

// setProviderItems sets the provider items in the list.
func (m *Models) setProviderItems() error {
	t := m.com.Styles
	cfg := m.com.Config()

	var selectedItemID string
	selectedType := m.modelType.Config()
	currentModel := cfg.Models[selectedType]
	recentItems := cfg.RecentModels[selectedType]

	// Track providers already added to avoid duplicates
	addedProviders := make(map[string]bool)

	// Get a list of known providers to compare against
	knownProviders, err := config.Providers(cfg)
	if err != nil {
		return fmt.Errorf("failed to get providers: %w", err)
	}

	containsProviderFunc := func(id string) func(p catwalk.Provider) bool {
		return func(p catwalk.Provider) bool {
			return p.ID == catwalk.InferenceProvider(id)
		}
	}

	// itemsMap contains the keys of added model items.
	itemsMap := make(map[string]*ModelItem)
	groups := []ModelGroup{}
	if len(m.localModels) > 0 {
		provider := catwalk.Provider{ID: "ollama-local", Name: "Ollama (Local)"}
		group := NewModelGroup(t, "Ollama (Local)", true)
		for _, local := range m.localModels {
			model := catwalk.Model{ID: local.Name, Name: local.DisplayName, ContextWindow: local.ContextWindow, DefaultMaxTokens: local.MaxTokens, SupportsImages: slices.Contains(local.Capabilities, "vision")}
			group.AppendItems(NewModelItem(t, provider, model, m.modelType, false, local.CodingCapable))
		}
		groups = append(groups, group)
	}
	for id, p := range cfg.Providers.Seq2() {
		if p.Disable {
			continue
		}

		// Check if this provider is not in the known providers list
		if !slices.ContainsFunc(knownProviders, containsProviderFunc(id)) ||
			!slices.ContainsFunc(m.providers, containsProviderFunc(id)) {
			provider := p.ToProvider()

			// Add this unknown provider to the list
			name := cmp.Or(p.Name, id)

			addedProviders[id] = true

			group := NewModelGroup(t, name, true)
			for _, model := range p.Models {
				item := NewModelItem(t, provider, model, m.modelType, false)
				group.AppendItems(item)
				itemsMap[item.ID()] = item
				if model.ID == currentModel.Model && string(provider.ID) == currentModel.Provider {
					selectedItemID = item.ID()
				}
			}
			if len(group.Items) > 0 {
				groups = append(groups, group)
			}
		}
	}

	// Now add known providers from the predefined list.
	// Providers already has Hyper at the front of the list.
	for _, provider := range m.providers {
		providerID := string(provider.ID)
		if addedProviders[providerID] {
			continue
		}

		providerConfig, providerConfigured := cfg.Providers.Get(providerID)
		if providerConfigured && providerConfig.Disable {
			continue
		}

		displayProvider := provider
		if providerConfigured {
			displayProvider.Name = cmp.Or(providerConfig.Name, displayProvider.Name)
			modelIndex := make(map[string]int, len(displayProvider.Models))
			for i, model := range displayProvider.Models {
				modelIndex[model.ID] = i
			}
			for _, model := range providerConfig.Models {
				if model.ID == "" {
					continue
				}
				if idx, ok := modelIndex[model.ID]; ok {
					if model.Name != "" {
						displayProvider.Models[idx].Name = model.Name
					}
					continue
				}
				model.Name = cmp.Or(model.Name, model.ID)
				displayProvider.Models = append(displayProvider.Models, model)
				modelIndex[model.ID] = len(displayProvider.Models) - 1
			}
		}

		name := cmp.Or(displayProvider.Name, providerID)

		group := NewModelGroup(t, name, providerConfigured)
		for _, model := range displayProvider.Models {
			item := NewModelItem(t, provider, model, m.modelType, false)
			group.AppendItems(item)
			itemsMap[item.ID()] = item
			if model.ID == currentModel.Model && string(provider.ID) == currentModel.Provider {
				selectedItemID = item.ID()
			}
		}

		groups = append(groups, group)
	}

	if len(recentItems) > 0 {
		recentGroup := NewModelGroup(t, "Recently used", false)

		var validRecentItems []config.SelectedModel
		for _, recent := range recentItems {
			key := modelKey(recent.Provider, recent.Model)
			item, ok := itemsMap[key]
			if !ok {
				continue
			}

			// Show provider for recent items
			item = NewModelItem(t, item.prov, item.model, m.modelType, true)
			item.showProvider = true

			validRecentItems = append(validRecentItems, recent)
			recentGroup.AppendItems(item)
			if recent.Model == currentModel.Model && recent.Provider == currentModel.Provider {
				selectedItemID = item.ID()
			}
		}

		if len(validRecentItems) != len(recentItems) {
			// FIXME: Does this need to be here? Is it mutating the config during a read?
			if err := m.com.Workspace.SetConfigField(config.ScopeGlobal, fmt.Sprintf("recent_models.%s", selectedType), validRecentItems); err != nil {
				return fmt.Errorf("failed to update recent models: %w", err)
			}
		}

		if len(recentGroup.Items) > 0 {
			groups = append([]ModelGroup{recentGroup}, groups...)
		}
	}

	// Set model groups in the list.
	m.list.SetGroups(groups...)
	m.list.SetSelectedItem(selectedItemID)
	if selectedItemID != "" {
		m.list.ScrollToSelected()
	} else {
		m.list.ScrollToTop()
	}

	// Update placeholder based on model type
	if !m.isOnboarding {
		m.input.Placeholder = m.modelType.Placeholder()
	}

	return nil
}

func modelKey(providerID, modelID string) string {
	if providerID == "" || modelID == "" {
		return ""
	}
	return providerID + ":" + modelID
}

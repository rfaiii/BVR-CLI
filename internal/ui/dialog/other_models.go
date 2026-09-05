package dialog

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/config"
	"github.com/richavery/bvr-cli/internal/ui/common"
)

// OtherModelsID is the identifier for the "Other Models" dialog, which lists
// models from extra/secondary catalogs (currently the Cline hosted gateway)
// that may not appear in the primary model picker.
const OtherModelsID = "other-models"

// OtherModels is a dialog that lists models from secondary provider catalogs
// (currently Cline) and emits an ActionSelectModel when one is chosen. If the
// provider is not yet configured, the standard API key authentication dialog
// is opened by the UI layer, mirroring the main models dialog flow.
type OtherModels struct {
	com      *common.Common
	prov     catwalk.Provider
	models   []catwalk.Model
	addKey   bool // show the "＋ ADD CLINE API KEY" row
	selected int
}

// NewOtherModels creates the "Other Models" dialog. If the Cline provider is
// already configured (a key is stored), it shows the live model catalog
// fetched from the gateway (which includes Cline's free models) and hides the
// add-key row; otherwise it falls back to the curated default list and shows
// an "ADD CLINE API KEY" row so the user can configure Cline right from here.
func NewOtherModels(com *common.Common) *OtherModels {
	prov := config.ClineProvider()
	models := config.ClineModels()
	addKey := true
	if cfg := com.Config(); cfg != nil {
		if pc, ok := cfg.Providers.Get(string(prov.ID)); ok {
			if len(pc.Models) > 0 {
				models = pc.Models
			}
			addKey = false // a key is stored; the catalog is live
		}
	}
	return &OtherModels{
		com:    com,
		prov:   prov,
		models: models,
		addKey: addKey,
	}
}

// rowCount returns the total number of selectable rows (models + optional
// add-key row).
func (o *OtherModels) rowCount() int {
	n := len(o.models)
	if o.addKey {
		n++
	}
	return n
}

// isAddKeyRow reports whether row i is the "ADD CLINE API KEY" row (last).
func (o *OtherModels) isAddKeyRow(i int) bool {
	return o.addKey && i == len(o.models)
}

// ID implements [Dialog].
func (o *OtherModels) ID() string { return OtherModelsID }

// HandleMsg implements [Dialog].
func (o *OtherModels) HandleMsg(msg tea.Msg) Action {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch key.String() {
	case "esc":
		return ActionClose{}
	case "up", "k":
		o.selected = (o.selected + o.rowCount() - 1) % o.rowCount()
	case "down", "j":
		o.selected = (o.selected + 1) % o.rowCount()
	case "enter", "ctrl+y":
		if o.isAddKeyRow(o.selected) {
			return ActionAddClineAPIKey{}
		}
		model := o.models[o.selected]
		return ActionSelectModel{
			Provider: o.prov,
			Model: config.SelectedModel{
				Model:           model.ID,
				Provider:        string(o.prov.ID),
				ReasoningEffort: model.DefaultReasoningEffort,
				MaxTokens:       model.DefaultMaxTokens,
			},
			ModelType: ModelTypeLarge.Config(),
		}
	}
	return nil
}

// Draw implements [Dialog].
func (o *OtherModels) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	lines := []string{
		"OTHER MODELS — CLINE CATALOG",
		"",
		"↑/↓ choose  enter select  esc close",
		"",
	}
	for i, model := range o.models {
		marker := "  "
		if i == o.selected && !o.isAddKeyRow(i) {
			marker = "● "
		}
		line := fmt.Sprintf("%s%s", marker, model.Name)
		if model.ContextWindow > 0 {
			line += fmt.Sprintf("  (%dk ctx)", model.ContextWindow/1000)
		}
		lines = append(lines, line)
	}
	if o.addKey {
		marker := "  "
		if o.selected == o.rowCount()-1 {
			marker = "● "
		}
		lines = append(lines, marker+"＋ ADD CLINE API KEY")
	}
	if len(o.models) == 0 && !o.addKey {
		lines = append(lines, "No models available.")
	}
	if o.addKey {
		lines = append(lines, "", "Add your Cline API key to load the full catalog, including free models.")
	} else {
		lines = append(lines, "", "Selecting a model will use your configured Cline key.")
	}

	view := o.com.Styles.Dialog.View.Width(min(64, max(1, area.Dx()-4))).Render(strings.Join(lines, "\n"))
	DrawCenter(scr, area, view)
	return nil
}

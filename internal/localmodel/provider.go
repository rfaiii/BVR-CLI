package localmodel

import (
	"charm.land/catwalk/pkg/catwalk"
	"github.com/richavery/bvr-cli/internal/config"
)

const ManagedOllamaProviderID = "ollama-local"

func ProviderConfig(model Model) config.ProviderConfig {
	return config.ProviderConfig{
		ID:                 ManagedOllamaProviderID,
		Name:               "Ollama (Local)",
		Type:               "ollama",
		BaseURL:            DefaultOllamaURL + "/v1",
		AutoDiscoverModels: new(true),
		Models:             []catwalk.Model{{ID: model.Name, Name: model.DisplayName, ContextWindow: model.ContextWindow, SupportsImages: contains(model.Capabilities, "vision")}},
	}
}

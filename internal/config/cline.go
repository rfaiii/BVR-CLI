package config

import (
	"charm.land/catwalk/pkg/catwalk"
)

// Cline provider integration.
//
// Cline hosts an OpenAI-compatible API gateway that proxies a curated set of
// frontier models for Cline users. It is integrated as a first-class known
// provider so it appears in the model picker/onboarding like any other
// catalog provider.
//
// - Base URL: https://api.cline.bot/v1
// - Auth:     X-Api-Key header with a Cline API token (created from the
//             Cline extension account page: https://account.cline.bot)
// - Key resolution: $CLINE_API_KEY, or providers.cline.api_key in config.
//
// The default model list below can be extended/overridden per-user via the
// providers.cline.models config entry (known-provider config merge handles
// that automatically).

// ClineProviderID is the ID of the built-in Cline provider.
const ClineProviderID = "cline"

// ClineBaseURL is the default Cline gateway endpoint.
const ClineBaseURL = "https://api.cline.bot/v1"

// ClineAPIKeyEnv is the environment variable read for the Cline API key.
const ClineAPIKeyEnv = "CLINE_API_KEY"

// ClineAPIKeyHeader is the HTTP header used to authenticate with the Cline
// gateway. Unlike standard OpenAI-compatible endpoints it is X-Api-Key rather
// than Authorization: Bearer.
const ClineAPIKeyHeader = "X-Api-Key"

// ClineModels returns the default model catalog exposed by the Cline gateway.
func ClineModels() []catwalk.Model {
	return []catwalk.Model{
		{
			ID:               "anthropic/claude-sonnet-4.5",
			Name:             "Claude Sonnet 4.5 (Cline)",
			ContextWindow:    200000,
			DefaultMaxTokens: 64000,
			CanReason:        true,
			SupportsImages:   true,
		},
		{
			ID:               "anthropic/claude-haiku-4.5",
			Name:             "Claude Haiku 4.5 (Cline)",
			ContextWindow:    200000,
			DefaultMaxTokens: 64000,
			CanReason:        true,
			SupportsImages:   true,
		},
		{
			ID:               "openai/gpt-5",
			Name:             "GPT-5 (Cline)",
			ContextWindow:    400000,
			DefaultMaxTokens: 128000,
			CanReason:        true,
			SupportsImages:   true,
		},
		{
			ID:               "openai/gpt-5-mini",
			Name:             "GPT-5 Mini (Cline)",
			ContextWindow:    400000,
			DefaultMaxTokens: 128000,
			CanReason:        true,
			SupportsImages:   true,
		},
		{
			ID:               "google/gemini-3-pro",
			Name:             "Gemini 3 Pro (Cline)",
			ContextWindow:    1048576,
			DefaultMaxTokens: 65536,
			CanReason:        true,
			SupportsImages:   true,
		},
		{
			ID:               "qwen/qwen3-coder",
			Name:             "Qwen3 Coder (Cline)",
			ContextWindow:    262144,
			DefaultMaxTokens: 65536,
			CanReason:        false,
			SupportsImages:   false,
		},
	}
}

// ClineProvider returns the built-in Cline provider definition.
func ClineProvider() catwalk.Provider {
	return catwalk.Provider{
		Name:                "Cline",
		ID:                  catwalk.InferenceProvider(ClineProviderID),
		APIEndpoint:         ClineBaseURL,
		Type:                catwalk.TypeOpenAICompat,
		DefaultLargeModelID: "anthropic/claude-sonnet-4.5",
		DefaultSmallModelID: "anthropic/claude-haiku-4.5",
		Models:              ClineModels(),
	}
}

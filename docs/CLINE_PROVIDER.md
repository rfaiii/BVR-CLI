# Cline Provider (api.cline.bot)

Cline's hosted API gateway is integrated into BVR-CLI as a first-class
provider. It is OpenAI-compatible and fronts a curated set of frontier models
(Claude, GPT, Gemini, Qwen) using a Cline account token.

## Setup

1. Create a Cline API token at https://account.cline.bot (or via the Cline
   extension's account page).
2. Either export it:

   ```sh
   export CLINE_API_KEY=your-token-here
   ```

   or run `bvr-cli`, open the model picker, choose the **Cline** provider,
   and paste the token when prompted (stored under
   `providers.cline.api_key` in the global config).

3. Select a model from the Cline group (e.g. `anthropic/claude-sonnet-4.5`).

## Usage notes

- The Cline catalog is reachable two ways:
  1. The `/` command palette → **Other Models** (aliases: `other`, `cline`,
     `external`). This opens the "Other Models" dialog
     (`internal/ui/dialog/other_models.go`) listing the Cline models; picking
     one prompts for the API key when none is stored.
  2. The main model picker (`ctrl+l`) when the provider is configured (i.e. a
     key was set via env or config).
- **"ADD CLINE API KEY" flow:** when the Cline provider is not yet configured,
  the Other Models dialog shows an "ADD CLINE API KEY" row. Selecting it opens
  the API key input; once the key is saved (verified with the `X-Api-Key`
  header), bvr fetches the full live catalog from the gateway — including all
  of Cline's **free models** — and shows it without a restart.
- Selecting a Cline model sets the large model; the small model defaults to
  the provider's `DefaultSmallModelID`.

## Details

- Base URL: `https://api.cline.bot/v1`
- Auth header: `X-Api-Key: <token>` (set automatically; the gateway does not
  use the standard `Authorization: Bearer` scheme)
- Provider type: `openai-compat`
- Default models: defined in `internal/config/cline.go` (`ClineModels()`).
  Add or override models per-user via the `providers.cline.models` config
  entry; user models are merged with the defaults.
- Endpoint override: set `providers.cline.base_url` in config if you need to
  proxy the gateway (honored by the known-provider config merge).

## Implementation notes

- `internal/config/cline.go` — provider + default model definitions.
- `internal/config/provider.go` — Cline is appended to the known provider
  list returned by `Providers()`.
- `internal/config/load.go` — `configureProviders` case resolves the API key
  (`$CLINE_API_KEY` or stored config) and attaches the `X-Api-Key` header;
  the header flows into the agent coordinator's OpenAI-compatible provider
  build (`internal/agent/coordinator.go`).

## Validation

`go build ./...`, `go vet ./internal/config/`, and
`go test ./internal/config/` pass. Live API validation requires a Cline token.

# bvr-cli: Multi-LLM Provider Integration Spec

## Context
This document serves as an implementation blueprint for `bvr-cli`. The objective is to refactor the application to support multiple LLM providers (Ollama, OpenRouter, Gemini) using the Strategy Pattern, driven by a local JSON configuration file. 

## Target Architecture
- **Language:** Go
- **Configuration Path:** OS-agnostic user config directory (e.g., `~/.config/bvr/config.json` on macOS/Linux).
- **Core Pattern:** Unified `LLMClient` interface decoupling commands from specific API implementations.

## Required Changes

### 1. Configuration Package (`config/config.go`)
- Implement a struct for the JSON config:
  - `active_provider` (string)
  - `providers` (map of provider names to credentials/endpoints)
- Implement a `LoadConfig()` function:
  - Use `os.UserConfigDir()` to resolve the path.
  - Ensure compatibility across different dev environments (macOS/Linux) by relying on standard library pathing.
  - Read and parse `bvr/config.json`.

### 2. LLM Interface (`llm/client.go`)
- Define the core interface:
  ```go
  type LLMClient interface {
      GenerateResponse(prompt string) (string, error)
  }
  ```

### 3. Provider Implementations (`llm/`)
- **`llm/ollama.go`**: Implement `OllamaClient` struct. Target local endpoint (default: `http://localhost:11434/api/generate`) for seamless local AI integration. 
- **`llm/openrouter.go`**: Implement `OpenRouterClient` struct. Needs to hit standard OpenAI compatible endpoints using OpenRouter's base URL and the user's API key.
- **`llm/gemini.go`**: Implement `GeminiClient` struct for Google's API endpoints.

### 4. Client Factory (`main.go` or `llm/factory.go`)
- Create an `InitializeClient(cfg *config.Config)` function.
- Use a `switch` statement on `cfg.ActiveProvider` to instantiate and return the correct implementation of `LLMClient`.

### 5. Dependency Management
- Ensure standard HTTP clients are used for Ollama.
- Add `github.com/sashabaranov/go-openai` for handling OpenRouter/OpenAI-compatible endpoints easily.
- Update `go.mod` accordingly.

## Execution Instructions for AI Agent
1. Read the current directory structure of the repository.
2. Create or update `config/config.go`.
3. Create the `llm` package and the `LLMClient` interface.
4. Implement the Ollama client first to ensure local model testing works immediately.
5. Wire the configuration loader and the client factory into the main entry point.
6. Refactor any existing hardcoded LLM calls to use the new `LLMClient` interface.

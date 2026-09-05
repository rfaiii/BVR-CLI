# Ollama How-To

BVR can discover and use local Ollama models without manually maintaining a
provider entry. Use `/ollama` to open **Switch Model**, or press `ctrl+l` from
the main screen. BVR checks the local API asynchronously and lists installed
models in the **Ollama (Local)** group.

## Install Ollama

Download Ollama from https://ollama.com and verify that the command is
available:

    ollama --version

macOS: install the Ollama application; it normally starts the service.
Windows: install the Ollama application; it normally runs in the background.
Linux: install the package and use your service manager, or run `ollama serve` in
a separate terminal. If Ollama is already running, BVR will use it instead of
starting a second server.

The default API endpoint is `http://127.0.0.1:11434`. Set `OLLAMA_HOST` when
using another endpoint:

    export OLLAMA_HOST=http://127.0.0.1:11434

PowerShell:

    $env:OLLAMA_HOST = "http://127.0.0.1:11434"

## Use Ollama in BVR

Keys in the model picker:

  enter  select a model
  r      refresh runtimes and models
  p      pull the highlighted Ollama model
  s      start Ollama if it is offline
  c      cancel an active pull
  esc    close

Highlight an installed model and press `enter`. BVR creates or updates the
managed `ollama-local` provider, saves the selected model, and refreshes the
agent. Prompts, tools, and normal agent coordination then use that local model.

Highlight a model and press `p` to download it. Pull progress appears in the
picker; press `c` to cancel. BVR refreshes the model list when the pull ends.

Manual provider configuration remains available for remote Ollama hosts, TLS,
proxies, custom headers, and other advanced settings.

Troubleshooting:

  command -v ollama       confirm Ollama on macOS/Linux
  where.exe ollama        confirm Ollama on Windows PowerShell
  ollama serve            start the daemon in a separate terminal
  ollama pull qwen2.5:7b  download a model directly

If Ollama is installed but offline, start it and press `r` in the model picker.
If a model is missing, press `p` on the model entry or run `ollama pull` in a
separate terminal. If the endpoint is remote, check `OLLAMA_HOST` and any
required proxy or TLS settings.
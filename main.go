// Package main is the entry point for the BVR CLI.
//
//	@title			BVR API
//	@version		1.2.1
//	@description	BVR is a terminal-based AI coding assistant. This API is served over a Unix socket (or Windows named pipe) and provides programmatic access to workspaces, sessions, agents, LSP, MCP, and more.
//	@contact.name	BVR
//	@contact.url	https://bvr-cli.com
//	@license.name	MIT
//	@license.url	https://github.com/richavery/bvr-cli/blob/main/LICENSE
//	@BasePath		/v1
package main

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/richavery/bvr-cli/internal/cmd"
	_ "github.com/richavery/bvr-cli/internal/dns"
)

func main() {
	if os.Getenv("BVR_PROFILE") != "" {
		go func() {
			slog.Info("Serving pprof at localhost:6060")
			if httpErr := http.ListenAndServe("localhost:6060", nil); httpErr != nil {
				slog.Error("Failed to pprof listen", "error", httpErr)
			}
		}()
	}

	cmd.Execute()
}

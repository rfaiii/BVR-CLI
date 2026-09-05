package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/richavery/bvr-cli/internal/node"
	"github.com/spf13/cobra"
)

var (
	nodeAgentHost  string
	nodeAgentToken string
)

func init() {
	nodeCmd.AddCommand(nodeServeCmd)
	nodeServeCmd.Flags().StringVar(&nodeAgentHost, "host", "127.0.0.1:7777", "NODE HTTP agent listen address")
	nodeServeCmd.Flags().StringVar(&nodeAgentToken, "token", "", "Bearer token required by NODE clients")
	rootCmd.AddCommand(nodeCmd)
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage NODE devices",
}

var nodeServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a NODE HTTP agent",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if strings.TrimSpace(nodeAgentToken) == "" && !strings.HasPrefix(nodeAgentHost, "127.0.0.1:") && !strings.HasPrefix(nodeAgentHost, "localhost:") {
			return fmt.Errorf("--token is required when NODE agent binds outside localhost")
		}
		agentHandler := node.NodeAgentHandler(nodeAgentToken, nil)
		server := &http.Server{Addr: nodeAgentHost, Handler: agentHandler, ReadHeaderTimeout: 5 * time.Second}
		errCh := make(chan error, 1)
		go func() { errCh <- server.ListenAndServe() }()
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		case err := <-errCh:
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}
	},
}

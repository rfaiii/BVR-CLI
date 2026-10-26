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
	nodeAgentHost      string
	nodeAgentToken     string
	nodeAgentAdvertise bool
	nodeScanWait       time.Duration
)

func init() {
	nodeCmd.AddCommand(nodeServeCmd, nodeScanCmd)
	nodeServeCmd.Flags().StringVar(&nodeAgentHost, "host", "127.0.0.1:7777", "NODE HTTP agent listen address")
	nodeServeCmd.Flags().StringVar(&nodeAgentToken, "token", "", "Bearer token required by NODE clients")
	nodeServeCmd.Flags().BoolVar(&nodeAgentAdvertise, "advertise", false, "Publish _bvr._tcp over mDNS so companions can find this host")
	nodeScanCmd.Flags().DurationVar(&nodeScanWait, "wait", node.BrowseWindow, "How long to listen for advertised NODEs")
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
		// The agent runs arbitrary commands for any client that authenticates,
		// so a published agent has to be an authenticated one.
		if nodeAgentAdvertise && strings.TrimSpace(nodeAgentToken) == "" {
			return fmt.Errorf("--advertise requires --token: a published NODE agent is reachable by anyone on the local network")
		}
		agentHandler := node.NodeAgentHandler(nodeAgentToken, nil)
		server := &http.Server{Addr: nodeAgentHost, Handler: agentHandler, ReadHeaderTimeout: 5 * time.Second}
		errCh := make(chan error, 1)
		go func() { errCh <- server.ListenAndServe() }()
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if nodeAgentAdvertise {
			advertisement, err := node.Advertise(node.AdvertiseOptions{Addr: nodeAgentHost})
			if err != nil {
				return err
			}
			defer advertisement.Close()
			fmt.Printf("NODE agent listening on %s, advertised as %s._bvr._tcp.local\n", nodeAgentHost, advertisement.Instance)
		}
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

var nodeScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Find NODE agents advertised on the local network",
	RunE: func(cmd *cobra.Command, _ []string) error {
		nodes, err := node.Scan(cmd.Context(), nodeScanWait)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			fmt.Printf("No NODE agents found. Start one with `bvr-cli node serve --advertise --token <token>`.\n")
			return nil
		}
		for _, found := range nodes {
			fmt.Printf("%-24s %-22s proto=%-2s id=%s\n", found.DisplayName(), found.HostPort(), found.Protocol, found.ID)
		}

		// Recording the results is what makes them show up in the device
		// picker on the next run, so a scan is worth doing even when the
		// printed list is not what the user came for.
		recorded, err := node.RecordScannedNodes(nodes)
		if err != nil {
			return fmt.Errorf("scan found %d node(s) but could not be recorded: %w", len(nodes), err)
		}
		fmt.Printf("\n%d of %d node(s) recorded in %s\n", len(recorded), len(nodes), node.DefaultDevicesPath())
		return nil
	},
}

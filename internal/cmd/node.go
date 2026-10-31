package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/richavery/bvr-cli/internal/node"
	"github.com/spf13/cobra"
)

var (
	nodeAgentHost         string
	nodeAgentToken        string
	nodeAgentAdvertise    bool
	nodeAgentProjects     []string
	nodeAgentAllowExec    []string
	nodeAgentUnrestricted bool
	nodeScanWait          time.Duration
	nodeConnectToken      string
	nodeConnectName       string
)

func init() {
	nodeCmd.AddCommand(nodeServeCmd, nodeScanCmd, nodeConnectCmd)
	nodeServeCmd.Flags().StringVar(&nodeAgentHost, "host", "127.0.0.1:7777", "NODE HTTP agent listen address")
	nodeServeCmd.Flags().StringVar(&nodeAgentToken, "token", "", "Bearer token required by NODE clients")
	nodeServeCmd.Flags().BoolVar(&nodeAgentAdvertise, "advertise", false, "Publish _bvr._tcp over mDNS so companions can find this host")
	nodeServeCmd.Flags().StringArrayVar(&nodeAgentProjects, "project", nil, "Project root to share with authenticated peers (repeatable)")
	nodeServeCmd.Flags().StringArrayVar(&nodeAgentAllowExec, "allow-exec", nil, "Extra command name the execute endpoint may run (repeatable)")
	nodeServeCmd.Flags().BoolVar(&nodeAgentUnrestricted, "unrestricted-exec", false, "Allow every command name. This is the previous execute behavior")
	nodeScanCmd.Flags().DurationVar(&nodeScanWait, "wait", node.BrowseWindow, "How long to listen for advertised NODEs")
	nodeConnectCmd.Flags().StringVar(&nodeConnectToken, "token", "", "Bearer token for the remote NODE agent")
	nodeConnectCmd.Flags().StringVar(&nodeConnectName, "name", "", "Nickname stored for this NODE")
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
		if strings.TrimSpace(nodeAgentToken) == "" && !node.ListenHostIsLoopback(nodeAgentHost) {
			return fmt.Errorf("--token is required when NODE agent binds outside localhost")
		}
		// The agent runs commands for any client that authenticates, so a
		// published agent has to be an authenticated one.
		if nodeAgentAdvertise && strings.TrimSpace(nodeAgentToken) == "" {
			return fmt.Errorf("--advertise requires --token: a published NODE agent is reachable by anyone on the local network")
		}
		if nodeAgentAdvertise && node.ListenHostIsLoopback(nodeAgentHost) {
			return fmt.Errorf("--advertise publishes a LAN address; bind a non-loopback address such as 0.0.0.0:7777")
		}
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		roots := make([]string, 0, len(nodeAgentProjects))
		for _, root := range nodeAgentProjects {
			if !filepath.IsAbs(root) {
				root = filepath.Join(cwd, root)
			}
			roots = append(roots, root)
		}
		share, err := node.NewFileShare(roots)
		if err != nil {
			return err
		}
		defer share.Close()
		policy := node.ExecPolicy{Unrestricted: nodeAgentUnrestricted, Extra: nodeAgentAllowExec}
		handler := node.NewAgentHandler(node.AgentOptions{
			Token: nodeAgentToken, Exec: policy, Share: share, NodeID: node.LocalNodeName(),
		})
		server := &http.Server{Addr: nodeAgentHost, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
		errCh := make(chan error, 1)
		go func() { errCh <- server.ListenAndServe() }()
		fmt.Fprint(cmd.OutOrStdout(), node.StartupInfo{
			Addr: nodeAgentHost, Exec: policy, Projects: share.Projects(), WatchWarning: share.WatchWarning(),
		}.Banner())
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if nodeAgentAdvertise {
			advertisement, err := node.Advertise(node.AdvertiseOptions{Addr: nodeAgentHost})
			if err != nil {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = server.Shutdown(shutdownCtx)
				return err
			}
			defer advertisement.Close()
			fmt.Fprintf(cmd.OutOrStdout(), "Advertised as %s._bvr._tcp.local\n", advertisement.Instance)
			if runtime.GOOS == "darwin" {
				fmt.Fprint(cmd.OutOrStdout(), "This Mac does not see its own Bonjour ad. Another process on this machine should use `bvr-cli node connect 127.0.0.1:<port> --token <token>`.\n")
			}
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

var nodeConnectCmd = &cobra.Command{
	Use:   "connect <host:port>",
	Short: "Dial a NODE agent and save it",
	Long: `Check a NODE agent's health and protocol, save it in nodes.json, and
register a live transport. The bearer token is stored in
node-credentials.json, not in the device record.

A second process on the same Mac should dial 127.0.0.1. macOS does not
deliver a machine's own Bonjour advertisement, so node scan on the Mac
that is serving will not list that agent.`,
	Example: `
# Same Mac, second terminal, agent bound on 0.0.0.0:7777 or 127.0.0.1:7777
bvr-cli node connect 127.0.0.1:7777 --token "$BVR_NODE_TOKEN"

# Another machine on the same LAN, or a Tailscale address
bvr-cli node connect 100.x.y.z:7777 --token "$BVR_NODE_TOKEN"
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := node.Connect(cmd.Context(), args[0], nodeConnectToken, nodeConnectName)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Connected to %s (%s) protocol %s\n", device.DisplayName(), device.Address, node.ProtocolVersion)
		fmt.Fprintf(cmd.OutOrStdout(), "Saved %s\n", node.DefaultDevicesPath())
		fmt.Fprintf(cmd.OutOrStdout(), "Token stored in %s\n", node.DefaultCredentialsPath())
		return nil
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
			fmt.Fprint(cmd.OutOrStdout(), "No NODE agents found. Start one with `bvr-cli node serve --advertise --token <token>`.\n")
			if runtime.GOOS == "darwin" {
				fmt.Fprint(cmd.OutOrStdout(), "A Mac does not receive its own Bonjour advertisement. From another terminal on that Mac, dial it with `bvr-cli node connect 127.0.0.1:7777 --token <token>`.\n")
			}
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

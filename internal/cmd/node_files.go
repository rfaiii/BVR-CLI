package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/richavery/bvr-cli/internal/node"
	"github.com/spf13/cobra"
)

var (
	filesNode    string
	filesProject string
	filesFrom    string
)

func init() {
	nodeFilesCmd.PersistentFlags().StringVar(&filesNode, "node", "", "Saved NODE id, name, or host:port")
	nodeFilesCmd.PersistentFlags().StringVar(&filesProject, "project", "", "Shared project id (default: the only project)")
	nodeFilesPutCmd.Flags().StringVar(&filesFrom, "from", "", "Local file to upload (default: stdin)")
	nodeFilesCmd.AddCommand(
		nodeFilesLsCmd,
		nodeFilesCatCmd,
		nodeFilesPutCmd,
		nodeFilesRmCmd,
		nodeFilesMvCmd,
		nodeFilesMkdirCmd,
		nodeFilesWatchCmd,
	)
	nodeCmd.AddCommand(nodeFilesCmd)
}

var nodeFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Read and write files on a connected NODE",
	Long: `Files stay on the machine running node serve --project. This command
is the BACKUP-side client: it uses the NODE saved by node connect.

Writes send the etag from a prior read. If MAIN's copy changed, the
command fails instead of overwriting it.`,
}

var nodeFilesLsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List a directory on MAIN",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		list, err := client.List(cmd.Context(), rel, 1)
		if err != nil {
			return err
		}
		for _, entry := range list.Entries {
			name := entry.Path
			if entry.Type == "dir" {
				name += "/"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", entry.Type, name, entry.ETag)
		}
		return nil
	},
}

var nodeFilesCatCmd = &cobra.Command{
	Use:   "cat <path>",
	Short: "Print a file from MAIN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		body, err := client.Read(cmd.Context(), rel)
		if err != nil {
			return err
		}
		_, err = io.WriteString(cmd.OutOrStdout(), body.Content)
		return err
	},
}

var nodeFilesPutCmd = &cobra.Command{
	Use:   "put <path>",
	Short: "Write a file on MAIN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		var data []byte
		if filesFrom != "" {
			data, err = os.ReadFile(filesFrom)
		} else {
			data, err = io.ReadAll(io.LimitReader(cmd.InOrStdin(), node.MaxFileBytes+1))
		}
		if err != nil {
			return err
		}
		if len(data) > node.MaxFileBytes {
			return fmt.Errorf("file exceeds %d bytes", node.MaxFileBytes)
		}
		entry, err := client.Put(cmd.Context(), rel, data)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s etag=%s\n", entry.Path, entry.ETag)
		return nil
	},
}

var nodeFilesRmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Delete a file or empty directory on MAIN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		if err := client.Remove(cmd.Context(), rel); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", rel)
		return nil
	},
}

var nodeFilesMvCmd = &cobra.Command{
	Use:   "mv <from> <to>",
	Short: "Rename a file on MAIN",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := filesClient(cmd, nil)
		if err != nil {
			return err
		}
		entry, err := client.Rename(cmd.Context(), args[0], args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "renamed %s -> %s\n", args[0], entry.Path)
		return nil
	},
}

var nodeFilesMkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "Create a directory on MAIN",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		if _, err := client.Mkdir(cmd.Context(), rel); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", rel)
		return nil
	},
}

var nodeFilesWatchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Print filesystem events from MAIN",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, rel, err := filesClient(cmd, args)
		if err != nil {
			return err
		}
		events, err := client.Watch(cmd.Context(), rel)
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), "watching\n")
		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case event, ok := <-events:
				if !ok {
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", event.Op, event.Path, event.ETag)
			}
		}
	},
}

func filesClient(cmd *cobra.Command, args []string) (*node.FilesClient, string, error) {
	rel := ""
	if len(args) > 0 {
		rel = args[0]
	}
	device, token, err := node.ResolveSaved(filesNode)
	if err != nil {
		return nil, "", err
	}
	base, err := node.HTTPBase(device.TransportURL)
	if err != nil {
		return nil, "", err
	}
	httpClient := node.NewHTTPTransport(base, token)
	project := strings.TrimSpace(filesProject)
	if project == "" {
		projects, err := httpClient.ListProjects(cmd.Context())
		if err != nil {
			return nil, "", err
		}
		switch len(projects) {
		case 0:
			return nil, "", errors.New("the NODE is not sharing a project. Start it with --project <path>")
		case 1:
			project = projects[0].ID
		default:
			return nil, "", errors.New("the NODE shares more than one project; pass --project")
		}
	}
	return &node.FilesClient{
		HTTP:         httpClient,
		WebSocketURL: device.TransportURL,
		Token:        token,
		Project:      project,
	}, rel, nil
}

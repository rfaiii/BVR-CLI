package cmd

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/richavery/bvr-cli/internal/shader"
	"github.com/spf13/cobra"
)

var shadersCmd = &cobra.Command{
	Use:   "shaders",
	Short: "Manage bundled shaders",
	Long:  "List, preview, and install the bundled cursor/screen shaders.",
	Example: `# List bundled shaders
bvr-cli shaders list

# Install default shaders into Ghostty
bvr-cli shaders install

# Install with custom shader names
bvr-cli shaders install cursor_warp.glsl cursor_tail.glsl`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(shadersCmd)
}

var shadersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bundled shaders",
	RunE: func(cmd *cobra.Command, _ []string) error {
		items, err := shader.List()
		if err != nil {
			return err
		}
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			for _, item := range items {
				cmd.Println(item)
			}
			return nil
		}
		for _, item := range items {
			cmd.Println("• " + item)
		}
		return nil
	},
}

func init() {
	shadersCmd.AddCommand(shadersListCmd)
}

var shadersInstallCmd = &cobra.Command{
	Use:   "install [shader names...]",
	Short: "Install bundled shaders into Ghostty",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		if err := shader.InstallGhostty(args, force); err != nil {
			return err
		}
		cmd.Println("Installed shaders into", shader.TargetDir())
		return nil
	},
}

func init() {
	shadersInstallCmd.Flags().BoolP("force", "f", false, "Overwrite existing custom-shader config")
	shadersCmd.AddCommand(shadersInstallCmd)
}

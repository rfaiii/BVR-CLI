package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/richavery/bvr-cli/internal/event"
	"github.com/richavery/bvr-cli/internal/ui/common"
	"github.com/richavery/bvr-cli/internal/ui/model"
	"github.com/spf13/cobra"
)

var codeCmd = &cobra.Command{
	Use:   "code",
	Short: "Start a coder-only session",
	Long:  `Start a BVR session in coder mode with coding-focused tools and system prompt.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		native, err := cmd.Flags().GetBool("native")
		if err != nil {
			return err
		}
		if native {
			return runNativeCode(cmd, args)
		}

		ws, cleanup, err := setupWorkspaceWithProgressBar(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		event.AppInitialized()

		if err := ws.InitCoderAgent(cmd.Context()); err != nil {
			return fmt.Errorf("failed to initialize coder agent: %w", err)
		}

		com := common.DefaultCommon(ws)
		m := model.New(com, "", false)

		var env uv.Environ = os.Environ()
		program := tea.NewProgram(
			m,
			tea.WithEnvironment(env),
			tea.WithContext(cmd.Context()),
		)
		go ws.Subscribe(program)

		if _, err := program.Run(); err != nil {
			event.Error(err)
			slog.Error("TUI run error", "error", err)
			return fmt.Errorf("BVR crashed: %w", err)
		}
		return nil
	},
}

func init() {
	codeCmd.Flags().Bool("native", false, "Use the fantasy-free native Ollama coding agent")
	codeCmd.Flags().String("model", "", "Ollama model for --native (defaults to codetool's default)")
	codeCmd.Flags().Bool("stream", true, "Stream model output for --native")
	rootCmd.AddCommand(codeCmd)
}

func runNativeCode(cmd *cobra.Command, args []string) error {
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return err
	}
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to determine working directory: %w", err)
		}
	} else if cwd, err = filepath.Abs(cwd); err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}

	tool, err := nativeCodeTool()
	if err != nil {
		return err
	}
	toolArgs := []string{"-cwd", cwd}
	if model, err := cmd.Flags().GetString("model"); err != nil {
		return err
	} else if model != "" {
		toolArgs = append(toolArgs, "-model", model)
	}
	if stream, err := cmd.Flags().GetBool("stream"); err != nil {
		return err
	} else if stream {
		toolArgs = append(toolArgs, "-stream")
	}
	toolArgs = append(toolArgs, args...)

	process := exec.CommandContext(cmd.Context(), tool, toolArgs...)
	process.Dir = cwd
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
	if err := process.Run(); err != nil {
		return fmt.Errorf("native coding agent failed: %w", err)
	}
	return nil
}

func nativeCodeTool() (string, error) {
	if path := os.Getenv("BVR_CODETOOL"); path != "" {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("BVR_CODETOOL does not point to an executable: %s", path)
	}
	path, err := exec.LookPath("codetool")
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("native coding agent not found; build cmd/codetool or set BVR_CODETOOL")
}

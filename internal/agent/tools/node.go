package tools

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"github.com/richavery/bvr-cli/internal/node"
	"github.com/richavery/bvr-cli/internal/permission"
)

const (
	NodeToolName = "node"
	NPMToolName  = "npm"
)

type NodeParams struct {
	Description string   `json:"description" description:"Brief description of the Node command"`
	DeviceID    string   `json:"device_id,omitempty" description:"Target NODE device ID (defaults to local)"`
	Command     string   `json:"command" description:"JavaScript file or Node command to execute"`
	Args        []string `json:"args,omitempty" description:"Arguments for the Node command"`
	WorkingDir  string   `json:"working_dir,omitempty" description:"Working directory"`
}

type NPMParams struct {
	Description    string   `json:"description" description:"Brief description of the package script"`
	DeviceID       string   `json:"device_id,omitempty" description:"Target NODE device ID (defaults to local)"`
	Script         string   `json:"script" description:"Package script name"`
	Arguments      []string `json:"arguments,omitempty" description:"Arguments passed to the script"`
	PackageManager string   `json:"package_manager,omitempty" description:"npm, pnpm, yarn, or bun"`
	WorkingDir     string   `json:"working_dir,omitempty" description:"Working directory"`
}

type NodeToolMetadata struct {
	DeviceID string `json:"device_id"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
}

func NewNodeTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(NodeToolName, "Execute a Node command on a selected NODE device.", func(ctx context.Context, params NodeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
		if params.Command == "" {
			return fantasy.NewTextErrorResponse("missing command"), nil
		}
		request := node.BuildNodeRequest(params.Command, params.Args, params.WorkingDir)
		return executeNodeTool(ctx, permissions, call, NodeToolName, params.DeviceID, params.WorkingDir, workingDir, params, request)
	})
}

func NewNPMTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(NPMToolName, "Run an npm-compatible package script on a selected NODE device.", func(ctx context.Context, params NPMParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
		if params.Script == "" {
			return fantasy.NewTextErrorResponse("missing script"), nil
		}
		request := node.BuildNPMRequest(params.PackageManager, params.Script, params.Arguments, params.WorkingDir)
		return executeNodeTool(ctx, permissions, call, NPMToolName, params.DeviceID, params.WorkingDir, workingDir, params, request)
	})
}

func executeNodeTool(ctx context.Context, permissions permission.Service, call fantasy.ToolCall, name, deviceID, _, defaultDir string, params any, request node.CommandRequest) (fantasy.ToolResponse, error) {
	if deviceID == "" {
		deviceID = "local"
	}
	if request.WorkingDir == "." || request.WorkingDir == "" {
		request.WorkingDir = defaultDir
	}
	sessionID := GetSessionFromContext(ctx)
	if sessionID == "" {
		return fantasy.ToolResponse{}, fmt.Errorf("session ID is required")
	}
	allowed, err := permissions.Request(ctx, permission.CreatePermissionRequest{
		SessionID: sessionID, ToolCallID: call.ID, ToolName: name, Action: "execute",
		Path: request.WorkingDir, Description: fmt.Sprintf("Execute %s on NODE %s", name, deviceID), Params: params,
	})
	if err != nil {
		return fantasy.ToolResponse{}, err
	}
	if !allowed {
		return NewPermissionDeniedResponse(), nil
	}
	result, err := node.ExecuteDevice(ctx, deviceID, request)
	metadata := NodeToolMetadata{DeviceID: deviceID, Command: request.Command, ExitCode: result.ExitCode}
	output := result.Stdout
	if result.Stderr != "" {
		output += "\n" + result.Stderr
	}
	if err != nil {
		return fantasy.WithResponseMetadata(fantasy.NewTextErrorResponse(output), metadata), nil
	}
	return fantasy.WithResponseMetadata(fantasy.NewTextResponse(output), metadata), nil
}

package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrTransportUnavailable = errors.New("node transport unavailable")

type CommandRequest struct {
	Command    string   `json:"command"`
	Args       []string `json:"args,omitempty"`
	WorkingDir string   `json:"working_dir,omitempty"`
}

type CommandResult struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code"`
}

type Transport interface {
	Execute(context.Context, CommandRequest) (CommandResult, error)
	Ping(context.Context) error
}

type LocalTransport struct{}

func (LocalTransport) Execute(ctx context.Context, request CommandRequest) (CommandResult, error) {
	if request.Command == "" {
		return CommandResult{}, errors.New("missing command")
	}
	cmd := exec.CommandContext(ctx, request.Command, request.Args...)
	if request.WorkingDir != "" {
		cmd.Dir = request.WorkingDir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	result := CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil {
		return result, nil
	}
	result.ExitCode = 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	return result, fmt.Errorf("command failed: %w", err)
}

func (LocalTransport) Ping(context.Context) error { return nil }

type TransportManager struct {
	devices    ManagerDevices
	transports map[string]Transport
}

// ManagerDevices is the small registry contract required by orchestration.
type ManagerDevices interface {
	Devices() []Device
	SetStatus(string, DeviceStatus)
}

func NewTransportManager(devices ManagerDevices) *TransportManager {
	return &TransportManager{devices: devices, transports: make(map[string]Transport)}
}

func (m *TransportManager) Register(id string, transport Transport) {
	if transport != nil {
		m.transports[id] = transport
	}
}

func (m *TransportManager) Execute(ctx context.Context, deviceID string, request CommandRequest) (CommandResult, error) {
	device, ok := findDevice(m.devices.Devices(), deviceID)
	if !ok {
		return CommandResult{}, fmt.Errorf("device %q not found", deviceID)
	}
	transport := m.transports[device.ID]
	if transport == nil {
		if device.ConnectionType == "local" || device.ID == "local" {
			transport = LocalTransport{}
		} else {
			return CommandResult{}, fmt.Errorf("device %q: %w", device.ID, ErrTransportUnavailable)
		}
	}
	result, err := transport.Execute(ctx, request)
	if err != nil {
		m.devices.SetStatus(device.ID, DeviceStatusError)
		return result, err
	}
	m.devices.SetStatus(device.ID, DeviceStatusOnline)
	return result, nil
}

func (m *TransportManager) Ping(ctx context.Context, deviceID string) error {
	device, ok := findDevice(m.devices.Devices(), deviceID)
	if !ok {
		return fmt.Errorf("device %q not found", deviceID)
	}
	transport := m.transports[device.ID]
	if transport == nil && (device.ConnectionType == "local" || device.ID == "local") {
		transport = LocalTransport{}
	}
	if transport == nil {
		m.devices.SetStatus(device.ID, DeviceStatusOffline)
		return ErrTransportUnavailable
	}
	if err := transport.Ping(ctx); err != nil {
		m.devices.SetStatus(device.ID, DeviceStatusError)
		return err
	}
	m.devices.SetStatus(device.ID, DeviceStatusOnline)
	return nil
}

func findDevice(devices []Device, id string) (Device, bool) {
	for _, device := range devices {
		if device.ID == id {
			return device, true
		}
	}
	return Device{}, false
}

func BuildNodeRequest(command string, args []string, workingDir string) CommandRequest {
	name := "node"
	if runtime.GOOS == "windows" {
		name = "node.exe"
	}
	return CommandRequest{Command: name, Args: append([]string{command}, args...), WorkingDir: filepath.Clean(workingDir)}
}

func BuildNPMRequest(manager string, script string, args []string, workingDir string) CommandRequest {
	if manager == "" {
		manager = "npm"
	}
	if strings.EqualFold(manager, "npm") {
		manager = "npm"
	}
	return CommandRequest{Command: manager, Args: append([]string{"run", script, "--"}, args...), WorkingDir: filepath.Clean(workingDir)}
}

var defaultTransportManager = NewTransportManager(defaultManager)

func RegisterTransport(id string, transport Transport) {
	defaultTransportManager.Register(id, transport)
}
func ExecuteDevice(ctx context.Context, id string, request CommandRequest) (CommandResult, error) {
	return defaultTransportManager.Execute(ctx, id, request)
}
func PingDevice(ctx context.Context, id string) error { return defaultTransportManager.Ping(ctx, id) }

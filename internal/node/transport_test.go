package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeTransport struct {
	result CommandResult
	err    error
}

func (f fakeTransport) Execute(context.Context, CommandRequest) (CommandResult, error) {
	return f.result, f.err
}
func (f fakeTransport) Ping(context.Context) error { return f.err }

func TestTransportManagerRoutesAndUpdatesStatus(t *testing.T) {
	devices := NewManager()
	devices.Upsert(Device{ID: "remote", Name: "Remote", ConnectionType: "test", Status: DeviceStatusOffline})
	transport := NewTransportManager(devices)
	transport.Register("remote", fakeTransport{result: CommandResult{Stdout: "ok"}})
	result, err := transport.Execute(context.Background(), "remote", CommandRequest{Command: "node"})
	require.NoError(t, err)
	require.Equal(t, "ok", result.Stdout)
	require.Equal(t, DeviceStatusOnline, devices.Devices()[0].Status)
}

func TestBuildRequests(t *testing.T) {
	nodeRequest := BuildNodeRequest("script.js", []string{"--watch"}, "/tmp/project")
	require.Equal(t, []string{"script.js", "--watch"}, nodeRequest.Args)
	npmRequest := BuildNPMRequest("pnpm", "build", []string{"--prod"}, "/tmp/project")
	require.Equal(t, "pnpm", npmRequest.Command)
	require.Equal(t, []string{"run", "build", "--", "--prod"}, npmRequest.Args)
}

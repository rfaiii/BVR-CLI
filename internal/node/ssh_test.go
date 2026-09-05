package node

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSSHTransportRequiresHostVerification(t *testing.T) {
	_, err := NewSSHTransport(SSHConfig{Address: "host:22", User: "user", Password: "password"})
	require.ErrorIs(t, err, ErrSSHHostKeyRequired)
}

func TestNewSSHTransportAllowsExplicitInsecureMode(t *testing.T) {
	transport, err := NewSSHTransport(SSHConfig{Address: "host:22", User: "user", Password: "password", InsecureSkipHostKey: true})
	require.NoError(t, err)
	require.NotNil(t, transport)
}

func TestNewSSHTransportRequiresCredentials(t *testing.T) {
	_, err := NewSSHTransport(SSHConfig{Address: "host:22", User: "user", InsecureSkipHostKey: true})
	require.Error(t, err)
}

func TestRemoteCommandShellQuotesArguments(t *testing.T) {
	command := remoteCommand(CommandRequest{Command: "node", Args: []string{"script.js", "it's safe"}, WorkingDir: "/tmp/project dir"})
	require.Equal(t, "cd -- '/tmp/project dir' && 'node' 'script.js' 'it'\\''s safe'", command)
}

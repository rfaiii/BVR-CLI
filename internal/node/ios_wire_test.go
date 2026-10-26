package node

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// The payloads below were captured from Swift's JSONEncoder running the real
// structs in mobile-cli/ios/BvrCompanion/BvrCompanion/NodeClient.swift. Field
// names are the entire contract between the two languages, so renaming one on
// either side fails here instead of showing up as a silent empty command on the
// phone.

func TestIOSExecuteFrameDecodes(t *testing.T) {
	var frame WSFrame
	require.NoError(t, json.Unmarshal([]byte(`{"request":{"command":"uname","args":["-sr"],"working_dir":"~\/projects"},"type":"execute","id":"A1"}`), &frame))
	require.Equal(t, "A1", frame.ID)
	require.Equal(t, "execute", frame.Type)
	require.NotNil(t, frame.Request)
	require.Equal(t, CommandRequest{Command: "uname", Args: []string{"-sr"}, WorkingDir: "~/projects"}, *frame.Request)
}

// Swift omits every nil optional, so a ping arrives with only id and type.
func TestIOSPingFrameDecodes(t *testing.T) {
	var frame WSFrame
	require.NoError(t, json.Unmarshal([]byte(`{"id":"A2","type":"ping"}`), &frame))
	require.Equal(t, "ping", frame.Type)
	require.Nil(t, frame.Request)
	require.Nil(t, frame.Result)
}

func TestResultFrameCarriesTheKeysSwiftReads(t *testing.T) {
	payload, err := json.Marshal(WSFrame{ID: "A1", Type: "result", Result: &CommandResult{Stdout: "Darwin 25.5.0\n"}})
	require.NoError(t, err)

	decoded := map[string]any{}
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, "A1", decoded["id"], "Swift matches each reply by the echoed frame id")
	result, ok := decoded["result"].(map[string]any)
	require.True(t, ok, "result object missing from %s", payload)
	require.Equal(t, "Darwin 25.5.0\n", result["stdout"])
	require.Contains(t, result, "exit_code")
}

func TestStreamAndErrorFramesMatchSwiftDecoder(t *testing.T) {
	stream, err := json.Marshal(WSFrame{ID: "A1", Type: "stdout", Stream: "stdout", Data: "line\n"})
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"A1","type":"stdout","stream":"stdout","data":"line\n"}`, string(stream))

	failure, err := json.Marshal(WSFrame{ID: "A2", Type: "error", Error: "unauthorized"})
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"A2","type":"error","error":"unauthorized"}`, string(failure))
}

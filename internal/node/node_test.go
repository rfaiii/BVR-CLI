package node

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiscoverDevicesProbesLocalhost(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	address := ln.Addr().String()
	devices := Discover([]Probe{{ID: "localhost", Name: address, Address: address, ConnectionType: "tcp"}}, TCPProbe(250*time.Millisecond))
	require.Len(t, devices, 1)
	require.Equal(t, address, devices[0].Address)
	require.Equal(t, DeviceStatusOnline, devices[0].Status)
}

func TestDeviceManagerConnectivityStateTransitions(t *testing.T) {
	m := NewManager()
	m.Upsert(Device{ID: "host", Name: "Host", Address: "127.0.0.1:7777", ConnectionType: "http", Status: DeviceStatusOffline})

	require.Equal(t, DeviceStatusOffline, m.Devices()[0].Status)
	m.SetStatus("host", DeviceStatusOnline)
	require.Equal(t, DeviceStatusOnline, m.Devices()[0].Status)
	require.False(t, m.Devices()[0].LastSeen.IsZero())

	m.SetStatus("host", DeviceStatusError)
	require.Equal(t, DeviceStatusError, m.Devices()[0].Status)

	m.SetTransportURL("host", "http://127.0.0.1:7777/v1/node/execute")
	require.Equal(t, "http://127.0.0.1:7777/v1/node/execute", m.Devices()[0].TransportURL)
}

func TestExecuteDeviceRoutesToLocalTransport(t *testing.T) {
	m := NewManager()
	m.Upsert(Device{ID: "local", Name: "Local", ConnectionType: "local", Status: DeviceStatusOnline})
	tm := NewTransportManager(m)

	result, err := tm.Execute(context.Background(), "local", CommandRequest{Command: "printf", Args: []string{"hello"}})
	require.NoError(t, err)
	require.Equal(t, "hello", strings.TrimSpace(result.Stdout))
}

func TestDiscoverDevicesProbesConfiguredAddress(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	address := ln.Addr().String()
	devices := Discover([]Probe{{ID: "test", Name: address, Address: address, ConnectionType: "tcp"}}, TCPProbe(250*time.Millisecond))
	require.Len(t, devices, 1)
	require.Equal(t, address, devices[0].Address)
	require.Equal(t, DeviceStatusOnline, devices[0].Status)
}

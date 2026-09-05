package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var ErrSSHHostKeyRequired = errors.New("ssh known_hosts verification is required")

type SSHConfig struct {
	Address             string
	User                string
	Password            string
	PrivateKey          []byte
	PrivateKeyPath      string
	KnownHostsPath      string
	InsecureSkipHostKey bool
	Timeout             time.Duration
}

type SSHTransport struct {
	config SSHConfig
	mu     sync.Mutex
	client *ssh.Client
}

func NewSSHTransport(config SSHConfig) (*SSHTransport, error) {
	if config.Address == "" {
		return nil, errors.New("missing SSH address")
	}
	if config.User == "" {
		return nil, errors.New("missing SSH user")
	}
	if config.Password == "" && len(config.PrivateKey) == 0 && config.PrivateKeyPath == "" {
		return nil, errors.New("SSH password or private key is required")
	}
	if !config.InsecureSkipHostKey && config.KnownHostsPath == "" {
		return nil, ErrSSHHostKeyRequired
	}
	if config.Timeout <= 0 {
		config.Timeout = 15 * time.Second
	}
	if config.PrivateKeyPath != "" && len(config.PrivateKey) == 0 {
		key, err := os.ReadFile(filepath.Clean(config.PrivateKeyPath))
		if err != nil {
			return nil, fmt.Errorf("read SSH private key: %w", err)
		}
		config.PrivateKey = key
	}
	return &SSHTransport{config: config}, nil
}

func (t *SSHTransport) clientConfig() (*ssh.ClientConfig, error) {
	auth := make([]ssh.AuthMethod, 0, 2)
	if t.config.Password != "" {
		auth = append(auth, ssh.Password(t.config.Password))
	}
	if len(t.config.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(t.config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("parse SSH private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	var callback ssh.HostKeyCallback
	if t.config.InsecureSkipHostKey {
		callback = ssh.InsecureIgnoreHostKey()
	} else {
		var err error
		callback, err = knownhosts.New(t.config.KnownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("load SSH known_hosts: %w", err)
		}
	}
	return &ssh.ClientConfig{User: t.config.User, Auth: auth, HostKeyCallback: callback, Timeout: t.config.Timeout}, nil
}

func (t *SSHTransport) connect(ctx context.Context) (*ssh.Client, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.client != nil {
		return t.client, nil
	}
	config, err := t.clientConfig()
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: t.config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", t.config.Address)
	if err != nil {
		return nil, err
	}
	clientConn, chans, reqs, err := ssh.NewClientConn(conn, t.config.Address, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	t.client = ssh.NewClient(clientConn, chans, reqs)
	return t.client, nil
}

func (t *SSHTransport) Execute(ctx context.Context, request CommandRequest) (CommandResult, error) {
	result, err := t.StreamExecute(ctx, request, func(string, string) error { return nil })
	return result, err
}

func (t *SSHTransport) StreamExecute(ctx context.Context, request CommandRequest, emit func(string, string) error) (CommandResult, error) {
	if request.Command == "" {
		return CommandResult{}, errors.New("missing command")
	}
	client, err := t.connect(ctx)
	if err != nil {
		return CommandResult{}, err
	}
	session, err := client.NewSession()
	if err != nil {
		t.Close()
		return CommandResult{}, err
	}
	defer session.Close()
	stdout, err := session.StdoutPipe()
	if err != nil {
		return CommandResult{}, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return CommandResult{}, err
	}
	command := remoteCommand(request)
	if err := session.Start(command); err != nil {
		return CommandResult{}, err
	}
	var wg sync.WaitGroup
	var streamErr error
	var mu sync.Mutex
	copyStream := func(name string, reader io.Reader) {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, readErr := reader.Read(buf)
			if n > 0 {
				mu.Lock()
				if streamErr == nil {
					streamErr = emit(name, string(buf[:n]))
				}
				mu.Unlock()
			}
			if readErr != nil {
				if readErr != io.EOF && streamErr == nil {
					streamErr = readErr
				}
				return
			}
		}
	}
	wg.Add(2)
	go copyStream("stdout", stdout)
	go copyStream("stderr", stderr)
	done := make(chan error, 1)
	go func() { done <- session.Wait() }()
	select {
	case err = <-done:
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return CommandResult{}, ctx.Err()
	}
	wg.Wait()
	result := CommandResult{}
	if exitErr, ok := err.(*ssh.ExitError); ok {
		result.ExitCode = exitErr.ExitStatus()
	} else if err != nil {
		result.ExitCode = 1
	}
	if streamErr != nil {
		return result, streamErr
	}
	if err != nil {
		return result, fmt.Errorf("SSH command failed: %w", err)
	}
	return result, nil
}

func (t *SSHTransport) Ping(ctx context.Context) error {
	_, err := t.Execute(ctx, CommandRequest{Command: "true"})
	return err
}

func (t *SSHTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.client == nil {
		return nil
	}
	err := t.client.Close()
	t.client = nil
	return err
}

func remoteCommand(request CommandRequest) string {
	parts := []string{shellQuote(request.Command)}
	for _, arg := range request.Args {
		parts = append(parts, shellQuote(arg))
	}
	command := strings.Join(parts, " ")
	if request.WorkingDir != "" {
		command = "cd -- " + shellQuote(request.WorkingDir) + " && " + command
	}
	return command
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

// RegisterSSHTransport creates and registers an SSH transport for a device.
// Credentials remain in the transport and are never copied into Device.
func RegisterSSHTransport(id string, config SSHConfig) error {
	transport, err := NewSSHTransport(config)
	if err != nil {
		return err
	}
	RegisterTransport(id, transport)
	return nil
}

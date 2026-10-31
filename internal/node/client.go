package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// FilesClient talks to one shared project on a saved NODE.
type FilesClient struct {
	HTTP         *HTTPTransport
	WebSocketURL string
	Token        string
	Project      string
}

// Connect checks /v1/node/health, saves the device, stores the token
// outside the device record, and registers a live transport in this
// process.
func Connect(ctx context.Context, address, token, name string) (Device, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	httpBase, wsURL, hostport, err := parseDial(address)
	if err != nil {
		return Device{}, err
	}
	health, err := fetchHealth(ctx, httpBase, token)
	if err != nil {
		return Device{}, err
	}
	if health.Protocol != ProtocolVersion {
		return Device{}, fmt.Errorf("node protocol %s, this client speaks %s", health.Protocol, ProtocolVersion)
	}
	id := firstNonEmpty(health.ID, hostport)
	if name == "" {
		name = defaultManager.nicknameOf(id)
	}
	device := Device{
		ID:             id,
		Name:           firstNonEmpty(health.Name, id),
		Nickname:       name,
		ConnectionType: "node",
		Address:        hostport,
		Status:         DeviceStatusOnline,
		LastSeen:       time.Now().UTC(),
		TransportURL:   wsURL,
	}
	defaultManager.Upsert(device)
	if err := defaultManager.Save(DefaultDevicesPath()); err != nil {
		return Device{}, err
	}
	if err := SaveToken(DefaultCredentialsPath(), id, token); err != nil {
		return Device{}, err
	}
	rememberToken(id, token)
	if err := RegisterNodeTransport(id, wsURL, token); err != nil {
		return Device{}, err
	}
	if err := PingDevice(ctx, id); err != nil {
		SetDeviceStatus(id, DeviceStatusError)
		return device, fmt.Errorf("health succeeded but the live transport failed: %w", err)
	}
	return device, nil
}

// ProbeSavedNodes dials every saved NODE and updates its status.
// Local rows stay online. A saved token is used for an HTTP health
// check; otherwise the saved address is probed with TCP.
func ProbeSavedNodes(ctx context.Context) {
	_ = AttachSavedTransports()
	for _, device := range Devices() {
		if device.ID == "local" || device.ConnectionType == "local" {
			SetDeviceStatus(device.ID, DeviceStatusOnline)
			continue
		}
		if device.TransportURL == "" && device.Address == "" {
			SetDeviceStatus(device.ID, DeviceStatusOffline)
			continue
		}
		if err := probeOne(ctx, device); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || isUnreachable(err) {
				SetDeviceStatus(device.ID, DeviceStatusOffline)
				continue
			}
			SetDeviceStatus(device.ID, DeviceStatusError)
			continue
		}
		SetDeviceStatus(device.ID, DeviceStatusOnline)
	}
}

func probeOne(ctx context.Context, device Device) error {
	token, ok := TokenFor(device.ID)
	if ok && device.TransportURL != "" {
		base, err := HTTPBase(device.TransportURL)
		if err != nil {
			return err
		}
		_, err = fetchHealth(ctx, base, token)
		return err
	}
	if transport := defaultTransportManager.transports[device.ID]; transport != nil {
		return transport.Ping(ctx)
	}
	address := device.Address
	if address == "" {
		return fmt.Errorf("device %s has no address", device.ID)
	}
	dialer := net.Dialer{Timeout: 500 * time.Millisecond}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	return conn.Close()
}

func isUnreachable(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, os.ErrDeadlineExceeded)
}

func fetchHealth(ctx context.Context, httpBase, token string) (healthResponse, error) {
	client := NewHTTPTransport(httpBase, token)
	client.Client = &http.Client{Timeout: 10 * time.Second}
	var health healthResponse
	if err := client.request(ctx, http.MethodGet, "/v1/node/health", nil, &health); err != nil {
		return healthResponse{}, err
	}
	if !health.OK {
		return healthResponse{}, errors.New("node reported unhealthy")
	}
	return health, nil
}

func parseDial(spec string) (httpBase, wsURL, hostport string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", "", errors.New("missing host")
	}
	if strings.Contains(spec, "://") {
		parsed, err := url.Parse(spec)
		if err != nil || parsed.Host == "" {
			return "", "", "", fmt.Errorf("invalid node address %q", spec)
		}
		hostport = parsed.Host
		if parsed.Port() == "" {
			hostport = net.JoinHostPort(parsed.Hostname(), "7777")
		}
		switch parsed.Scheme {
		case "http", "ws":
			return "http://" + hostport, "ws://" + hostport + websocketPath, hostport, nil
		case "https", "wss":
			return "https://" + hostport, "wss://" + hostport + websocketPath, hostport, nil
		default:
			return "", "", "", fmt.Errorf("unsupported node scheme %q", parsed.Scheme)
		}
	}
	if _, _, err := net.SplitHostPort(spec); err != nil {
		hostport = net.JoinHostPort(spec, "7777")
	} else {
		hostport = spec
	}
	return "http://" + hostport, "ws://" + hostport + websocketPath, hostport, nil
}

// HTTPBase converts a saved transport URL into the HTTP origin used
// for health checks and file calls.
func HTTPBase(transportURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(transportURL))
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("invalid node endpoint %q", transportURL)
	}
	switch parsed.Scheme {
	case "ws":
		parsed.Scheme = "http"
	case "wss":
		parsed.Scheme = "https"
	case "http", "https":
	default:
		return "", fmt.Errorf("unsupported node scheme %q", parsed.Scheme)
	}
	parsed.Path = ""
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

// ResolveSaved picks a NODE recorded by Connect or scan. An empty ref
// selects the only remote. The token comes from the credentials file,
// never from nodes.json.
func ResolveSaved(nodeRef string) (Device, string, error) {
	if _, err := LoadDeviceRegistry(); err != nil {
		return Device{}, "", err
	}
	if err := AttachSavedTransports(); err != nil {
		return Device{}, "", err
	}
	var remotes []Device
	for _, device := range Devices() {
		if device.ConnectionType == "local" || device.ID == "local" {
			continue
		}
		remotes = append(remotes, device)
	}
	var match Device
	switch {
	case strings.TrimSpace(nodeRef) == "" && len(remotes) == 1:
		match = remotes[0]
	case strings.TrimSpace(nodeRef) == "" && len(remotes) == 0:
		return Device{}, "", errors.New("no saved NODE. Run `bvr-cli node connect <host:port> --token <token>`")
	case strings.TrimSpace(nodeRef) == "":
		return Device{}, "", errors.New("more than one saved NODE; pass --node")
	default:
		for _, device := range remotes {
			if device.ID == nodeRef || device.Address == nodeRef || device.Name == nodeRef || device.Nickname == nodeRef {
				match = device
				break
			}
		}
	}
	if match.ID == "" {
		return Device{}, "", fmt.Errorf("saved NODE %q not found", nodeRef)
	}
	token, ok := TokenFor(match.ID)
	if !ok {
		return Device{}, "", fmt.Errorf("no token saved for NODE %s. Run `bvr-cli node connect` again", match.ID)
	}
	return match, token, nil
}

// ListProjects asks MAIN which project roots it opted into.
func (t *HTTPTransport) ListProjects(ctx context.Context) ([]ProjectInfo, error) {
	var response struct {
		Projects []ProjectInfo `json:"projects"`
	}
	if err := t.request(ctx, http.MethodGet, "/v1/node/projects", nil, &response); err != nil {
		return nil, err
	}
	return response.Projects, nil
}

// List lists a directory on MAIN.
func (c *FilesClient) List(ctx context.Context, rel string, depth int) (FileList, error) {
	query := url.Values{}
	query.Set("project", c.Project)
	query.Set("path", rel)
	if depth > 0 {
		query.Set("depth", fmt.Sprintf("%d", depth))
	}
	var list FileList
	if err := c.HTTP.request(ctx, http.MethodGet, "/v1/node/files?"+query.Encode(), nil, &list); err != nil {
		return FileList{}, err
	}
	return list, nil
}

// Read reads a file from MAIN.
func (c *FilesClient) Read(ctx context.Context, rel string) (FileContent, error) {
	query := url.Values{}
	query.Set("project", c.Project)
	query.Set("path", rel)
	var body FileContent
	if err := c.HTTP.request(ctx, http.MethodGet, "/v1/node/files/read?"+query.Encode(), nil, &body); err != nil {
		return FileContent{}, err
	}
	return body, nil
}

// Put writes content. It reads the current etag first so two devices
// cannot silently overwrite each other. A missing file is created.
func (c *FilesClient) Put(ctx context.Context, rel string, content []byte) (FileEntry, error) {
	expected := ""
	current, err := c.Read(ctx, rel)
	switch {
	case err == nil:
		expected = current.ETag
	case IsNotFound(err):
		expected = ""
	default:
		return FileEntry{}, err
	}
	return c.Write(ctx, rel, content, expected)
}

// Write sends content with an explicit expected etag.
func (c *FilesClient) Write(ctx context.Context, rel string, content []byte, expected string) (FileEntry, error) {
	payload := map[string]any{
		"project":       c.Project,
		"path":          rel,
		"content":       string(content),
		"expected_etag": expected,
	}
	var entry FileEntry
	if err := c.HTTP.request(ctx, http.MethodPut, "/v1/node/files", payload, &entry); err != nil {
		return FileEntry{}, err
	}
	return entry, nil
}

// Mkdir creates a directory on MAIN.
func (c *FilesClient) Mkdir(ctx context.Context, rel string) (FileEntry, error) {
	payload := map[string]string{"project": c.Project, "path": rel, "kind": "dir"}
	var entry FileEntry
	if err := c.HTTP.request(ctx, http.MethodPost, "/v1/node/files", payload, &entry); err != nil {
		return FileEntry{}, err
	}
	return entry, nil
}

// Remove deletes a file, symlink, or empty directory. Files and
// symlinks send the etag from a depth-0 listing.
func (c *FilesClient) Remove(ctx context.Context, rel string) error {
	list, err := c.List(ctx, rel, 0)
	if err != nil {
		return err
	}
	if len(list.Entries) != 1 {
		return fmt.Errorf("path %q not found", rel)
	}
	payload := map[string]any{"project": c.Project, "path": rel}
	if list.Entries[0].Type != "dir" {
		payload["expected_etag"] = list.Entries[0].ETag
	}
	return c.HTTP.request(ctx, http.MethodDelete, "/v1/node/files", payload, nil)
}

// Rename moves a file inside the project.
func (c *FilesClient) Rename(ctx context.Context, from, to string) (FileEntry, error) {
	payload := map[string]any{"project": c.Project, "from": from, "to": to}
	list, err := c.List(ctx, from, 0)
	if err != nil {
		return FileEntry{}, err
	}
	if len(list.Entries) == 1 && list.Entries[0].Type != "dir" {
		payload["expected_etag"] = list.Entries[0].ETag
	}
	var entry FileEntry
	if err := c.HTTP.request(ctx, http.MethodPost, "/v1/node/files/rename", payload, &entry); err != nil {
		return FileEntry{}, err
	}
	return entry, nil
}

// Watch subscribes to change events until ctx is cancelled.
func (c *FilesClient) Watch(ctx context.Context, rel string) (<-chan FSEvent, error) {
	dialer := websocket.DefaultDialer
	header := http.Header{}
	if c.Token != "" {
		header.Set("Authorization", "Bearer "+c.Token)
	}
	conn, resp, err := dialer.DialContext(ctx, websocketDialURL(c.WebSocketURL), header)
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	frame := WSFrame{
		ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Type: "watch",
		Watch: &WatchRequest{
			ProjectID: c.Project,
			Path:      rel,
		},
	}
	if err := conn.WriteJSON(frame); err != nil {
		_ = conn.Close()
		return nil, err
	}
	events := make(chan FSEvent, 16)
	ready := make(chan error, 1)
	go func() {
		defer close(events)
		defer conn.Close()
		sawReady := false
		for {
			var incoming WSFrame
			if err := conn.ReadJSON(&incoming); err != nil {
				if !sawReady {
					ready <- err
				}
				return
			}
			if incoming.Type == "error" {
				if !sawReady {
					ready <- errors.New(incoming.Error)
				}
				return
			}
			if incoming.Event == nil {
				continue
			}
			if incoming.Event.Op == "ready" && !sawReady {
				sawReady = true
				ready <- nil
				continue
			}
			select {
			case <-ctx.Done():
				return
			case events <- *incoming.Event:
			}
		}
	}()
	select {
	case <-ctx.Done():
		_ = conn.Close()
		return nil, ctx.Err()
	case err := <-ready:
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	return events, nil
}

// IsNotFound reports whether err is a missing path or project.
func IsNotFound(err error) bool {
	var status *RemoteError
	if errors.As(err, &status) {
		return status.Status == http.StatusNotFound
	}
	return false
}

// IsConflict reports an etag mismatch.
func IsConflict(err error) bool {
	var status *RemoteError
	return errors.As(err, &status) && status.Status == http.StatusConflict
}

// RemoteError is an HTTP error from a NODE agent.
type RemoteError struct {
	Status  int
	Body    string
	Code    string
	ETag    string
	Message string
}

func (e *RemoteError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" && e.ETag != "" {
		return e.Code + ": current etag " + e.ETag
	}
	if e.Code != "" {
		return e.Code
	}
	if e.Body != "" {
		return e.Body
	}
	return fmt.Sprintf("node HTTP status %d", e.Status)
}

func decodeRemoteError(status int, body []byte) *RemoteError {
	remote := &RemoteError{Status: status, Body: strings.TrimSpace(string(body))}
	var payload struct {
		Error   string `json:"error"`
		ETag    string `json:"etag"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		remote.Code = payload.Error
		remote.ETag = payload.ETag
		remote.Message = payload.Message
	}
	return remote
}

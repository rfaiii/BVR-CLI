// Package ggwave provides the GGWave audio transport for node pairing.
package ggwave

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// SampleFormat represents the audio sample format.
type SampleFormat int

const (
	SampleFormatInt16 SampleFormat = iota
	SampleFormatFloat32
)

// AudioDevice represents an available audio device.
type AudioDevice struct {
	ID           string
	Name         string
	IsInput      bool
	IsOutput     bool
	IsDefaultIn  bool
	IsDefaultOut bool
	MaxChannels  int
	SampleRates  []int
}

// AudioConfig holds the configuration for audio capture/playback.
type AudioConfig struct {
	SampleRate     int
	Channels       int
	Format         SampleFormat
	BufferSize     int    // frames per buffer
	DeviceID       string // empty for default device
	AutoReconnect  bool   // attempt to reconnect on device error
	ReconnectDelay time.Duration
	MaxReconnects  int
	EnableVAD      bool    // voice activity detection
	VADThreshold   float64 // VAD threshold (0.0-1.0)
	Logger         *slog.Logger
}

// DefaultConfig returns a default audio configuration suitable for GGWave.
func DefaultConfig() AudioConfig {
	return AudioConfig{
		SampleRate:     48000,
		Channels:       1,
		Format:         SampleFormatInt16,
		BufferSize:     1024,
		DeviceID:       "",
		AutoReconnect:  true,
		ReconnectDelay: 500 * time.Millisecond,
		MaxReconnects:  3,
		EnableVAD:      false,
		VADThreshold:   0.01,
		Logger:         nil,
	}
}

// Validate checks the configuration for valid values.
func (c *AudioConfig) Validate() error {
	if c.SampleRate <= 0 {
		return errors.New("sample rate must be positive")
	}
	if c.Channels <= 0 || c.Channels > 2 {
		return errors.New("channels must be 1 or 2")
	}
	if c.BufferSize <= 0 {
		return errors.New("buffer size must be positive")
	}
	if c.Format != SampleFormatInt16 && c.Format != SampleFormatFloat32 {
		return errors.New("unsupported sample format")
	}
	if c.AutoReconnect && c.ReconnectDelay <= 0 {
		return errors.New("reconnect delay must be positive when auto-reconnect enabled")
	}
	if c.MaxReconnects < 0 {
		return errors.New("max reconnects must be non-negative")
	}
	if c.VADThreshold < 0 || c.VADThreshold > 1 {
		return errors.New("VAD threshold must be between 0 and 1")
	}
	return nil
}

// AudioDeviceProvider provides platform-specific device enumeration.
var AudioDeviceProvider func() ([]AudioDevice, error) = defaultAudioDeviceProvider

func defaultAudioDeviceProvider() ([]AudioDevice, error) {
	// This will be overridden by platform-specific implementations
	return nil, errors.New("no audio device provider registered")
}

// RegisterAudioDeviceProvider registers a platform-specific device enumeration function.
func RegisterAudioDeviceProvider(f func() ([]AudioDevice, error)) {
	AudioDeviceProvider = f
}

// ListAudioDevices returns available audio devices.
func ListAudioDevices() ([]AudioDevice, error) {
	return AudioDeviceProvider()
}

// GetDefaultInputDevice returns the default input device.
func GetDefaultInputDevice() (*AudioDevice, error) {
	devices, err := ListAudioDevices()
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].IsDefaultIn {
			return &devices[i], nil
		}
	}
	if len(devices) > 0 {
		for i := range devices {
			if devices[i].IsInput {
				return &devices[i], nil
			}
		}
	}
	return nil, errors.New("no input device available")
}

// GetDefaultOutputDevice returns the default output device.
func GetDefaultOutputDevice() (*AudioDevice, error) {
	devices, err := ListAudioDevices()
	if err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].IsDefaultOut {
			return &devices[i], nil
		}
	}
	if len(devices) > 0 {
		for i := range devices {
			if devices[i].IsOutput {
				return &devices[i], nil
			}
		}
	}
	return nil, errors.New("no output device available")
}

// CaptureSource represents an audio input stream (microphone).
type CaptureSource interface {
	io.Reader
	io.Closer
	Start() error
	Stop() error
	Config() AudioConfig
	Device() *AudioDevice
	Err() error
	// AudioLevel returns the current audio level (RMS) normalized 0.0-1.0
	AudioLevel() float64
	// SetVAD enables/disables voice activity detection
	SetVAD(enabled bool, threshold float64) error
}

// PlaybackSink represents an audio output stream (speakers).
type PlaybackSink interface {
	io.Writer
	io.Closer
	Start() error
	Stop() error
	Config() AudioConfig
	Device() *AudioDevice
	Err() error
	// AudioLevel returns the current output audio level (RMS) normalized 0.0-1.0
	AudioLevel() float64
}

// Transport manages both capture and playback for GGWave pairing.
type Transport struct {
	capture  CaptureSource
	playback PlaybackSink
	config   AudioConfig
	mu       sync.Mutex
	closed   bool

	// Reconnection state
	reconnectMu     sync.Mutex
	reconnectCount  int
	reconnecting    bool
	reconnectCancel context.CancelFunc
}

// NewTransport creates a new GGWave audio transport.
func NewTransport(config AudioConfig) (*Transport, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	capture, err := newCaptureSource(config)
	if err != nil {
		return nil, err
	}

	playback, err := newPlaybackSink(config)
	if err != nil {
		capture.Close()
		return nil, err
	}

	return &Transport{
		capture:  capture,
		playback: playback,
		config:   config,
	}, nil
}

// Capture returns the capture source for reading audio input.
func (t *Transport) Capture() CaptureSource {
	return t.capture
}

// Playback returns the playback sink for writing audio output.
func (t *Transport) Playback() PlaybackSink {
	return t.playback
}

// Start begins audio capture and playback.
func (t *Transport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return errors.New("transport closed")
	}

	return t.startWithRetry()
}

// startWithRetry attempts to start the transport with retry logic.
func (t *Transport) startWithRetry() error {
	maxAttempts := 1
	if t.config.AutoReconnect {
		maxAttempts = t.config.MaxReconnects + 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if t.closed {
			return errors.New("transport closed")
		}

		if err := t.capture.Start(); err != nil {
			lastErr = err
			if t.config.Logger != nil {
				t.config.Logger.Warn("capture start failed", "attempt", attempt+1, "error", err)
			}
			if attempt < maxAttempts-1 {
				t.sleepWithContext(t.config.ReconnectDelay)
				continue
			}
			return fmt.Errorf("capture start failed after %d attempts: %w", maxAttempts, lastErr)
		}

		if err := t.playback.Start(); err != nil {
			lastErr = err
			t.capture.Stop()
			if t.config.Logger != nil {
				t.config.Logger.Warn("playback start failed", "attempt", attempt+1, "error", err)
			}
			if attempt < maxAttempts-1 {
				t.sleepWithContext(t.config.ReconnectDelay)
				continue
			}
			return fmt.Errorf("playback start failed after %d attempts: %w", maxAttempts, lastErr)
		}

		return nil
	}

	return fmt.Errorf("start failed after %d attempts: %w", maxAttempts, lastErr)
}

func (t *Transport) sleepWithContext(delay time.Duration) {
	select {
	case <-time.After(delay):
	case <-time.After(0): // Allow immediate continuation if needed
	}
}

// Stop stops audio capture and playback.
func (t *Transport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.playback.Stop()
	t.capture.Stop()
	return nil
}

// Close releases all resources.
func (t *Transport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()

	t.Stop()
	t.playback.Close()
	t.capture.Close()
	return nil
}

// IsClosed returns true if the transport has been closed.
func (t *Transport) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

// Config returns the transport configuration.
func (t *Transport) Config() AudioConfig {
	return t.config
}

// PairingSessionConfig holds configuration for a pairing session.
type PairingSessionConfig struct {
	// Timeout for the entire pairing operation.
	Timeout time.Duration
	// Deadline for the pairing operation (alternative to Timeout).
	Deadline time.Time
	// Minimum frames to accumulate before attempting decode.
	MinFrames int
	// Maximum frames to accumulate before giving up.
	MaxFrames int
	// Enable metrics collection for this session.
	EnableMetrics bool
}

// DefaultPairingSessionConfig returns default session configuration.
func DefaultPairingSessionConfig() PairingSessionConfig {
	return PairingSessionConfig{
		Timeout:       30 * time.Second,
		MinFrames:     1024,      // One GGWave frame
		MaxFrames:     1024 * 10, // 10 frames max
		EnableMetrics: true,
	}
}

// PairingSession handles a single GGWave pairing exchange.
type PairingSession struct {
	transport *Transport
	ctx       context.Context
	cancel    context.CancelFunc
	config    PairingSessionConfig
	mu        sync.Mutex
	closed    bool
	// Metrics
	startTime      time.Time
	framesReceived int64
	bytesReceived  int64
}

// NewPairingSession creates a new pairing session with default configuration.
func NewPairingSession(transport *Transport) *PairingSession {
	return NewPairingSessionWithConfig(transport, DefaultPairingSessionConfig())
}

// NewPairingSessionWithConfig creates a new pairing session with custom configuration.
func NewPairingSessionWithConfig(transport *Transport, config PairingSessionConfig) *PairingSession {
	ctx, cancel := context.WithCancel(context.Background())

	// Timeout and deadline layers each own a timer that only stops when their
	// cancel func is called. Discarding those funcs leaks the timer for the
	// lifetime of the process, and leaves the deadline armed after the session
	// is cancelled. Collect them so the session releases all of them together.
	var layers []context.CancelFunc

	if config.Timeout > 0 {
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, config.Timeout)
		ctx, layers = timeoutCtx, append(layers, timeoutCancel)
	}

	if !config.Deadline.IsZero() {
		deadlineCtx, deadlineCancel := context.WithDeadline(ctx, config.Deadline)
		ctx, layers = deadlineCtx, append(layers, deadlineCancel)
	}

	// Cancel innermost layers first so their timers stop before the parent is
	// torn down, then release the parent.
	cancelAll := func() {
		for _, layer := range layers {
			layer()
		}
		cancel()
	}

	s := &PairingSession{
		transport: transport,
		ctx:       ctx,
		cancel:    cancelAll,
		config:    config,
		startTime: time.Now(),
	}

	// Record metrics
	if config.EnableMetrics {
		GlobalMetrics().RecordSessionCreate()
	}

	return s
}

// Transmit sends a pairing envelope over audio.
func (s *PairingSession) Transmit(waveform []int16) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors.New("session closed")
	}
	s.mu.Unlock()

	start := time.Now()

	// Write the waveform to the playback sink
	n, err := s.transport.Playback().Write(int16ToBytes(waveform))
	if err != nil {
		if s.config.EnableMetrics {
			GlobalMetrics().RecordSessionError()
		}
		return err
	}

	// Record metrics
	if s.config.EnableMetrics {
		GlobalMetrics().RecordBytesTransmitted(int64(n))
		GlobalMetrics().RecordFramesTransmitted(int64(len(waveform)))
		GlobalMetrics().RecordLatency(time.Since(start))
	}

	return nil
}

// Receive listens for a pairing envelope from audio.
func (s *PairingSession) Receive(timeout time.Duration) ([]int16, error) {
	// Use session context with optional additional timeout
	ctx := s.ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(s.ctx, timeout)
		// Note: we don't defer cancel here because we want the timeout
		// to be relative to the session start, not this call.
		// The context will be cancelled when the parent is cancelled.
		_ = cancel
	}

	// Read audio data from capture source
	buffer := make([]byte, s.transport.config.BufferSize*2) // int16 = 2 bytes

	// We need to accumulate enough data for GGWave decoding
	// GGWave needs at least one frame (1024 samples) but preferably more
	var accumulated []int16
	minFrames := s.config.MinFrames
	if minFrames <= 0 {
		minFrames = s.transport.config.BufferSize
	}
	maxFrames := s.config.MaxFrames
	if maxFrames <= 0 {
		maxFrames = s.transport.config.BufferSize * 10
	}

	receiveStart := time.Now()

	for {
		select {
		case <-ctx.Done():
			if s.config.EnableMetrics {
				GlobalMetrics().RecordSessionError()
			}
			return nil, ctx.Err()
		default:
			n, err := s.transport.Capture().Read(buffer)
			if err != nil {
				if s.config.EnableMetrics {
					GlobalMetrics().RecordSessionError()
				}
				return nil, err
			}

			if n == 0 {
				continue
			}

			samples := bytesToInt16(buffer[:n])
			accumulated = append(accumulated, samples...)
			s.framesReceived += int64(len(samples))
			s.bytesReceived += int64(n)

			// Check if we have enough data for decoding
			if len(accumulated) >= minFrames {
				if s.config.EnableMetrics {
					GlobalMetrics().RecordSessionComplete()
					GlobalMetrics().RecordBytesReceived(s.bytesReceived)
					GlobalMetrics().RecordFramesReceived(s.framesReceived)
					GlobalMetrics().RecordLatency(time.Since(receiveStart))
				}
				return accumulated, nil
			}

			// Check max frames limit
			if len(accumulated) >= maxFrames {
				if s.config.EnableMetrics {
					GlobalMetrics().RecordSessionError()
				}
				return nil, errors.New("max frames exceeded without successful decode")
			}
		}
	}
}

// Cancel cancels the pairing session.
func (s *PairingSession) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		s.cancel()
	}
}

// IsClosed returns true if the session has been cancelled.
func (s *PairingSession) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func int16ToBytes(samples []int16) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		buf[i*2] = byte(s)
		buf[i*2+1] = byte(s >> 8)
	}
	return buf
}

func bytesToInt16(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(buf[i*2]) | int16(buf[i*2+1])<<8
	}
	return samples
}

// ErrNoCaptureSource is returned when no capture source is registered.
var ErrNoCaptureSource = errors.New("no audio capture source registered")

// ErrNoPlaybackSink is returned when no playback sink is registered.
var ErrNoPlaybackSink = errors.New("no audio playback sink registered")

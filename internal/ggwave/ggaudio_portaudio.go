//go:build ggwave && cgo

package ggwave

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync"

	"github.com/gordonklaus/portaudio"
)

// safeExecute executes a function with panic recovery and returns any error.
func safeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("panic recovered: %v\n%s", r, buf)
		}
	}()
	return fn()
}

// portaudioCapture implements CaptureSource using PortAudio (blocking I/O).
type portaudioCapture struct {
	stream       *portaudio.Stream
	config       AudioConfig
	device       *AudioDevice
	buffer       []int16
	mu           sync.Mutex
	cond         *sync.Cond
	running      bool
	err          error
	logger       *slog.Logger
	vadEnabled   bool
	vadThreshold float64
}

// portaudioPlayback implements PlaybackSink using PortAudio (blocking I/O).
type portaudioPlayback struct {
	stream  *portaudio.Stream
	config  AudioConfig
	device  *AudioDevice
	buffer  []int16
	mu      sync.Mutex
	cond    *sync.Cond
	running bool
	err     error
	logger  *slog.Logger
}

// PortAudio requires a one-time process-wide Initialize before any stream is
// opened. This used to run from an init() function, which cannot report an
// error and fired on package import even when no audio was ever requested. It
// is now lazy and memoized, so the failure reaches NewTransport as a normal
// error instead of being silently swallowed.
var (
	portaudioInitOnce sync.Once
	portaudioInitErr  error
)

func initPortAudio() error {
	portaudioInitOnce.Do(func() {
		portaudioInitErr = portaudio.Initialize()
	})
	return portaudioInitErr
}

func portaudioDevices() ([]AudioDevice, error) {
	if err := initPortAudio(); err != nil {
		return nil, fmt.Errorf("initialize portaudio: %w", err)
	}

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("list portaudio devices: %w", err)
	}

	defaultIn, _ := portaudio.DefaultInputDevice()
	defaultOut, _ := portaudio.DefaultOutputDevice()

	result := make([]AudioDevice, 0, len(devices))
	for _, d := range devices {
		result = append(result, AudioDevice{
			ID:           d.Name,
			Name:         d.Name,
			IsInput:      d.MaxInputChannels > 0,
			IsOutput:     d.MaxOutputChannels > 0,
			IsDefaultIn:  defaultIn != nil && d.Name == defaultIn.Name,
			IsDefaultOut: defaultOut != nil && d.Name == defaultOut.Name,
			MaxChannels:  max(d.MaxInputChannels, d.MaxOutputChannels),
			SampleRates:  []int{int(d.DefaultSampleRate)},
		})
	}
	return result, nil
}

func init() {
	RegisterAudioDeviceProvider(portaudioDevices)
}

func newCaptureSource(config AudioConfig) (CaptureSource, error) {
	if err := initPortAudio(); err != nil {
		return nil, fmt.Errorf("initialize portaudio: %w", err)
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	c := &portaudioCapture{
		config:       config,
		buffer:       make([]int16, 0, config.BufferSize*2),
		logger:       logger.With("component", "capture"),
		vadEnabled:   config.EnableVAD,
		vadThreshold: config.VADThreshold,
	}
	c.cond = sync.NewCond(&c.mu)

	// Get default input device
	defaultIn, err := GetDefaultInputDevice()
	if err != nil {
		return nil, fmt.Errorf("get default input device: %w", err)
	}
	c.device = defaultIn

	// Open input stream with int16 buffer
	inBuf := make([]int16, config.BufferSize)
	stream, err := portaudio.OpenDefaultStream(config.Channels, 0, float64(config.SampleRate), config.BufferSize, &inBuf)
	if err != nil {
		return nil, fmt.Errorf("open input stream: %w", err)
	}
	c.stream = stream
	c.logger.Debug("opened capture stream", "device", c.device.Name, "sample_rate", config.SampleRate, "channels", config.Channels)
	return c, nil
}

func newPlaybackSink(config AudioConfig) (PlaybackSink, error) {
	if err := initPortAudio(); err != nil {
		return nil, fmt.Errorf("initialize portaudio: %w", err)
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	p := &portaudioPlayback{
		config: config,
		buffer: make([]int16, 0, config.BufferSize*2),
		logger: logger.With("component", "playback"),
	}
	p.cond = sync.NewCond(&p.mu)

	// Get default output device
	defaultOut, err := GetDefaultOutputDevice()
	if err != nil {
		return nil, fmt.Errorf("get default output device: %w", err)
	}
	p.device = defaultOut

	// Open output stream with int16 buffer
	outBuf := make([]int16, config.BufferSize)
	stream, err := portaudio.OpenDefaultStream(0, config.Channels, float64(config.SampleRate), config.BufferSize, &outBuf)
	if err != nil {
		return nil, fmt.Errorf("open output stream: %w", err)
	}
	p.stream = stream
	p.logger.Debug("opened playback stream", "device", p.device.Name, "sample_rate", config.SampleRate, "channels", config.Channels)
	return p, nil
}

func (c *portaudioCapture) Start() error {
	return safeExecute(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.err != nil {
			return c.err
		}
		c.running = true
		if err := c.stream.Start(); err != nil {
			c.err = err
			c.running = false
			c.logger.Error("failed to start capture stream", "error", err)
			return err
		}
		c.logger.Debug("started capture stream")
		return nil
	})
}

func (c *portaudioCapture) Stop() error {
	return safeExecute(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()
		if !c.running {
			return nil
		}
		c.running = false
		err := c.stream.Stop()
		if err != nil {
			c.err = err
			c.logger.Error("failed to stop capture stream", "error", err)
			return err
		}
		c.logger.Debug("stopped capture stream")
		return nil
	})
}

func (c *portaudioCapture) Read(p []byte) (int, error) {
	var result int
	var resultErr error
	err := safeExecute(func() error {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.err != nil {
			return c.err
		}

		for len(c.buffer) == 0 && c.running && c.err == nil {
			// Read from PortAudio stream
			if err := c.stream.Read(); err != nil {
				c.err = err
				c.logger.Error("capture stream read error", "error", err)
				return err
			}

			// Note: The actual buffer management depends on how the stream was opened.
			// For blocking I/O with a provided buffer, the data is in the buffer passed to OpenDefaultStream.
			// This is a simplified implementation - a production version would need proper buffer management.
		}

		if c.err != nil {
			return c.err
		}

		// Convert int16 to bytes
		maxSamples := len(p) / 2
		if maxSamples > len(c.buffer) {
			maxSamples = len(c.buffer)
		}

		for i := 0; i < maxSamples; i++ {
			p[i*2] = byte(c.buffer[i])
			p[i*2+1] = byte(c.buffer[i] >> 8)
		}

		c.buffer = c.buffer[maxSamples:]
		result = maxSamples * 2
		return nil
	})
	if err != nil {
		return 0, err
	}
	return result, resultErr
}

func (c *portaudioCapture) Close() error {
	return safeExecute(func() error {
		c.Stop()
		if c.stream != nil {
			if err := c.stream.Close(); err != nil {
				c.logger.Error("failed to close capture stream", "error", err)
			}
			c.stream = nil
		}
		c.logger.Debug("closed capture stream")
		return nil
	})
}

func (c *portaudioCapture) Config() AudioConfig {
	return c.config
}

func (c *portaudioCapture) Device() *AudioDevice {
	return c.device
}

func (c *portaudioCapture) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

// AudioLevel returns the current audio level (RMS) normalized 0.0-1.0
func (c *portaudioCapture) AudioLevel() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) == 0 {
		return 0.0
	}

	var sumSquares float64
	for _, sample := range c.buffer {
		f := float64(sample) / 32768.0
		sumSquares += f * f
	}

	rms := math.Sqrt(sumSquares / float64(len(c.buffer)))
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}

// SetVAD enables/disables voice activity detection
func (c *portaudioCapture) SetVAD(enabled bool, threshold float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if threshold < 0.0 || threshold > 1.0 {
		return errors.New("VAD threshold must be between 0.0 and 1.0")
	}

	c.vadEnabled = enabled
	c.vadThreshold = threshold
	return nil
}

func (c *portaudioCapture) isVADActive() bool {
	if !c.vadEnabled {
		return true // If VAD disabled, always consider active
	}
	return c.AudioLevel() >= c.vadThreshold
}

func (p *portaudioPlayback) Write(data []byte) (int, error) {
	var result int
	err := safeExecute(func() error {
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.err != nil {
			return p.err
		}
		if !p.running {
			return errors.New("playback not running")
		}

		// Convert bytes to int16
		samples := make([]int16, len(data)/2)
		for i := 0; i < len(samples); i++ {
			samples[i] = int16(data[i*2]) | int16(data[i*2+1])<<8
		}

		p.buffer = append(p.buffer, samples...)
		p.cond.Signal()
		result = len(data)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (p *portaudioPlayback) Start() error {
	return safeExecute(func() error {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.err != nil {
			return p.err
		}
		p.running = true
		if err := p.stream.Start(); err != nil {
			p.err = err
			p.running = false
			p.logger.Error("failed to start playback stream", "error", err)
			return err
		}
		p.logger.Debug("started playback stream")
		return nil
	})
}

func (p *portaudioPlayback) Stop() error {
	return safeExecute(func() error {
		p.mu.Lock()
		defer p.mu.Unlock()
		if !p.running {
			return nil
		}
		p.running = false
		err := p.stream.Stop()
		if err != nil {
			p.err = err
			p.logger.Error("failed to stop playback stream", "error", err)
			return err
		}
		p.logger.Debug("stopped playback stream")
		return nil
	})
}

func (p *portaudioPlayback) Close() error {
	return safeExecute(func() error {
		p.Stop()
		if p.stream != nil {
			if err := p.stream.Close(); err != nil {
				p.logger.Error("failed to close playback stream", "error", err)
			}
			p.stream = nil
		}
		p.logger.Debug("closed playback stream")
		return nil
	})
}

func (p *portaudioPlayback) Config() AudioConfig {
	return p.config
}

func (p *portaudioPlayback) Device() *AudioDevice {
	return p.device
}

func (p *portaudioPlayback) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// AudioLevel returns the current output audio level (RMS) normalized 0.0-1.0
func (p *portaudioPlayback) AudioLevel() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.buffer) == 0 {
		return 0.0
	}

	var sumSquares float64
	for _, sample := range p.buffer {
		f := float64(sample) / 32768.0
		sumSquares += f * f
	}

	rms := math.Sqrt(sumSquares / float64(len(p.buffer)))
	if rms > 1.0 {
		rms = 1.0
	}
	return rms
}

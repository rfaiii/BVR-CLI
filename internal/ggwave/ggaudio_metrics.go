package ggwave

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects metrics for GGWave audio transport.
type MetricsCollector struct {
	mu sync.RWMutex

	// Transport metrics
	transportStartCount     atomic.Int64
	transportStopCount      atomic.Int64
	transportErrorCount     atomic.Int64
	transportReconnectCount atomic.Int64

	// Session metrics
	sessionCreateCount   atomic.Int64
	sessionCompleteCount atomic.Int64
	sessionErrorCount    atomic.Int64
	sessionCancelCount   atomic.Int64

	// Audio metrics
	bytesTransmitted  atomic.Int64
	bytesReceived     atomic.Int64
	framesTransmitted atomic.Int64
	framesReceived    atomic.Int64
	underrunCount     atomic.Int64
	overrunCount      atomic.Int64

	// Latency metrics (in microseconds)
	totalLatencyUs     atomic.Int64
	latencySampleCount atomic.Int64

	// VAD metrics
	vadActiveCount   atomic.Int64
	vadInactiveCount atomic.Int64

	// Timing
	startTime time.Time
}

var globalMetrics = &MetricsCollector{
	startTime: time.Now(),
}

// GlobalMetrics returns the global metrics collector.
func GlobalMetrics() *MetricsCollector {
	return globalMetrics
}

// RecordTransportStart increments the transport start counter.
func (m *MetricsCollector) RecordTransportStart() {
	m.transportStartCount.Add(1)
}

// RecordTransportStop increments the transport stop counter.
func (m *MetricsCollector) RecordTransportStop() {
	m.transportStopCount.Add(1)
}

// RecordTransportError increments the transport error counter.
func (m *MetricsCollector) RecordTransportError() {
	m.transportErrorCount.Add(1)
}

// RecordTransportReconnect increments the transport reconnect counter.
func (m *MetricsCollector) RecordTransportReconnect() {
	m.transportReconnectCount.Add(1)
}

// RecordSessionCreate increments the session create counter.
func (m *MetricsCollector) RecordSessionCreate() {
	m.sessionCreateCount.Add(1)
}

// RecordSessionComplete increments the session complete counter.
func (m *MetricsCollector) RecordSessionComplete() {
	m.sessionCompleteCount.Add(1)
}

// RecordSessionError increments the session error counter.
func (m *MetricsCollector) RecordSessionError() {
	m.sessionErrorCount.Add(1)
}

// RecordSessionCancel increments the session cancel counter.
func (m *MetricsCollector) RecordSessionCancel() {
	m.sessionCancelCount.Add(1)
}

// RecordBytesTransmitted adds to the bytes transmitted counter.
func (m *MetricsCollector) RecordBytesTransmitted(n int64) {
	m.bytesTransmitted.Add(n)
}

// RecordBytesReceived adds to the bytes received counter.
func (m *MetricsCollector) RecordBytesReceived(n int64) {
	m.bytesReceived.Add(n)
}

// RecordFramesTransmitted adds to the frames transmitted counter.
func (m *MetricsCollector) RecordFramesTransmitted(n int64) {
	m.framesTransmitted.Add(n)
}

// RecordFramesReceived adds to the frames received counter.
func (m *MetricsCollector) RecordFramesReceived(n int64) {
	m.framesReceived.Add(n)
}

// RecordUnderrun increments the underrun counter.
func (m *MetricsCollector) RecordUnderrun() {
	m.underrunCount.Add(1)
}

// RecordOverrun increments the overrun counter.
func (m *MetricsCollector) RecordOverrun() {
	m.overrunCount.Add(1)
}

// RecordLatency records a latency sample in microseconds.
func (m *MetricsCollector) RecordLatency(latency time.Duration) {
	m.totalLatencyUs.Add(latency.Microseconds())
	m.latencySampleCount.Add(1)
}

// RecordVADActive increments the VAD active counter.
func (m *MetricsCollector) RecordVADActive() {
	m.vadActiveCount.Add(1)
}

// RecordVADInactive increments the VAD inactive counter.
func (m *MetricsCollector) RecordVADInactive() {
	m.vadInactiveCount.Add(1)
}

// Snapshot returns a copy of the current metrics.
type MetricsSnapshot struct {
	TransportStartCount     int64
	TransportStopCount      int64
	TransportErrorCount     int64
	TransportReconnectCount int64

	SessionCreateCount   int64
	SessionCompleteCount int64
	SessionErrorCount    int64
	SessionCancelCount   int64

	BytesTransmitted  int64
	BytesReceived     int64
	FramesTransmitted int64
	FramesReceived    int64
	UnderrunCount     int64
	OverrunCount      int64

	AvgLatencyUs       int64
	LatencySampleCount int64

	VADActiveCount   int64
	VADInactiveCount int64

	Uptime time.Duration
}

// Snapshot returns a point-in-time snapshot of metrics.
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	latencyCount := m.latencySampleCount.Load()
	var avgLatency int64
	if latencyCount > 0 {
		avgLatency = m.totalLatencyUs.Load() / latencyCount
	}

	return MetricsSnapshot{
		TransportStartCount:     m.transportStartCount.Load(),
		TransportStopCount:      m.transportStopCount.Load(),
		TransportErrorCount:     m.transportErrorCount.Load(),
		TransportReconnectCount: m.transportReconnectCount.Load(),

		SessionCreateCount:   m.sessionCreateCount.Load(),
		SessionCompleteCount: m.sessionCompleteCount.Load(),
		SessionErrorCount:    m.sessionErrorCount.Load(),
		SessionCancelCount:   m.sessionCancelCount.Load(),

		BytesTransmitted:  m.bytesTransmitted.Load(),
		BytesReceived:     m.bytesReceived.Load(),
		FramesTransmitted: m.framesTransmitted.Load(),
		FramesReceived:    m.framesReceived.Load(),
		UnderrunCount:     m.underrunCount.Load(),
		OverrunCount:      m.overrunCount.Load(),

		AvgLatencyUs:       avgLatency,
		LatencySampleCount: latencyCount,

		VADActiveCount:   m.vadActiveCount.Load(),
		VADInactiveCount: m.vadInactiveCount.Load(),

		Uptime: time.Since(m.startTime),
	}
}

// String returns a human-readable summary of metrics.
func (s MetricsSnapshot) String() string {
	return fmt.Sprintf(
		"GGWave Metrics:\n"+
			"  Transport: starts=%d stops=%d errors=%d reconnects=%d\n"+
			"  Sessions: created=%d completed=%d errors=%d cancelled=%d\n"+
			"  Audio: tx_bytes=%d rx_bytes=%d tx_frames=%d rx_frames=%d\n"+
			"  Issues: underruns=%d overruns=%d\n"+
			"  Latency: avg=%dµs samples=%d\n"+
			"  VAD: active=%d inactive=%d\n"+
			"  Uptime: %s\n",
		s.TransportStartCount, s.TransportStopCount, s.TransportErrorCount, s.TransportReconnectCount,
		s.SessionCreateCount, s.SessionCompleteCount, s.SessionErrorCount, s.SessionCancelCount,
		s.BytesTransmitted, s.BytesReceived, s.FramesTransmitted, s.FramesReceived,
		s.UnderrunCount, s.OverrunCount,
		s.AvgLatencyUs, s.LatencySampleCount,
		s.VADActiveCount, s.VADInactiveCount,
		s.Uptime.Round(time.Second))
}

// Reset resets all metrics to zero.
func (m *MetricsCollector) Reset() {
	m.transportStartCount.Store(0)
	m.transportStopCount.Store(0)
	m.transportErrorCount.Store(0)
	m.transportReconnectCount.Store(0)

	m.sessionCreateCount.Store(0)
	m.sessionCompleteCount.Store(0)
	m.sessionErrorCount.Store(0)
	m.sessionCancelCount.Store(0)

	m.bytesTransmitted.Store(0)
	m.bytesReceived.Store(0)
	m.framesTransmitted.Store(0)
	m.framesReceived.Store(0)
	m.underrunCount.Store(0)
	m.overrunCount.Store(0)

	m.totalLatencyUs.Store(0)
	m.latencySampleCount.Store(0)

	m.vadActiveCount.Store(0)
	m.vadInactiveCount.Store(0)

	m.startTime = time.Now()
}

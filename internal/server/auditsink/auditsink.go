package auditsink

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AuditEvent holds information that represents an audit internally.
type AuditEvent struct {
	ResourceName string `json:"resourceName"`
}

// AuditSink is the abstraction for various audit sink configurations
// that Flipt will support.
type AuditSink interface {
	SendAudits([]AuditEvent) error
	fmt.Stringer
}

// SinkPublisher holds information about the configured sinks that we are going
// to send audit events to.
type SinkPublisher struct {
	mtx         sync.Mutex
	logger      *zap.Logger
	capacity    int
	sinks       []AuditSink
	auditEvents []AuditEvent
	ticker      *time.Ticker
}

// NewSinkPublisher is the constructor for a Publisher.
func NewSinkPublisher(logger *zap.Logger, capacity int, sinks []AuditSink, tickerDuration time.Duration) *SinkPublisher {
	p := &SinkPublisher{
		logger:      logger,
		capacity:    capacity,
		sinks:       sinks,
		auditEvents: make([]AuditEvent, 0, capacity),
		ticker:      time.NewTicker(tickerDuration),
	}

	go p.flushWhenNecessary()

	return p
}

// flush flushes the buffer to the configured sinks.
func (p *SinkPublisher) flush() {
	copiedEvents := make([]AuditEvent, len(p.auditEvents))
	copy(copiedEvents, p.auditEvents)
	p.auditEvents = p.auditEvents[:0]
	for _, sink := range p.sinks {
		go func(s AuditSink) {
			err := s.SendAudits(copiedEvents)
			if err != nil {
				p.logger.Warn("Failed to send audits to sink", zap.Stringer("sink", s))
			}
		}(sink)
	}
}

// flushWhenNecessary flushes the buffer to the configured sinks if a tick elapses
// and there are elements in the buffer, to prevent things from staying in the buffer
// for an indefinite amount of time.
func (p *SinkPublisher) flushWhenNecessary() {
	for {
		<-p.ticker.C
		p.mtx.Lock()
		p.flush()
		p.mtx.Unlock()
	}
}

// Publish sends audit events over to the configured sinks when the buffer sized is reached.
// The shared state here are the audit events which are initialized when a Publisher is constructed.
//
// This Publish method has to be concurrent-safe, due to the nature of gRPC requests to the server.
func (p *SinkPublisher) Publish(auditEvent *AuditEvent) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if len(p.auditEvents) >= p.capacity {
		p.flush()
	}

	p.auditEvents = append(p.auditEvents, *auditEvent)
}

// Close releases all the resources for the Publisher.
func (p *SinkPublisher) Close() {
	p.ticker.Stop()
}

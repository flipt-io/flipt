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

// Publisher holds information about the configured sinks that we are going
// to send audit events to.
type Publisher struct {
	mtx         sync.Mutex
	logger      *zap.Logger
	bufferSize  int
	sinks       []AuditSink
	auditEvents []AuditEvent
	ticker      *time.Ticker
}

// NewPublisher is the constructor for a Publisher.
func NewPublisher(logger *zap.Logger, bufferSize int, sinks []AuditSink, tickerDuration time.Duration) *Publisher {
	p := &Publisher{
		logger:      logger,
		bufferSize:  bufferSize,
		sinks:       sinks,
		auditEvents: make([]AuditEvent, 0),
		ticker:      time.NewTicker(tickerDuration),
	}

	go p.flushWhenNecessary()

	return p
}

// flush flushes the buffer to the configured sinks.
func (p *Publisher) flush(auditEvent *AuditEvent) {
	copiedEvents := make([]AuditEvent, len(p.auditEvents))
	copy(copiedEvents, p.auditEvents)
	p.auditEvents = make([]AuditEvent, 0)
	if auditEvent != nil {
		p.auditEvents = append(p.auditEvents, *auditEvent)
	}
	p.mtx.Unlock()
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
func (p *Publisher) flushWhenNecessary() {
	for {
		<-p.ticker.C
		p.mtx.Lock()
		p.flush(nil)
	}
}

// Publish sends audit events over to the configured sinks when the buffer sized is reached.
// The shared state here are the audit events which are initialized when a Publisher is constructed.
//
// This Publish method has to be concurrent-safe, due to the nature of gRPC requests to the server.
func (p *Publisher) Publish(auditEvent *AuditEvent) {
	p.mtx.Lock()
	if len(p.auditEvents) < p.bufferSize {
		defer p.mtx.Unlock()
		p.auditEvents = append(p.auditEvents, *auditEvent)
		return
	}

	p.flush(auditEvent)
}

// Close releases all the resources for the Publisher.
func (p *Publisher) Close() {
	p.ticker.Stop()
}

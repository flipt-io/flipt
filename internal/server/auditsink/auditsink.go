package auditsink

import (
	"fmt"
)

// AuditEvents is a shared package variable that will be mutated during
// RPC requests. This allows us to send audit events to the configured
// sinks asynchronously.
var AuditEvents []*AuditEvent

// AuditEvent holds information that represents an audit internally.
type AuditEvent struct {
	ResourceName string `json:"resourceName"`
}

// AuditSink is the abstraction for various audit sink configurations
// that Flipt will support.
type AuditSink interface {
	SendAudits([]*AuditEvent) error
	fmt.Stringer
}

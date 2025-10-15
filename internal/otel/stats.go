package otel

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

var _ stats.Handler = (*statsHandler)(nil)

type statsHandler struct {
	stats.Handler
}

// HandleRPC processes the RPC stats.
// It ignores client stats to avoid double spans with the same GRPC call.
func (w *statsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	if s.IsClient() {
		return
	}
	w.Handler.HandleRPC(ctx, s)
}

// NewInprocStatsHandler creates a new gRPC stats handler for telemetry.
func NewInprocStatsHandler() stats.Handler {
	return &statsHandler{Handler: otelgrpc.NewServerHandler()}
}

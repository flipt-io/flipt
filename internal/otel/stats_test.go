package otel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/stats"
)

func TestStatsHandler_HandleRPC_ClientStats(t *testing.T) {
	ctx := t.Context()
	mockHandler := NewMockHandler(t)

	handler := &statsHandler{Handler: mockHandler}

	clientStats := &stats.Begin{
		Client:    true,
		BeginTime: time.Now(),
	}

	handler.HandleRPC(ctx, clientStats)

	mockHandler.AssertNotCalled(t, "HandleRPC", mock.Anything, mock.Anything)
}

func TestStatsHandler_HandleRPC_ServerStats(t *testing.T) {
	ctx := t.Context()
	mockHandler := NewMockHandler(t)
	handler := &statsHandler{Handler: mockHandler}

	serverStats := &stats.Begin{
		Client:    false,
		BeginTime: time.Now(),
	}

	mockHandler.On("HandleRPC", ctx, serverStats).Return()

	handler.HandleRPC(ctx, serverStats)

	mockHandler.AssertCalled(t, "HandleRPC", ctx, serverStats)
}

func TestStatsHandler_HandleRPC_MultipleServerCalls(t *testing.T) {
	ctx := t.Context()
	mockHandler := NewMockHandler(t)
	handler := &statsHandler{Handler: mockHandler}

	serverStats1 := &stats.Begin{
		Client:    false,
		BeginTime: time.Now(),
	}
	serverStats2 := &stats.End{
		Client:  false,
		EndTime: time.Now(),
	}

	mockHandler.On("HandleRPC", ctx, serverStats1).Return()
	mockHandler.On("HandleRPC", ctx, serverStats2).Return()

	handler.HandleRPC(ctx, serverStats1)
	handler.HandleRPC(ctx, serverStats2)

	mockHandler.AssertNumberOfCalls(t, "HandleRPC", 2)
}

func TestStatsHandler_HandleRPC_MixedClientServerCalls(t *testing.T) {
	ctx := t.Context()
	mockHandler := NewMockHandler(t)
	handler := &statsHandler{Handler: mockHandler}

	clientStats := &stats.Begin{
		Client:    true,
		BeginTime: time.Now(),
	}

	serverStats := &stats.Begin{
		Client:    false,
		BeginTime: time.Now(),
	}

	mockHandler.On("HandleRPC", ctx, serverStats).Return()

	handler.HandleRPC(ctx, clientStats)

	handler.HandleRPC(ctx, serverStats)

	mockHandler.AssertNumberOfCalls(t, "HandleRPC", 1)
}

func TestStatsHandler_HandleRPC_DifferentStatsTypes(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name         string
		stat         stats.RPCStats
		isClient     bool
		shouldHandle bool
	}{
		{
			name: "client Begin",
			stat: &stats.Begin{
				Client:    true,
				BeginTime: time.Now(),
			},
			isClient:     true,
			shouldHandle: false,
		},
		{
			name: "server Begin",
			stat: &stats.Begin{
				Client:    false,
				BeginTime: time.Now(),
			},
			isClient:     false,
			shouldHandle: true,
		},
		{
			name: "client End",
			stat: &stats.End{
				Client:  true,
				EndTime: time.Now(),
			},
			isClient:     true,
			shouldHandle: false,
		},
		{
			name: "server End",
			stat: &stats.End{
				Client:  false,
				EndTime: time.Now(),
			},
			isClient:     false,
			shouldHandle: true,
		},
		{
			name: "client InPayload",
			stat: &stats.InPayload{
				Client:     true,
				RecvTime:   time.Now(),
				WireLength: 100,
			},
			isClient:     true,
			shouldHandle: false,
		},
		{
			name: "server InPayload",
			stat: &stats.InPayload{
				Client:     false,
				RecvTime:   time.Now(),
				WireLength: 100,
			},
			isClient:     false,
			shouldHandle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)
			handler := &statsHandler{Handler: mockHandler}

			if tt.shouldHandle {
				mockHandler.On("HandleRPC", ctx, tt.stat).Return()
			}

			handler.HandleRPC(ctx, tt.stat)

			if tt.shouldHandle {
				mockHandler.AssertCalled(t, "HandleRPC", ctx, tt.stat)
			} else {
				mockHandler.AssertNotCalled(t, "HandleRPC", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestNewInprocStatsHandler(t *testing.T) {
	handler := NewInprocStatsHandler()
	assert.NotNil(t, handler)
	sh, ok := handler.(*statsHandler)
	assert.True(t, ok, "handler should be of type *statsHandler")
	assert.NotNil(t, sh.Handler)
}

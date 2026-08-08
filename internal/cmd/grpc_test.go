package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAuditLogsRecordedWhenTracingDisabled(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.Default()
	cfg.Database.URL = "file:" + filepath.ToSlash(filepath.Join(tmpDir, "flipt.db"))
	cfg.Audit.Sinks.Log.Enabled = true
	cfg.Audit.Sinks.Log.File = filepath.Join(tmpDir, "audit.log")
	cfg.Server.GRPCPort = 0

	logger := zaptest.NewLogger(t)
	ctx := t.Context()

	ipch := &inprocgrpc.Channel{}

	server, err := NewGRPCServer(ctx, logger, cfg, ipch, info.Flipt{Version: t.Name()}, true)
	require.NoError(t, err)

	startServerChan := make(chan struct{})
	go func() {
		close(startServerChan)
		err := server.Run()
		assert.NoError(t, err)
	}()

	runtime.Gosched()
	<-startServerChan

	conn, err := grpc.NewClient(
		server.ln.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	client := flipt.NewFliptClient(conn)
	_, err = client.CreateFlag(ctx, &flipt.CreateFlagRequest{
		Key:  "test-flag",
		Name: "Test Flag",
		Type: flipt.FlagType_VARIANT_FLAG_TYPE,
	})
	require.NoError(t, err)

	_ = server.Shutdown(ctx)

	data, err := os.ReadFile(cfg.Audit.Sinks.Log.File)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type": "flag"`)
	assert.Contains(t, string(data), `"action": "created"`)
}

package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/environments/evaluation"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap/zaptest"
)

// Test_Environment_SnapshotStoredBeforePublish is a regression test for
// https://github.com/flipt-io/flipt/issues/6588.
//
// A subscriber that is notified of a new snapshot and immediately reads the
// environment's current snapshot (what the GET snapshot endpoint serves) must
// see the snapshot it was just notified about.
func Test_Environment_SnapshotStoredBeforePublish(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := t.Context()
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)

	publisher := evaluation.NewSnapshotPublisher(ctx, logger, evaluation.WithTimeout(5*time.Second))
	env, err := NewEnvironmentFromRepo(ctx, logger, &config.EnvironmentConfig{Name: "production"}, repo, fs.NewStorage(logger), publisher, config.TemplatesConfig{})
	require.NoError(t, err)

	const ns = "team-a"
	// ch1 = the SSE client; ch2 = another (slower) subscriber.
	ch1 := make(chan *rpcevaluation.EvaluationNamespaceSnapshot)
	ch2 := make(chan *rpcevaluation.EvaluationNamespaceSnapshot)
	c1, err := env.EvaluationNamespaceSnapshotSubscribe(ctx, ns, ch1)
	require.NoError(t, err)
	defer c1.Close()
	c2, err := env.EvaluationNamespaceSnapshotSubscribe(ctx, ns, ch2)
	require.NoError(t, err)
	defer c2.Close()

	rev := ""
	if resp, err := env.ListNamespaces(ctx); err == nil && resp != nil {
		rev = resp.Revision
	}
	_, err = env.CreateNamespace(ctx, rev, &rpcenvironments.Namespace{Key: ns, Name: "Team A"})
	require.NoError(t, err)

	type observed struct {
		published string
		served    string
		getErr    error
	}
	done := make(chan observed, 1)
	go func() {
		pub := <-ch1 // the "refetchEvaluation" hint arrives
		// the client immediately GETs the snapshot endpoint
		got, gerr := env.EvaluationNamespaceSnapshot(ctx, ns)
		o := observed{published: pub.Digest, getErr: gerr}
		if got != nil {
			o.served = got.Digest
		}
		<-ch2 // let the other subscriber drain so Publish can return
		done <- o
	}()

	require.NoError(t, env.updateSnapshot(ctx))

	o := <-done
	t.Logf("published digest=%q served digest=%q getErr=%v", o.published, o.served, o.getErr)
	require.NoError(t, o.getErr, "GET after publish should serve the published snapshot")
	assert.Equal(t, o.published, o.served, "GET after publish served a stale snapshot")
}

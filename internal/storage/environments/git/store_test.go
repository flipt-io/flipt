package git

import (
	"context"
	"strings"
	"testing"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage/environments/evaluation"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
)

func Test_Environment_messageForChanges(t *testing.T) {
	for _, test := range []struct {
		name            string
		changes         []storagefs.Change
		tmpl            *template.Template
		expectedMessage string
		expectedErr     error
	}{
		{
			name:        "no changes",
			expectedErr: git.ErrEmptyCommit,
		},
		{
			name: "single change",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbCreate,
					Resource: storagefs.Resource{
						Type: environments.NewResourceType("flipt.config", "Namespace"),
						Key:  "default",
					},
				},
			},
			expectedMessage: `create namespace default`,
		},
		{
			name: "multiple changes",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbCreate,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Flag"),
						Namespace: "default",
						Key:       "someFeature",
					},
				},
				{
					Verb: storagefs.VerbDelete,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Segment"),
						Namespace: "default",
						Key:       "someSegment",
					},
				},
			},
			expectedMessage: `updated multiple resources

create flag default/someFeature
delete segment default/someSegment`,
		},
		{
			name: "single change",
			changes: []storagefs.Change{
				{
					Verb: storagefs.VerbUpdate,
					Resource: storagefs.Resource{
						Type:      environments.NewResourceType("flipt.core", "Flag"),
						Namespace: "default",
						Key:       "some-flag",
					},
				},
			},
			tmpl: template.Must(template.New("commitMessage").Parse(`{{- if eq (len .Changes) 1 }}
        {{- printf "feat(flipt/%s): %s" .Environment.Name (index .Changes 0) }}
        {{- else -}}
        updated multiple resources
        {{ range $change := .Changes }}
        {{- printf "feat(flipt/%s): %s" .Environment.Name $change }}
        {{- end }}
        {{- end }}`)),
			expectedMessage: `feat(flipt/production): update flag default/some-flag`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := &Environment{
				cfg: &config.EnvironmentConfig{
					Name: "production",
				},
			}

			template := test.tmpl
			if template == nil {
				template = storagefs.DefaultFliptConfig().Templates.CommitMessageTemplate
			}

			msg, err := env.messageForChanges(template, test.changes...)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedMessage, msg)
		})
	}
}

//nolint:unparam
func newTestEnvironment(t *testing.T, envName string) *Environment {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)
	storage := fs.NewStorage(logger)
	cfg := &config.EnvironmentConfig{Name: envName}
	env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, evaluation.NoopPublisher)
	require.NoError(t, err)
	return env
}

func Test_Environment_BranchCreation(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	branchEnv, err := env.Branch(ctx, "flipt/production/testbranch")
	require.NoError(t, err)
	assert.NotNil(t, branchEnv)
	assert.NotEqual(t, env.Key(), branchEnv.Key())
	assert.Contains(t, branchEnv.(*Environment).currentBranch, "flipt/production/")
}

func Test_Environment_ListBranches(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// No branches yet
	resp, err := env.ListBranches(ctx)
	require.NoError(t, err)
	assert.Empty(t, resp.Branches)

	// Create a branch
	_, err = env.Branch(ctx, "flipt/production/testbranch")
	require.NoError(t, err)

	resp, err = env.ListBranches(ctx)
	require.NoError(t, err)
	assert.Len(t, resp.Branches, 1)
	assert.Contains(t, resp.Branches[0].Ref, "flipt/production/")
}

func Test_Environment_ProposeNotImplemented(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	_, err := env.Propose(ctx, nil, environments.ProposalOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotImplemented("Propose not implemented in non-commercial version"))
}

func Test_Environment_ListBranchedChangesNotImplemented(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	_, err := env.ListBranchedChanges(ctx, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrNotImplemented("ListBranchedChanges not implemented in non-commercial version"))
}

func Test_Environment_RefreshEnvironment(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Simulate creating a branch directly in the repo
	branchName := "flipt/production/testbranch"
	err := env.repo.CreateBranchIfNotExists(ctx, branchName, storagegit.WithBase(env.currentBranch))
	require.NoError(t, err)

	// Simulate refs map as would be passed by the repository
	references, err := env.repo.References()
	require.NoError(t, err)
	refs := map[string]string{}
	_ = references.ForEach(func(r *plumbing.Reference) error {
		if r.Name().IsRemote() {
			refs[strings.TrimPrefix(r.Name().String(), "refs/remotes/origin/")] = r.Hash().String()
		}
		return nil
	})

	newBranches, err := env.RefreshEnvironment(ctx, refs)
	require.NoError(t, err)
	assert.NotEmpty(t, newBranches)
	found := false
	for _, b := range newBranches {
		if b.Key() == "testbranch" {
			found = true
		}
	}
	assert.True(t, found, "expected to find new branch environment 'testbranch'")
}

func Test_Environment_GetAndListNamespaces(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Initially, only the default namespace should exist
	resp, err := env.ListNamespaces(ctx)
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "default", resp.Items[0].Key)

	// GetNamespace for default
	nsResp, err := env.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, "default", nsResp.Namespace.Key)
}

func Test_Environment_CreateUpdateDeleteNamespace(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Get current revision
	rev := ""
	resp, err := env.ListNamespaces(ctx)
	require.NoError(t, err)
	if resp != nil {
		rev = resp.Revision
	}

	// Create a new namespace
	ns := &rpcenvironments.Namespace{
		Key:         "team-a",
		Name:        "Team A",
		Description: ptr("Team A namespace"),
	}
	newRev, err := env.CreateNamespace(ctx, rev, ns)
	require.NoError(t, err)
	assert.NotEmpty(t, newRev)

	// ListNamespaces should now include the new namespace
	resp, err = env.ListNamespaces(ctx)
	require.NoError(t, err)
	found := false
	for _, n := range resp.Items {
		if n.Key == "team-a" {
			found = true
			assert.Equal(t, "Team A", n.Name)
			assert.Equal(t, "Team A namespace", n.GetDescription())
		}
	}
	assert.True(t, found)

	// Update the namespace
	ns.Name = "Team A Updated"
	ns.Description = ptr("Updated desc")
	updatedRev, err := env.UpdateNamespace(ctx, resp.Revision, ns)
	require.NoError(t, err)
	assert.NotEmpty(t, updatedRev)

	// GetNamespace should reflect the update
	nsResp, err := env.GetNamespace(ctx, "team-a")
	require.NoError(t, err)
	assert.Equal(t, "Team A Updated", nsResp.Namespace.Name)
	assert.Equal(t, "Updated desc", nsResp.Namespace.GetDescription())

	// Delete the namespace
	finalRev, err := env.DeleteNamespace(ctx, nsResp.Revision, "team-a")
	require.NoError(t, err)
	assert.NotEmpty(t, finalRev)

	// ListNamespaces should not include the deleted namespace
	resp, err = env.ListNamespaces(ctx)
	require.NoError(t, err)
	for _, n := range resp.Items {
		assert.NotEqual(t, "team-a", n.Key)
	}
}

// Minimal ResourceType and ResourceStorage for View/Update tests

type testResourceStorage struct {
	getCalled, listCalled, putCalled, deleteCalled bool
}

func (t *testResourceStorage) ResourceType() environments.ResourceType {
	return environments.NewResourceType("test", "Test")
}
func (t *testResourceStorage) GetResource(_ context.Context, _ fs.Filesystem, ns, key string) (*rpcenvironments.Resource, error) {
	t.getCalled = true
	return &rpcenvironments.Resource{NamespaceKey: ns, Key: key}, nil
}
func (t *testResourceStorage) ListResources(_ context.Context, _ fs.Filesystem, ns string) ([]*rpcenvironments.Resource, error) {
	t.listCalled = true
	return []*rpcenvironments.Resource{{NamespaceKey: ns, Key: "foo"}}, nil
}
func (t *testResourceStorage) PutResource(_ context.Context, _ fs.Filesystem, r *rpcenvironments.Resource) error {
	t.putCalled = true
	return nil
}
func (t *testResourceStorage) DeleteResource(_ context.Context, _ fs.Filesystem, ns, key string) error {
	t.deleteCalled = true
	return nil
}

func Test_Environment_ViewAndUpdate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)
	storage := fs.NewStorage(logger)
	cfg := &config.EnvironmentConfig{Name: "production"}
	env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, evaluation.NoopPublisher)
	require.NoError(t, err)

	// Inject test resource storage
	rs := &testResourceStorage{}
	env.storage = fs.NewStorage(logger, rs)

	// View
	err = env.View(ctx, rs.ResourceType(), func(_ context.Context, s environments.ResourceStoreView) error {
		resp, err := s.GetResource(ctx, "default", "foo")
		require.NoError(t, err)
		assert.Equal(t, "foo", resp.Resource.Key)
		return nil
	})
	require.NoError(t, err)
	assert.True(t, rs.getCalled)

	// Update
	resp, err := env.ListNamespaces(ctx)
	require.NoError(t, err)
	rev := resp.Revision

	_, err = env.Update(ctx, rev, rs.ResourceType(), func(_ context.Context, s environments.ResourceStore) error {
		err := s.CreateResource(ctx, &rpcenvironments.Resource{NamespaceKey: "default", Key: "bar"})
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)
	assert.True(t, rs.putCalled)
}

func Test_store_UpdateResource_and_DeleteResource(t *testing.T) {
	ctx := context.Background()
	rs := &testResourceStorage{}
	st := &store{
		typ:    rs.ResourceType(),
		rstore: rs,
		fs:     nil, // not used in mock
	}

	// Test UpdateResource
	resource := &rpcenvironments.Resource{NamespaceKey: "default", Key: "bar"}
	err := st.UpdateResource(ctx, resource)
	require.NoError(t, err)
	assert.True(t, rs.putCalled)
	assert.Len(t, st.changes, 1)
	assert.Equal(t, storagefs.VerbUpdate, st.changes[0].Verb)
	assert.Equal(t, "bar", st.changes[0].Resource.Key)

	// Test DeleteResource
	rs.putCalled = false // reset
	rs.deleteCalled = false
	st.changes = nil
	err = st.DeleteResource(ctx, "default", "baz")
	require.NoError(t, err)
	assert.True(t, rs.deleteCalled)
	assert.Len(t, st.changes, 1)
	assert.Equal(t, storagefs.VerbDelete, st.changes[0].Verb)
	assert.Equal(t, "baz", st.changes[0].Resource.Key)
}

func Test_store_ListResources(t *testing.T) {
	ctx := context.Background()
	rs := &testResourceStorage{}
	st := &store{
		typ:    rs.ResourceType(),
		rstore: rs,
		fs:     nil, // not used in mock
		base:   plumbing.NewHash("deadbeef"),
	}

	resp, err := st.ListResources(ctx, "default")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Resources, 1)
	assert.Equal(t, "foo", resp.Resources[0].Key)
	assert.Equal(t, "deadbeef00000000000000000000000000000000", resp.Revision)
	assert.True(t, rs.listCalled)
}

func Test_NewEnvironmentFromRepo_InitialSnapshot(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	// Create a test repository
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)

	// Create environment configuration
	cfg := &config.EnvironmentConfig{
		Name:    "test",
		Storage: "local",
	}

	// Create environment storage
	storage := fs.NewStorage(logger)

	// Create publisher
	publisher := evaluation.NoopPublisher

	// Create environment from repository
	env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, publisher)
	require.NoError(t, err)

	// Verify that the snapshot was built (not nil) - this is the key issue
	// Even for an empty repository, the snapshot should be initialized
	evaluationStore, err := env.EvaluationStore()
	require.NoError(t, err)
	require.NotNil(t, evaluationStore)

	// The snapshot should be accessible even if empty
	// This verifies the fix - that updateSnapshot() was called during initialization
	snapshot, ok := evaluationStore.(*storagefs.Snapshot)
	require.True(t, ok, "EvaluationStore should return a Snapshot")
	require.NotNil(t, snapshot)
}

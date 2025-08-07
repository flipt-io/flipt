package git

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage/environments/evaluation"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	rpcevaluation "go.flipt.io/flipt/rpc/v2/evaluation"
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

func Test_Environment_Branches(t *testing.T) {
	env := newTestEnvironment(t, "production")

	branches := env.Branches()
	assert.Len(t, branches, 2)
	assert.Contains(t, branches, env.currentBranch)
	assert.Contains(t, branches, "flipt/production/*")
}

func Test_Environment_Key(t *testing.T) {
	env := newTestEnvironment(t, "production")
	assert.Equal(t, "production", env.Key())
}

func Test_Environment_Default(t *testing.T) {
	tests := []struct {
		name      string
		envName   string
		isDefault bool
	}{
		{
			name:      "default environment",
			envName:   "production",
			isDefault: true,
		},
		{
			name:      "non-default environment",
			envName:   "development",
			isDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			ctx := context.Background()
			repo, err := storagegit.NewRepository(ctx, logger)
			require.NoError(t, err)
			storage := fs.NewStorage(logger)
			cfg := &config.EnvironmentConfig{
				Name:    tt.envName,
				Default: tt.isDefault,
			}
			env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, evaluation.NoopPublisher)
			require.NoError(t, err)
			assert.Equal(t, tt.isDefault, env.Default())
		})
	}
}

func Test_Environment_Repository(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)
	storage := fs.NewStorage(logger)
	cfg := &config.EnvironmentConfig{Name: "test"}
	env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, evaluation.NoopPublisher)
	require.NoError(t, err)

	assert.Equal(t, repo, env.Repository())
}

func Test_Environment_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.EnvironmentConfig
		base           string
		currentBranch  string
		repoRemote     string
		expectedConfig func(*rpcenvironments.EnvironmentConfiguration)
	}{
		{
			name: "basic configuration",
			cfg: &config.EnvironmentConfig{
				Name: "test",
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.Equal(t, "main", ec.Ref)
				assert.Nil(t, ec.Remote)
				assert.Nil(t, ec.Directory)
				assert.Nil(t, ec.Base)
				assert.Nil(t, ec.Scm)
			},
		},
		{
			name: "configuration with directory",
			cfg: &config.EnvironmentConfig{
				Name:      "test",
				Directory: "config/flipt",
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.Equal(t, "main", ec.Ref)
				assert.NotNil(t, ec.Directory)
				assert.Equal(t, "config/flipt", *ec.Directory)
			},
		},
		{
			name: "configuration with base",
			cfg: &config.EnvironmentConfig{
				Name: "test",
			},
			base:          "production",
			currentBranch: "flipt/production/feature",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.Equal(t, "flipt/production/feature", ec.Ref)
				assert.NotNil(t, ec.Base)
				assert.Equal(t, "production", *ec.Base)
			},
		},
		{
			name: "configuration with GitHub SCM",
			cfg: &config.EnvironmentConfig{
				Name: "test",
				SCM: &config.SCMConfig{
					Type: config.GitHubSCMType,
				},
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.NotNil(t, ec.Scm)
				assert.Equal(t, rpcenvironments.SCM_SCM_GITHUB, *ec.Scm)
			},
		},
		{
			name: "configuration with GitLab SCM",
			cfg: &config.EnvironmentConfig{
				Name: "test",
				SCM: &config.SCMConfig{
					Type: config.GitLabSCMType,
				},
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.NotNil(t, ec.Scm)
				assert.Equal(t, rpcenvironments.SCM_SCM_GITLAB, *ec.Scm)
			},
		},
		{
			name: "configuration with Gitea SCM",
			cfg: &config.EnvironmentConfig{
				Name: "test",
				SCM: &config.SCMConfig{
					Type: config.GiteaSCMType,
				},
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.NotNil(t, ec.Scm)
				assert.Equal(t, rpcenvironments.SCM_SCM_GITEA, *ec.Scm)
			},
		},
		{
			name: "configuration with Azure SCM",
			cfg: &config.EnvironmentConfig{
				Name: "test",
				SCM: &config.SCMConfig{
					Type: config.AzureSCMType,
				},
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.NotNil(t, ec.Scm)
				assert.Equal(t, rpcenvironments.SCM_SCM_AZURE, *ec.Scm)
			},
		},
		{
			name: "configuration with BitBucket SCM",
			cfg: &config.EnvironmentConfig{
				Name: "test",
				SCM: &config.SCMConfig{
					Type: config.BitBucketSCMType,
				},
			},
			currentBranch: "main",
			expectedConfig: func(ec *rpcenvironments.EnvironmentConfiguration) {
				assert.NotNil(t, ec.Scm)
				assert.Equal(t, rpcenvironments.SCM_SCM_BITBUCKET, *ec.Scm)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			ctx := context.Background()

			var opts []containers.Option[storagegit.Repository]
			if tt.repoRemote != "" {
				opts = append(opts, storagegit.WithRemote("origin", tt.repoRemote))
			}

			repo, err := storagegit.NewRepository(ctx, logger, opts...)
			require.NoError(t, err)

			storage := fs.NewStorage(logger)
			env, err := NewEnvironmentFromRepo(ctx, logger, tt.cfg, repo, storage, evaluation.NoopPublisher)
			require.NoError(t, err)

			env.currentBranch = tt.currentBranch
			env.base = tt.base

			config := env.Configuration()
			tt.expectedConfig(config)
		})
	}
}

func Test_Environment_DeleteBranch(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Create a branch first
	branchEnv, err := env.Branch(ctx, "flipt/production/test-delete")
	require.NoError(t, err)
	assert.NotNil(t, branchEnv)

	// Verify branch exists in branches map
	env.mu.RLock()
	_, exists := env.branches["test-delete"]
	env.mu.RUnlock()
	assert.True(t, exists)

	// Delete the branch
	err = env.DeleteBranch(ctx, "test-delete")
	require.NoError(t, err)

	// Verify branch is removed from branches map
	env.mu.RLock()
	_, exists = env.branches["test-delete"]
	env.mu.RUnlock()
	assert.False(t, exists)
}

func Test_Environment_EvaluationNamespaceSnapshot(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Get namespace snapshot
	_, err := env.EvaluationNamespaceSnapshot(ctx, "default")
	// Will error because empty snapshot, but shouldn't panic
	assert.Error(t, err)
}

func Test_Environment_EvaluationNamespaceSnapshotSubscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	repo, err := storagegit.NewRepository(ctx, logger)
	require.NoError(t, err)
	storage := fs.NewStorage(logger)
	cfg := &config.EnvironmentConfig{Name: "production"}

	// Use a real snapshot publisher instead of NoopPublisher for this test
	publisher := evaluation.NewSnapshotPublisher(logger)
	env, err := NewEnvironmentFromRepo(ctx, logger, cfg, repo, storage, publisher)
	require.NoError(t, err)

	ch := make(chan *rpcevaluation.EvaluationNamespaceSnapshot)
	closer, err := env.EvaluationNamespaceSnapshotSubscribe(ctx, "default", ch)
	require.NoError(t, err)
	assert.NotNil(t, closer)

	// Clean up
	err = closer.Close()
	assert.NoError(t, err)
}

func Test_Environment_Notify(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Test with no hash change
	env.refs[env.currentBranch] = "abc123"
	err := env.Notify(ctx, map[string]string{env.currentBranch: "abc123"})
	require.NoError(t, err)

	// Test with hash change
	err = env.Notify(ctx, map[string]string{env.currentBranch: "def456"})
	require.NoError(t, err)
	assert.Equal(t, "def456", env.refs[env.currentBranch])
}

func Test_Environment_updateSnapshot(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Initial snapshot update
	err := env.updateSnapshot(ctx)
	require.NoError(t, err)

	// Get the current head
	env.mu.RLock()
	initialHead := env.head
	env.mu.RUnlock()

	// Update again with same head (should skip)
	err = env.updateSnapshot(ctx)
	require.NoError(t, err)

	env.mu.RLock()
	secondHead := env.head
	env.mu.RUnlock()

	assert.Equal(t, initialHead, secondHead)
}

func Test_branchEnvIterator(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Create some test branches
	for i := 1; i <= 3; i++ {
		branchName := fmt.Sprintf("flipt/production/test-branch-%d", i)
		err := env.repo.CreateBranchIfNotExists(ctx, branchName, storagegit.WithBase(env.currentBranch))
		require.NoError(t, err)
	}

	// Get branch iterator
	iter, err := env.listBranchEnvs(ctx)
	require.NoError(t, err)

	// Collect all branches
	var branches []*branchEnvConfig
	for cfg := range iter.All() {
		branches = append(branches, cfg)
	}

	// Should have found 3 branches
	assert.Len(t, branches, 3)

	// Check that all branches have correct prefix
	for _, b := range branches {
		assert.Contains(t, b.branch, "flipt/production/")
	}

	// Check for any errors
	assert.NoError(t, iter.Err())
}

func Test_Environment_Branch_GeneratesNameIfEmpty(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Branch with empty name (should generate a random name)
	branchEnv, err := env.Branch(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, branchEnv)

	// Check that a name was generated
	assert.NotEqual(t, "production", branchEnv.Key())
	assert.NotEmpty(t, branchEnv.Key())

	// Check that the branch was created with the correct prefix
	assert.Contains(t, branchEnv.(*Environment).currentBranch, "flipt/production/")
}

func Test_Environment_Branch_WithSpaces(t *testing.T) {
	env := newTestEnvironment(t, "production")
	ctx := context.Background()

	// Branch with just the name including spaces should be trimmed
	branchEnv, err := env.Branch(ctx, "  spaced-branch  ")
	require.NoError(t, err)
	assert.NotNil(t, branchEnv)
	// The Key() returns just the name part without the prefix
	assert.Equal(t, "spaced-branch", branchEnv.Key())
	// The full branch name should be properly constructed
	assert.Equal(t, "flipt/production/spaced-branch", branchEnv.(*Environment).currentBranch)
}

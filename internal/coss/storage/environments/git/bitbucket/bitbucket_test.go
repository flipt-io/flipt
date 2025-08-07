// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package bitbucket

import (
	"context"
	"fmt"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/coss/storage/environments/git"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func TestNewSCM(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		opts    []ClientOption
		wantErr bool
	}{
		{
			name:    "basic creation",
			owner:   "testowner",
			repo:    "testrepo",
			opts:    nil,
			wantErr: false,
		},
		{
			name:  "with custom API URL",
			owner: "testowner",
			repo:  "testrepo",
			opts: []ClientOption{
				WithApiURL("https://bitbucket.company.com/api/v2.0"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			ctx := context.Background()

			scm, err := NewSCM(ctx, logger, tt.owner, tt.repo, tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, scm)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, scm)
				assert.Equal(t, tt.owner, scm.owner)
				assert.Equal(t, tt.repo, scm.repository)
			}
		})
	}
}

func TestSCM_Propose(t *testing.T) {
	tests := []struct {
		name     string
		request  git.ProposalRequest
		mockResp any
		mockErr  error
		want     *environments.EnvironmentProposalDetails
		wantErr  bool
	}{
		{
			name: "successful PR creation",
			request: git.ProposalRequest{
				Base:  "main",
				Head:  "feature-branch",
				Title: "Test PR",
				Body:  "Test description",
				Draft: false,
			},
			mockResp: map[string]any{
				"links": map[string]any{
					"html": map[string]any{
						"href": "https://bitbucket.org/owner/repo/pull-requests/1",
					},
				},
				"state": "OPEN",
			},
			want: &environments.EnvironmentProposalDetails{
				Url:   "https://bitbucket.org/owner/repo/pull-requests/1",
				State: environments.ProposalState_PROPOSAL_STATE_OPEN,
			},
			wantErr: false,
		},
		{
			name: "draft PR creation",
			request: git.ProposalRequest{
				Base:  "main",
				Head:  "feature-branch",
				Title: "Test PR",
				Body:  "Test description",
				Draft: true,
			},
			mockResp: map[string]any{
				"links": map[string]any{
					"html": map[string]any{
						"href": "https://bitbucket.org/owner/repo/pull-requests/2",
					},
				},
				"state": "OPEN",
			},
			want: &environments.EnvironmentProposalDetails{
				Url:   "https://bitbucket.org/owner/repo/pull-requests/2",
				State: environments.ProposalState_PROPOSAL_STATE_OPEN,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRs := NewMockPullRequestsService(t)
			mockCommits := NewMockCommitsService(t)

			// Set up mock expectations
			mockPRs.On("Create", mock.MatchedBy(func(po *bitbucket.PullRequestsOptions) bool {
				// Verify the request parameters
				assert.Equal(t, "testowner", po.Owner)
				assert.Equal(t, "testrepo", po.RepoSlug)
				assert.Equal(t, tt.request.Base, po.DestinationBranch)
				assert.Equal(t, tt.request.Head, po.SourceBranch)
				assert.Equal(t, tt.request.Body, po.Description)

				// Check title - should have [DRAFT] prefix if draft
				expectedTitle := tt.request.Title
				if tt.request.Draft {
					expectedTitle = "[DRAFT] " + tt.request.Title
				}
				assert.Equal(t, expectedTitle, po.Title)

				return true
			})).Return(tt.mockResp, tt.mockErr)

			scm := &SCM{
				logger:     zap.NewNop(),
				owner:      "testowner",
				repository: "testrepo",
				prs:        mockPRs,
				commits:    mockCommits,
			}

			ctx := context.Background()
			got, err := scm.Propose(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSCM_ListChanges(t *testing.T) {
	tests := []struct {
		name     string
		request  git.ListChangesRequest
		mockResp any
		mockErr  error
		want     *environments.ListBranchedEnvironmentChangesResponse
		wantErr  bool
	}{
		{
			name: "successful changes list",
			request: git.ListChangesRequest{
				Base:  "main",
				Head:  "feature-branch",
				Limit: 10,
			},
			mockResp: map[string]any{
				"values": []any{
					map[string]any{
						"hash":    "abc123",
						"message": "feat: add new feature",
						"date":    "2024-01-01T12:00:00Z",
						"links": map[string]any{
							"html": map[string]any{
								"href": "https://bitbucket.org/owner/repo/commits/abc123",
							},
						},
						"author": map[string]any{
							"user": map[string]any{
								"display_name": "John Doe",
							},
						},
					},
					map[string]any{
						"hash":    "def456",
						"message": "fix: bug fix",
						"date":    "2024-01-01T11:00:00Z",
						"links": map[string]any{
							"html": map[string]any{
								"href": "https://bitbucket.org/owner/repo/commits/def456",
							},
						},
						"author": map[string]any{
							"user": map[string]any{
								"display_name": "Jane Smith",
							},
						},
					},
				},
			},
			want: &environments.ListBranchedEnvironmentChangesResponse{
				Changes: []*environments.Change{
					{
						Revision:   "abc123",
						Message:    "feat: add new feature",
						Timestamp:  "2024-01-01T12:00:00Z",
						ScmUrl:     stringPtr("https://bitbucket.org/owner/repo/commits/abc123"),
						AuthorName: stringPtr("John Doe"),
					},
					{
						Revision:   "def456",
						Message:    "fix: bug fix",
						Timestamp:  "2024-01-01T11:00:00Z",
						ScmUrl:     stringPtr("https://bitbucket.org/owner/repo/commits/def456"),
						AuthorName: stringPtr("Jane Smith"),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRs := NewMockPullRequestsService(t)
			mockCommits := NewMockCommitsService(t)

			// Set up mock expectations
			mockCommits.On("GetCommits", mock.MatchedBy(func(cmo *bitbucket.CommitsOptions) bool {
				// Verify the request parameters
				assert.Equal(t, "testowner", cmo.Owner)
				assert.Equal(t, "testrepo", cmo.RepoSlug)
				assert.Equal(t, tt.request.Head, cmo.Include)
				assert.Equal(t, tt.request.Base, cmo.Exclude)
				return true
			})).Return(tt.mockResp, tt.mockErr)

			scm := &SCM{
				logger:     zap.NewNop(),
				owner:      "testowner",
				repository: "testrepo",
				prs:        mockPRs,
				commits:    mockCommits,
			}

			ctx := context.Background()
			got, err := scm.ListChanges(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSCM_ListProposals(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		baseRef  string
		mockResp any
		mockErr  error
		want     map[string]*environments.EnvironmentProposalDetails
		wantErr  bool
	}{
		{
			name:    "successful proposals list",
			envKey:  "production",
			baseRef: "main",
			mockResp: map[string]any{
				"values": []any{
					map[string]any{
						"source": map[string]any{
							"branch": map[string]any{
								"name": "flipt/production/feature-1",
							},
						},
						"destination": map[string]any{
							"branch": map[string]any{
								"name": "main",
							},
						},
						"state": "OPEN",
						"links": map[string]any{
							"html": map[string]any{
								"href": "https://bitbucket.org/owner/repo/pull-requests/1",
							},
						},
					},
					map[string]any{
						"source": map[string]any{
							"branch": map[string]any{
								"name": "flipt/production/feature-2",
							},
						},
						"destination": map[string]any{
							"branch": map[string]any{
								"name": "main",
							},
						},
						"state": "MERGED",
						"links": map[string]any{
							"html": map[string]any{
								"href": "https://bitbucket.org/owner/repo/pull-requests/2",
							},
						},
					},
					// This PR should be filtered out (wrong environment prefix)
					map[string]any{
						"source": map[string]any{
							"branch": map[string]any{
								"name": "flipt/staging/feature-3",
							},
						},
						"destination": map[string]any{
							"branch": map[string]any{
								"name": "main",
							},
						},
						"state": "OPEN",
						"links": map[string]any{
							"html": map[string]any{
								"href": "https://bitbucket.org/owner/repo/pull-requests/3",
							},
						},
					},
				},
			},
			want: map[string]*environments.EnvironmentProposalDetails{
				"flipt/production/feature-1": {
					Url:   "https://bitbucket.org/owner/repo/pull-requests/1",
					State: environments.ProposalState_PROPOSAL_STATE_OPEN,
				},
				"flipt/production/feature-2": {
					Url:   "https://bitbucket.org/owner/repo/pull-requests/2",
					State: environments.ProposalState_PROPOSAL_STATE_MERGED,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRs := NewMockPullRequestsService(t)
			mockCommits := NewMockCommitsService(t)
			mockEnv := serverenvs.NewMockEnvironment(t)

			// Set up mock expectations
			mockPRs.On("Gets", mock.MatchedBy(func(po *bitbucket.PullRequestsOptions) bool {
				assert.Equal(t, "testowner", po.Owner)
				assert.Equal(t, "testrepo", po.RepoSlug)
				assert.Contains(t, po.States, "OPEN")
				assert.Contains(t, po.States, "MERGED")
				assert.Equal(t, fmt.Sprintf(`source.branch.name ~ "flipt/" AND destination.branch.name = "%s"`, tt.baseRef), po.Query)
				return true
			})).Return(tt.mockResp, tt.mockErr)

			mockEnv.On("Key").Return(tt.envKey)
			mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{
				Ref: tt.baseRef,
			})

			scm := &SCM{
				logger:     zap.NewNop(),
				owner:      "testowner",
				repository: "testrepo",
				prs:        mockPRs,
				commits:    mockCommits,
			}

			ctx := context.Background()
			got, err := scm.ListProposals(ctx, mockEnv)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestListProposalsWithMultiplePRs tests ListProposals with multiple PRs (tests automatic pagination)
func TestListProposalsWithMultiplePRs(t *testing.T) {
	mockPRs := NewMockPullRequestsService(t)
	mockCommits := NewMockCommitsService(t)
	mockEnv := serverenvs.NewMockEnvironment(t)

	// Mock a large response that would require pagination in a real scenario
	// The go-bitbucket library handles pagination automatically, so we simulate
	// receiving all results in a single response
	largeMockResp := map[string]any{
		"values": []any{
			// Simulate many PRs to test pagination behavior
			map[string]any{
				"source": map[string]any{
					"branch": map[string]any{
						"name": "flipt/production/feature-1",
					},
				},
				"destination": map[string]any{
					"branch": map[string]any{
						"name": "main",
					},
				},
				"state": "OPEN",
				"links": map[string]any{
					"html": map[string]any{
						"href": "https://bitbucket.org/owner/repo/pull-requests/1",
					},
				},
			},
			map[string]any{
				"source": map[string]any{
					"branch": map[string]any{
						"name": "flipt/production/feature-2",
					},
				},
				"destination": map[string]any{
					"branch": map[string]any{
						"name": "main",
					},
				},
				"state": "MERGED",
				"links": map[string]any{
					"html": map[string]any{
						"href": "https://bitbucket.org/owner/repo/pull-requests/2",
					},
				},
			},
		},
		"size": 2,
		// No "next" field indicates this is the final page
	}

	// Set up mock expectations
	mockPRs.On("Gets", mock.MatchedBy(func(po *bitbucket.PullRequestsOptions) bool {
		// Verify that the correct options are passed to the API call
		assert.Equal(t, "testowner", po.Owner)
		assert.Equal(t, "testrepo", po.RepoSlug)
		assert.Contains(t, po.States, "OPEN")
		assert.Contains(t, po.States, "MERGED")
		assert.Contains(t, po.States, "DECLINED")
		assert.Contains(t, po.States, "SUPERSEDED")
		assert.Equal(t, `source.branch.name ~ "flipt/" AND destination.branch.name = "main"`, po.Query)
		return true
	})).Return(largeMockResp, nil)

	mockEnv.On("Key").Return("production")
	mockEnv.On("Configuration").Return(&environments.EnvironmentConfiguration{
		Ref: "main",
	})

	scm := &SCM{
		logger:     zap.NewNop(),
		owner:      "testowner",
		repository: "testrepo",
		prs:        mockPRs,
		commits:    mockCommits,
	}

	ctx := context.Background()
	result, err := scm.ListProposals(ctx, mockEnv)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify that both PRs were processed correctly
	assert.Contains(t, result, "flipt/production/feature-1")
	assert.Contains(t, result, "flipt/production/feature-2")

	assert.Equal(t, environments.ProposalState_PROPOSAL_STATE_OPEN, result["flipt/production/feature-1"].State)
	assert.Equal(t, environments.ProposalState_PROPOSAL_STATE_MERGED, result["flipt/production/feature-2"].State)
}

func stringPtr(s string) *string {
	return &s
}

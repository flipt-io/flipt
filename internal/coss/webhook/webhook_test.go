package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
)

const testSecret = "s3cr3t"

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	return b
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// auth applies a provider's authentication to a request.
type auth func(r *http.Request, body []byte)

func header(key string, value func(body []byte) string) auth {
	return func(r *http.Request, body []byte) {
		r.Header.Set(key, value(body))
	}
}

func githubSig(secret string) auth {
	return header(headerGitHubSignature, func(b []byte) string { return "sha256=" + sign(secret, b) })
}

func giteaSig(secret string) auth {
	return header(headerGiteaSignature, func(b []byte) string { return sign(secret, b) })
}

func bitbucketSig(secret string) auth {
	return header(headerBitbucketSig, func(b []byte) string { return "sha256=" + sign(secret, b) })
}

func gitlabToken(token string) auth {
	return header(headerGitLabToken, func([]byte) string { return token })
}

func azureBasic(password string) auth {
	return func(r *http.Request, _ []byte) { r.SetBasicAuth("flipt", password) }
}

func noAuth(*http.Request, []byte) {}

type testCase struct {
	name     string
	scm      config.SCMType
	fixture  string
	headers  map[string]string
	auth     auth
	wantErr  func(t *testing.T, err error)
	wantKind Kind
	wantName string
	wantRefs []string
}

func unauthenticated(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errs.AsMatch[errs.ErrUnauthenticated](err), "expected ErrUnauthenticated, got %T: %v", err, err)
}

func runCases(t *testing.T, cases []testCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := readFixture(t, tc.fixture)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/production", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			tc.auth(req, body)

			ev, err := handle(tc.scm, []byte(testSecret), req)
			if tc.wantErr != nil {
				tc.wantErr(t, err)
				assert.Equal(t, Event{}, ev)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantKind, ev.Kind)
			assert.Equal(t, tc.wantName, ev.Name)
			assert.Equal(t, tc.wantRefs, ev.Branches)
		})
	}
}

func TestHandle_GitHub(t *testing.T) {
	pushHeaders := map[string]string{headerGitHubEvent: "push"}

	runCases(t, []testCase{
		{name: "valid push", scm: config.GitHubSCMType, fixture: "github/push.json", headers: pushHeaders, auth: githubSig(testSecret), wantKind: KindPush, wantName: "push", wantRefs: []string{"main"}},
		{name: "tag push has no branches", scm: config.GitHubSCMType, fixture: "github/push_tag.json", headers: pushHeaders, auth: githubSig(testSecret), wantKind: KindPush, wantName: "push"},
		{name: "ping", scm: config.GitHubSCMType, fixture: "github/ping.json", headers: map[string]string{headerGitHubEvent: "ping"}, auth: githubSig(testSecret), wantKind: KindPing, wantName: "ping"},
		{name: "bad signature", scm: config.GitHubSCMType, fixture: "github/push.json", headers: pushHeaders, auth: githubSig("wrong"), wantErr: unauthenticated},
		{name: "missing sha256 prefix", scm: config.GitHubSCMType, fixture: "github/push.json", headers: pushHeaders, auth: giteaSig(testSecret), wantErr: unauthenticated},
		{name: "signature in X-Hub-Signature", scm: config.GitHubSCMType, fixture: "github/push.json", headers: pushHeaders, auth: bitbucketSig(testSecret), wantErr: unauthenticated},
		{name: "missing signature", scm: config.GitHubSCMType, fixture: "github/push.json", headers: pushHeaders, auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_Gitea(t *testing.T) {
	pushHeaders := map[string]string{headerGiteaEvent: "push"}

	runCases(t, []testCase{
		{name: "valid push", scm: config.GiteaSCMType, fixture: "gitea/push.json", headers: pushHeaders, auth: giteaSig(testSecret), wantKind: KindPush, wantName: "push", wantRefs: []string{"main"}},
		{name: "non-push event", scm: config.GiteaSCMType, fixture: "gitea/create.json", headers: map[string]string{headerGiteaEvent: "create"}, auth: giteaSig(testSecret), wantKind: KindOther, wantName: "create"},
		{name: "bad signature", scm: config.GiteaSCMType, fixture: "gitea/push.json", headers: pushHeaders, auth: giteaSig("wrong"), wantErr: unauthenticated},
		{name: "prefixed signature rejected", scm: config.GiteaSCMType, fixture: "gitea/push.json", headers: pushHeaders, auth: header(headerGiteaSignature, func(b []byte) string { return "sha256=" + sign(testSecret, b) }), wantErr: unauthenticated},
		{name: "missing signature", scm: config.GiteaSCMType, fixture: "gitea/push.json", headers: pushHeaders, auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_GitLab(t *testing.T) {
	pushHeaders := map[string]string{headerGitLabEvent: "Push Hook"}

	runCases(t, []testCase{
		{name: "valid push", scm: config.GitLabSCMType, fixture: "gitlab/push.json", headers: pushHeaders, auth: gitlabToken(testSecret), wantKind: KindPush, wantName: "Push Hook", wantRefs: []string{"main"}},
		{name: "tag push hook ignored", scm: config.GitLabSCMType, fixture: "gitlab/tag_push.json", headers: map[string]string{headerGitLabEvent: "Tag Push Hook"}, auth: gitlabToken(testSecret), wantKind: KindOther, wantName: "Tag Push Hook"},
		{name: "wrong token", scm: config.GitLabSCMType, fixture: "gitlab/push.json", headers: pushHeaders, auth: gitlabToken("wrong"), wantErr: unauthenticated},
		{name: "token prefix", scm: config.GitLabSCMType, fixture: "gitlab/push.json", headers: pushHeaders, auth: gitlabToken(testSecret[:3]), wantErr: unauthenticated},
		{name: "missing token", scm: config.GitLabSCMType, fixture: "gitlab/push.json", headers: pushHeaders, auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_BitbucketCloud(t *testing.T) {
	pushHeaders := map[string]string{headerBitbucketEvent: "repo:push"}

	runCases(t, []testCase{
		{name: "valid push skips tag", scm: config.BitBucketSCMType, fixture: "bitbucket-cloud/push.json", headers: pushHeaders, auth: bitbucketSig(testSecret), wantKind: KindPush, wantName: "repo:push", wantRefs: []string{"main"}},
		{name: "non-push event", scm: config.BitBucketSCMType, fixture: "bitbucket-cloud/pullrequest_created.json", headers: map[string]string{headerBitbucketEvent: "pullrequest:created"}, auth: bitbucketSig(testSecret), wantKind: KindOther, wantName: "pullrequest:created"},
		{name: "bad signature", scm: config.BitBucketSCMType, fixture: "bitbucket-cloud/push.json", headers: pushHeaders, auth: bitbucketSig("wrong"), wantErr: unauthenticated},
		{name: "missing signature", scm: config.BitBucketSCMType, fixture: "bitbucket-cloud/push.json", headers: pushHeaders, auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_BitbucketServer(t *testing.T) {
	pushHeaders := map[string]string{headerBitbucketEvent: "repo:refs_changed"}

	runCases(t, []testCase{
		{name: "valid refs changed skips tag", scm: config.BitBucketSCMType, fixture: "bitbucket-server/refs_changed.json", headers: pushHeaders, auth: bitbucketSig(testSecret), wantKind: KindPush, wantName: "repo:refs_changed", wantRefs: []string{"main"}},
		{name: "ping", scm: config.BitBucketSCMType, fixture: "bitbucket-server/ping.json", headers: map[string]string{headerBitbucketEvent: "diagnostics:ping"}, auth: bitbucketSig(testSecret), wantKind: KindPing, wantName: "diagnostics:ping"},
		{name: "bad signature", scm: config.BitBucketSCMType, fixture: "bitbucket-server/refs_changed.json", headers: pushHeaders, auth: bitbucketSig("wrong"), wantErr: unauthenticated},
		{name: "missing signature", scm: config.BitBucketSCMType, fixture: "bitbucket-server/refs_changed.json", headers: pushHeaders, auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_Azure(t *testing.T) {
	runCases(t, []testCase{
		{name: "valid push", scm: config.AzureSCMType, fixture: "azure/push.json", auth: azureBasic(testSecret), wantKind: KindPush, wantName: "git.push", wantRefs: []string{"main"}},
		{name: "non-push event", scm: config.AzureSCMType, fixture: "azure/pullrequest_created.json", auth: azureBasic(testSecret), wantKind: KindOther, wantName: "git.pullrequest.created"},
		{name: "wrong password", scm: config.AzureSCMType, fixture: "azure/push.json", auth: azureBasic("wrong"), wantErr: unauthenticated},
		{name: "secret as username only", scm: config.AzureSCMType, fixture: "azure/push.json", auth: func(r *http.Request, _ []byte) { r.SetBasicAuth(testSecret, "") }, wantErr: unauthenticated},
		{name: "missing authorization", scm: config.AzureSCMType, fixture: "azure/push.json", auth: noAuth, wantErr: unauthenticated},
	})
}

func TestHandle_EmptySecretNeverAuthenticates(t *testing.T) {
	body := readFixture(t, "gitlab/push.json")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(headerGitLabEvent, "Push Hook")
	req.Header.Set(headerGitLabToken, "")

	_, err := handle(config.GitLabSCMType, nil, req)
	unauthenticated(t, err)

	// An HMAC keyed with an empty secret is forgeable, so it must be rejected too.
	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(headerGitHubEvent, "push")
	req.Header.Set(headerGitHubSignature, "sha256="+sign("", body))

	_, err = handle(config.GitHubSCMType, []byte{}, req)
	unauthenticated(t, err)
}

func TestHandle_UnsupportedSCM(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader("{}"))

	_, err := handle(config.SCMType("svn"), []byte(testSecret), req)
	require.Error(t, err)
	assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
}

func TestHandle_BodyTooLarge(t *testing.T) {
	tests := []struct {
		name string
		scm  config.SCMType
		auth auth
	}{
		{name: "github", scm: config.GitHubSCMType, auth: githubSig(testSecret)},
		{name: "gitea", scm: config.GiteaSCMType, auth: giteaSig(testSecret)},
		{name: "bitbucket", scm: config.BitBucketSCMType, auth: bitbucketSig(testSecret)},
		{name: "authenticated gitlab", scm: config.GitLabSCMType, auth: gitlabToken(testSecret)},
		{name: "authenticated azure", scm: config.AzureSCMType, auth: azureBasic(testSecret)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("a"), int(MaxBodySize)+1)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", bytes.NewReader(body))
			tt.auth(req, body)

			_, err := handle(tt.scm, []byte(testSecret), req)
			require.ErrorIs(t, err, ErrBodyTooLarge)

			// a body of exactly MaxBodySize is read in full and reaches the
			// parser, which rejects it as malformed rather than too large
			body = body[:MaxBodySize]
			req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", bytes.NewReader(body))
			tt.auth(req, body)
			// GitHub, Gitea and GitLab only decode push payloads
			req.Header.Set(headerGitHubEvent, "push")
			req.Header.Set(headerGiteaEvent, "push")
			req.Header.Set(headerGitLabEvent, "Push Hook")
			req.Header.Set(headerBitbucketEvent, "repo:push")

			_, err = handle(tt.scm, []byte(testSecret), req)
			require.Error(t, err)
			require.NotErrorIs(t, err, ErrBodyTooLarge)
			assert.True(t, errs.AsMatch[errs.ErrInvalid](err), "expected ErrInvalid, got %T: %v", err, err)
		})
	}
}

// recordingReader records whether the request body was read.
type recordingReader struct {
	read bool
}

func (r *recordingReader) Read([]byte) (int, error) {
	r.read = true
	return 0, errors.New("body must not be read")
}

func TestHandle_HeaderTokenAuthFailureDoesNotReadBody(t *testing.T) {
	tests := []struct {
		name string
		scm  config.SCMType
		auth auth
	}{
		{name: "gitlab wrong token", scm: config.GitLabSCMType, auth: gitlabToken("wrong")},
		{name: "gitlab missing token", scm: config.GitLabSCMType, auth: noAuth},
		{name: "azure wrong password", scm: config.AzureSCMType, auth: azureBasic("wrong")},
		{name: "azure missing authorization", scm: config.AzureSCMType, auth: noAuth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &recordingReader{}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", body)
			tt.auth(req, nil)

			_, err := handle(tt.scm, []byte(testSecret), req)
			unauthenticated(t, err)
			assert.False(t, body.read, "body was read before authenticating")
		})
	}
}

func TestHandle_MalformedAuthenticPayload(t *testing.T) {
	body := []byte(`{"ref":`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(headerGitHubEvent, "push")
	req.Header.Set(headerGitHubSignature, "sha256="+sign(testSecret, body))

	_, err := handle(config.GitHubSCMType, []byte(testSecret), req)
	require.Error(t, err)
	assert.True(t, errs.AsMatch[errs.ErrInvalid](err))
}

func TestParse_BitbucketCloudBranchDeletion(t *testing.T) {
	body := []byte(`{"push":{"changes":[{"old":{"type":"branch","name":"release"},"new":null}]}}`)

	ev, err := parse(config.BitBucketSCMType, http.Header{headerBitbucketEvent: {"repo:push"}}, body)
	require.NoError(t, err)
	assert.Equal(t, Event{Kind: KindPush, Name: "repo:push", Branches: []string{"release"}}, ev)
}

func TestParse_AzureMultipleRefsDeduplicated(t *testing.T) {
	body := []byte(`{"eventType":"git.push","resource":{"refUpdates":[
		{"name":"refs/heads/main"},{"name":"refs/tags/v1"},{"name":"refs/heads/feature/x"},{"name":"refs/heads/main"}]}}`)

	ev, err := parse(config.AzureSCMType, http.Header{}, body)
	require.NoError(t, err)
	assert.Equal(t, []string{"main", "feature/x"}, ev.Branches)
}

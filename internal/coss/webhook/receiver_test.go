package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap/zaptest"
)

const (
	metricRequests    = "flipt_incoming_webhook_requests_total"
	metricFetchErrors = "flipt_incoming_webhook_fetch_errors_total"
	metricSync        = "flipt_incoming_webhook_sync_duration"
)

var testMetricReader *sdkmetric.ManualReader

func TestMain(m *testing.M) {
	testMetricReader = sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(testMetricReader)))
	os.Exit(m.Run())
}

func collect(t *testing.T) metricdata.ResourceMetrics {
	t.Helper()

	var data metricdata.ResourceMetrics
	require.NoError(t, testMetricReader.Collect(t.Context(), &data))

	return data
}

func findMetric(data metricdata.ResourceMetrics, name string) (metricdata.Metrics, bool) {
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m, true
			}
		}
	}

	return metricdata.Metrics{}, false
}

// counterValue returns the value of the named counter for exactly attrs.
func counterValue(t *testing.T, name string, attrs ...attribute.KeyValue) int64 {
	t.Helper()

	m, ok := findMetric(collect(t), name)
	if !ok {
		return 0
	}

	sum, ok := m.Data.(metricdata.Sum[int64])
	require.True(t, ok, "%s is not an int64 sum", name)

	set := attribute.NewSet(attrs...)
	for _, dp := range sum.DataPoints {
		if dp.Attributes.Equals(&set) {
			return dp.Value
		}
	}

	return 0
}

// syncCount returns the number of sync duration observations for exactly
// attrs.
func syncCount(t *testing.T, attrs ...attribute.KeyValue) uint64 {
	t.Helper()

	m, ok := findMetric(collect(t), metricSync)
	if !ok {
		return 0
	}

	hist, ok := m.Data.(metricdata.Histogram[float64])
	require.True(t, ok, "%s is not a float64 histogram", metricSync)

	set := attribute.NewSet(attrs...)
	for _, dp := range hist.DataPoints {
		if dp.Attributes.Equals(&set) {
			return dp.Count
		}
	}

	return 0
}

func requestAttrs(environment string, scm config.SCMType, res result) []attribute.KeyValue {
	return []attribute.KeyValue{
		attrEnvironment.String(environment),
		attrSCM.String(string(scm)),
		attrResult.String(string(res)),
	}
}

func syncAttrs(environment string, scm config.SCMType) []attribute.KeyValue {
	return []attribute.KeyValue{
		attrEnvironment.String(environment),
		attrSCM.String(string(scm)),
	}
}

type provider struct {
	scm     config.SCMType
	fixture string
	headers map[string]string
	auth    func(secret string) auth
}

var providers = map[string]provider{
	"github": {
		scm:     config.GitHubSCMType,
		fixture: "github/push.json",
		headers: map[string]string{headerGitHubEvent: "push"},
		auth:    githubSig,
	},
	"gitlab": {
		scm:     config.GitLabSCMType,
		fixture: "gitlab/push.json",
		headers: map[string]string{headerGitLabEvent: "Push Hook"},
		auth:    gitlabToken,
	},
	"gitea": {
		scm:     config.GiteaSCMType,
		fixture: "gitea/push.json",
		headers: map[string]string{headerGiteaEvent: "push"},
		auth:    giteaSig,
	},
	"bitbucket-cloud": {
		scm:     config.BitBucketSCMType,
		fixture: "bitbucket-cloud/push.json",
		headers: map[string]string{headerBitbucketEvent: "repo:push"},
		auth:    bitbucketSig,
	},
	"bitbucket-server": {
		scm:     config.BitBucketSCMType,
		fixture: "bitbucket-server/refs_changed.json",
		headers: map[string]string{headerBitbucketEvent: "repo:refs_changed"},
		auth:    bitbucketSig,
	},
	"azure": {
		scm:     config.AzureSCMType,
		fixture: "azure/push.json",
		auth:    azureBasic,
	},
}

func newRequest(t *testing.T, environment, fixture string, headers map[string]string, a auth) *http.Request {
	t.Helper()

	body := readFixture(t, fixture)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/"+environment, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	a(req, body)

	return req
}

func serve(r *Receiver, req *http.Request, environment string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeWebhook(rec, req, environment)
	return rec
}

func newTestReceiver(t *testing.T, environment string, scm config.SCMType, repo Repository) *Receiver {
	t.Helper()

	return NewReceiver(t.Context(), zaptest.NewLogger(t), map[string]Target{
		environment: {SCM: scm, Secret: []byte(testSecret), Repository: repo},
	})
}

// TestReceiver_AzurePush covers Azure DevOps end to end through the
// receiver: a push with valid Basic auth fetches the tracked branch once.
func TestReceiver_AzurePush(t *testing.T) {
	const environment = "azure-production"

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Once()
	repo.EXPECT().Fetch(mock.Anything).Return(nil).Once()

	r := newTestReceiver(t, environment, config.AzureSCMType, repo)

	req := newRequest(t, environment, "azure/push.json", nil, azureBasic(testSecret))
	rec := serve(r, req, environment)
	r.Wait()

	assert.Equal(t, http.StatusAccepted, rec.Code)
	repo.AssertExpectations(t)
	repo.AssertNumberOfCalls(t, "Fetch", 1)

	assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, config.AzureSCMType, resultAccepted)...))
	assert.Equal(t, uint64(1), syncCount(t, syncAttrs(environment, config.AzureSCMType)...))
	assert.Equal(t, int64(0), counterValue(t, metricFetchErrors, syncAttrs(environment, config.AzureSCMType)...))
}

func TestReceiver_ValidPushFetchesOnce(t *testing.T) {
	for name, p := range providers {
		t.Run(name, func(t *testing.T) {
			environment := "push-" + name

			repo := NewMockRepository(t)
			repo.EXPECT().Tracks("main").Return(true).Once()
			repo.EXPECT().Fetch(mock.Anything).Return(nil).Once()

			r := newTestReceiver(t, environment, p.scm, repo)

			rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
			r.Wait()

			assert.Equal(t, http.StatusAccepted, rec.Code)
			repo.AssertExpectations(t)
			repo.AssertNumberOfCalls(t, "Fetch", 1)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultAccepted)...))
			assert.Equal(t, uint64(1), syncCount(t, syncAttrs(environment, p.scm)...))
		})
	}
}

func TestReceiver_UnauthenticatedNeverFetches(t *testing.T) {
	for name, p := range providers {
		for _, tc := range []struct {
			name string
			auth auth
		}{
			{name: "wrong secret", auth: p.auth("wrong")},
			{name: "missing auth", auth: noAuth},
		} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				environment := "unauth-" + name

				// no expectations: any Tracks or Fetch call fails the test
				repo := NewMockRepository(t)
				r := newTestReceiver(t, environment, p.scm, repo)

				before := counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultUnauthorized)...)

				rec := serve(r, newRequest(t, environment, p.fixture, p.headers, tc.auth), environment)
				r.Wait()

				assert.Equal(t, http.StatusUnauthorized, rec.Code)
				repo.AssertExpectations(t)
				repo.AssertNotCalled(t, "Fetch", mock.Anything)

				after := counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultUnauthorized)...)
				assert.Equal(t, int64(1), after-before)
				assert.Equal(t, int64(0), counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultAccepted)...))
			})
		}
	}
}

func TestReceiver_UnknownEnvironmentNeverFetches(t *testing.T) {
	repo := NewMockRepository(t)
	r := newTestReceiver(t, "production", config.GitHubSCMType, repo)

	notFound := requestAttrs(unknownLabel, unknownLabel, resultNotFound)
	before := counterValue(t, metricRequests, notFound...)

	p := providers["github"]
	rec := serve(r, newRequest(t, "staging", p.fixture, p.headers, p.auth(testSecret)), "staging")
	r.Wait()

	assert.Equal(t, http.StatusNotFound, rec.Code)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Fetch", mock.Anything)

	assert.Equal(t, int64(1), counterValue(t, metricRequests, notFound...)-before)
	// the requested environment never becomes a label value
	assert.Equal(t, int64(0), counterValue(t, metricRequests, requestAttrs("staging", unknownLabel, resultNotFound)...))
}

// TestReceiver_InvalidNeverFetches asserts malformed, unsupported and
// oversized requests for a configured environment are counted as invalid and
// never fetch.
func TestReceiver_InvalidNeverFetches(t *testing.T) {
	tests := []struct {
		name    string
		scm     config.SCMType
		body    []byte
		headers map[string]string
		auth    auth
		code    int
	}{
		{
			name:    "malformed authentic payload",
			scm:     config.GitHubSCMType,
			body:    []byte(`{"ref":`),
			headers: map[string]string{headerGitHubEvent: "push"},
			auth:    githubSig(testSecret),
			code:    http.StatusBadRequest,
		},
		{
			name: "unsupported scm",
			scm:  config.SCMType("svn"),
			body: []byte(`{}`),
			auth: noAuth,
			code: http.StatusBadRequest,
		},
		{
			name: "body too large",
			scm:  config.GitLabSCMType,
			body: bytes.Repeat([]byte("a"), int(MaxBodySize)+1),
			headers: map[string]string{
				headerGitLabEvent: "Push Hook",
			},
			auth: gitlabToken(testSecret),
			code: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const environment = "invalid"

			// no expectations: any Tracks or Fetch call fails the test
			repo := NewMockRepository(t)
			r := newTestReceiver(t, environment, tt.scm, repo)

			before := counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultInvalid)...)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/"+environment, bytes.NewReader(tt.body))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			tt.auth(req, tt.body)

			rec := serve(r, req, environment)
			r.Wait()

			assert.Equal(t, tt.code, rec.Code)
			repo.AssertExpectations(t)
			repo.AssertNotCalled(t, "Fetch", mock.Anything)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultInvalid)...)-before)
			for _, res := range []result{resultAccepted, resultIgnored, resultUnauthorized} {
				assert.Equal(t, int64(0), counterValue(t, metricRequests, requestAttrs(environment, tt.scm, res)...), res)
			}
		})
	}
}

func TestReceiver_IgnoredEvents(t *testing.T) {
	tests := []struct {
		name    string
		scm     config.SCMType
		fixture string
		headers map[string]string
		auth    auth
		tracks  map[string]bool
	}{
		{
			name:    "github ping",
			scm:     config.GitHubSCMType,
			fixture: "github/ping.json",
			headers: map[string]string{headerGitHubEvent: "ping"},
			auth:    githubSig(testSecret),
		},
		{
			name:    "github tag push",
			scm:     config.GitHubSCMType,
			fixture: "github/push_tag.json",
			headers: map[string]string{headerGitHubEvent: "push"},
			auth:    githubSig(testSecret),
		},
		{
			name:    "azure pull request",
			scm:     config.AzureSCMType,
			fixture: "azure/pullrequest_created.json",
			auth:    azureBasic(testSecret),
		},
		{
			name:    "push to untracked branch",
			scm:     config.GitLabSCMType,
			fixture: "gitlab/push.json",
			headers: map[string]string{headerGitLabEvent: "Push Hook"},
			auth:    gitlabToken(testSecret),
			tracks:  map[string]bool{"main": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const environment = "ignored"

			repo := NewMockRepository(t)
			for branch, tracked := range tt.tracks {
				repo.EXPECT().Tracks(branch).Return(tracked).Once()
			}

			r := newTestReceiver(t, environment, tt.scm, repo)

			before := counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultIgnored)...)

			rec := serve(r, newRequest(t, environment, tt.fixture, tt.headers, tt.auth), environment)
			r.Wait()

			assert.Equal(t, http.StatusOK, rec.Code)
			repo.AssertExpectations(t)
			repo.AssertNotCalled(t, "Fetch", mock.Anything)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultIgnored)...)-before)
		})
	}
}

func TestReceiver_FetchError(t *testing.T) {
	const environment = "fetch-error"

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Once()
	repo.EXPECT().Fetch(mock.Anything).Return(errors.New("connection refused")).Once()

	r := newTestReceiver(t, environment, config.GiteaSCMType, repo)

	p := providers["gitea"]
	rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
	r.Wait()

	assert.Equal(t, http.StatusAccepted, rec.Code)
	repo.AssertExpectations(t)

	assert.Equal(t, int64(1), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GiteaSCMType)...))
	assert.Equal(t, uint64(0), syncCount(t, syncAttrs(environment, config.GiteaSCMType)...))
}

// TestReceiver_CollapsesBursts asserts webhooks arriving while a fetch runs
// are served by a single follow-up fetch rather than parallel fetches.
func TestReceiver_CollapsesBursts(t *testing.T) {
	const (
		environment = "burst"
		burst       = 20
	)

	var (
		started  = make(chan struct{})
		release  = make(chan struct{})
		mu       sync.Mutex
		calls    int
		inFlight int
		maxIn    int
	)

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true)
	repo.EXPECT().Fetch(mock.Anything).RunAndReturn(func(context.Context, ...string) error {
		mu.Lock()
		calls++
		inFlight++
		maxIn = max(maxIn, inFlight)
		first := calls == 1
		mu.Unlock()

		if first {
			close(started)
			<-release
		}

		mu.Lock()
		inFlight--
		mu.Unlock()

		return nil
	})

	r := newTestReceiver(t, environment, config.GitHubSCMType, repo)
	p := providers["github"]

	// the first webhook starts a fetch which blocks until released
	assert.Equal(t, http.StatusAccepted, serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment).Code)

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first fetch never started")
	}

	// a burst arrives while the first fetch is running
	var wg sync.WaitGroup
	for range burst {
		wg.Go(func() {
			rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
			assert.Equal(t, http.StatusAccepted, rec.Code)
		})
	}
	wg.Wait()

	close(release)
	r.Wait()

	repo.AssertNumberOfCalls(t, "Fetch", 2)
	assert.Equal(t, 1, maxIn, "fetches must never run in parallel")

	assert.Equal(t, int64(burst+1), counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...))
	assert.Equal(t, uint64(burst+1), syncCount(t, syncAttrs(environment, config.GitHubSCMType)...))
}

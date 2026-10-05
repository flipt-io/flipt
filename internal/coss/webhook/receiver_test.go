package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

// OTel instrument names. The Prometheus exporter appends unit suffixes, so
// the sync histogram is exported as flipt_incoming_webhook_sync_duration_seconds
// (see TestMetrics_PrometheusExportedNames).
const (
	metricRequests    = "flipt_incoming_webhook_requests_total"
	metricFetchErrors = "flipt_incoming_webhook_fetch_errors_total"
	metricSync        = "flipt_incoming_webhook_sync_duration"
)

var (
	testMetricReader *sdkmetric.ManualReader
	// testPromRegistry holds the same metrics as exported by the Prometheus
	// exporter Flipt uses for metrics.exporter: prometheus.
	testPromRegistry *prometheus.Registry
)

func TestMain(m *testing.M) {
	testMetricReader = sdkmetric.NewManualReader()
	testPromRegistry = prometheus.NewRegistry()

	promExporter, err := otelprom.New(otelprom.WithRegisterer(testPromRegistry))
	if err != nil {
		panic(err)
	}

	otel.SetMeterProvider(sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(testMetricReader),
		sdkmetric.WithReader(promExporter),
	))
	os.Exit(m.Run())
}

// TestMetrics_PrometheusExportedNames checks the series names operators see
// when scraping /metrics through the Prometheus exporter.
func TestMetrics_PrometheusExportedNames(t *testing.T) {
	const environment = "prometheus-names"

	m := newReceiverMetrics()
	m.recordRequest(t.Context(), environment, string(config.GitHubSCMType), resultAccepted)
	m.recordFetchError(t.Context(), environment, string(config.GitHubSCMType))
	m.recordSync(t.Context(), environment, string(config.GitHubSCMType), 300*time.Millisecond)

	families, err := testPromRegistry.Gather()
	require.NoError(t, err)

	var names []string
	for _, f := range families {
		if strings.HasPrefix(f.GetName(), "flipt_incoming_webhook_") {
			names = append(names, f.GetName())
		}
	}

	assert.ElementsMatch(t, []string{
		"flipt_incoming_webhook_requests_total",
		"flipt_incoming_webhook_fetch_errors_total",
		"flipt_incoming_webhook_sync_duration_seconds",
	}, names)

	rec := httptest.NewRecorder()
	promhttp.HandlerFor(testPromRegistry, promhttp.HandlerOpts{}).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	for _, series := range []string{
		"flipt_incoming_webhook_sync_duration_seconds_bucket{",
		"flipt_incoming_webhook_sync_duration_seconds_sum{",
		"flipt_incoming_webhook_sync_duration_seconds_count{",
	} {
		assert.Contains(t, body, series)
	}
	assert.Contains(t, body, `flipt_incoming_webhook_sync_duration_seconds_bucket{environment="prometheus-names"`)
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

	// metrics are global, so assert deltas to stay correct under -count
	var (
		acceptedBefore = counterValue(t, metricRequests, requestAttrs(environment, config.AzureSCMType, resultAccepted)...)
		syncBefore     = syncCount(t, syncAttrs(environment, config.AzureSCMType)...)
		errorsBefore   = counterValue(t, metricFetchErrors, syncAttrs(environment, config.AzureSCMType)...)
	)

	req := newRequest(t, environment, "azure/push.json", nil, azureBasic(testSecret))
	rec := serve(r, req, environment)
	r.Wait()

	assert.Equal(t, http.StatusAccepted, rec.Code)
	repo.AssertExpectations(t)
	repo.AssertNumberOfCalls(t, "Fetch", 1)

	assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, config.AzureSCMType, resultAccepted)...)-acceptedBefore)
	assert.Equal(t, uint64(1), syncCount(t, syncAttrs(environment, config.AzureSCMType)...)-syncBefore)
	assert.Equal(t, int64(0), counterValue(t, metricFetchErrors, syncAttrs(environment, config.AzureSCMType)...)-errorsBefore)
}

func TestReceiver_ValidPushFetchesOnce(t *testing.T) {
	for name, p := range providers {
		t.Run(name, func(t *testing.T) {
			environment := "push-" + name

			repo := NewMockRepository(t)
			repo.EXPECT().Tracks("main").Return(true).Once()
			repo.EXPECT().Fetch(mock.Anything).Return(nil).Once()

			r := newTestReceiver(t, environment, p.scm, repo)

			var (
				acceptedBefore = counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultAccepted)...)
				syncBefore     = syncCount(t, syncAttrs(environment, p.scm)...)
			)

			rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
			r.Wait()

			assert.Equal(t, http.StatusAccepted, rec.Code)
			repo.AssertExpectations(t)
			repo.AssertNumberOfCalls(t, "Fetch", 1)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultAccepted)...)-acceptedBefore)
			assert.Equal(t, uint64(1), syncCount(t, syncAttrs(environment, p.scm)...)-syncBefore)
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

// TestReceiver_HeaderTokenUnauthenticatedSkipsBody asserts a GitLab or Azure
// DevOps request failing authentication gets 401 without its body being read.
func TestReceiver_HeaderTokenUnauthenticatedSkipsBody(t *testing.T) {
	for _, name := range []string{"gitlab", "azure"} {
		p := providers[name]

		t.Run(name, func(t *testing.T) {
			environment := "unread-" + name

			// no expectations: any Tracks or Fetch call fails the test
			repo := NewMockRepository(t)
			r := newTestReceiver(t, environment, p.scm, repo)

			before := counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultUnauthorized)...)

			body := &recordingReader{}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/"+environment, body)
			for k, v := range p.headers {
				req.Header.Set(k, v)
			}
			p.auth("wrong")(req, nil)

			rec := serve(r, req, environment)
			r.Wait()

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, body.read, "body was read before authenticating")
			repo.AssertNotCalled(t, "Fetch", mock.Anything)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultUnauthorized)...)-before)
			assert.Equal(t, int64(0), counterValue(t, metricRequests, requestAttrs(environment, p.scm, resultInvalid)...))
		})
	}
}

// TestReceiver_UnknownEnvironmentNeverFetches asserts a request for an
// environment without a webhook gets the same 401 as an authentication
// failure, without its body being read, and is counted as not_found.
func TestReceiver_UnknownEnvironmentNeverFetches(t *testing.T) {
	repo := NewMockRepository(t)
	r := newTestReceiver(t, "production", config.GitHubSCMType, repo)

	notFound := requestAttrs(unknownLabel, unknownLabel, resultNotFound)
	before := counterValue(t, metricRequests, notFound...)

	p := providers["github"]
	body := &recordingReader{}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/staging", body)
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	rec := serve(r, req, "staging")
	r.Wait()

	// indistinguishable from an authentication failure for a configured environment
	unauthorized := serve(r, newRequest(t, "production", p.fixture, p.headers, p.auth("wrong")), "production")
	r.Wait()

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, unauthorized.Code, rec.Code)
	assert.Equal(t, unauthorized.Body.String(), rec.Body.String())
	assert.False(t, body.read, "body was read for an unknown environment")
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
			name: "hmac body too large",
			scm:  config.GitHubSCMType,
			body: bytes.Repeat([]byte("a"), int(MaxBodySize)+1),
			headers: map[string]string{
				headerGitHubEvent: "push",
			},
			auth: githubSig(testSecret),
			code: http.StatusRequestEntityTooLarge,
		},
		{
			name: "authenticated header-token body too large",
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

	var (
		errorsBefore = counterValue(t, metricFetchErrors, syncAttrs(environment, config.GiteaSCMType)...)
		syncBefore   = syncCount(t, syncAttrs(environment, config.GiteaSCMType)...)
	)

	p := providers["gitea"]
	rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
	r.Wait()

	assert.Equal(t, http.StatusAccepted, rec.Code)
	repo.AssertExpectations(t)

	assert.Equal(t, int64(1), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GiteaSCMType)...)-errorsBefore)
	assert.Equal(t, uint64(0), syncCount(t, syncAttrs(environment, config.GiteaSCMType)...)-syncBefore)
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

	var (
		acceptedBefore = counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...)
		syncBefore     = syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)
	)

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

	assert.Equal(t, int64(burst+1), counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...)-acceptedBefore)
	assert.Equal(t, uint64(burst+1), syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)-syncBefore)
}

// TestReceiver_SharedRepositoryCollapsesBursts asserts that environments
// backed by the same repository share one syncer: a burst of webhooks to
// both while a fetch runs never fetches in parallel and is served by a single
// follow-up fetch, while the requests and sync metrics are recorded for each
// environment separately.
func TestReceiver_SharedRepositoryCollapsesBursts(t *testing.T) {
	const (
		githubEnvironment = "shared-github"
		gitlabEnvironment = "shared-gitlab"
		burst             = 10
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

	r := NewReceiver(t.Context(), zaptest.NewLogger(t), map[string]Target{
		githubEnvironment: {SCM: config.GitHubSCMType, Secret: []byte(testSecret), Repository: repo},
		gitlabEnvironment: {SCM: config.GitLabSCMType, Secret: []byte(testSecret), Repository: repo},
	})
	require.Len(t, r.syncers, 1, "environments sharing a repository must share a syncer")

	var (
		github = providers["github"]
		gitlab = providers["gitlab"]

		githubAcceptedBefore = counterValue(t, metricRequests, requestAttrs(githubEnvironment, config.GitHubSCMType, resultAccepted)...)
		gitlabAcceptedBefore = counterValue(t, metricRequests, requestAttrs(gitlabEnvironment, config.GitLabSCMType, resultAccepted)...)
		githubSyncBefore     = syncCount(t, syncAttrs(githubEnvironment, config.GitHubSCMType)...)
		gitlabSyncBefore     = syncCount(t, syncAttrs(gitlabEnvironment, config.GitLabSCMType)...)
	)

	sendGitHub := func() int {
		return serve(r, newRequest(t, githubEnvironment, github.fixture, github.headers, github.auth(testSecret)), githubEnvironment).Code
	}
	sendGitLab := func() int {
		return serve(r, newRequest(t, gitlabEnvironment, gitlab.fixture, gitlab.headers, gitlab.auth(testSecret)), gitlabEnvironment).Code
	}

	// the first webhook starts a fetch which blocks until released
	assert.Equal(t, http.StatusAccepted, sendGitHub())
	awaitClosed(t, started, "first fetch never started")

	// a burst to both environments arrives while the first fetch is running
	var wg sync.WaitGroup
	for range burst {
		wg.Go(func() { assert.Equal(t, http.StatusAccepted, sendGitHub()) })
		wg.Go(func() { assert.Equal(t, http.StatusAccepted, sendGitLab()) })
	}
	wg.Wait()

	close(release)
	r.Wait()

	repo.AssertNumberOfCalls(t, "Fetch", 2)
	assert.Equal(t, 1, maxIn, "fetches of a shared repository must never run in parallel")

	assert.Equal(t, int64(burst+1), counterValue(t, metricRequests, requestAttrs(githubEnvironment, config.GitHubSCMType, resultAccepted)...)-githubAcceptedBefore)
	assert.Equal(t, int64(burst), counterValue(t, metricRequests, requestAttrs(gitlabEnvironment, config.GitLabSCMType, resultAccepted)...)-gitlabAcceptedBefore)
	assert.Equal(t, uint64(burst+1), syncCount(t, syncAttrs(githubEnvironment, config.GitHubSCMType)...)-githubSyncBefore)
	assert.Equal(t, uint64(burst), syncCount(t, syncAttrs(gitlabEnvironment, config.GitLabSCMType)...)-gitlabSyncBefore)
}

// TestReceiver_BranchDeletionFetchesOnce asserts a push deleting a tracked
// branch is checked with Tracks and schedules exactly one fetch of every
// tracked branch (no branch list is passed to Fetch).
func TestReceiver_BranchDeletionFetchesOnce(t *testing.T) {
	tests := []struct {
		name    string
		scm     config.SCMType
		body    string
		headers map[string]string
		auth    auth
	}{
		{
			name:    "bitbucket cloud new null",
			scm:     config.BitBucketSCMType,
			body:    `{"push":{"changes":[{"old":{"type":"branch","name":"release"},"new":null}]}}`,
			headers: map[string]string{headerBitbucketEvent: "repo:push"},
			auth:    bitbucketSig(testSecret),
		},
		{
			name:    "github deleted",
			scm:     config.GitHubSCMType,
			body:    `{"ref":"refs/heads/release","before":"6113728f27ae82c7b1a177c8d03f9e96e0adf246","after":"0000000000000000000000000000000000000000","created":false,"deleted":true}`,
			headers: map[string]string{headerGitHubEvent: "push"},
			auth:    githubSig(testSecret),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const environment = "branch-deletion"

			repo := NewMockRepository(t)
			repo.EXPECT().Tracks("release").Return(true).Once()
			repo.EXPECT().Fetch(mock.Anything).RunAndReturn(func(_ context.Context, heads ...string) error {
				assert.Empty(t, heads, "a webhook-triggered fetch updates every tracked branch")
				return nil
			}).Once()

			r := newTestReceiver(t, environment, tt.scm, repo)

			var (
				acceptedBefore = counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultAccepted)...)
				syncBefore     = syncCount(t, syncAttrs(environment, tt.scm)...)
			)

			body := []byte(tt.body)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v2/webhooks/"+environment, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			tt.auth(req, body)

			rec := serve(r, req, environment)
			r.Wait()

			assert.Equal(t, http.StatusAccepted, rec.Code)
			repo.AssertExpectations(t)
			repo.AssertNumberOfCalls(t, "Fetch", 1)

			assert.Equal(t, int64(1), counterValue(t, metricRequests, requestAttrs(environment, tt.scm, resultAccepted)...)-acceptedBefore)
			assert.Equal(t, uint64(1), syncCount(t, syncAttrs(environment, tt.scm)...)-syncBefore)
		})
	}
}

// TestReceiver_CancelledContextSkipsFetch asserts that once the receiver's
// context is cancelled (Flipt is shutting down) an authentic push responds
// 503 without scheduling a fetch or counting an accepted request or a fetch
// error.
func TestReceiver_CancelledContextSkipsFetch(t *testing.T) {
	const environment = "cancelled"

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Once()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	r := NewReceiver(ctx, zaptest.NewLogger(t), map[string]Target{
		environment: {SCM: config.GitHubSCMType, Secret: []byte(testSecret), Repository: repo},
	})

	p := providers["github"]
	rec := serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment)
	require.NoError(t, r.Shutdown(t.Context()))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	repo.AssertNotCalled(t, "Fetch", mock.Anything)

	assert.Equal(t, int64(0), counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...))
	assert.Equal(t, int64(0), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...))
}

// awaitClosed fails the test unless ch is closed within 5s.
func awaitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal(what)
	}
}

// TestReceiver_ShutdownCompletesInFlightFetch asserts that cancelling the
// receiver's context neither aborts a running fetch nor drops a webhook
// accepted while it runs: new webhooks get 503, and Shutdown waits for both
// fetches to complete and record their sync durations.
func TestReceiver_ShutdownCompletesInFlightFetch(t *testing.T) {
	const environment = "shutdown-complete"

	var (
		started = make(chan struct{})
		release = make(chan struct{})
		calls   atomic.Int32
	)

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Times(3)
	repo.EXPECT().Fetch(mock.Anything).RunAndReturn(func(ctx context.Context, _ ...string) error {
		if calls.Add(1) == 1 {
			close(started)
			<-release
		}
		// non-nil only if the fetch was cancelled, which would make it an
		// aborted fetch recording no sync duration
		return ctx.Err()
	}).Twice()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	r := NewReceiver(ctx, zaptest.NewLogger(t), map[string]Target{
		environment: {SCM: config.GitHubSCMType, Secret: []byte(testSecret), Repository: repo},
	})

	var (
		acceptedBefore = counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...)
		errorsBefore   = counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)
		syncBefore     = syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)
	)

	p := providers["github"]
	send := func() int {
		return serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment).Code
	}

	assert.Equal(t, http.StatusAccepted, send())
	awaitClosed(t, started, "fetch never started")

	// accepted while the first fetch runs, so still pending at shutdown
	assert.Equal(t, http.StatusAccepted, send())

	cancel()

	assert.Equal(t, http.StatusServiceUnavailable, send(), "new webhooks are refused once shutting down")

	done := make(chan error, 1)
	go func() { done <- r.Shutdown(t.Context()) }()

	select {
	case err := <-done:
		t.Fatalf("Shutdown returned before the fetch completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown never returned after the fetch completed")
	}

	repo.AssertExpectations(t)
	assert.Equal(t, int64(2), counterValue(t, metricRequests, requestAttrs(environment, config.GitHubSCMType, resultAccepted)...)-acceptedBefore)
	assert.Equal(t, int64(0), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)-errorsBefore)
	assert.Equal(t, uint64(2), syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)-syncBefore)
}

// TestReceiver_ShutdownFetchFailureIsAnError asserts that a fetch failing on
// its own while Flipt shuts down is still logged and counted as a fetch error.
func TestReceiver_ShutdownFetchFailureIsAnError(t *testing.T) {
	const environment = "shutdown-failure"

	var (
		started = make(chan struct{})
		release = make(chan struct{})
	)

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Once()
	repo.EXPECT().Fetch(mock.Anything).RunAndReturn(func(context.Context, ...string) error {
		close(started)
		<-release
		return errors.New("connection refused")
	}).Once()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	core, logs := observer.New(zapcore.DebugLevel)
	r := NewReceiver(ctx, zap.New(core), map[string]Target{
		environment: {SCM: config.GitHubSCMType, Secret: []byte(testSecret), Repository: repo},
	})

	var (
		errorsBefore = counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)
		syncBefore   = syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)
	)

	p := providers["github"]
	assert.Equal(t, http.StatusAccepted, serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment).Code)
	awaitClosed(t, started, "fetch never started")

	cancel()
	close(release)

	shutdownCtx, shutdownCancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer shutdownCancel()
	require.NoError(t, r.Shutdown(shutdownCtx))

	repo.AssertExpectations(t)
	assert.Equal(t, int64(1), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)-errorsBefore)
	assert.Equal(t, uint64(0), syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)-syncBefore)
	assert.Equal(t, 1, logs.FilterLevelExact(zapcore.ErrorLevel).FilterMessage("webhook-triggered fetch failed").Len())
}

// TestReceiver_ShutdownTimeoutCancelsFetch asserts that when Shutdown's
// context expires first, it cancels the running fetch, drops the pending
// webhook, waits for the fetch to unwind and returns the context's error.
// The aborted fetch is logged at Debug and isn't counted as a fetch error.
func TestReceiver_ShutdownTimeoutCancelsFetch(t *testing.T) {
	const environment = "shutdown-timeout"

	var (
		started  = make(chan struct{})
		unwound  = make(chan struct{})
		fetchErr error
	)

	repo := NewMockRepository(t)
	repo.EXPECT().Tracks("main").Return(true).Twice()
	repo.EXPECT().Fetch(mock.Anything).RunAndReturn(func(ctx context.Context, _ ...string) error {
		defer close(unwound)
		close(started)
		<-ctx.Done()
		fetchErr = ctx.Err()
		return fetchErr
	}).Once()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	core, logs := observer.New(zapcore.DebugLevel)
	r := NewReceiver(ctx, zap.New(core), map[string]Target{
		environment: {SCM: config.GitHubSCMType, Secret: []byte(testSecret), Repository: repo},
	})

	var (
		errorsBefore = counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)
		syncBefore   = syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)
	)

	p := providers["github"]
	send := func() int {
		return serve(r, newRequest(t, environment, p.fixture, p.headers, p.auth(testSecret)), environment).Code
	}

	assert.Equal(t, http.StatusAccepted, send())
	awaitClosed(t, started, "fetch never started")

	// pending behind the running fetch; dropped once it is cancelled
	assert.Equal(t, http.StatusAccepted, send())

	cancel()

	expired, expiredCancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer expiredCancel()
	require.ErrorIs(t, r.Shutdown(expired), context.DeadlineExceeded)

	// Shutdown returned only after the cancelled fetch unwound
	select {
	case <-unwound:
	default:
		t.Fatal("Shutdown returned before the cancelled fetch unwound")
	}
	require.ErrorIs(t, fetchErr, context.Canceled)

	repo.AssertExpectations(t)
	assert.Equal(t, int64(0), counterValue(t, metricFetchErrors, syncAttrs(environment, config.GitHubSCMType)...)-errorsBefore)
	assert.Equal(t, uint64(0), syncCount(t, syncAttrs(environment, config.GitHubSCMType)...)-syncBefore)
	assert.Equal(t, 0, logs.FilterLevelExact(zapcore.ErrorLevel).Len())
	assert.Equal(t, 1, logs.FilterLevelExact(zapcore.DebugLevel).FilterMessage("webhook-triggered fetch aborted").Len())
}

package cmd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/coss/webhook"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/secrets"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

// gitEnvironment is an environment backed by a storage git repository.
type gitEnvironment struct {
	serverenvironments.Environment
	repo *storagegit.Repository
}

func (e gitEnvironment) Repository() *storagegit.Repository { return e.repo }

type fakeEnvironments map[string]serverenvironments.Environment

func (f fakeEnvironments) Get(_ context.Context, key string) (serverenvironments.Environment, error) {
	env, ok := f[key]
	if !ok {
		return nil, errors.New("not found")
	}

	return env, nil
}

func licenseFor(t *testing.T, p product.Product) license.Manager {
	t.Helper()

	lm := license.NewMockManager(t)
	lm.EXPECT().Product().Return(p).Maybe()

	return lm
}

func webhookConfig(webhooks map[string]*config.IncomingWebhookConfig) *config.Config {
	cfg := &config.Config{Environments: config.EnvironmentsConfig{}}
	for name, wh := range webhooks {
		cfg.Environments[name] = &config.EnvironmentConfig{
			Name:    name,
			Storage: "default",
			SCM:     &config.SCMConfig{Type: config.GitLabSCMType, Webhook: wh},
		}
	}

	return cfg
}

func TestNewWebhookReceiver(t *testing.T) {
	envs := fakeEnvironments{"production": gitEnvironment{}}

	t.Run("no webhooks configured", func(t *testing.T) {
		cfg := webhookConfig(nil)
		cfg.Environments["production"] = &config.EnvironmentConfig{Name: "production", SCM: &config.SCMConfig{Type: config.GitHubSCMType}}

		receiver, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, nil, licenseFor(t, product.Pro))
		require.NoError(t, err)
		assert.Nil(t, receiver)
	})

	t.Run("without pro license warns and disables", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)

		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {Secret: "s3cr3t"}})

		receiver, err := newWebhookReceiver(t.Context(), zap.New(core), cfg, envs, nil, licenseFor(t, product.OSS))
		require.NoError(t, err)
		assert.Nil(t, receiver)

		warnings := logs.FilterMessageSnippet("require a paid license").All()
		require.Len(t, warnings, 1)
		assert.Equal(t, zapcore.WarnLevel, warnings[0].Level)
	})

	t.Run("plain secret", func(t *testing.T) {
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {Secret: "s3cr3t"}})

		receiver, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, nil, licenseFor(t, product.Pro))
		require.NoError(t, err)
		require.NotNil(t, receiver)
		assert.Equal(t, []string{"production"}, receiver.Environments())
	})

	t.Run("secret_ref resolved through secrets manager", func(t *testing.T) {
		ref := &config.SecretReference{Provider: "file", Path: "webhooks", Key: "gitlab"}
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {SecretRef: ref}})

		sm := secrets.NewMockManager(t)
		sm.EXPECT().GetSecretValue(mock.Anything, mock.MatchedBy(func(r secrets.Reference) bool {
			return r.Provider == "file" && r.Path == "webhooks" && r.Key == "gitlab"
		})).Return([]byte("s3cr3t"), nil).Once()

		receiver, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, sm, licenseFor(t, product.Pro))
		require.NoError(t, err)
		require.NotNil(t, receiver)
		sm.AssertExpectations(t)
	})

	t.Run("secret_ref with unconfigured provider fails startup", func(t *testing.T) {
		ref := &config.SecretReference{Provider: "file", Path: "webhooks", Key: "gitlab"}
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {SecretRef: ref}})

		// the file provider is disabled, so it is never registered
		sm, err := secrets.NewManager(zaptest.NewLogger(t), &config.Config{})
		require.NoError(t, err)

		receiver, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, sm, licenseFor(t, product.Pro))
		require.Error(t, err)
		assert.Nil(t, receiver)
		assert.Contains(t, err.Error(), `environment "production"`)
		assert.Contains(t, err.Error(), `provider "file" not found`)
	})

	t.Run("secret_ref without secrets manager fails startup", func(t *testing.T) {
		ref := &config.SecretReference{Provider: "vault", Path: "webhooks", Key: "gitlab"}
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {SecretRef: ref}})

		_, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, nil, licenseFor(t, product.Pro))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `provider "vault"`)
	})

	t.Run("secret_ref resolving to empty value fails startup", func(t *testing.T) {
		ref := &config.SecretReference{Provider: "file", Path: "webhooks", Key: "gitlab"}
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {SecretRef: ref}})

		sm := secrets.NewMockManager(t)
		sm.EXPECT().GetSecretValue(mock.Anything, mock.Anything).Return(nil, nil).Once()

		_, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg, envs, sm, licenseFor(t, product.Pro))
		require.ErrorContains(t, err, "empty value")
	})

	t.Run("non-git environment fails startup", func(t *testing.T) {
		cfg := webhookConfig(map[string]*config.IncomingWebhookConfig{"production": {Secret: "s3cr3t"}})

		_, err := newWebhookReceiver(t.Context(), zaptest.NewLogger(t), cfg,
			fakeEnvironments{"production": serverenvironments.Environment(nil)}, nil, licenseFor(t, product.Pro))
		require.ErrorContains(t, err, "require git storage")
	})
}

// TestWebhookRoute asserts the webhook route is registered only with a
// receiver, sits outside Flipt authentication and bypasses cross-origin
// protection without exempting other routes.
func TestWebhookRoute(t *testing.T) {
	const (
		secret = "s3cr3t"
		body   = `{"ref":"refs/heads/main"}`
	)

	newCfg := func() *config.Config {
		cfg := &config.Config{Server: config.ServerConfig{Host: "localhost"}}
		// enables cross-origin protection
		cfg.Authentication.Session.CSRF.Key = "abcdefghijklmnopqrstuvwxyz123456"
		cfg.Authentication.Required = true
		return cfg
	}

	newReq := func(t *testing.T) *http.Request {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://localhost/api/v2/webhooks/production", strings.NewReader(body))
		req.Header.Set("X-Gitlab-Event", "Push Hook")
		req.Header.Set("X-Gitlab-Token", secret)
		// what a browser would send for a cross-site request
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Header.Set("Origin", "https://evil.example")
		return req
	}

	t.Run("with receiver", func(t *testing.T) {
		repo := webhook.NewMockRepository(t)
		repo.EXPECT().Tracks("main").Return(true).Once()
		repo.EXPECT().Fetch(mock.Anything).Return(nil).Once()

		receiver := webhook.NewReceiver(t.Context(), zaptest.NewLogger(t), map[string]webhook.Target{
			"production": {SCM: config.GitLabSCMType, Secret: []byte(secret), Repository: repo},
		})

		server, err := NewHTTPServer(t.Context(), zaptest.NewLogger(t), newCfg(), nil, info.Flipt{}, WithWebhookReceiver(receiver))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		server.Handler.ServeHTTP(res, newReq(t))
		receiver.Wait()

		assert.Equal(t, http.StatusAccepted, res.Code, "body: %s", res.Body.String())
		repo.AssertExpectations(t)

		// the exemption covers only the webhook route: a cross-site POST to
		// any other API path is still rejected.
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://localhost/api/v2/environments/production/namespaces", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Header.Set("Origin", "https://evil.example")

		res = httptest.NewRecorder()
		server.Handler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code, "body: %s", res.Body.String())
	})

	t.Run("without receiver", func(t *testing.T) {
		server, err := NewHTTPServer(t.Context(), zaptest.NewLogger(t), newCfg(), nil, info.Flipt{}, WithWebhookReceiver(nil))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		server.Handler.ServeHTTP(res, newReq(t))

		assert.NotEqual(t, http.StatusAccepted, res.Code)
		assert.GreaterOrEqual(t, res.Code, http.StatusBadRequest)
	})
}

package cmd

import (
	"context"
	"fmt"
	"slices"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/webhook"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/secrets"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	"go.uber.org/zap"
)

// webhookPathPattern is the route serving incoming SCM webhooks.
const webhookPathPattern = "/api/v2/webhooks/{environment}"

// webhookCrossOriginExemptPattern exempts incoming webhooks from cross-origin
// protection. Webhooks are authenticated by their own signature or token and
// carry no ambient (cookie) authority.
const webhookCrossOriginExemptPattern = "POST /api/v2/webhooks/"

type environmentGetter interface {
	Get(ctx context.Context, key string) (serverenvironments.Environment, error)
}

// newWebhookReceiver builds the incoming webhook receiver for every
// environment with scm.webhook configured. It returns nil when no
// environment configures a webhook, or when the license isn't Pro.
//
// A webhook secret_ref that can't be resolved (for example because it names
// an unconfigured or disabled secrets provider) fails startup.
func newWebhookReceiver(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	envs environmentGetter,
	secretsManager secrets.Manager,
	licenseManager interface{ Product() product.Product },
) (*webhook.Receiver, error) {
	var names []string
	for _, envConf := range cfg.Environments {
		if envConf.SCM != nil && envConf.SCM.Webhook != nil {
			names = append(names, envConf.Name)
		}
	}

	if len(names) == 0 {
		return nil, nil
	}

	slices.Sort(names)

	if licenseManager == nil || licenseManager.Product() != product.Pro {
		logger.Warn("incoming scm webhooks require a paid license; webhook receiver disabled.", zap.Strings("environments", names))
		return nil, nil
	}

	targets := make(map[string]webhook.Target, len(names))
	for _, envConf := range cfg.Environments {
		if envConf.SCM == nil || envConf.SCM.Webhook == nil {
			continue
		}

		secret, err := resolveWebhookSecret(ctx, envConf.SCM.Webhook, secretsManager)
		if err != nil {
			return nil, fmt.Errorf("environment %q: resolving scm webhook secret: %w", envConf.Name, err)
		}

		env, err := envs.Get(ctx, envConf.Name)
		if err != nil {
			return nil, fmt.Errorf("environment %q: %w", envConf.Name, err)
		}

		repoEnv, ok := env.(interface {
			Repository() *storagegit.Repository
		})
		if !ok {
			return nil, fmt.Errorf("environment %q: scm webhooks require git storage", envConf.Name)
		}

		targets[envConf.Name] = webhook.Target{
			SCM:        envConf.SCM.Type,
			Secret:     secret,
			Repository: repoEnv.Repository(),
		}
	}

	receiver := webhook.NewReceiver(ctx, logger, targets)

	logger.Info("incoming scm webhook receiver enabled", zap.Strings("environments", receiver.Environments()))

	return receiver, nil
}

// resolveWebhookSecret returns the configured plain secret, or resolves
// secret_ref through the secrets manager.
func resolveWebhookSecret(ctx context.Context, cfg *config.IncomingWebhookConfig, secretsManager secrets.Manager) ([]byte, error) {
	if cfg.SecretRef == nil {
		if cfg.Secret == "" {
			return nil, fmt.Errorf("secret is empty")
		}

		return []byte(cfg.Secret), nil
	}

	ref := secrets.Reference{
		Provider: cfg.SecretRef.Provider,
		Path:     cfg.SecretRef.Path,
		Key:      cfg.SecretRef.Key,
	}

	if secretsManager == nil {
		return nil, fmt.Errorf("secret_ref provider %q: secrets manager not configured", ref.Provider)
	}

	value, err := secretsManager.GetSecretValue(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("secret_ref provider %q (is it configured and enabled?): %w", ref.Provider, err)
	}

	if len(value) == 0 {
		return nil, fmt.Errorf("secret_ref provider %q path %q key %q resolved to an empty value", ref.Provider, ref.Path, ref.Key)
	}

	return value, nil
}

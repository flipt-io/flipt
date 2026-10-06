package cmd

import (
	"context"
	"fmt"
	"slices"

	"go.uber.org/zap"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/webhook"
	"go.flipt.io/flipt/internal/product"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	storagegit "go.flipt.io/flipt/internal/storage/git"
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
// Webhook secrets may be ${secret:provider:key} references; those are
// resolved before this runs (see resolveSecretsInConfig in cmd/flipt). A
// secret that is empty after resolution fails startup.
func newWebhookReceiver(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	envs environmentGetter,
	licenseManager interface{ Product() product.Product },
) (*webhook.Receiver, error) {
	var envConfs []*config.EnvironmentConfig
	for _, envConf := range cfg.Environments {
		if envConf.SCM != nil && envConf.SCM.Webhook != nil {
			envConfs = append(envConfs, envConf)
		}
	}

	if len(envConfs) == 0 {
		return nil, nil
	}

	if licenseManager == nil || licenseManager.Product() != product.Pro {
		names := make([]string, 0, len(envConfs))
		for _, envConf := range envConfs {
			names = append(names, envConf.Name)
		}
		slices.Sort(names)

		logger.Warn("incoming scm webhooks require a paid license; webhook receiver disabled.", zap.Strings("environments", names))
		return nil, nil
	}

	targets := make(map[string]webhook.Target, len(envConfs))
	for _, envConf := range envConfs {
		if envConf.SCM.Webhook.Secret == "" {
			return nil, fmt.Errorf("environment %q: scm webhook secret is empty", envConf.Name)
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
			Secret:     []byte(envConf.SCM.Webhook.Secret),
			Repository: repoEnv.Repository(),
		}
	}

	receiver := webhook.NewReceiver(ctx, logger, targets)

	logger.Info("incoming scm webhook receiver enabled", zap.Strings("environments", receiver.Environments()))

	return receiver, nil
}

package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	metadataKeyNamespace          = "io.flipt.auth.k8s.namespace"
	metadataKeyPodName            = "io.flipt.auth.k8s.pod.name"
	metadataKeyPodUID             = "io.flipt.auth.k8s.pod.uid"
	metadataKeyServiceAccountName = "io.flipt.auth.k8s.serviceaccount.name"
	metadataKeyServiceAccountUID  = "io.flipt.auth.k8s.serviceaccount.uid"
)

// Authenticator is the interface for authenticating Kubernetes service account tokens
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*auth.Authentication, error)
}

// KubernetesAuthenticator handles authentication of Kubernetes service account tokens
type KubernetesAuthenticator struct {
	logger   *zap.Logger
	verifier tokenVerifier
	cache    *cache.Cache
	config   config.AuthenticationMethodKubernetesConfig
}

// New creates a new KubernetesAuthenticator instance
func New(logger *zap.Logger, config config.AuthenticationMethodKubernetesConfig) (Authenticator, error) {
	verifier, err := newKubernetesOIDCVerifier(logger, config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes OIDC verifier: %w", err)
	}

	// Default cache TTL of 5 minutes, cleanup every minute
	c := cache.New(5*time.Minute, time.Minute)

	return &KubernetesAuthenticator{
		logger:   logger,
		verifier: verifier,
		cache:    c,
		config:   config,
	}, nil
}

// Authenticate validates a Kubernetes service account token and returns authentication details
func (ka *KubernetesAuthenticator) Authenticate(ctx context.Context, token string) (*auth.Authentication, error) {
	if ka.cache != nil {
		// Check cache first
		if cached, found := ka.cache.Get(token); found {
			ka.logger.Debug("using cached authentication")
			return cached.(*auth.Authentication), nil
		}
	}

	// Verify token
	claims, err := ka.verifier.verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid service account token: %w", err)
	}

	expiresAt := time.Unix(claims.Expiration, 0)
	authentication := &auth.Authentication{
		Method:    auth.Method_METHOD_KUBERNETES,
		ExpiresAt: timestamppb.New(expiresAt),
		Metadata: map[string]string{
			metadataKeyNamespace:          claims.Identity.Namespace,
			metadataKeyPodName:            claims.Identity.Pod.Name,
			metadataKeyPodUID:             claims.Identity.Pod.UID,
			metadataKeyServiceAccountName: claims.Identity.ServiceAccount.Name,
			metadataKeyServiceAccountUID:  claims.Identity.ServiceAccount.UID,
		},
	}

	if ka.cache != nil {
		// Cache until token expiration
		ka.cache.Set(token, authentication, time.Until(expiresAt))
	}

	return authentication, nil
}

// tokenVerifier interface from v1 package
type tokenVerifier interface {
	verify(ctx context.Context, jwt string) (claims, error)
}

// claims struct from v1 package
type claims struct {
	Expiration int64    `json:"exp"`
	Identity   identity `json:"kubernetes.io"`
}

type identity struct {
	Namespace      string   `json:"namespace"`
	Pod            resource `json:"pod"`
	ServiceAccount resource `json:"serviceaccount"`
}

type resource struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

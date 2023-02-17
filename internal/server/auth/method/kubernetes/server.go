package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
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

type resource struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type claims struct {
	Expiration int64    `jsom:"exp"`
	Identity   identity `json:"kubernetes.io"`
}

type identity struct {
	Namespace      string   `json:"namespace"`
	Pod            resource `json:"pod"`
	ServiceAccount resource `json:"serviceaccount"`
}

type tokenVerifier interface {
	verify(jwt string) (claims, error)
}

// Server is the core server-side implementation of the "kubernetes" authentication method.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	verifier tokenVerifier

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// NewServer constructs a new Server instance based on the provided logger, store and configuration.
func NewServer(logger *zap.Logger, store storageauth.Store, config config.AuthenticationConfig) *Server {
	return &Server{
		logger:   logger,
		store:    store,
		config:   config,
		verifier: noopTokenVerifier{},
	}
}

func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	claims, err := s.verifier.verify(req.ServiceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("verifying service account: %w", err)
	}

	clientToken, authentication, err := s.store.CreateAuthentication(
		ctx,
		&storageauth.CreateAuthenticationRequest{
			Method:    auth.Method_METHOD_KUBERNETES,
			ExpiresAt: timestamppb.New(time.Unix(claims.Expiration, 0)),
			Metadata: map[string]string{
				metadataKeyNamespace:          claims.Identity.Namespace,
				metadataKeyPodName:            claims.Identity.Pod.Name,
				metadataKeyPodUID:             claims.Identity.Pod.UID,
				metadataKeyServiceAccountName: claims.Identity.ServiceAccount.Name,
				metadataKeyServiceAccountUID:  claims.Identity.ServiceAccount.UID,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verifying service account: %w", err)
	}

	return &auth.VerifyServiceAccountResponse{
		ClientToken:    clientToken,
		Authentication: authentication,
	}, nil
}

type noopTokenVerifier struct{}

func (n noopTokenVerifier) verify(string) (claims, error) {
	return claims{}, errors.New("invalid token")
}

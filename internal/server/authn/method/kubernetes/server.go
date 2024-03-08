package kubernetes

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	verify(ctx context.Context, jwt string) (claims, error)
}

// Server is the core server-side implementation of the "kubernetes" authentication method.
//
// The method allows services deployed into the same Kubernetes cluster as Flipt to leverage
// their service account token in order to obtain access to Flipt itself.
// When enabled, this authentication method grants any service in the same cluster access to Flipt.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	verifier tokenVerifier

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// New constructs a new Server instance based on the provided logger, store and configuration.
func New(logger *zap.Logger, store storageauth.Store, config config.AuthenticationConfig) (*Server, error) {
	s := &Server{
		logger: logger,
		store:  store,
		config: config,
	}

	var err error
	s.verifier, err = newKubernetesOIDCVerifier(logger, config.Methods.Kubernetes.Method)

	return s, err
}

// RegisterGRPC registers the server instnace on the provided gRPC server.
func (s *Server) RegisterGRPC(srv *grpc.Server) {
	auth.RegisterAuthenticationMethodKubernetesServiceServer(srv, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

// VerifyServiceAccount takes a service account token, configured by a kubernetes environment,
// validates it's authenticity and (if valid) creates a Flipt client token and returns it.
// The returned client token is valid for the lifetime of the service account JWT.
// The token tracks the source service account and pod identity of the provided token.
func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	claims, err := s.verifier.verify(ctx, req.ServiceAccountToken)
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

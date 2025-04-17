package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method"
	rpcflipt "go.flipt.io/flipt/rpc/flipt"
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

// Server is the core server-side implementation of the "kubernetes" authentication method.
//
// The method allows services deployed into the same Kubernetes cluster as Flipt to leverage
// their service account token in order to obtain access to Flipt itself.
// When enabled, this authentication method grants any service in the same cluster access to Flipt.
type Server struct {
	logger *zap.Logger
	config config.AuthenticationConfig
	// now is a function that returns the current time
	// and is overridden in tests
	now       func() *timestamppb.Timestamp
	validator method.JWTValidator

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// NewServer constructs a new Server instance based on the provided logger and configuration.
func NewServer(logger *zap.Logger, config config.AuthenticationConfig, validator method.JWTValidator) (*Server, error) {
	s := &Server{
		logger:    logger,
		config:    config,
		now:       rpcflipt.Now,
		validator: validator,
	}

	return s, nil
}

// RegisterGRPC registers the server instnace on the provided gRPC server.
func (s *Server) RegisterGRPC(srv *grpc.Server) {
	auth.RegisterAuthenticationMethodKubernetesServiceServer(srv, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

// VerifyServiceAccount takes a service account token, configured by a kubernetes environment,
// validates it's authenticity and (if valid) returns it.
func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	resp, err := s.validator.Validate(ctx, req.ServiceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("verifying service account: %w", err)
	}

	claims := &claims{}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshalling service account token: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, claims); err != nil {
		return nil, fmt.Errorf("unmarshalling service account token: %w", err)
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
		CreatedAt: s.now(),
		UpdatedAt: s.now(),
	}

	return &auth.VerifyServiceAccountResponse{
		ClientToken:    req.ServiceAccountToken,
		Authentication: authentication,
	}, nil
}

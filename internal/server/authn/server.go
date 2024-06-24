package authn

import (
	"context"
	"fmt"
	"strings"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/authn/method"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	ipKey  = "x-forwarded-for"
	subKey = "sub"
)

var _ auth.AuthenticationServiceServer = &Server{}

func firstNonEmptyString(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}

	return ""
}

func GetActorFromAuthentication(authentication *auth.Authentication, ip string) *audit.Actor {
	actor := func(m string) *audit.Actor {
		return &audit.Actor{
			Authentication: m,
			Email:          firstNonEmptyString(authentication.Metadata[method.StorageMetadataEmail], authentication.Metadata[fmt.Sprintf("io.flipt.auth.%s.email", m)]),
			Name:           firstNonEmptyString(authentication.Metadata[method.StorageMetadataName], authentication.Metadata[fmt.Sprintf("io.flipt.auth.%s.name", m)]),
			Picture:        firstNonEmptyString(authentication.Metadata[method.StorageMetadataPicture], authentication.Metadata[fmt.Sprintf("io.flipt.auth.%s.picture", m)]),
			IP:             ip,
		}
	}

	switch authentication.Method {
	case auth.Method_METHOD_OIDC:
		return actor("oidc")
	case auth.Method_METHOD_JWT:
		return actor("jwt")
	case auth.Method_METHOD_GITHUB:
		return actor("github")

	default:
		authName := strings.ToLower(strings.TrimPrefix(authentication.Method.String(), "METHOD_"))
		return &audit.Actor{
			Authentication: authName,
			IP:             ip,
		}
	}
}

func ActorFromContext(ctx context.Context) *audit.Actor {
	var (
		ipAddress string
	)

	md, _ := metadata.FromIncomingContext(ctx)
	if len(md[ipKey]) > 0 {
		ipAddress = md[ipKey][0]
	}

	auth := authmiddlewaregrpc.GetAuthenticationFrom(ctx)
	if auth != nil {
		return GetActorFromAuthentication(auth, ipAddress)
	}

	return &audit.Actor{
		Authentication: "none",
		IP:             ipAddress,
	}
}

// Server is the core AuthenticationServiceServer implementations.
//
// It is the service which presents all Authentications created in the backing auth store.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store

	enableAuditLogging bool

	auth.UnimplementedAuthenticationServiceServer
}

type Option func(*Server)

// WithAuditLoggingEnabled sets the option for enabling audit logging for the auth server.
func WithAuditLoggingEnabled(enabled bool) Option {
	return func(s *Server) {
		s.enableAuditLogging = enabled
	}
}

func NewServer(logger *zap.Logger, store storageauth.Store, opts ...Option) *Server {
	s := &Server{
		logger: logger,
		store:  store,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationServiceServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return false
}

// GetAuthenticationSelf returns the Authentication which was derived from the request context.
func (s *Server) GetAuthenticationSelf(ctx context.Context, _ *emptypb.Empty) (*auth.Authentication, error) {
	if auth := authmiddlewaregrpc.GetAuthenticationFrom(ctx); auth != nil {
		s.logger.Debug("GetAuthentication", zap.String("id", auth.Id))

		return auth, nil
	}

	return nil, errors.ErrUnauthenticatedf("request was not authenticated")
}

// GetAuthentication returns the Authentication identified by the supplied id.
func (s *Server) GetAuthentication(ctx context.Context, r *auth.GetAuthenticationRequest) (*auth.Authentication, error) {
	return s.store.GetAuthenticationByID(ctx, r.Id)
}

// ListAuthentications produces a set of authentications for the provided method filter and pagination parameters.
func (s *Server) ListAuthentications(ctx context.Context, r *auth.ListAuthenticationsRequest) (*auth.ListAuthenticationsResponse, error) {
	req := &storage.ListRequest[storageauth.ListAuthenticationsPredicate]{
		QueryParams: storage.QueryParams{
			Limit:     uint64(r.Limit),
			PageToken: r.PageToken,
		},
	}

	if r.Method != auth.Method_METHOD_NONE {
		req.Predicate.Method = &r.Method
	}

	results, err := s.store.ListAuthentications(ctx, req)
	if err != nil {
		s.logger.Error("listing authentication", zap.Error(err))

		return nil, fmt.Errorf("listing authentications: %w", err)
	}

	return &auth.ListAuthenticationsResponse{
		Authentications: results.Results,
		NextPageToken:   results.NextPageToken,
	}, nil
}

// DeleteAuthentication deletes the authentication with the supplied ID.
func (s *Server) DeleteAuthentication(ctx context.Context, req *auth.DeleteAuthenticationRequest) (*emptypb.Empty, error) {
	s.logger.Debug("DeleteAuthentication", zap.String("id", req.Id))

	if s.enableAuditLogging {
		actor := ActorFromContext(ctx)

		a, err := s.GetAuthentication(ctx, &auth.GetAuthenticationRequest{
			Id: req.Id,
		})
		if err != nil {
			s.logger.Error("failed to get authentication for audit events", zap.Error(err))
			return nil, err
		}
		if a.Method == auth.Method_METHOD_TOKEN {
			request := flipt.NewRequest(flipt.ResourceAuthentication, flipt.ActionDelete, flipt.WithSubject(flipt.SubjectToken))
			event := audit.NewEvent(request, actor, a.Metadata)
			event.AddToSpan(ctx)
		}
	}

	return &emptypb.Empty{}, s.store.DeleteAuthentications(ctx, storageauth.Delete(storageauth.WithID(req.Id)))
}

// ExpireAuthenticationSelf expires the Authentication which was derived from the request context.
// If no expire_at is provided, the current time is used. This is useful for logging out a user.
// If the expire_at is greater than the current expiry time, the expiry time is extended.
func (s *Server) ExpireAuthenticationSelf(ctx context.Context, req *auth.ExpireAuthenticationSelfRequest) (*emptypb.Empty, error) {
	if auth := authmiddlewaregrpc.GetAuthenticationFrom(ctx); auth != nil {
		s.logger.Debug("ExpireAuthentication", zap.String("id", auth.Id))

		if req.ExpiresAt == nil || !req.ExpiresAt.IsValid() {
			req.ExpiresAt = flipt.Now()
		}

		return &emptypb.Empty{}, s.store.ExpireAuthenticationByID(ctx, auth.Id, req.ExpiresAt)
	}

	return nil, errors.ErrUnauthenticatedf("request was not authenticated")
}

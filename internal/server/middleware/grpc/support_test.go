package grpc_middleware

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ storageauth.Store = &authStoreMock{}

type authStoreMock struct {
	mock.Mock
}

func (a *authStoreMock) CreateAuthentication(ctx context.Context, r *storageauth.CreateAuthenticationRequest) (string, *auth.Authentication, error) {
	args := a.Called(ctx, r)
	return args.String(0), args.Get(1).(*auth.Authentication), args.Error(2)
}

func (a *authStoreMock) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*auth.Authentication, error) {
	return nil, nil
}

func (a *authStoreMock) GetAuthenticationByID(ctx context.Context, id string) (*auth.Authentication, error) {
	return nil, nil
}

func (a *authStoreMock) ListAuthentications(ctx context.Context, r *storage.ListRequest[storageauth.ListAuthenticationsPredicate]) (set storage.ResultSet[*auth.Authentication], err error) {
	return set, err
}

func (a *authStoreMock) DeleteAuthentications(ctx context.Context, r *storageauth.DeleteAuthenticationsRequest) error {
	args := a.Called(ctx, r)
	return args.Error(0)
}

func (a *authStoreMock) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) error {
	return nil
}

func (a *authStoreMock) Shutdown(ctx context.Context) error {
	return nil
}

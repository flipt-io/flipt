package grpc_middleware

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
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

type cacheSpy struct {
	cache.Cacher

	getKeys   map[string]struct{}
	getCalled int

	setItems  map[string][]byte
	setCalled int

	deleteKeys   map[string]struct{}
	deleteCalled int
}

func newCacheSpy(c cache.Cacher) *cacheSpy {
	return &cacheSpy{
		Cacher:     c,
		getKeys:    make(map[string]struct{}),
		setItems:   make(map[string][]byte),
		deleteKeys: make(map[string]struct{}),
	}
}

func (c *cacheSpy) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.getCalled++
	c.getKeys[key] = struct{}{}
	return c.Cacher.Get(ctx, key)
}

func (c *cacheSpy) Set(ctx context.Context, key string, value []byte) error {
	c.setCalled++
	c.setItems[key] = value
	return c.Cacher.Set(ctx, key, value)
}

func (c *cacheSpy) Delete(ctx context.Context, key string) error {
	c.deleteCalled++
	c.deleteKeys[key] = struct{}{}
	return c.Cacher.Delete(ctx, key)
}

type auditSinkSpy struct {
	sendAuditsCalled int
	events           []audit.Event
	fmt.Stringer
}

func (a *auditSinkSpy) SendAudits(ctx context.Context, es []audit.Event) error {
	a.sendAuditsCalled++
	a.events = append(a.events, es...)
	return nil
}

func (a *auditSinkSpy) String() string {
	return "auditSinkSpy"
}

func (a *auditSinkSpy) Close() error { return nil }

type auditExporterSpy struct {
	audit.EventExporter
	sinkSpy *auditSinkSpy
}

func newAuditExporterSpy(logger *zap.Logger) *auditExporterSpy {
	aspy := &auditSinkSpy{events: make([]audit.Event, 0)}
	as := []audit.Sink{aspy}

	return &auditExporterSpy{
		EventExporter: audit.NewSinkSpanExporter(logger, as),
		sinkSpy:       aspy,
	}
}

func (a *auditExporterSpy) GetSendAuditsCalled() int {
	return a.sinkSpy.sendAuditsCalled
}

func (a *auditExporterSpy) GetEvents() []audit.Event {
	return a.sinkSpy.events
}

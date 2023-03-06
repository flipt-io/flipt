package common

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	panic("not implemented")
}

func (s *Store) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	panic("not implemented")
}

func (s *Store) CountNamespaces(ctx context.Context) (uint64, error) {
	panic("not implemented")
}

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	panic("not implemented")
}

func (s *Store) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	panic("not implemented")
}

func (s *Store) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	panic("not implemented")
}

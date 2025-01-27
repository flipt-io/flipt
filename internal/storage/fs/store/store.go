package store

import (
	"context"
	"errors"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap"
)

// NewStore is a constructor that handles all the known declarative backend storage types
// Given the provided storage type is know, the relevant backend is configured and returned
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config) (_ storage.Store, err error) {
	return nil, errors.New("not implemented")
}

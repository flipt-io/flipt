package main

import (
	"fmt"
	"net/url"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func fliptServer(logger *zap.Logger, cfg *config.Config) (*server.Server, func(), error) {
	db, driver, err := sql.Open(*cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("opening db: %w", err)
	}

	logger.Debug("constructing builder", zap.Bool("prepared_statements", cfg.Database.PreparedStatementsEnabled))

	builder := sql.BuilderFor(db, driver, cfg.Database.PreparedStatementsEnabled)

	var store storage.Store

	switch driver {
	case sql.SQLite, sql.LibSQL:
		store = sqlite.NewStore(db, builder, logger)
	case sql.Postgres, sql.CockroachDB:
		store = postgres.NewStore(db, builder, logger)
	case sql.MySQL:
		store = mysql.NewStore(db, builder, logger)
	}

	return server.New(logger, store), func() { _ = db.Close() }, nil
}

func fliptSDK(address, token string) (*sdk.SDK, error) {
	addr, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("export address is invalid %w", err)
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http", "https":
		transport = sdkhttp.NewTransport(address)
	case "grpc":
		conn, err := grpc.Dial(addr.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to dial Flipt %w", err)
		}

		transport = sdkgrpc.NewTransport(conn)
	default:
		return nil, fmt.Errorf("unexpected protocol %s", addr.Scheme)
	}

	var opts []sdk.Option
	if token != "" {
		opts = append(opts, sdk.WithAuthenticationProvider(
			sdk.StaticTokenAuthenticationProvider(token),
		))
	}
	s := sdk.New(transport, opts...)
	return &s, nil
}

func fliptClient(address, token string) (*sdk.Flipt, error) {
	s, err := fliptSDK(address, token)
	if err != nil {
		return nil, err
	}
	return s.Flipt(), nil
}

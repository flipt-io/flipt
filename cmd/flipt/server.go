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

	var store storage.Store

	switch driver {
	case sql.SQLite:
		store = sqlite.NewStore(db, logger)
	case sql.Postgres, sql.CockroachDB:
		store = postgres.NewStore(db, logger)
	case sql.MySQL:
		store = mysql.NewStore(db, logger)
	}

	return server.New(logger, store), func() { _ = db.Close() }, nil
}

func fliptClient(logger *zap.Logger, address, token string) *sdk.Flipt {
	addr, err := url.Parse(address)
	if err != nil {
		logger.Fatal("export address is invalid", zap.Error(err))
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http":
		transport = sdkhttp.NewTransport(address)
	case "grpc":
		conn, err := grpc.Dial(addr.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("failed to dial Flipt", zap.Error(err))
		}

		transport = sdkgrpc.NewTransport(conn)
	default:
		logger.Fatal("unexpected protocol", zap.String("address", address))
	}

	var opts []sdk.Option
	if token != "" {
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(token),
		))
	}

	return sdk.New(transport, opts...).Flipt()
}

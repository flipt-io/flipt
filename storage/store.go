package storage

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/pkg/errors"
)

type Store struct {
	dbType dbType
	uri    string

	builder sq.StatementBuilderType
	db      *sql.DB
}

func Open(url string) (*Store, error) {
	dbType, uri, err := parse(url)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(dbType.String(), uri)
	if err != nil {
		return nil, err
	}

	var (
		cacher  = sq.NewStmtCacher(db)
		builder sq.StatementBuilderType
	)

	switch dbType {
	case dbSQLite:
		builder = sq.StatementBuilder.RunWith(cacher)
	case dbPostgres:
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(cacher)
	}

	return &Store{
		dbType:  dbType,
		uri:     uri,
		builder: builder,
		db:      db,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(path string) error {
	var (
		d   database.Driver
		err error
	)

	switch s.dbType {
	case dbSQLite:
		d, err = sqlite3.WithInstance(s.db, &sqlite3.Config{})
	case dbPostgres:
		d, err = postgres.WithInstance(s.db, &postgres.Config{})
	}

	if err != nil {
		return errors.Wrap(err, "getting db driver for migrations")
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", path, s.dbType))
	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), s.dbType.String(), d)
	if err != nil {
		return errors.Wrap(err, "opening migrations")
	}

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "running migrations")
	}

	return nil
}

var (
	dbTypeToString = map[dbType]string{
		dbSQLite:   "sqlite3",
		dbPostgres: "postgres",
	}

	stringToDBType = map[string]dbType{
		"sqlite3":  dbSQLite,
		"postgres": dbPostgres,
	}
)

type dbType uint8

func (d dbType) String() string {
	return dbTypeToString[d]
}

const (
	_ dbType = iota
	dbSQLite
	dbPostgres
)

func parse(url string) (dbType, string, error) {
	parts := strings.SplitN(url, "://", 2)
	// TODO: check parts

	dbType := stringToDBType[parts[0]]
	if dbType == 0 {
		return 0, "", fmt.Errorf("unknown database type: %s", parts[0])
	}

	uri := parts[1]

	switch dbType {
	case dbSQLite:
		uri = fmt.Sprintf("%s?cache=shared&_fk=true", parts[1])
	case dbPostgres:
		uri = fmt.Sprintf("postgres://%s", parts[1])
	}

	return dbType, uri, nil
}

// RuleStore ...
type RuleStore interface {
	GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error)
	ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error)
	CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error)
	UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error)
	DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error
	OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error
	CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error)
	DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error
	Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error)
}

// FlagStore ...
type FlagStore interface {
	GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error)
	ListFlags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error)
	CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error)
	UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error
	CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error)
	UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error
}

// SegmentStore ...
type SegmentStore interface {
	GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error)
	ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) ([]*flipt.Segment, error)
	CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error)
	UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error)
	DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error
	CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
	UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error)
	DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error
}

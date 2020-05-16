package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	sq "github.com/Masterminds/squirrel"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage/db"
	"gopkg.in/yaml.v2"
)

var (
	dropBeforeImport bool
	importStdin      bool
)

func runImport(args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	sql, driver, err := db.Open(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}

	defer sql.Close()

	var (
		builder    sq.StatementBuilderType
		stmtCacher = sq.NewStmtCacher(sql)
	)

	switch driver {
	case db.SQLite:
		builder = sq.StatementBuilder.RunWith(stmtCacher)

	case db.Postgres:
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(stmtCacher)
	}

	var in io.ReadCloser = os.Stdin

	if !importStdin {
		importFilename := args[0]
		if importFilename == "" {
			return errors.New("import filename required")
		}

		f := filepath.Clean(importFilename)

		logger.Debugf("importing from %q", f)

		in, err = os.Open(f)
		if err != nil {
			return fmt.Errorf("opening import file: %w", err)
		}
	}

	defer in.Close()

	// drop tables if specified
	if dropBeforeImport {
		logger.Debug("dropping tables before import")

		tables := []string{"schema_migrations", "distributions", "rules", "constraints", "variants", "segments", "flags"}

		for _, table := range tables {
			if _, err := sql.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
				return fmt.Errorf("dropping tables: %w", err)
			}
		}
	}

	migrator, err := db.NewMigrator(cfg)
	if err != nil {
		return err
	}

	defer migrator.Close()

	canAutoMigrate := false

	// check if any migrations are pending
	currentVersion, err := migrator.CurrentVersion()
	if err != nil {
		// if first run then it's safe to migrate
		if err == db.ErrMigrationsNilVersion {
			canAutoMigrate = true
		} else {
			return fmt.Errorf("checking migration status: %w", err)
		}
	}

	if currentVersion < expectedMigrationVersion {
		logger.Debugf("migrations pending: [current version=%d, want version=%d]", currentVersion, expectedMigrationVersion)

		if !canAutoMigrate {
			return errors.New("migrations pending, please backup your database and run `flipt migrate`")
		}

		logger.Debug("running migrations...")

		if err := migrator.Run(); err != nil {
			return err
		}

		logger.Debug("finished migrations")
	} else {
		logger.Debug("migrations up to date")
	}

	migrator.Close()

	var (
		flagStore    = db.NewFlagStore(builder)
		segmentStore = db.NewSegmentStore(builder)
		ruleStore    = db.NewRuleStore(builder, sql)

		dec = yaml.NewDecoder(in)
		doc = new(Document)
	)

	if err := dec.Decode(doc); err != nil {
		return fmt.Errorf("importing: %w", err)
	}

	var (
		// map flagKey => *flag
		createdFlags = make(map[string]*flipt.Flag)
		// map segmentKey => *segment
		createdSegments = make(map[string]*flipt.Segment)
		// map flagKey:variantKey => *variant
		createdVariants = make(map[string]*flipt.Variant)
	)

	// create flags/variants
	for _, f := range doc.Flags {
		flag, err := flagStore.CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         f.Key,
			Name:        f.Name,
			Description: f.Description,
			Enabled:     f.Enabled,
		})

		if err != nil {
			return fmt.Errorf("importing flag: %w", err)
		}

		for _, v := range f.Variants {
			variant, err := flagStore.CreateVariant(ctx, &flipt.CreateVariantRequest{
				FlagKey:     f.Key,
				Key:         v.Key,
				Name:        v.Name,
				Description: v.Description,
			})

			if err != nil {
				return fmt.Errorf("importing variant: %w", err)
			}

			createdVariants[fmt.Sprintf("%s:%s", flag.Key, variant.Key)] = variant
		}

		createdFlags[flag.Key] = flag
	}

	// create segments/constraints
	for _, s := range doc.Segments {
		segment, err := segmentStore.CreateSegment(ctx, &flipt.CreateSegmentRequest{
			Key:         s.Key,
			Name:        s.Name,
			Description: s.Description,
		})

		if err != nil {
			return fmt.Errorf("importing segment: %w", err)
		}

		for _, c := range s.Constraints {
			_, err := segmentStore.CreateConstraint(ctx, &flipt.CreateConstraintRequest{
				SegmentKey: s.Key,
				Type:       flipt.ComparisonType(flipt.ComparisonType_value[c.Type]),
				Property:   c.Property,
				Operator:   c.Operator,
				Value:      c.Value,
			})

			if err != nil {
				return fmt.Errorf("importing constraint: %w", err)
			}
		}

		createdSegments[segment.Key] = segment
	}

	// create rules/distributions
	for _, f := range doc.Flags {
		// loop through rules
		for _, r := range f.Rules {
			rule, err := ruleStore.CreateRule(ctx, &flipt.CreateRuleRequest{
				FlagKey:    f.Key,
				SegmentKey: r.SegmentKey,
				Rank:       int32(r.Rank),
			})

			if err != nil {
				return fmt.Errorf("importing rule: %w", err)
			}

			for _, d := range r.Distributions {
				variant, found := createdVariants[fmt.Sprintf("%s:%s", f.Key, d.VariantKey)]
				if !found {
					return fmt.Errorf("finding variant: %s; flag: %s", d.VariantKey, f.Key)
				}

				_, err := ruleStore.CreateDistribution(ctx, &flipt.CreateDistributionRequest{
					FlagKey:   f.Key,
					RuleId:    rule.Id,
					VariantId: variant.Id,
					Rollout:   d.Rollout,
				})

				if err != nil {
					return fmt.Errorf("importing distribution: %w", err)
				}
			}
		}
	}

	return nil
}

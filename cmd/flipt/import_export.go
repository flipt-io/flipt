package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sq "github.com/Masterminds/squirrel"
	"github.com/markphelps/flipt/storage/db"
	"gopkg.in/yaml.v2"
)

type Flag struct {
	Key         string     `yaml:"key,omitempty"`
	Name        string     `yaml:"name,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Enabled     bool       `yaml:"enabled"`
	Variants    []*Variant `yaml:"variants,omitempty"`
	Rules       []*Rule    `yaml:"rules,omitempty"`
}

type Variant struct {
	Key         string `yaml:"key,omitempty"`
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Rule struct {
	FlagKey       string          `yaml:"flag,omitempty"`
	SegmentKey    string          `yaml:"segment,omitempty"`
	Rank          uint            `yaml:"rank,omitempty"`
	Distributions []*Distribution `yaml:"distributions,omitempty"`
}

type Distribution struct {
	VariantKey string  `yaml:"variant,omitempty"`
	Rollout    float32 `yaml:"rollout,omitempty"`
}

type Segment struct {
	Key         string        `yaml:"key,omitempty"`
	Name        string        `yaml:"name,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Constraints []*Constraint `yaml:"constraints,omitempty"`
}

type Constraint struct {
	Type     string `yaml:"type,omitempty"`
	Property string `yaml:"property,omitempty"`
	Operator string `yaml:"operator,omitempty"`
	Value    string `yaml:"value,omitempty"`
}

func runExport() error {
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

	if err != nil {
		return fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	out, err := os.Create("export.yaml")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	defer out.Close()

	enc := yaml.NewEncoder(out)
	defer enc.Close()

	var (
		flagStore    = db.NewFlagStore(builder)
		segmentStore = db.NewSegmentStore(builder)
		// ruleStore    = db.NewRuleStore(builder, sql)
	)

	// export flags/variants
	flagList := &struct {
		Flags []*Flag `yaml:"flags,omitempty"`
	}{}

	flags, err := flagStore.ListFlags(ctx, 0, 0)
	if err != nil {
		return fmt.Errorf("getting flags: %w", err)
	}

	for _, f := range flags {
		flag := &Flag{
			Key:         f.Key,
			Name:        f.Name,
			Description: f.Description,
			Enabled:     f.Enabled,
		}

		for _, v := range f.Variants {
			flag.Variants = append(flag.Variants, &Variant{
				Key:         v.Key,
				Name:        v.Name,
				Description: v.Description,
			})
		}

		flagList.Flags = append(flagList.Flags, flag)
	}

	if err := enc.Encode(flagList); err != nil {
		return fmt.Errorf("exporting flags: %w", err)
	}

	// export segments/constraints
	segmentList := &struct {
		Segments []*Segment `yaml:"segments,omitempty"`
	}{}

	segments, err := segmentStore.ListSegments(ctx, 0, 0)
	if err != nil {
		return fmt.Errorf("getting segments: %w", err)
	}

	for _, s := range segments {
		segment := &Segment{
			Key:         s.Key,
			Name:        s.Name,
			Description: s.Description,
		}

		for _, c := range s.Constraints {
			segment.Constraints = append(segment.Constraints, &Constraint{
				Type:     c.Type.String(),
				Property: c.Property,
				Operator: c.Operator,
				Value:    c.Value,
			})
		}

		segmentList.Segments = append(segmentList.Segments, segment)
	}

	if err := enc.Encode(segmentList); err != nil {
		return fmt.Errorf("exporting segments: %w", err)
	}

	return nil
}

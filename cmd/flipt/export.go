package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/db"
	"github.com/markphelps/flipt/storage/db/postgres"
	"github.com/markphelps/flipt/storage/db/sqlite"
	"gopkg.in/yaml.v2"
)

type Document struct {
	Flags    []*Flag    `yaml:"flags,omitempty"`
	Segments []*Segment `yaml:"segments,omitempty"`
}

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

const batchSize = 25

var exportFilename string

func runExport(_ []string) error {
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

	var storeProvider storage.Provider

	switch driver {
	case db.SQLite, db.MySQL:
		storeProvider = sqlite.NewProvider(sql)
	case db.Postgres:
		storeProvider = postgres.NewProvider(sql)
	}

	// default to stdout
	var out io.WriteCloser = os.Stdout

	// export to file
	if exportFilename != "" {
		logger.Debugf("exporting to %q", exportFilename)

		out, err = os.Create(exportFilename)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}

		fmt.Fprintf(out, "# exported by Flipt (%s) on %s\n\n", version, time.Now().UTC().Format(time.RFC3339))
	}

	defer out.Close()

	var (
		flagStore    = storeProvider.FlagStore()
		segmentStore = storeProvider.SegmentStore()
		ruleStore    = storeProvider.RuleStore()

		enc = yaml.NewEncoder(out)
		doc = new(Document)
	)

	defer enc.Close()

	var remaining = true

	// export flags/variants in batches
	for batch := uint64(0); remaining; batch++ {
		flags, err := flagStore.ListFlags(ctx, storage.WithOffset(batch*batchSize), storage.WithLimit(batchSize))
		if err != nil {
			return fmt.Errorf("getting flags: %w", err)
		}

		remaining = len(flags) == batchSize

		for _, f := range flags {
			flag := &Flag{
				Key:         f.Key,
				Name:        f.Name,
				Description: f.Description,
				Enabled:     f.Enabled,
			}

			// map variant id => variant key
			variantKeys := make(map[string]string)

			for _, v := range f.Variants {
				flag.Variants = append(flag.Variants, &Variant{
					Key:         v.Key,
					Name:        v.Name,
					Description: v.Description,
				})

				variantKeys[v.Id] = v.Key
			}

			// export rules for flag
			rules, err := ruleStore.ListRules(ctx, flag.Key)
			if err != nil {
				return fmt.Errorf("getting rules for flag %q: %w", flag.Key, err)
			}

			for _, r := range rules {
				rule := &Rule{
					SegmentKey: r.SegmentKey,
					Rank:       uint(r.Rank),
				}

				for _, d := range r.Distributions {
					rule.Distributions = append(rule.Distributions, &Distribution{
						VariantKey: variantKeys[d.VariantId],
						Rollout:    d.Rollout,
					})
				}

				flag.Rules = append(flag.Rules, rule)
			}

			doc.Flags = append(doc.Flags, flag)
		}
	}

	remaining = true

	// export segments/constraints in batches
	for batch := uint64(0); remaining; batch++ {
		segments, err := segmentStore.ListSegments(ctx, storage.WithOffset(batch*batchSize), storage.WithLimit(batchSize))
		if err != nil {
			return fmt.Errorf("getting segments: %w", err)
		}

		remaining = len(segments) == batchSize

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

			doc.Segments = append(doc.Segments, segment)
		}
	}

	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sq "github.com/Masterminds/squirrel"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage/db"
	"gopkg.in/yaml.v2"
)

type FlagList struct {
	Flags []*Flag `yaml:"flags"`
}

type Flag struct {
	Key         string     `yaml:"key"`
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Enabled     bool       `yaml:"enabled"`
	Variants    []*Variant `yaml:variants,flow"`
}

type Variant struct {
	Key         string `yaml:"key"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
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

	var (
		flagStore = db.NewFlagStore(builder)
		//segmentStore = db.NewSegmentStore(builder)
		//ruleStore    = db.NewRuleStore(builder, sql)
	)

	flags, err := flagStore.ListFlags(ctx, &flipt.ListFlagRequest{})

	if err != nil {
		return fmt.Errorf("getting flags: %w", err)
	}

	out, err := os.Create("export.yaml")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	defer out.Close()

	flagList := &FlagList{}

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

	enc := yaml.NewEncoder(out)
	defer enc.Close()

	if err := enc.Encode(flagList); err != nil {
		return fmt.Errorf("exporting flags: %w", err)
	}

	return nil
}

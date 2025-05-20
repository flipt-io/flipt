// Flipt Enterprise-Only Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid Enterprise license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package git

import (
	"bytes"
	"context"

	"github.com/go-git/go-git/v5/plumbing"
	"go.flipt.io/flipt/errors"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/internal/storage/environments/git"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ serverenvs.Environment = (*Environment)(nil)

type SCM interface {
	Propose(context.Context, ProposalRequest) (*environments.ProposeEnvironmentResponse, error)
}

type ProposalRequest struct {
	Base  string
	Head  string
	Title string
	Body  string
}

type Environment struct {
	logger *zap.Logger
	*git.Environment
	SCM SCM
}

func (e *Environment) Propose(ctx context.Context, branch serverenvs.Environment) (resp *environments.ProposeEnvironmentResponse, err error) {
	var (
		baseCfg   = e.Configuration()
		branchCfg = branch.Configuration()
	)

	if branchCfg.Base != nil && *branchCfg.Base != e.Key() {
		return nil, errors.ErrInvalidf("environment %q is not a based on environment %q", branch.Key(), e.Key())
	}

	type templateContext struct {
		Base   *environments.EnvironmentConfiguration
		Branch *environments.EnvironmentConfiguration
	}

	if err := e.Repository().View(ctx, branchCfg.Branch, func(hash plumbing.Hash, src environmentsfs.Filesystem) error {
		// chroot our filesystem to the configured directory
		src = environmentsfs.SubFilesystem(src, baseCfg.Directory)

		conf, err := storagefs.GetConfig(e.logger, environmentsfs.ToFS(src))
		if err != nil {
			return err
		}

		tmplCtx := templateContext{Base: baseCfg, Branch: branchCfg}

		title := &bytes.Buffer{}
		if err := conf.Templates.ProposalTitleTemplate.Execute(title, tmplCtx); err != nil {
			return err
		}

		body := &bytes.Buffer{}
		if err := conf.Templates.ProposalBodyTemplate.Execute(body, tmplCtx); err != nil {
			return err
		}

		resp, err = e.SCM.Propose(ctx, ProposalRequest{
			Base:  baseCfg.Branch,
			Head:  branchCfg.Branch,
			Title: title.String(),
			Body:  body.String(),
		})

		return err
	}); err != nil {
		return nil, err
	}

	return
}

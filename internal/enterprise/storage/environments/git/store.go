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
	Propose(context.Context, ProposalRequest) (*environments.EnvironmentProposalDetails, error)
	ListChanges(context.Context, ListChangesRequest) (*environments.ListBranchedEnvironmentChangesResponse, error)
	ListProposals(context.Context, serverenvs.Environment) (map[string]*environments.EnvironmentProposalDetails, error)
}

type ProposalRequest struct {
	Base  string
	Head  string
	Title string
	Body  string
	Draft bool
}

type ListChangesRequest struct {
	Base  string
	Head  string
	Since string
	Limit int32
}

type Environment struct {
	logger *zap.Logger
	*git.Environment
	SCM SCM
}

func NewEnvironment(logger *zap.Logger, env *git.Environment, scm SCM) *Environment {
	return &Environment{
		logger:      logger,
		Environment: env,
		SCM:         scm,
	}
}

func (e *Environment) ListBranchedChanges(ctx context.Context, branch serverenvs.Environment) (resp *environments.ListBranchedEnvironmentChangesResponse, err error) {
	var (
		baseCfg   = e.Configuration()
		branchCfg = branch.Configuration()
	)

	if branchCfg.Base != nil && *branchCfg.Base != e.Key() {
		return nil, errors.ErrInvalidf("environment %q is not a based on environment %q", e.Key(), branch.Key())
	}

	return e.SCM.ListChanges(ctx, ListChangesRequest{
		Base:  baseCfg.Branch,
		Head:  branchCfg.Branch,
		Limit: 10,
	})
}

func (e *Environment) Propose(ctx context.Context, base serverenvs.Environment, opts serverenvs.ProposalOptions) (resp *environments.EnvironmentProposalDetails, err error) {
	var (
		baseCfg   = e.Configuration()
		branchCfg = base.Configuration()
	)

	if branchCfg.Base != nil && *branchCfg.Base != e.Key() {
		return nil, errors.ErrInvalidf("environment %q is not a based on environment %q", e.Key(), base.Key())
	}

	type templateContext struct {
		Base   *environments.EnvironmentConfiguration
		Branch *environments.EnvironmentConfiguration
	}

	if err := e.Repository().View(ctx, branchCfg.Branch, func(hash plumbing.Hash, src environmentsfs.Filesystem) error {
		// chroot our filesystem to the configured directory
		dir := ""
		if baseCfg.Directory != nil {
			dir = *baseCfg.Directory
		}
		src = environmentsfs.SubFilesystem(src, dir)

		conf, err := storagefs.GetConfig(e.logger, environmentsfs.ToFS(src))
		if err != nil {
			return err
		}

		var (
			title = &bytes.Buffer{}
			body  = &bytes.Buffer{}
		)

		tmplCtx := templateContext{Base: baseCfg, Branch: branchCfg}

		if opts.Title != "" {
			title.WriteString(opts.Title)
		} else {
			if err := conf.Templates.ProposalTitleTemplate.Execute(title, tmplCtx); err != nil {
				return err
			}
		}

		if opts.Body != "" {
			body.WriteString(opts.Body)
		} else {
			if err := conf.Templates.ProposalBodyTemplate.Execute(body, tmplCtx); err != nil {
				return err
			}
		}

		resp, err = e.SCM.Propose(ctx, ProposalRequest{
			Base:  baseCfg.Branch,
			Head:  branchCfg.Branch,
			Title: title.String(),
			Body:  body.String(),
			Draft: opts.Draft,
		})

		return err
	}); err != nil {
		return nil, err
	}

	return
}

func (e *Environment) ListBranches(ctx context.Context) (*environments.ListEnvironmentBranchesResponse, error) {
	proposals, err := e.SCM.ListProposals(ctx, e)
	if err != nil {
		return nil, err
	}

	branches, err := e.Environment.ListBranches(ctx)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches.Branches {
		branch.Proposal = proposals[branch.Branch]
	}

	return branches, nil
}

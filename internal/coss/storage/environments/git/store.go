// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

package git

import (
	"bytes"
	"context"

	"github.com/go-git/go-git/v6/plumbing"
	"go.flipt.io/flipt/errors"
	serverenvs "go.flipt.io/flipt/internal/server/environments"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/internal/storage/environments/git"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

var _ serverenvs.Environment = (*Environment)(nil)

// templateContext is the data made available to the proposal title and body
// templates when they are rendered.
type templateContext struct {
	Base   *environments.EnvironmentConfiguration
	Branch *environments.EnvironmentConfiguration
}

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

	resp, err = e.SCM.ListChanges(ctx, ListChangesRequest{
		Base:  baseCfg.Ref,
		Head:  branchCfg.Ref,
		Limit: 10,
	})
	if err != nil {
		return nil, err
	}

	// Render the hydrated proposal defaults so the UI can pre-fill (and let the
	// user override) the title and body. Failures here should not prevent the
	// caller from listing changes, so degrade gracefully.
	title, body, rerr := e.renderProposalDefaults(ctx, baseCfg, branchCfg)
	if rerr != nil {
		e.logger.Warn("rendering proposal defaults", zap.Error(rerr))
	} else {
		resp.ProposalTitle = title
		resp.ProposalBody = body
	}

	return resp, nil
}

// renderProposalDefaults renders the default proposal title and body for the
// given branch by executing the hydrated templates (built-in defaults overlaid
// with server-level and repository-level overrides) against the branch context.
func (e *Environment) renderProposalDefaults(ctx context.Context, baseCfg, branchCfg *environments.EnvironmentConfiguration) (title, body string, err error) {
	err = e.Repository().View(ctx, branchCfg.Ref, func(hash plumbing.Hash, src environmentsfs.Filesystem) error {
		// chroot our filesystem to the configured directory
		dir := ""
		if baseCfg.Directory != nil {
			dir = *baseCfg.Directory
		}
		src = environmentsfs.SubFilesystem(src, dir)

		conf, err := storagefs.GetConfig(e.logger, environmentsfs.ToFS(src), e.ServerTemplates())
		if err != nil {
			return err
		}

		tmplCtx := templateContext{Base: baseCfg, Branch: branchCfg}

		var titleBuf, bodyBuf bytes.Buffer
		if err := conf.Templates.ProposalTitleTemplate.Execute(&titleBuf, tmplCtx); err != nil {
			return err
		}
		if err := conf.Templates.ProposalBodyTemplate.Execute(&bodyBuf, tmplCtx); err != nil {
			return err
		}

		title, body = titleBuf.String(), bodyBuf.String()
		return nil
	})

	return title, body, err
}

func (e *Environment) Propose(ctx context.Context, base serverenvs.Environment, opts serverenvs.ProposalOptions) (resp *environments.EnvironmentProposalDetails, err error) {
	var (
		baseCfg   = e.Configuration()
		branchCfg = base.Configuration()
	)

	if branchCfg.Base != nil && *branchCfg.Base != e.Key() {
		return nil, errors.ErrInvalidf("environment %q is not a based on environment %q", e.Key(), base.Key())
	}

	// Start from the hydrated defaults, then let any caller-supplied title/body
	// override them. Supplied values are used verbatim (already rendered by the
	// UI), matching the previous behavior.
	title, body, err := e.renderProposalDefaults(ctx, baseCfg, branchCfg)
	if err != nil {
		return nil, err
	}

	if opts.Title != "" {
		title = opts.Title
	}

	if opts.Body != "" {
		body = opts.Body
	}

	return e.SCM.Propose(ctx, ProposalRequest{
		Base:  baseCfg.Ref,
		Head:  branchCfg.Ref,
		Title: title,
		Body:  body,
		Draft: opts.Draft,
	})
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
		branch.Proposal = proposals[branch.Ref]
	}

	return branches, nil
}

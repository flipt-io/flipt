//go:build mage
// +build mage

package main

import (
	"context"

	"github.com/magefile/mage/mg"
	"go.flipt.io/build/release"
)

type Release mg.Namespace

func (r Release) Next(ctx context.Context, module, versionParts string) error {
	return release.Next(module, versionParts)
}

func (r Release) Latest(ctx context.Context, module string) error {
	return release.Latest(module, false)
}

func (r Release) LatestRC(ctx context.Context, module string) error {
	return release.Latest(module, true)
}

func (r Release) Changelog(ctx context.Context, module, version string) error {
	return release.UpdateChangelog(module, version)
}

func (r Release) Tag(ctx context.Context, module, version string) error {
	return release.Tag(ctx, module, version)
}

package release

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v50/github"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

const rootModule = "go.flipt.io/flipt"

func PrepareChangelog(module, version string) error {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mod, err := os.ReadFile(path.Join(curDir, "go.mod"))
	if err != nil {
		panic(err)
	}

	if modfile.ModulePath(mod) == path.Join(rootModule, "build") {
		if err := os.Chdir("../.."); err != nil {
			return err
		}

		defer os.Chdir("hack/build")
	}

	if !strings.HasPrefix(module, rootModule) {
		return fmt.Errorf("Expected module %q to be prefixed with %q", module, rootModule)
	}

	prefix := module[len(rootModule):]
	if len(prefix) > 0 && prefix[0] == '/' {
		prefix = prefix[1:]
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	tags, err := versionTags(repo, prefix)
	if err != nil {
		return err
	}

	// go-git doesn't support half the functionality of git log
	// It ends up brining along a lot of extra commits.
	// So for now, I am just going to pop a shell.
	rng := "HEAD"
	if len(tags) > 0 {
		latest := tags[len(tags)-1]
		if semver.Compare(version, latest) < 1 {
			return fmt.Errorf("requested version %v must be newer than latest %v", version, latest)
		}

		rng = latest + ".." + rng
	}

	logCmd := fmt.Sprintf(`git --no-pager log --format="%%s" %s -- %s`, rng, prefix)
	cmd := exec.Command("sh", "-c", logCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	messages := strings.Split(strings.TrimSpace(string(out)), "\n")

	return insertChangeLogEntryIntoFile(
		path.Join(prefix, "CHANGELOG.md"),
		parseChangeLogVersion(prefix, version, time.Now(), messages...),
	)
}

func versionTags(repo *git.Repository, prefix string) (v []string, err error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		name := strings.TrimPrefix(string(ref.Name()), "refs/tags/")
		if strings.HasPrefix(name, prefix) {
			v = append(v, name)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	semver.Sort(v)

	// filter out all prerelease versions
	for i := 0; i < len(v)-1; {
		if semver.Prerelease(v[i]) != "" {
			if len(v) == i+1 {
				v = v[:i]
				continue
			}

			v = append(v[:i], v[i+1:]...)
			continue
		}
		i++
	}

	return
}

func Submodules(ctx context.Context, client *dagger.Client, tag string) error {
	var (
		appID, _          = strconv.ParseInt(os.Getenv("FLIPT_RELEASE_BOT_APP_ID"), 10, 64)
		installationID, _ = strconv.ParseInt(os.Getenv("FLIPT_RELEASE_BOT_INSTALLATION_ID"), 10, 64)
		privateKey        = os.Getenv("FLIPT_RELEASE_BOT_APP_PEM")
	)

	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(privateKey))
	if err != nil {
		return err
	}

	ghClient := github.NewClient(&http.Client{Transport: itr})

	base := client.Container().From("golang:1.20-alpine3.17").
		WithExec([]string{"apk", "add", "-U", "--no-cache", "git"}).
		WithExec([]string{"git", "clone", "https://github.com/flipt-io/flipt.git", "/src/flipt"}).
		WithWorkdir("/src/flipt")

	tagList, err := base.WithExec([]string{"git", "tag", "--list", "v*"}).Stdout(ctx)
	if err != nil {
		return err
	}

	tags := strings.Split(tagList, "\n")
	semver.Sort(tags)

	lastTwo := tags[len(tags)-2:]

	// ensure we only attempt to publish updates when requested
	// tag matches the latest semver in default branch.
	if lastTwo[1] != tag {
		return fmt.Errorf("Publish tag %q does not match latest semver tag %q. Aborting.", tag, lastTwo[1])
	}

	for _, submodule := range []string{
		"errors",
		"rpc/flipt",
	} {
		diff, err := base.WithExec([]string{"git", "diff", "--shortstat", lastTwo[0], lastTwo[1], "--", submodule}).Stdout(ctx)
		if err != nil {
			return err
		}

		if strings.TrimSpace(diff) == "" {
			fmt.Printf("Nothing changed between %q and %q. Skipping sub-module release.", lastTwo[0], lastTwo[1])
			continue
		}

		if err := tagSubmodule(ctx, ghClient, submodule, lastTwo[1]); err != nil {
			return err
		}
	}

	return nil
}

func tagSubmodule(ctx context.Context, client *github.Client, submodule, tag string) error {
	const (
		owner = "flipt-io"
		repo  = "flipt"
	)
	ref, _, err := client.Git.GetRef(ctx, owner, repo, "tags/"+tag)
	if err != nil {
		return err
	}

	if ref.GetObject().GetType() != "tag" && ref.GetObject().GetType() != "commit" {
		return fmt.Errorf("unexpected object type %q", ref.GetObject().GetType())
	}

	target := path.Join(submodule, tag)

	// create Tag object based on existing repo tag SHA
	created, _, err := client.Git.CreateTag(ctx, owner, repo, &github.Tag{
		Tag:     github.String(target),
		Message: github.String(fmt.Sprintf("Releasing go.flipt.io/flipt/%s version %s", submodule, tag)),
		Tagger: &github.CommitAuthor{
			Name:  github.String("flipt-release[bot]"),
			Email: github.String("dev@flipt.io"),
			Date:  &github.Timestamp{Time: time.Now()},
		},
		Object: &github.GitObject{
			Type: github.String("commit"),
			SHA:  ref.GetObject().SHA,
		},
	})
	if err != nil {
		return err
	}

	// create tag ref: https://docs.github.com/en/rest/git/tags?apiVersion=2022-11-28#create-a-tag-object
	_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String("tags/" + target),
		Object: created.Object,
	})

	return err
}

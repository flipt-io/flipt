package release

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

const rootModule = "go.flipt.io/flipt"

type VersionPart string

const (
	RC    VersionPart = "RC"
	PATCH VersionPart = "PATCH"
	MINOR VersionPart = "MINOR"
	MAJOR VersionPart = "MAJOR"
)

func parseVersionPart(v string) (VersionPart, error) {
	vp := VersionPart(strings.ToUpper(v))
	switch vp {
	case RC, PATCH, MINOR, MAJOR:
		return vp, nil
	default:
		return "", fmt.Errorf("unexpected version part: %q", v)
	}
}

func Next(module string, parts string) error {
	defer chdirRoot()()

	if len(parts) == 0 {
		return errors.New("expected at least one version part")
	}

	_, latest, err := latestRelease(module, true)
	if err != nil {
		return err
	}

	nextVersion := latest
	for _, part := range strings.Split(parts, "+") {
		vp, err := parseVersionPart(part)
		if err != nil {
			return err
		}

		nextVersion, err = next(nextVersion, vp)
		if err != nil {
			return err
		}
	}

	fmt.Println(nextVersion)

	return nil
}

func next(current string, vp VersionPart) (string, error) {
	version := strings.SplitN(semver.Canonical(current), ".", 3)
	var (
		major, _ = strconv.ParseInt(version[0][1:], 10, 64)
		minor, _ = strconv.ParseInt(version[1], 10, 64)
		patch, _ = strconv.ParseInt(version[2], 10, 64)
		pre      = semver.Prerelease(current)
	)

	if vp == RC {
		if pre != "" && !strings.HasPrefix(pre, "-rc") {
			return "", fmt.Errorf("unexpected existing prerelease on latest: %q", current)
		}

		if strings.HasPrefix(pre, "-rc") {
			rc, err := strconv.ParseInt(pre[3:], 10, 64)
			if err != nil {
				return "", err
			}

			pre = fmt.Sprintf("-rc%d", rc+1)
		} else {
			pre = "-rc1"
		}

		return fmt.Sprintf("v%d.%d.%d%s", major, minor, patch, pre), nil
	}

	if pre != "" {
		// latest release was a pre-release and so we can consider
		// the next release just drops the "pre" part for any version.
		return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
	}

	switch vp {
	case PATCH:
		patch++
	case MINOR:
		patch = 0
		minor++
	case MAJOR:
		patch = 0
		minor = 0
		major++
	default:
		return "", fmt.Errorf("unexpected version part: %q", vp)
	}

	return fmt.Sprintf("v%d.%d.%d%s", major, minor, patch, pre), nil
}

func Latest(module string, includePre bool) error {
	defer chdirRoot()()

	_, latest, err := latestRelease(module, includePre)
	if err != nil {
		return err
	}

	fmt.Println(latest)

	return nil
}

func UpdateChangelog(module, version string) error {
	defer chdirRoot()()

	// If we're generating a pre-release then we compare to the last pre-release.
	// Else, we compare with the latest non-prerelease to get maximum detail.
	prefix, latest, err := latestRelease(module, semver.Prerelease(version) != "")
	if err != nil {
		return err
	}

	// go-git doesn't support half the functionality of git log
	// It ends up bringing along a lot of extra commits.
	// So for now, I am just going to pop a shell.
	rng := "HEAD"
	if latest != "" {
		if err := ensureVersionAfter(version, latest); err != nil {
			return err
		}

		rng = path.Join(prefix, latest) + ".." + rng
	}

	logCmd := fmt.Sprintf(`git --no-pager log --format="%%s" %s -- %s`, rng, prefix)
	cmd := exec.Command("sh", "-c", logCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	messages := strings.Split(strings.TrimSpace(string(out)), "\n")

	return insertChangeLogEntryIntoFile(
		path.Join(prefix, "..", "CHANGELOG.md"),
		parseChangeLogVersion(prefix, version, time.Now(), messages...),
	)
}

func Tag(ctx context.Context, module, version string) error {
	defer chdirRoot()()

	prefix, latest, err := latestRelease(module, true)
	if err != nil {
		return err
	}

	if err := ensureVersionAfter(version, latest); err != nil {
		return err
	}

	diff, err := exec.Command("git", "diff", "--shortstat", path.Join(prefix, latest), "--", prefix).CombinedOutput()
	if err != nil {
		return err
	}

	if strings.TrimSpace(string(diff)) == "" {
		fmt.Printf(
			"Nothing changed since %q. Skipping sub-module release.",
			latest,
		)
		return nil
	}

	return tagModule(ctx, module, version)
}

func ensureVersionAfter(requested, latest string) error {
	if semver.Compare(requested, latest) < 1 {
		return fmt.Errorf("requested version %v must be newer than latest %v", requested, latest)
	}

	return nil
}

func tagModule(_ context.Context, module, version string) error {
	prefix, err := moduleTagPrefix(module)
	if err != nil {
		return err
	}

	var (
		tag     = path.Join(prefix, version)
		message = fmt.Sprintf("Releasing %s version %s", module, version)
		cmd     = exec.Command("git", "tag", "-a", "-m", message, tag)
	)

	cmd.Env = append(cmd.Env,
		"GIT_AUTHOR_NAME=flipt-release[bot]",
		"GIT_AUTHOR_EMAIL=dev@flipt.io",
		fmt.Sprintf("GIT_AUTHOR_DATE=%s", time.Now().UTC().Format(time.RFC3339)),
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Println(strings.TrimSpace(string(out)))

	return err
}

func chdirRoot() func() {
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	mod, err := os.ReadFile(path.Join(curDir, "go.mod"))
	if err != nil {
		panic(err)
	}

	if modfile.ModulePath(mod) == path.Join(rootModule, "build") {
		if err := os.Chdir(".."); err != nil {
			panic(err)
		}

		return func() { _ = os.Chdir("build") }
	}

	return func() {}
}

func moduleTagPrefix(module string) (prefix string, _ error) {
	if !strings.HasPrefix(module, rootModule) {
		return "", fmt.Errorf("expected module %q to be prefixed with %q", module, rootModule)
	}

	prefix = module[len(rootModule):]
	if len(prefix) > 0 && prefix[0] == '/' {
		prefix = prefix[1:]
	}

	return
}

func latestRelease(module string, includePre bool) (prefix string, version string, err error) {
	prefix, err = moduleTagPrefix(module)
	if err != nil {
		return
	}

	repo, err := git.PlainOpen("../.")
	if err != nil {
		return "", "", err
	}

	tags, err := versionTags(repo, prefix, includePre)
	if err != nil {
		return "", "", err
	}

	if len(tags) < 1 {
		return prefix, "", nil
	}

	return prefix, tags[len(tags)-1], nil
}

func versionTags(repo *git.Repository, prefix string, includePre bool) (v []string, err error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	fix := func(v string) string {
		left, right, match := strings.Cut(v, "-")
		if !match {
			return semver.Canonical(v)
		}

		if left = semver.Canonical(left); left == "" {
			return left
		}

		return left + "-" + right
	}

	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		name := strings.TrimPrefix(string(ref.Name()), "refs/tags/")
		if prefix == "" {
			if strings.HasPrefix(name, "v") {
				v = append(v, fix(name))
			}

			return nil
		}

		if strings.HasPrefix(name, prefix) {
			v = append(v, fix(name[len(prefix)+1:]))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	semver.Sort(v)

	// filter out all prerelease versions
	if !includePre {
		for i := 0; i < len(v); {
			if semver.Prerelease(v[i]) != "" {
				if i+1 == len(v) {
					v = v[:i]
					continue
				}

				v = append(v[:i], v[i+1:]...)
				continue
			}
			i++
		}
	}

	return
}

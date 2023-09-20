package release

import (
	"context"
	"fmt"
	"regexp"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v32/github"
)

type Info struct {
	CurrentVersion   string
	LatestVersion    string
	LatestVersionURL string
	UpdateAvailable  bool
}

type githubReleaseChecker interface {
	getLatestRelease(ctx context.Context) (*github.RepositoryRelease, error)
}

type repositoryService interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

type githubReleaseCheckerImpl struct {
	client repositoryService
}

func (c *githubReleaseCheckerImpl) getLatestRelease(ctx context.Context) (*github.RepositoryRelease, error) {
	release, _, err := c.client.GetLatestRelease(ctx, "flipt-io", "flipt")
	if err != nil {
		return nil, fmt.Errorf("checking for latest version: %w", err)
	}

	return release, nil
}

var (
	devVersionRegex              = regexp.MustCompile(`dev$`)
	snapshotVersionRegex         = regexp.MustCompile(`snapshot$`)
	releaseCandidateVersionRegex = regexp.MustCompile(`rc.*$`)

	// defaultReleaseChecker checks for the latest release
	// can be overridden for testing
	defaultReleaseChecker githubReleaseChecker = &githubReleaseCheckerImpl{
		client: github.NewClient(nil).Repositories,
	}
)

// Check checks for the latest release and returns an Info struct containing
// the current version, latest version, if the current version is a release, and
// if an update is available.
func Check(ctx context.Context, version string) (Info, error) {
	return check(ctx, defaultReleaseChecker, version)
}

// visible for testing
func check(ctx context.Context, rc githubReleaseChecker, version string) (Info, error) {
	i := Info{
		CurrentVersion: version,
	}

	cv, err := semver.ParseTolerant(version)
	if err != nil {
		return i, fmt.Errorf("parsing current version: %w", err)
	}

	release, err := rc.getLatestRelease(ctx)
	if err != nil {
		return i, fmt.Errorf("checking for latest release: %w", err)
	}

	if release != nil {
		var err error
		lv, err := semver.ParseTolerant(release.GetTagName())
		if err != nil {
			return i, fmt.Errorf("parsing latest version: %w", err)
		}

		i.LatestVersion = lv.String()

		// if current version is less than latest version, an update is available
		if cv.Compare(lv) < 0 {
			i.UpdateAvailable = true
			i.LatestVersionURL = release.GetHTMLURL()
		}
	}

	return i, nil
}

func Is(version string) bool {
	return !devVersionRegex.MatchString(version) && !snapshotVersionRegex.MatchString(version) && !releaseCandidateVersionRegex.MatchString(version)
}

# Release

This document describes the release process for Flipt.

The release process is managed by the [Release GitHub Action](.github/workflows/release.yml) and is triggered by pushing a tag to the repository that matches the pattern `v*` (e.g. `v1.0.0`).

Flipt is currently released as a single binary for Linux (arm64 and amd64) and macOS (arm64). The release process builds the binary for each platform and uploads them to the [Flipt GitHub Releases page](https://github.com/flipt-io/flipt/releases).

We also publish a Docker image for each release to [Docker Hub](https://hub.docker.com/r/flipt/flipt/tags) and [GitHub Container Registry](https://github.com/flipt-io/flipt/pkgs/container/flipt).

## Release Process

We use GitHub Actions and [GoReleaser Pro](https://goreleaser.com/) to build and publish Flipt releases.

We support three main types of releases:

- Stable releases (e.g. `v1.0.0`)
- Snapshot releases (e.g. `v1.0.0-snapshot`)
- Nightly releases (e.g. `v1.0.0-nightly`)

The process for each looks roughly the same using GitHub Actions and GoReleaser Pro, but there are some differences in how the release is tagged and how the release is published.

![Release Process](.github/images/release-process.png)

### Stable Releases

1. Create a new branch from `main` named `release/vX.Y.Z` (e.g. `release/v1.0.0`).
2. Update the CHANGELOG.md with the release notes for the new version. (We have a `mage` task to help with this in the `build` directory.)
3. Commit the changes and push the branch to GitHub.
4. Create a pull request from the release branch to `main`.
5. Once the pull request is merged, create a new tag on `main` with the version number (e.g. `v1.0.0`).
6. Push the tag to GitHub.
7. CI will build and publish the release to GitHub and Docker Hub.

### Snapshot Releases

Snapshot releases are created on-demand and are used to test the build and release process for stable releases.

To create a snapshot release, we run the [Snapshot Release GitHub Action](.github/workflows/snapshot.yml) workflow manually from the GitHub Actions UI.

The snapshot release does not create an actual release, it simply builds the binaries and uploads them as artifacts to the workflow run. The artifacts can then be downloaded and tested.

### Nightly Releases

Nightly releases are created automatically every night at 23:30 UTC via the [Nightly Release GitHub Action](.github/workflows/nightly.yml).

The nightly release also does not create an actual release. Unlike the snapshot release, the nightly release does not upload the binaries as artifacts. Instead, the nightly release builds a Docker image and pushes it to Docker Hub and GitHub Container Registry (Linux arm and x64 only). The Docker image is tagged with the version number (e.g. `v1.0.0-nightly`).

Currently the nightly release only builds the Docker image for Linux. We do not build a nightly release for macOS.

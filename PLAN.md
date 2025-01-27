## Development Plan

This document is a high-level plan for the development of Flipt v2. This is a working document and will be updated as we progress.

Please note that this is a high-level plan and will be refined as we progress.

If you'd like to discuss this plan or add any additional ideas, please open an issue and tag it with `v2`.

### Goals

- Remove database dependencies for storing flag state and instead use a declarative storage backend
- Support GitOps workflows completely, including write operations
- Maintain compatibility with the current Flipt Evaluation APIs
- Consolidate some configuration options and remove some that are no longer needed
- Support authentication backends such as in-memory and Redis only
- Create new declarative API for managing flag configuration
- Remove legacy evaluation APIs
- Make UI improvements where needed
- Tackle any existing v1 issues that could be resolved by v2
- (Optional) Support write operations to object storage backends
- (Optional) Support approval git-based workflows

### Non-Goals

- Maintain compatibility with the current Flipt Management APIs
- Maintain backward compatibility with configuration files from v1
- Change v1 base types (flags, segments, etc) as this would require new evaluation APIs

## TODO

- [ ] Implement new declarative API for managing flag configuration
- [ ] Update UI to support new API
- [x] Remove legacy evaluation APIs
- [x] Remove database dependencies except for analytics
- [ ] Port over git storage backend
- [ ] Implement Redis and in-memory authentication backends
- [ ] Refactor and consolidate configuration options
- [ ] Fix and improve unit test coverage
- [ ] Fix and improve integration test coverage
- [ ] Package and release
    - [ ] Binary
    - [ ] Docker image
    - [ ] Helm chart
    - [ ] Homebrew tap
- [ ] Documentation
    - [ ] Create v2 docs site
    - [ ] Migrate applicable docs from v1
- [ ] Update examples

# Versioning Strategy

Chotko follows [Semantic Versioning 2.0.0](https://semver.org/).

## Version Format

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

Examples:
- `0.1.0-alpha.1` - First alpha release
- `0.1.0-beta.1` - First beta release
- `0.1.0-rc.1` - First release candidate
- `0.1.0` - First stable release
- `1.0.0` - First major release

## Version Stages

### Pre-1.0 Development (Current)

During initial development (0.x.x), the API is not considered stable:

- **0.1.x-alpha.N** - Early development, expect breaking changes
- **0.1.x-beta.N** - Feature complete for minor version, bug fixes only
- **0.1.x-rc.N** - Release candidates, final testing
- **0.1.x** - Stable release of minor version

### Post-1.0 Releases

Once we reach 1.0.0, semantic versioning rules apply strictly:

- **MAJOR** - Incompatible API changes, breaking config changes
- **MINOR** - New features, backward compatible
- **PATCH** - Bug fixes, backward compatible

## Release Process

### Creating a Release

1. Ensure all tests pass on `main` branch
2. Update CHANGELOG.md with release notes
3. Create and push a git tag:

```bash
# For alpha releases
git tag -a v0.1.0-alpha.1 -m "Release v0.1.0-alpha.1"
git push origin v0.1.0-alpha.1

# For stable releases
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

4. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Upload release assets

### Tag Format

Tags must follow the pattern: `v{MAJOR}.{MINOR}.{PATCH}[-{PRERELEASE}]`

Valid examples:
- `v0.1.0-alpha.1`
- `v0.1.0-alpha.2`
- `v0.1.0-beta.1`
- `v0.1.0-rc.1`
- `v0.1.0`
- `v0.2.0`
- `v1.0.0`

### Pre-release Identifiers

- `alpha.N` - Alpha releases (unstable, incomplete features)
- `beta.N` - Beta releases (feature complete, may have bugs)
- `rc.N` - Release candidates (final testing)

## Supported Platforms

Binary releases are built for:

| OS | Architecture | Binary Name |
|----|--------------|-------------|
| macOS | arm64 (Apple Silicon) | `chotko-darwin-arm64` |
| macOS | amd64 (Intel) | `chotko-darwin-amd64` |
| Linux | arm64 | `chotko-linux-arm64` |
| Linux | amd64 | `chotko-linux-amd64` |
| Windows | arm64 | `chotko-windows-arm64.exe` |
| Windows | amd64 | `chotko-windows-amd64.exe` |

## Version Information in Binary

The binary includes embedded version information:

```bash
$ chotko --version
chotko v0.1.0-alpha.1 (commit: abc1234, built: 2024-01-15T10:30:00Z)
```

This is set at build time via ldflags:
- `version` - Git tag or "dev"
- `commit` - Git commit SHA
- `date` - Build timestamp

## Changelog

All notable changes are documented in [CHANGELOG.md](CHANGELOG.md).

Format follows [Keep a Changelog](https://keepachangelog.com/).

## Deprecation Policy

- Deprecation warnings will be issued at least one minor version before removal
- Breaking changes are only allowed in major version bumps (post-1.0)
- Configuration format changes will include migration guides

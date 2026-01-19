# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] - 2025-01-19

### Added

- **Custom CA certificate support**: New `ca_cert` config option and `WithCustomCA()` client option for secure TLS with internal/self-signed certificates
- **Password file support**: New `--password-file` flag for secure password input (preferred over `-p` flag)
- **API rate limiting**: New `rate_limit` config option and `WithRateLimit()` client option to prevent server overload
- **Structured logging**: New `internal/logging` package using Go's `log/slog` with configurable levels, JSON output, and file logging
- **Dockerfile**: Multi-stage Docker build for containerized deployments
- **Graceful shutdown**: Proper handling of SIGTERM/SIGINT signals for container environments
- **Shared ListView component**: New `internal/components/listview` package for consistent list navigation
- **Consolidated Severity type**: New `zabbix.Severity` type with `Name()`, `ShortCode()`, and `Emoji()` methods

### Changed

- Split `update.go` (1342 lines) into 6 focused files for better maintainability
- Alerts, hosts, and events components now use shared `listview.State` for navigation
- Session token is now always cleared on logout (even if server logout fails)
- HTTP client now uses connection pooling for better performance
- Host lookups now use O(1) map index instead of O(n) linear search
- Sparkline regeneration is now selective (only changed items)
- Severity mappings consolidated from scattered locations into `zabbix.Severity`
- Config file permissions are now checked on load (warns if too permissive)

### Fixed

- Theme name in README corrected from `high-contrast` to `highcontrast`
- Added missing package comment to graphs component
- Removed unused method parameters in model.go
- Test files now use `strings.Contains` instead of custom helpers

### Security

- Added deprecation warning when using `-p` flag (visible in process list)
- CI security scan now fails on dependency vulnerabilities (not just informational)
- User-Agent header added to API requests for better audit trails

### Documentation

- Added documentation for new features (CA certs, password-file, rate limiting, logging)
- Synced AGENTS.md key bindings table with README
- Updated LICENSE copyright year to 2024-2025

## [0.6.1] - 2025-01-19

### Changed

- Updated `interface{}` to `any` throughout zabbix package (Go 1.18+ style)
- Moved `truncate` utility function to shared `format` package
- Expanded `format` package documentation

### Fixed

- API client now limits response size to 10MB (prevents memory exhaustion)
- Ignore list lookup optimized with nested map structure (eliminates string allocation)
- Alert severity counts now cached during filtering (reduces duplicate iteration)
- Added slice preallocation hints for better memory efficiency
- SECURITY.md now shows correct supported versions (0.6.x, 0.5.x)
- README now documents all config options (window_title, emoji_title, graphs section)
- README now includes high-contrast theme in themes list
- README now documents TLS/insecure_skip_verify configuration
- CHANGELOG v0.3.0 now has release date

### Documentation

- Added missing key bindings to README (i, I for ignore management)

## [0.6.0] - 2025-01-09

### Added

- **Close alert**: Press `c` to manually close/resolve alerts when the trigger allows manual close
- Dynamic window/tab title showing alert counts by severity (💥 Disaster, 🔥 High, 🚨 Average, ⚠️ Warning, ⓘ Info)
- New config options: `window_title`, `emoji_title`, `title_min_severity`
- Text fallback mode for terminals that don't support emoji in titles

### Changed

- Detail pane actions hint now conditionally shows `[c]lose` only when manual close is available
- Status bar displays message when manual close is not permitted for a trigger

## [0.5.0] - 2025-01-06

### Added

- **Runtime theme switching**: Change themes without restarting via `:theme` command or `:theme <name>`
- **High contrast theme**: WCAG AAA compliant theme for accessibility (`--theme high-contrast`)
- **Artifact signing**: Release binaries are now signed with Sigstore cosign for verification
- **Security scan**: govulncheck added to CI pipeline
- Minimum terminal size enforcement (60x12) with friendly error message
- Loading indicator during initial Zabbix connection
- Distinct severity symbols for better differentiation (⬤▲◆■●○)
- Adaptive pane proportions based on terminal width
- Environment variables documentation in README (CHOTKO_SERVER, CHOTKO_TOKEN, CHOTKO_PASSWORD)

### Changed

- Help modal now uses 2-column layout to fit smaller screens
- Host counts now calculated from loaded hosts on Hosts tab (eliminates redundant API call)

### Fixed

- Help modal now correctly shows F1-F4 tab shortcuts
- O(1) lookup performance for ignore list matching

### Security

- Password input in setup wizard is now masked
- Warning displayed when using `--password` flag (visible in process list)
- Warning displayed when `insecure_skip_verify` is enabled
- CI job timeouts added to prevent runaway builds

## [0.4.2] - 2025-01-02

### Added

- Makefile with standard Go targets (`build`, `test`, `lint`, `fmt`, `clean`, `update`)

### Changed

- Updated `.golangci.yml` to v2 format with comprehensive linter coverage
- Improved code style compliance with stricter linting rules
- More restrictive file permissions (0600 for files, 0750 for directories)
- Require TLS 1.2 minimum for Zabbix API connections

### Fixed

- Type assertion safety checks in tests
- Variable shadowing issues across multiple packages
- US English spelling consistency (canceled, canceling)
- Switch statement exhaustiveness
- String comparison idioms (`s == ""` instead of `len(s) == 0`)

## [0.4.1] - 2024-12-23

### Fixed

- Auto-refresh (30-second timer) no longer stops after opening and closing modals (editor, help, or error)

## [0.4.0] - 2024-12-13

### Added

- Host trigger editing: enable/disable triggers directly from the TUI (`t` key)
- Host macro editing: view, edit, and delete macros (`m` key)
- Toggle host monitoring on/off (`e` key on Hosts tab)
- New modal editor component for trigger and macro management
- Trigger API methods for Zabbix integration

### Fixed

- Row highlighting now displays consistently across all list components
- Command bar no longer shows duplicate prompt characters

## [0.3.0] - 2024-11-15

### Added

- Initial release of Chotko - Zabbix Terminal UI
- Real-time problem/alert monitoring with auto-refresh
- Support for Zabbix 7.x API
- Multiple authentication methods (API token, username/password)
- 7 built-in themes (default, nord, dracula, gruvbox, catppuccin, tokyonight, solarized)
- Custom theme support via YAML configuration
- Severity filtering (0-5)
- Text search filtering
- Alert acknowledgment
- Host status overview
- Events history tab with problem/recovery tracking
- Graphs tab with time series charts for numeric metrics
- Mouse support using BubbleZone library:
  - Click tabs to switch between tabs
  - Click list items to select them
  - Click tree nodes to select and expand/collapse
  - Click panes to change focus
  - Scroll wheel scrolls pane under mouse cursor
- Keyboard-driven navigation
- Configuration wizard for first-time setup
- Cross-platform support (macOS, Linux, Windows)

### Security

- Secure credential storage in config file
- API token authentication support

## [0.1.0-alpha.1] - Unreleased

### Added

- Initial alpha release
- Core TUI functionality
- Zabbix API client
- Basic theme support

[Unreleased]: https://github.com/harpchad/chotko/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/harpchad/chotko/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/harpchad/chotko/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/harpchad/chotko/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/harpchad/chotko/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/harpchad/chotko/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/harpchad/chotko/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/harpchad/chotko/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/harpchad/chotko/compare/v0.1.0-alpha.1...v0.3.0
[0.1.0-alpha.1]: https://github.com/harpchad/chotko/releases/tag/v0.1.0-alpha.1

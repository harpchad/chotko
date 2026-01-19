# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/harpchad/chotko/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/harpchad/chotko/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/harpchad/chotko/compare/v0.4.2...v0.5.0
[0.4.2]: https://github.com/harpchad/chotko/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/harpchad/chotko/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/harpchad/chotko/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/harpchad/chotko/compare/v0.1.0-alpha.1...v0.3.0
[0.1.0-alpha.1]: https://github.com/harpchad/chotko/releases/tag/v0.1.0-alpha.1

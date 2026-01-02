# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Dynamic window/tab title showing alert counts by severity (💥 Disaster, 🔥 High, 🚨 Average, ⚠️ Warning, ⓘ Info)
- New config options: `window_title`, `emoji_title`, `title_min_severity`
- Text fallback mode for terminals that don't support emoji in titles

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

## [0.3.0]

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

[Unreleased]: https://github.com/harpchad/chotko/compare/v0.4.2...HEAD
[0.4.2]: https://github.com/harpchad/chotko/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/harpchad/chotko/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/harpchad/chotko/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/harpchad/chotko/compare/v0.1.0-alpha.1...v0.3.0
[0.1.0-alpha.1]: https://github.com/harpchad/chotko/releases/tag/v0.1.0-alpha.1

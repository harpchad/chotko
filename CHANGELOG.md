# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/harpchad/chotko/compare/v0.1.0-alpha.1...HEAD
[0.1.0-alpha.1]: https://github.com/harpchad/chotko/releases/tag/v0.1.0-alpha.1

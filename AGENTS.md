# Chotko Development Progress

## Current Status: MVP Functional

### Completed

- [x] Project structure and go.mod
- [x] Config loading (XDG paths, YAML parsing)
- [x] Setup wizard for first-run configuration
- [x] Zabbix API client with token and user/password auth
- [x] Theme system with 7 built-in themes (default, nord, dracula, gruvbox, catppuccin, tokyonight, solarized)
- [x] Custom theme loader from YAML
- [x] BubbleTea app skeleton with pane management
- [x] Status bar component (host counts with availability status)
- [x] Tabs component with keyboard navigation
- [x] Alerts list component with severity colors
- [x] Detail pane component
- [x] Command input component
- [x] Modal component (error/help overlays)
- [x] Key bindings and help overlay
- [x] CLI flag parsing
- [x] **Zabbix 7.x API compatibility** - Fixed!
  - Two-step problem fetching: `problem.get` for active problems, `event.get` for host details
  - Correct integer types for `source` and `object` parameters
  - Filter out problems from disabled triggers (matches web UI behavior)
- [x] Row highlighting in alerts list
- [x] Tab navigation (`[`/`]` or `H`/`L`, `F1`-`F3`)
- [x] Host availability status (fixed Unknown vs Unavailable mapping)

### In Progress

- [x] Hosts tab view (UI exists, needs data integration)
- [ ] Events/History tab view
- [ ] Graphs tab view

### TODO

- [ ] Add CI, codecov, and code quality badges to README
- [ ] Investigate mouse support

### Known Issues

1. Tabs switch visually but only Alerts tab has functional content

---

## Future Roadmap

### Phase 2: Enhanced Monitoring

- [ ] Problem Timeline - horizontal timeline showing problem duration
- [ ] Trigger Heatmap - visual grid of trigger activity
- [ ] Event Log Stream - real-time scrolling events
- [ ] Top N Panel - hosts with highest resource usage
- [ ] Sparkline graphs for key metrics

### Phase 3: Infrastructure Views

- [ ] Host Group Tree - collapsible navigation
- [ ] Template Dependency Graph - ASCII visualization
- [ ] Proxy Status panel
- [ ] Discovery Status panel
- [ ] Network map (ASCII topology)

### Phase 4: Advanced Operations

- [ ] Mass Operations - bulk ack/close/suppress
- [ ] Maintenance Windows - view/create/delete
- [ ] Script Execution - run scripts on hosts
- [ ] User Sessions - who's logged into Zabbix
- [ ] Recent Actions log

### Phase 5: Metrics & Graphs

- [ ] Item history viewer with pagination
- [ ] Graph rendering (ASCII/braille patterns)
- [ ] SLA Dashboard - availability percentages
- [ ] Calculated item viewer

### Phase 6: Advanced Features

- [ ] Multiple server support (switch between Zabbix instances)
- [ ] SSH tunnel support for secure connections
- [ ] Notification sounds (terminal bell patterns)
- [ ] Desktop notifications integration (via notify-send/osascript)
- [ ] Export to JSON/CSV
- [ ] Vim mode (full modal editing)
- [ ] Macro expansion viewer

### Theme Additions

- [ ] High contrast theme (accessibility)
- [ ] Light mode themes (solarized-light, catppuccin-latte)
- [ ] Auto-detect terminal background (dark/light)
- [ ] ANSI-256 fallback themes for limited terminals
- [ ] ANSI-16 basic fallback
- [ ] Custom theme hot-reload

### Integration Ideas

- [ ] PagerDuty integration
- [ ] Slack webhook notifications
- [ ] Prometheus metrics export
- [ ] Grafana deep links

### UX Improvements

- [ ] Fuzzy search across all data
- [ ] Bookmarked/pinned problems
- [ ] Problem grouping by host/hostgroup
- [ ] Customizable column layouts
- [ ] Persistent filters
- [ ] Session history (undo ack)

---

## Development Guidelines

### Git Workflow

**IMPORTANT: Never push directly to main. Always use pull requests.**

1. Create a feature branch for any changes:
   ```bash
   git checkout -b fix/description-of-fix
   # or
   git checkout -b feature/description-of-feature
   ```

2. Make changes and commit with conventional commit messages:
   ```bash
   git commit -m "fix: description of the fix"
   git commit -m "feat: description of the feature"
   ```

3. Push the branch and create a PR:
   ```bash
   git push -u origin fix/description-of-fix
   gh pr create --title "fix: description" --body "Details..."
   ```

4. Wait for CI to pass, then merge via GitHub UI or:
   ```bash
   gh pr merge --squash --delete-branch
   ```

### Code Formatting

Use `gofumpt` for all Go code formatting:

```bash
# Install gofumpt
go install mvdan.cc/gofumpt@latest

# Format all files
gofumpt -w .
```

### Pre-commit Hooks

This project uses pre-commit hooks. Install them with:

```bash
# Install pre-commit
brew install pre-commit  # or pip install pre-commit

# Install hooks
pre-commit install
```

### Linting

We use `golangci-lint` for comprehensive linting:

```bash
# Install golangci-lint
brew install golangci-lint  # or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

---

## Architecture Notes

### BubbleTea Value Semantics

Important: BubbleTea uses value semantics for models. When returning `tea.Cmd` functions:

- Capture all needed values *before* creating the closure
- Pass data back through messages (e.g., `ConnectedMsg` carries the client)
- Don't rely on pointer receiver modifications being visible

### Zabbix API Notes (7.x)

- `apiinfo.version` must be called WITHOUT authorization header
- `problem.get` only returns active/unresolved problems (by design)
- `problem.get` doesn't support `selectHosts` - use `event.get` with eventids instead
- `event.get` parameters `source` and `object` must be integers, not arrays
- Use `selectRelatedObject` to get trigger status and filter disabled triggers
- Host `active_available` values: 0=Unknown, 1=Available, 2=Unavailable
- `event.acknowledge` response returns eventids as numbers (not strings) in some versions

### Two-Step Problem Fetching

```
1. problem.get → Get active problem eventids
2. event.get(eventids) → Get full details with hosts and trigger status
3. Filter out problems where trigger status=1 (disabled)
```

---

## Key Bindings

| Key | Action |
|-----|--------|
| `j`/`↓` | Move down |
| `k`/`↑` | Move up |
| `]`/`L` | Next tab |
| `[`/`H` | Previous tab |
| `F1-F3` | Jump to tab |
| `Tab` | Next pane |
| `Shift+Tab` | Previous pane |
| `a` | Acknowledge problem |
| `A` | Acknowledge with message |
| `r` | Refresh |
| `/` | Filter mode |
| `0-5` | Filter by severity |
| `Ctrl+L` | Clear filter |
| `:` | Command mode |
| `?` | Help |
| `q` | Quit |

---

## File Structure

```
chotko/
├── cmd/chotko/main.go           # Entry point, CLI flags
├── internal/
│   ├── app/                      # BubbleTea application
│   │   ├── model.go             # Main model, Init, commands
│   │   ├── update.go            # Message handling
│   │   ├── view.go              # Rendering
│   │   ├── keys.go              # Key bindings
│   │   └── messages.go          # Message types
│   ├── components/               # UI components
│   │   ├── alerts/              # Alert list
│   │   ├── command/             # Command input
│   │   ├── detail/              # Detail pane
│   │   ├── modal/               # Modal dialogs
│   │   ├── statusbar/           # Status bar
│   │   └── tabs/                # Tab bar
│   ├── config/                   # Configuration
│   ├── theme/                    # Theming system
│   └── zabbix/                   # API client
│       ├── client.go            # HTTP client, auth
│       ├── types.go             # Data structures
│       ├── problems.go          # Problem/event fetching
│       └── hosts.go             # Host fetching
├── .github/
│   ├── workflows/
│   │   ├── ci.yml               # CI pipeline
│   │   └── release.yml          # Release automation
│   └── dependabot.yml           # Dependency updates
├── .pre-commit-config.yaml      # Pre-commit hooks
├── .golangci.yml                # Linter config (v2 format)
├── AGENTS.md                    # This file
├── CHANGELOG.md                 # Release changelog
├── VERSIONING.md                # Versioning strategy
└── README.md
```

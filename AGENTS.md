# Chotko - Agent Guidelines

## Commands
```bash
go build ./...                           # Build all
go test -race ./...                      # Test all
go test -race ./internal/zabbix -run TestClient  # Single test
golangci-lint run --timeout=5m           # Lint (uses .golangci.yml v2)
```

## Code Style
- **Formatter**: `gofumpt` (stricter than gofmt) - runs via pre-commit
- **Imports**: stdlib, blank line, external, blank line, `github.com/harpchad/chotko/...`
- **Errors**: Return `fmt.Errorf("context: %w", err)`, check all errors
- **Comments**: Package comments required, exported symbols must have doc comments
- **Naming**: MixedCaps, no underscores; receivers are single letter (e.g., `m`, `c`)

## Git Workflow

**CRITICAL RULES:**
1. **Never push directly to main** - Always use pull requests
2. **Never use `--admin` flag** - Don't bypass branch protection
3. **Wait for CI checks to pass** - Don't merge until all checks complete

```bash
# 1. Create feature branch
git checkout -b fix/description   # or feature/description

# 2. Make changes and commit
git add -A && git commit -m "fix: description"

# 3. Push branch and create PR
git push -u origin fix/description
gh pr create --title "fix: description" --body "## Summary
- Brief description of changes"

# 4. Wait for CI, then merge (use --auto if checks still running)
gh pr merge --squash --delete-branch --auto
```

---

# Development Progress

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
- [x] Events/History tab view
- [x] Graphs tab view
- [x] Mouse support (click tabs, items, tree nodes; scroll wheel)
- [x] Add badges to README (release, build, codecov, Go Report Card, license)
- [x] Add LICENSE file (MIT)
- [x] Add SECURITY.md for GitHub security tab

### TODO

None currently.

### Known Issues

None currently.

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

**CRITICAL RULES - NO EXCEPTIONS:**
1. **NEVER push directly to main** - Always use pull requests
2. **NEVER use `--admin` flag** - Don't bypass branch protection checks
3. **ALWAYS wait for CI checks** - Don't merge until all checks pass

Direct pushes and admin overrides bypass CI checks, code review, and can break
the build for everyone.

1. Create a feature branch for any changes:
   ```bash
   git checkout -b fix/description-of-fix
   # or
   git checkout -b feature/description-of-feature
   # or
   git checkout -b refactor/description
   ```

2. Make changes and commit with conventional commit messages:
   ```bash
   git add -A
   git commit -m "fix: description of the fix"
   git commit -m "feat: description of the feature"
   git commit -m "refactor: description of refactoring"
   ```

3. Push the branch and create a PR:
   ```bash
   git push -u origin fix/description-of-fix
   gh pr create --title "fix: description" --body "$(cat <<'EOF'
   ## Summary
   - Brief description of changes
   EOF
   )"
   ```

4. Wait for CI checks to pass, then merge:
   ```bash
   # If checks are still running, use --auto to merge when ready
   gh pr merge --squash --delete-branch --auto

   # If checks have passed
   gh pr merge --squash --delete-branch
   ```

**Branch naming conventions:**
- `fix/` - Bug fixes
- `feature/` or `feat/` - New features
- `refactor/` - Code refactoring
- `docs/` - Documentation changes
- `chore/` - Maintenance tasks

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
| `F1-F4` | Jump to tab |
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

### Graphs Tab Keys

| Key | Action |
|-----|--------|
| `Enter`/`Space` | Toggle expand/collapse |
| `E` | Expand all |
| `C` | Collapse all |

### Mouse Support

- **Click tabs** to switch between tabs
- **Click list items** to select them
- **Click tree nodes** to select and expand/collapse (Graphs tab)
- **Click panes** to change focus
- **Scroll wheel** scrolls the pane under the mouse cursor (3 lines per tick)

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
│   │   ├── events/              # Events history list
│   │   ├── graphs/              # Graphs tree view
│   │   ├── hosts/               # Hosts list
│   │   ├── modal/               # Modal dialogs
│   │   ├── statusbar/           # Status bar
│   │   └── tabs/                # Tab bar
│   ├── config/                   # Configuration
│   ├── theme/                    # Theming system
│   └── zabbix/                   # API client
│       ├── client.go            # HTTP client, auth
│       ├── types.go             # Data structures
│       ├── problems.go          # Problem/event fetching
│       ├── hosts.go             # Host fetching
│       └── items.go             # Item/history fetching
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

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build       # Build executables to ./dist folder (uses GoReleaser)
make test        # Run all tests
make coverage    # Run tests with coverage report (add html=true for browser view)
make generate    # Regenerate OpenHue API client from OpenAPI spec
make tidy        # Clean go.mod dependencies
make clean       # Remove dist folder and coverage files
```

To run a single test:
```bash
go test ./cmd/setup -run TestDiscover
```

## Architecture

This is a Cobra-based CLI for controlling Philips Hue smart lighting. It uses the [openhue-go](https://github.com/openhue/openhue-go) library for API communication.

### Core Structure

- **Entry point**: `main.go` → `cmd/root.go` (Execute function)
- **Command groups**: Two groups defined - `hue` (light control) and `config` (setup commands)
- **Context pattern**: `openhue.Context` holds shared state (IOStreams, BuildInfo, Home, Config) passed to all commands

### Key Packages

- `openhue/` - Core domain: Config management, Context, HomeModel (hierarchical data model)
- `cmd/` - Cobra commands organized by action (`get/`, `set/`, `setup/`, `version/`)
- `util/` - Helpers for output formatting, color conversion, logging

### HomeModel Pattern

The `HomeModel` in `openhue/home_model.go` creates a hierarchical view of Hue resources:
- `HomeModel` → `Room[]` → `Device[]` → `Light`
- Each resource embeds a `Resource` struct with Id, Name, Type, Parent pointer
- Resources wrap raw HueData from the API and add convenience methods
- Search functions (`SearchLights`, `SearchRooms`, `SearchScenes`) accept name OR id for flexibility

### Command Loading Behavior

Commands in the `hue` group trigger `LoadHomeIfNeeded` in PersistentPreRun, which loads the full HomeModel from the bridge. Config commands skip this overhead.

### Configuration

Config stored in `~/.openhue/config.yaml` (or `$XDG_CONFIG_HOME/openhue/`). Contains bridge IP and application key. Commands listed in `CommandsWithNoConfig` can run without setup.

### Testing Pattern

- Use `openhue.NewTestContext(home)` to create test contexts with mock HomeModel
- Test files use `openhue/test/assert` for custom assertions
- Tests use `stretchr/testify` for standard assertions

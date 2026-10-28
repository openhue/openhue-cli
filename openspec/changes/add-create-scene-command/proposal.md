# Proposal

## Why

The CLI can list and activate scenes but cannot create or edit them, so users must open the Hue app to define a lighting setup. A complete scene toolchain — create, update, delete, define in a file, export — makes scenes scriptable, reviewable and versionable, and matches what the Hue app offers: save the current look, or define each light individually, including transitions, gradients, effects and dynamic palettes.

## What Changes

- **`openhue create scene <name> --room <room>`** (new top-level `create` command, group `hue`): creates a scene in a room. By default it captures every light's current state.
  - Light by light: repeatable `--light "<light>:<settings>"`, settings being `on|off`, `brightness=`, `rgb=`, `x=`, `y=`, `color=`, `temperature=`, `transition=<duration>`, `effect=<name>`, `gradient=<color>;<color>;…`, `gradient-mode=<mode>`. Real lights are not changed.
  - Defaults for all lights: the `set light` flags (`--on/--off`, `-b`, `--rgb`, `-x/-y`, `-c`, `-t`, `--transition-time`); per-light settings override them.
  - Dynamic scenes: `--speed <0-100>`, `--auto-dynamic`, and palette flags `--palette-color`, `--palette-temperature`, `--palette-brightness`, `--palette-effect` (repeatable).
  - `--only-listed` turns off the lights that are not listed (the bridge requires every light of the room in a scene); `--dry-run` prints the request instead of sending it; `--json` prints the created scene as JSON.
- **`openhue update scene <scene> [--room <room>]`** (new `update` command): edits an existing scene with the same light, default, dynamic and palette flags, plus `--name`, `--remove-light` (the light is turned off in the scene), `--clear-palette`. Unmentioned lights keep their stored values (including effects/gradients).
- **`--from-file <path|->` on create and update**: a YAML scene definition (name, room, scene-level dynamics and palette, defaults, per-light entries) for large or versioned scenes. Flags override file values. On `update`, the file is a partial patch.
- **`openhue get scene <scene> --yaml`**: exports a scene in the same definition format, so a scene can be exported, edited and re-applied.
- **`openhue delete scene <scene> [--room <room>]`** (new `delete` command): deletes a scene.
- Creation fails if the room is unknown or if a scene with the same name exists in that room. Success prints the scene's name and ID.

## Capabilities

### New Capabilities
- `scene-creation`: creating a scene in a room from current state, per-light settings (including transitions, effects, gradients) and dynamic palette settings, via flags.
- `scene-update`: editing an existing scene (rename, change/add/remove lights, dynamics and palette) via flags.
- `scene-definition-file`: defining and patching scenes from a YAML file.
- `scene-export`: exporting an existing scene in the definition-file format.
- `scene-deletion`: deleting a scene.

### Modified Capabilities

(none; no specs exist yet under `openspec/specs/`)

## Impact

- New packages `cmd/create/`, `cmd/update/`, `cmd/delete/` (commands + tests), registered in `cmd/root.go`; `cmd/get/get_scene.go` gains `--yaml`. Shared flag/`--light`/file parsing lives in one place.
- New scene definition model, builder and `Room`/`Scene` methods in `openhue/` wrapping `openhue-go`'s `Home.CreateScene`, `UpdateScene` and `DeleteScene` (all in v1.0.0).
- `gopkg.in/yaml.v3` becomes a direct dependency (already in `go.sum` as indirect).
- README/command docs updated, with a sample definition file. MCP server is unchanged.

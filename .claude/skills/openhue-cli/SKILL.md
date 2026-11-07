---
name: openhue-cli
description: Reference for the `openhue` CLI (this repo) — every command, subcommand, flag, alias and behavior for controlling Philips Hue lights, rooms, scenes and switch controls, plus setup and the MCP server. Use whenever writing, running or explaining an `openhue ...` invocation, adding or changing a CLI command, or answering "how do I do X with openhue" / "what flags does openhue set light take". Trigger: /openhue-cli, "openhue CLI", or any `openhue <verb>` you are unsure of.
---

# openhue CLI

Cobra CLI that controls a Philips Hue Bridge through [openhue-go](https://github.com/openhue/openhue-go).
Source of truth is `cmd/` — if this file and `openhue <cmd> --help` disagree, trust `--help` and fix this file.

Run from source: `go run . <args>` (or `make build` → `./dist`).

## Command map

| Group | Command | Aliases | What it does |
|---|---|---|---|
| Configuration | `setup` | | Discover the bridge (mDNS), pair by pressing the bridge button, save config |
| | `config -b <ip> -k <key>` | | Save bridge IP + application key manually |
| | `discover` | | Print bridges found via mDNS |
| | `completion` | | Shell completion scripts |
| | `install-skill` | | Write the root `SKILL.md` (embedded in the binary) to `~/.claude/skills/openhue/` (`$CLAUDE_CONFIG_DIR/skills` if set). `--project` → `./.claude/skills`, `--dir <skills dir>`, `--print` → stdout. Overwrites; reports installed / updated / already up to date |
| | `version` | | Build info |
| Philips Hue | `get light\|room\|scene\|control\|events` | `get`=`g` | Read resources |
| | `set light\|room\|scene\|control` | `set`=`s` | Change state / activate / configure |
| | `create scene` | `create`=`c` | Create a scene |
| | `update scene` | | Edit a scene |
| | `delete scene` | | Delete one scene (irreversible) |
| MCP Server | `mcp` | | stdio MCP server for LLM clients |

Every resource subcommand accepts singular or plural (`light`/`lights`, `scene`/`scenes`, ...).

## Config & lifecycle

- Config file: `~/.openhue/config.yaml`, or `$XDG_CONFIG_HOME/openhue/config.yaml`. Keys: `Bridge`, `Key`, optional `log_level`.
- Log file: `openhue.log` next to the config.
- Without config, any command not in `CommandsWithNoConfig` (`setup`, `config`, `help`, `discover`, `auth`, `version`, `completion`, `install-skill`) prints "not configured yet, please run the 'setup' command" and exits 0.
- Commands in the `hue` group load the full `HomeModel` from the bridge in `PersistentPreRun` before running (one round-trip of all resources). Config/MCP commands skip it.

## Resolving resources

- Lights, rooms, scenes, controls, devices are accepted by **name or ID** everywhere.
- Names are not unique across rooms: use `-r/--room <name|id>` to disambiguate (`set light`, `set scene`, `get scene`, `update scene`, `delete scene`).
- Quote names with spaces: `"Living Room"`.

## get

`-j/--json` is a persistent flag on `get` (works on every subcommand). Bare `openhue get` lists all resources and also takes `-t/--type <light|scene|room...>` (not inherited by subcommands).
Default output is a table.

| Command | Args | Flags |
|---|---|---|
| `get light [id\|name...]` | optional | `-r/--room` filter by room |
| `get room [id\|name...]` | optional | — |
| `get scene [id\|name...]` | optional, multiple | `-r/--room`, `--yaml` (export a single scene as a YAML definition reusable by `create`/`update scene --from-file`) |
| `get control [id...]` | optional | `-r/--room`, `-d/--device <name\|id>`, `--kind button\|rotary` |
| `get events` | — | Streams bridge SSE events as JSON until Ctrl+C (alpha) |

`get control` shows per button the "Short press" / "Hold" assignment: `scene <name>`, `cycle: a, b, ... (n)`, `time-based`, `smart-scene <id>`, `dim-up`, or `-`. `--json` adds the raw `configuration`.

## set light / set room

Up to 10 targets at once. **`--on` is required to light up**, even when setting brightness/color.

| Flag | Meaning |
|---|---|
| `--on` / `--off` | Power |
| `-b, --brightness <0-100>` | float |
| `--rgb #RRGGBB` | hex color |
| `-x, --cie-x` / `-y, --cie-y` | CIE xy color |
| `-c, --color <name>` | named color (`white`, `lime`, `powder_blue`, ...) |
| `-t, --temperature <153-500>` | Mirek |
| `--transition-time <duration>` | e.g. `5s`, default `0s` |
| `-r, --room` | (`set light` only) force the room of the lights |

```bash
openhue set light "Hue Play TV" Desk --on --brightness 60 --rgb "#3399FF"
openhue set room Studio --off
openhue set light office1 --on -t 250 --transition-time 30s
```

## set scene

`openhue set scene <id|name> [-r room] [-a active|dynamic|static]` — activates a scene; `-a dynamic` plays its dynamic palette.

## set control

Configures a button (rotary dials are read-only). Select it by control ID, or by device name/ID + `--control <n>` (button number on the device). The accessory must already be set up on the bridge (e.g. via the Hue app). Anything the flags do not cover is left untouched.

| Flag | Meaning |
|---|---|
| `--scene <name\|id>` | Short press recalls the scene (from the button's room). Repeat to cycle through scenes in order |
| `--with-off` | Second press turns lights off |
| `--repeat-timeout <dur\|secs>` | Idle time before a scene cycle restarts (default 3s) |
| `--on-hold dim-up\|dim-down\|dim-alternate\|none` | Hold behavior |
| `--repeat-interval <dur\|ms>` | Repeat interval while held |
| `--clear` | Remove the short-press action |
| `--dry-run` | Print the configuration that would be sent; send nothing |

```bash
openhue set control "Hue dimmer switch" --control 1 --scene Cozy --with-off
openhue set control "Hue dimmer switch" --control 4 --scene Soho --scene Nebula --repeat-timeout 10s --dry-run
```

## create scene / update scene

Neither command changes the real lights — they only write the scene.

**create**: `openhue create scene <name> -r <room> [flags]`. Layering: current state of every light in the room → global flags → per-light `--light` → (and global flags override `--from-file`).

**update**: `openhue update scene <id|name> [-r room] [flags]`. Applies on top of the values **stored in the scene**; unmentioned lights keep their settings (effects, gradients). `--light` for a room light not yet in the scene adds it. `--from-file` is a patch.

Shared flags: `--on`, `--off`, `-b`, `--rgb`, `-x`, `-y`, `-c`, `-t`, `--transition-time`, `--light "<light>:<settings>"` (repeatable), `--from-file <file|->`, `--auto-dynamic`, `--speed <0-100>`, `--palette-color`, `--palette-brightness`, `--palette-temperature`, `--palette-effect` (all repeatable), `-j/--json`, `--dry-run`.
Create only: `--only-listed` (lights not given with `--light` are turned off in the scene).
Update only: `--name <new>`, `--remove-light <light>` (repeatable), `--clear-palette`.

Per-light settings (comma-separated): `on`, `off`, `brightness=<0-100>`, `rgb=<#hex>`, `x=`, `y=`, `color=<name>`, `temperature=<153-500>`, `transition=<duration>`, `effect=<name>`, `gradient=<c1>;<c2>;...`, `gradient-mode=<mode>`.

```bash
openhue create scene Work -r Office --light "Desk:on,brightness=100,temperature=250" --light "Ceiling:off"
openhue create scene Aurora -r Lounge --auto-dynamic --speed 30 --palette-color blue --palette-color "#00FF88"
openhue get scene Cozy -r "Living Room" --yaml > cozy.yaml   # edit, then:
openhue update scene Cozy -r "Living Room" --from-file cozy.yaml
openhue update scene Cozy --light "Shelf:on,brightness=30" --remove-light Ceiling
```

## delete scene

`openhue delete scene <id|name> [-r room]` — exactly one scene, no confirmation prompt, cannot be undone. Export first with `get scene <name> --yaml` if a copy is wanted.

## mcp

`openhue mcp` runs a stdio MCP server. It builds its own openhue-go client from the config (does not use the preloaded `HomeModel`). Tools (defined in `cmd/mcp/tools.go`): `list_lights`, `list_rooms`, `list_scenes`, `set_light`, `set_room`, `activate_scene`.

```json
{ "mcpServers": { "openhue": { "command": "openhue", "args": ["mcp"] } } }
```

## Adding or changing a command

1. Create `cmd/<verb>/<verb>_<resource>.go` with a `NewCmd...(ctx *openhue.Context)` constructor; follow an existing sibling (e.g. `cmd/get/get_control.go`).
2. Register it on its parent (`cmd/<verb>/<verb>.go`); new top-level verbs go in `addCommands` in `cmd/root.go` with `GroupID` `hue` (loads the home) or `config`.
3. Resolve resources via the `Search*` helpers in `openhue/home_service.go` (name or ID).
4. Shared scene flags live in `cmd/scenecmd/flags.go`; shared light/room setters in `cmd/set/common.go`.
5. Test with `openhue.NewTestContext(home)` and `openhue/test/assert`; run `go test ./cmd/<verb>`.
6. Update this skill (command map + flag table) and the `Example:` block in the command's help text.

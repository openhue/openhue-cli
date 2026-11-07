---
name: openhue
description: Control Philips Hue lights from the terminal with the `openhue` CLI — install it, pair it with a Hue Bridge, list and switch lights and rooms, set brightness and colors, activate, create, edit and delete scenes, configure switch buttons, and run the MCP server for LLM clients. Use when the user wants to control Hue lights, rooms or scenes, or asks how to run an `openhue ...` command.
---

# Using the openhue CLI

`openhue` controls a Philips Hue Bridge on the local network. Full docs: https://www.openhue.io/cli

When unsure about a flag, run `openhue <command> --help`. It is always right.

## 1. Install

```bash
brew tap openhue/cli
brew install openhue-cli
openhue version
```

Install this skill for Claude Code (re-run after each upgrade to refresh it):

```bash
openhue install-skill              # ~/.claude/skills/openhue/SKILL.md
openhue install-skill --project    # ./.claude/skills/openhue/SKILL.md
```

Docker alternative (mount the config folder):

```bash
docker run -v "${HOME}/.openhue:/.openhue" --rm -it openhue/cli get lights
```

## 2. Pair with the bridge (once)

```bash
openhue setup
```

`setup` finds the bridge via mDNS, then asks you to press the bridge's link button. Other options:

| Command | Use when |
|---|---|
| `openhue discover` | Only print the bridges found on the network |
| `openhue setup -b 192.168.1.68` | mDNS discovery fails; give the bridge IP |
| `openhue config -b <ip> -k <key>` | You already have an application key |

Config is saved to `~/.openhue/config.yaml` (or `$XDG_CONFIG_HOME/openhue/`). Until it exists, Hue commands print "not configured yet, please run the 'setup' command".

## 3. Find what you have

```bash
openhue get lights
openhue get rooms
openhue get scenes
openhue get light -r "Living Room"   # filter by room
openhue get lights --json            # machine-readable
```

- Aliases: `get` = `g`, `set` = `s`, `create` = `c`. Singular and plural both work (`light`/`lights`).
- Resources are accepted by **name or ID**. Quote names with spaces.
- Names repeat across rooms: add `-r/--room <name|id>` to pick the right one.

## 4. Control lights and rooms

**`--on` is required to light up**, even when you only set brightness or color.

```bash
openhue set light Desk --on
openhue set room Studio --off
openhue set light "Hue Play TV" Desk --on --brightness 60 --rgb "#3399FF"
openhue set light Desk --on -c powder_blue
openhue set room Bedroom --on -t 400 --transition-time 30s
```

| Flag | Meaning |
|---|---|
| `--on` / `--off` | Power |
| `-b, --brightness` | 0–100 |
| `--rgb` | Hex color, `#RRGGBB` |
| `-c, --color` | Named color (`white`, `lime`, `powder_blue`, ...) |
| `-x` / `-y` | CIE xy color |
| `-t, --temperature` | White temperature in Mirek, 153 (cool) – 500 (warm) |
| `--transition-time` | Fade duration, e.g. `5s` |

Up to 10 lights or rooms per command.

## 5. Scenes

Activate:

```bash
openhue set scene Relax -r "Living Room"
openhue set scene Aurora -a dynamic      # play the dynamic palette
```

Create (writes the scene only; lights do not change). It starts from the room's current light state:

```bash
openhue create scene Work -r Office \
  --light "Desk:on,brightness=100,temperature=250" \
  --light "Ceiling:off"
```

Edit or copy through YAML:

```bash
openhue get scene Cozy -r "Living Room" --yaml > cozy.yaml
openhue update scene Cozy -r "Living Room" --from-file cozy.yaml
openhue update scene Cozy --light "Shelf:on,brightness=30" --remove-light Ceiling
openhue update scene Cozy --name "Cozy Evening"
```

Delete (no prompt, cannot be undone — export with `--yaml` first if you want a copy):

```bash
openhue delete scene Cozy -r "Living Room"
```

Per-light settings for `--light "<light>:<settings>"`, comma-separated: `on`, `off`, `brightness=`, `rgb=`, `color=`, `temperature=`, `x=`, `y=`, `transition=`, `effect=`, `gradient=<c1>;<c2>`, `gradient-mode=`.

Add `--dry-run` to `create`/`update scene` to preview without saving.

## 6. Switches and dimmers

Accessories must already be added to the bridge (e.g. with the Hue app).

```bash
openhue get control                       # show button assignments
openhue set control "Hue dimmer switch" --control 1 --scene Cozy --with-off
openhue set control "Hue dimmer switch" --control 4 --scene Soho --scene Nebula --dry-run
```

Repeat `--scene` to cycle scenes on each press. Other flags: `--on-hold dim-up|dim-down|dim-alternate|none`, `--repeat-timeout`, `--clear`. Rotary dials are read-only.

## 7. Watch live events

```bash
openhue get events    # streams bridge events as JSON; Ctrl+C to stop
```

## 8. Use from an LLM (MCP server)

```json
{ "mcpServers": { "openhue": { "command": "openhue", "args": ["mcp"] } } }
```

Tools exposed: `list_lights`, `list_rooms`, `list_scenes`, `set_light`, `set_room`, `activate_scene`.

## Troubleshooting

| Symptom | Fix |
|---|---|
| "not configured yet" | Run `openhue setup` |
| Bridge not discovered | `openhue setup -b <bridge-ip>` |
| Light ignores brightness/color | Add `--on` |
| Wrong light or scene picked | Add `-r <room>` or use the ID from `get ... --json` |
| Need detail | Check `openhue.log` next to the config file |

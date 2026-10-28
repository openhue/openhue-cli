# Proposal

## Why

The Hue app lets users make a switch or dial press recall a scene, or cycle through several. The CLI cannot show or change these assignments, so setups can't be scripted, reviewed or restored. On the bridge this is not a property of the button but the configuration of a per-accessory `behavior_instance` (script "Hue Accessories"), which is why it is a change of its own, building on the control commands from `add-control-commands`.

## What Changes

- **`openhue get control`** gains a "Short press" and a "Hold" column (and the full configuration in `--json`), summarizing what each button does: `scene <name>`, `cycle: <scene>, <scene>, … (n)`, `time-based (read-only)`, `dim-up`, `-`.
- **`openhue set control <device|id> [--control <n>] --scene <scene>`**: makes a short press recall one scene (optional `--with-off` so a second press turns the lights off).
- **Repeatable `--scene`**: with two or more scenes, each press goes to the next one (`--repeat-timeout <duration>` resets the cycle).
- **`--on-hold dim-up|dim-down|dim-alternate|none`** sets the hold behavior; **`--clear`** removes the button's short-press action.
- **`--dry-run`** prints the resulting behavior configuration without sending it.
- Scenes are resolved by name or ID within the room the button is bound to. Untouched buttons and unsupported fields (e.g. time-based slots, long-press) are preserved on write.
- Requires `openhue-go` behavior wrappers (`GetBehaviorInstances`, `GetBehaviorInstanceById`, `UpdateBehaviorInstance`) in addition to the control ones from the first change.

## Capabilities

### New Capabilities
- `control-actions`: showing what each button of a switch is configured to do.
- `control-scene-assignment`: assigning a scene, a scene cycle or a hold behavior to a button.

### Modified Capabilities
<!-- `control-listing` and `control-configuration` come from add-control-commands and are not in openspec/specs yet; this change extends them and MUST be applied after it. -->

## Impact

- `cmd/get/get_control.go` (columns), `cmd/set/set_control.go` (flags); new behavior model in `openhue/` (parse/merge/build of the button configuration), tests.
- `openhue-go`: behavior instance wrappers and mock methods (separate repository), CLI `go.mod` bump.
- README. Out of scope: creating a behavior instance for an unconfigured accessory, time-based and smart-scene authoring, long-press actions, motion sensors/schedules/other behavior scripts.

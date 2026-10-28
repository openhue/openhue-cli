# Proposal

## Why

Hue switches, dimmers and dials expose their physical controls (buttons and rotary dials) through the bridge, but the CLI cannot show or configure them. Users can't check which controls a device has, what they last reported, or tune their behavior without the Hue app or raw API calls. This also gives scripts a way to inspect the controls of a room.

## What Changes

- **`openhue get control [controlId]`** (alias `controls`): lists the buttons and relative rotaries of the home with ID, kind (`button` | `rotary`), owning device, room, control number on the device (`control_id`) and the last reported event with its time. Filters: `--room/-r`, `--device/-d` (name or ID), `--kind button|rotary`. `--json` prints the raw resources.
- **`openhue set control <controlId>`**: updates the settings a control exposes. Today only buttons are configurable: `--repeat-interval <duration|ms>` (how often `repeat` events fire while held). Rotaries have no writable setting, so setting one is a clear error rather than a silent no-op.
- Controls are resolved by ID, or by `<device name>` plus `--control <n>` (the `control_id`), since controls have no name of their own. Ambiguity is an error.
- **Dependency**: `openhue-go` v1.0.0 only has `GetButtons`/`GetButtonById`; relative rotary reads and `UpdateButton` must be added there (the generated client already has them) and the CLI bumped to that release.

## Capabilities

### New Capabilities
- `control-listing`: discovering and inspecting buttons and rotary dials, per home, room or device.
- `control-configuration`: changing the configurable settings of a control.

### Modified Capabilities

(none; `openspec/specs/` has no specs yet)

## Impact

- New `cmd/get/get_control.go` and `cmd/set/set_control.go` (+ tests), registered in `get.go` / `set.go`.
- `openhue/`: home loading gains buttons and relative rotaries; new `Control` model attached to `Device`, with search helpers.
- `openhue-go`: new `Home.GetRelativeRotaries`, `GetRelativeRotaryById`, `UpdateButton` (and mock support) in a separate repository; CLI `go.mod` bump.
- README command docs. MCP server unchanged. Out of scope: creating/deleting controls (not possible on the bridge), rebinding what a button does (see the follow-up change `add-control-scene-assignment`), watching live events (already `get events`).

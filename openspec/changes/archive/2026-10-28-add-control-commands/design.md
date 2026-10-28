# Design

## Context

Hue exposes controls as `button` and `relative_rotary` resources; each has an `owner` (the device) and `metadata.control_id`. The home model (`openhue/home_context.go`) loads rooms, devices, lights, grouped lights and scenes, but not controls. `get`/`set` subcommands follow the pattern in `cmd/get/get_scene.go` and `cmd/set/set_scene.go`, with `openhue.NewTestContext` for tests. `openhue-go` v1.0.0 wraps only `GetButtons`/`GetButtonById`; the generated client already contains `GetRelativeRotaries`, `UpdateButton` and `UpdateRelativeRotary`. `ButtonPut` has `repeat_interval` (and `control_id`); `RelativeRotaryPut` has no writable field. See proposal.md for scope.

## Goals / Non-Goals

**Goals:**
- Read-only listing with room/device context, and a safe, scriptable way to set the button repeat interval.
- Consistent with existing `get`/`set` conventions (name-or-ID resolution, `--json`, `-r`).

**Non-Goals:**
- Rebinding button actions (behavior instances/scenes), live event streaming, pairing/removing devices.
- Writing `control_id` (read-only in practice; changing it would corrupt device layout).

## Decisions

- **Controls in the home model.** Load buttons and relative rotaries in `loadHueHomeCtx`, and attach `Controls []Control` to `Device` (matched by `owner.rid`). `Control` holds kind, control number, and the raw `ButtonGet`/`RelativeRotaryGet`; its parent is the device, so room is `Device.Parent`. *Alternative:* fetch on demand in the commands — rejected, it bypasses the shared model and the test context.
- **Failure tolerance.** A bridge without switches returns empty lists; if the controls endpoints fail, the home still loads and a warning is logged (controls are not essential to light/scene commands).
- **Command names:** `get control` (alias `controls`) and `set control`, matching singular-noun commands. `--kind` filters button vs rotary. *Alternative:* separate `get button`/`get rotary` — rejected, users think of "the controls on this switch" and `control_id` spans both.
- **Resolution:** controls have no names, so the argument is a control ID, or a device name/ID with `--control <n>`. Zero or multiple matches are errors listing candidates (never fan out a write).
- **Last event display:** prefer `button_report` / `rotary_report` over the deprecated `last_event`; fall back to it only if the report is absent. Time printed in the local zone like other tables; raw values in `--json`.
- **`--repeat-interval`** accepts a Go duration or plain milliseconds (plain integers = ms), must be > 0. Only flag today; the options struct is built so future writable fields slot in. Rotary targets error before any call.
- **`openhue-go` first.** Add `GetRelativeRotaries`, `GetRelativeRotaryById`, `UpdateButton` (and `UpdateRelativeRotary` for completeness) and mock methods to `openhue-go`, release it, then bump the CLI. During development a `replace`/`go.work` pointing at the local checkout may be used but MUST NOT be committed.

## Risks / Trade-offs

- [Cross-repo dependency blocks the CLI] → Land and tag `openhue-go` first; tasks are ordered accordingly.
- [Extra API calls at startup for every `hue` command] → Two small list calls; tolerated failure; revisit lazy loading if it shows in startup time.
- [Repeat-interval bounds unknown] → Only validate > 0; surface bridge rejection verbatim.
- [Rotary has no settings, so `set control` feels half-complete] → Explicit error now; the command shape is ready for future fields.

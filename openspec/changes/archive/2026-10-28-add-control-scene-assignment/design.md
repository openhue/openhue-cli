# Design

## Context

Verified on a real bridge: each configured accessory has one `behavior_instance` with script "Hue Accessories", whose `configuration` is `{device: {rid}, model_id, buttons: {<button id>: {where: [{group: {rid, rtype: room}}], on_short_release, on_long_press, on_repeat}}}`. `on_short_release` holds `recall_single_extended {actions:[{action:{recall:{rid, rtype: scene|smart_scene}}}], with_off:{enabled}}`, `scene_cycle_extended {slots:[[{action:{recall…}}]…], repeat_timeout:{seconds}, with_off}`, or `time_based_extended`. `on_repeat.action` is `dim_up|dim_down|dim_alternate`; `on_long_press.action` is e.g. `do_nothing`. `BehaviorInstanceGet/Put` expose `configuration` as an untyped map and `openhue-go` has no behavior wrappers. The control model and `set control` come from `add-control-commands`, which this change depends on.

## Goals / Non-Goals

**Goals:**
- Read and write short-press scene assignments and hold dimming safely.
- Never lose configuration the CLI does not understand.

**Non-Goals:**
- Creating behavior instances, time-based or smart-scene authoring, long-press actions, other scripts (motion, schedules).

## Decisions

- **Treat the configuration as an untyped tree and patch only known paths.** Decode `configuration` into `map[string]any`, modify `buttons[<id>].on_short_release` / `on_repeat` in place, and PUT the whole document. Typed structs would drop fields the bridge adds later. *Alternative:* typed model of the full schema — rejected, fragile.
- **Instance lookup:** list behavior instances once (in `Home` loading, tolerating failure like controls) and match `script == Hue Accessories` and `configuration.device.rid == button owner`. None found → explicit error.
- **Scene resolution** uses the room from the button's `where` (group room), falling back to the device's room; scenes must belong to it. Zones in `where` are accepted as-is when `--scene` is not used (hold/clear only).
- **Mapping:** one `--scene` → `recall_single_extended` (`with_off` from `--with-off`, default false); several → `scene_cycle_extended` with one slot per scene and `repeat_timeout` (default: keep existing, else 3 s); `--clear` removes `on_short_release`. Switching between single and cycle replaces the key; an existing `time_based_extended` is replaced only when `--scene`/`--clear` is explicit.
- **Summaries for `get control`:** a pure function from the stored short-press/repeat tree to text; unrecognized shapes display as `custom`, never as an error.
- **Safety:** `--dry-run` prints the full PUT body; all validation (target, scene, flag conflicts) happens before any call; read-modify-write uses the state loaded at startup (milliseconds old). Concurrent edits by the Hue app in that window can be overwritten — accepted and documented.
- **`openhue-go` first:** add `GetBehaviorInstances`, `GetBehaviorInstanceById`, `UpdateBehaviorInstance` and mocks, release, bump. Playground: `openhue-go/playground/controls -behaviors` dumps the real configurations.

## Risks / Trade-offs

- [A bad PUT can break a physical switch] → dry-run, validate first, preserve unknown keys, round-trip tests that an unchanged button re-serializes identically.
- [Bridge firmware changes the schema] → untyped patching; summaries degrade to `custom`.
- [Lost update if the app edits simultaneously] → documented; single short-lived command.
- [Only one script supported] → clear error for other accessory types.

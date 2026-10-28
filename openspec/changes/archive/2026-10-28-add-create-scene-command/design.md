# Design

## Context

Commands in the `hue` group get a preloaded `HomeModel` (`ctx.Home`) via `LoadHomeIfNeeded`; rooms expose `Devices[].Light` (`HueData *openhue.LightGet`) and `Scenes` (`HueData *openhue.SceneGet`, including stored `Actions`, `Palette`, `Speed`, `AutoDynamic`). `Scene.Activate` in `openhue/home_model.go` shows how to call the `openhue-go` client. `openhue-go` v1.0.0 provides `CreateScene(ScenePost)`, `UpdateScene(id, ScenePut)` and `DeleteScene(id)`; the API replaces the whole `actions` list on PUT. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Full scene lifecycle (create, update, delete) and a round-trippable text format (export → edit → apply).
- One internal scene model, so flags, files and export cannot drift apart.
- No side effects on real lights, and no bridge call when validation fails.

**Non-Goals:**
- Zones as scene targets, scene images/appdata, scheduling or automations.
- Editing a scene's `type`/internal Hue fields beyond those listed in the spec.
- MCP tools for scene authoring.

## Decisions

- **Single internal model (`openhue.SceneDefinition`).** Fields: name, room, `speed`, `auto_dynamic`, `palette` (colors, temperatures, brightness, effects), `defaults` (a light-settings block), and `lights[]` (each: `light` ref + a light-settings block). The light-settings block holds `on`, `brightness`, `rgb`/`x`+`y`/`color`, `temperature`, `transition`, `effect`, `gradient{mode, points[]}`. Flags and YAML both produce a `SceneDefinition` (or a patch of it); a pure builder turns it into `ScenePost`/`ScenePut`. *Why:* validation, precedence and tests live in one place.
- **Layered resolution:** baseline → `defaults` → per-light settings. Baseline is each light's current state for `create` (and for lights newly added by `update`), and the scene's stored actions for `update`. Merge flags over file: **file < flags**, and the positional name/`--room` override the file's `name`/`room`.
- **`--light "<light>:<settings>"`**, repeatable, split at the *last* `:`; settings are comma-separated `key=value` or bare `on|off`. Gradient points use `;` (`gradient=#FF0000;#00FF00`) because `,` separates settings. Keys reuse the `set light` validation (`CmdSetLightFlags.toSetLightOptions()`) and ranges. *Alternative:* order-sensitive interleaved flags — rejected, Cobra/pflag cannot associate them reliably.
- **Transitions** map to `Dynamics.Duration` (ms) on each action; `--transition-time` is the default, `transition=` the per-light override, same duration parsing as `set light`.
- **Color vs. color temperature vs. gradient:** a light action carries at most one of `color`, `color_temperature`, `gradient` (setting one clears the others from the baseline). **Capabilities** are checked against the light's `LightGet` (color, color_temperature, gradient, effects support): an unsupported *explicit* per-light setting is an error; the same setting from defaults is skipped with a warning.
- **Dynamic scenes:** `speed` is exposed as 0–100 and mapped to the API's 0–1; `auto_dynamic` sets `AutoDynamic`; palette maps to `ScenePalette`. A palette is validated for API minimums (at least one entry in some palette list when `--auto-dynamic` or a palette flag is used). `--clear-palette` removes it on update.
- **YAML via `gopkg.in/yaml.v3`** with `KnownFields(true)` so typos are errors with line numbers. `--from-file -` reads stdin. On `update`, the file is a *patch*: omitted lights/fields are untouched, `remove_lights: [...]` turns the listed lights off in the scene.
- **Export** (`get scene <scene> --yaml`) converts `SceneGet` (+ light names resolved through the home model) to a `SceneDefinition`, emitting explicit per-light entries and scene-level dynamics; a re-apply of an exported file to its own scene is a no-op. Light names are used (IDs when a name is ambiguous).
- **Every room light stays in the scene.** The bridge rejects a scene whose actions do not cover all the lights of its room (`Light action targets not matching lights in referenced group`, verified on a real bridge). So `--only-listed` and `--remove-light` do not drop actions: the light's action becomes `on: false` with no other attribute, and defaults are not applied to it.
- **Effects and palettes follow bridge rules** (verified on a real bridge): an action cannot carry `effects.effect` together with a color, temperature or gradient, so an effect replaces them and setting a color drops a stored effect; each palette color/temperature entry requires a `dimming` (100% by default). The bridge answers `201 Created` to a scene creation, which `openhue-go` v1.0.0 reports as an error: the CLI treats it as success and looks the scene up by name.
- **Full replacement on PUT:** `update` sends the complete merged action list. Stored `ActionGet` → `ActionPost` conversion preserves every field (effects, gradients, dynamics), so untouched lights lose nothing.
- **Resolution rules:** rooms and scenes by name or ID via existing search; zero matches, or several without a disambiguating room, are errors (ask for the ID). Light refs resolve within the room's devices. Duplicate scene name in the room is rejected on create and rename (the bridge allows it, but this CLI resolves scenes by name).
- **`--dry-run`** prints the JSON request body and exits 0 without calling the bridge; **`--json`** prints the resulting scene. **No-op protection:** `update` with no change is a usage error.
- **Delete:** `delete scene` resolves like `update`, deletes exactly one scene, and requires no prompt (scriptable); ambiguity is an error, never a multi-delete.
- **Placement:** model, parsing and builder in `openhue/` (unit-testable without Cobra); `cmd/create`, `cmd/update`, `cmd/delete` handle flags/output; shared flag registration in one helper. All new commands in group `hue`.

## Risks / Trade-offs

- [Large flag surface] → Everything is also expressible in the file; flags map 1:1 to file keys and are documented from one table.
- [Bridge rejects effects/gradients on unsupported lights] → Capability checks before the request; bridge errors surfaced verbatim.
- [Export/apply drift as the bridge adds fields] → Update preserves unknown stored action fields; export covers only documented keys and the README states this.
- [Stale state] → The snapshot comes from the state loaded at startup (milliseconds old); acceptable.
- [Delete is irreversible] → Single-target resolution only; `--dry-run` is not offered for delete, so the README recommends `get scene` first.

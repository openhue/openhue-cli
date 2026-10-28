# Tasks

## 1. Scene model and builder

- [x] 1.1 Add `openhue.SceneDefinition` (scene-level dynamics/palette, defaults, per-light settings incl. transition, effect, gradient) and the `--light` settings parser (split at last `:`, comma-separated, `;` for gradient points); verify with parser unit tests for valid input, names containing `:`, unknown keys, bad ranges, malformed values (`go test ./openhue`)
- [x] 1.2 Implement the pure builder: baseline (current state or stored actions via `ActionGet` → `ActionPost` preserving all fields) → defaults → per-light → `ScenePost`/`ScenePut`, with one-of color/temperature/gradient and capability checks (error for explicit, warning for defaults); verify with table-driven tests for precedence, partial settings, white-only lights, transitions (`Dynamics.Duration`), effects, gradients, speed 0–100 → 0–1 mapping and palette validation
- [x] 1.3 Add `Room`/`Scene` methods wrapping `Home.CreateScene`, `UpdateScene` (full action list, rename, remove lights, palette clear) and `DeleteScene`, with name/duplicate/ambiguity/empty-room rules; verify with unit tests on the `openhue-go` mock, including no bridge call on validation failure

## 2. YAML definition files

- [x] 2.1 Add YAML loading with `gopkg.in/yaml.v3` (`KnownFields(true)`, stdin via `-`, errors with line numbers) mapping to `SceneDefinition`, promote it to a direct dependency (`go mod tidy`); verify with tests for valid files, typos, wrong types and unreadable files
- [x] 2.2 Implement merge rules (file < flags; positional name and `--room` override; patch semantics and `remove_lights` for update); verify with unit tests for each precedence case and the room-mismatch error
- [x] 2.3 Implement scene → definition export, using light names (IDs when ambiguous); verify with round-trip tests: export then re-apply is a no-op, export then create from file yields an equivalent scene

## 3. Create command

- [x] 3.1 Create `cmd/create/create.go` (group `hue`) and `create_scene.go`: name arg, `--room/-r`, `--from-file`, `--light`, `--only-listed`, default flags, dynamics/palette flags, `--dry-run`, `--json`; shared flag helper in a reusable file; register in `cmd/root.go`; verify with Cobra tests via `openhue.NewTestContext` for snapshot, per-light, defaults+override, only-listed, dynamics/palette, file, dry-run, JSON, missing room/name, unknown room and duplicate name (`go test ./cmd/create`)

## 4. Update command

- [x] 4.1 Create `cmd/update/update.go` and `update_scene.go`: scene arg, `--room`, shared flags, `--from-file` patch, `--name`, `--remove-light`, `--clear-palette`, no-op usage error, `--dry-run`, `--json`; register in `cmd/root.go`; verify with Cobra tests for by-name/ID, ambiguous/missing scene, each change flag, preservation of untouched lights and add-light behavior (`go test ./cmd/update`)

## 5. Export and delete

- [x] 5.1 Add `--yaml` to `cmd/get/get_scene.go` (exclusive with `--json`); verify with a test that output parses back to the definition and errors on zero/multiple matches
- [x] 5.2 Create `cmd/delete/delete.go` and `delete_scene.go` (group `hue`), register in `cmd/root.go`; verify with tests for delete by name/ID, not found, ambiguous and bridge error (`go test ./cmd/delete`)

## 6. Docs and integration

- [x] 6.1 Document create, update, delete, `--from-file` and `get scene --yaml` in README and command help, with a sample definition file under `docs/` or `examples/`; verify each example runs with `--dry-run` or matches `--help`
- [x] 6.2 Run `make test` and `go vet ./...`; verify both pass, and optionally try against a real bridge: create → `get scene --yaml` → edit → `update --from-file` → `set scene` → `delete scene`

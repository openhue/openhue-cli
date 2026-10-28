# Spec Delta

## Purpose

Lets users modify an existing Hue scene from the command line, light by light, without recreating it or changing the real lights.

## ADDED Requirements

### Requirement: Update an existing scene
The CLI SHALL provide `openhue update scene <scene>` where `<scene>` is a scene name or ID, with optional `--room` to disambiguate scenes sharing a name, and SHALL print the scene's name and ID on success.

#### Scenario: Update by name
- **WHEN** the user runs `openhue update scene Cozy --light "Desk:brightness=20"` and exactly one scene "Cozy" exists
- **THEN** that scene is updated and the exit status is zero

#### Scenario: Scene not found
- **WHEN** no scene matches (within `--room` if given)
- **THEN** an error naming the scene is printed and the exit status is non-zero

#### Scenario: Ambiguous scene
- **WHEN** several scenes match and `--room` does not narrow it to one
- **THEN** an error asking for `--room` or the ID is printed and nothing is changed

### Requirement: Changes apply on top of the stored scene
An update SHALL start from the scene's stored per-light values, not from the lights' current state. Lights not mentioned SHALL keep their stored values, including effects, gradients and transitions the command did not touch, and the real lights SHALL NOT be modified.

#### Scenario: One light changed
- **WHEN** a three-light scene is updated with `--light "Desk:brightness=20"`
- **THEN** only Desk's brightness changes; the others keep their stored values, including any effect or gradient

### Requirement: Per-light and default settings on update
`update scene` SHALL accept the same `--light` settings, default flags and validation as `create scene`. Default flags apply to every light in the scene; per-light settings take precedence. A `--light` for a room light not yet in the scene SHALL add it, starting from that light's current state.

#### Scenario: Add a light
- **WHEN** Shelf is in the room but not the scene and the user passes `--light "Shelf:on,brightness=30"`
- **THEN** the scene now includes Shelf with those settings

#### Scenario: Defaults on all scene lights
- **WHEN** the user runs `openhue update scene Cozy --brightness 10`
- **THEN** every light in the scene is stored at 10%

#### Scenario: Light outside the room
- **WHEN** `--light` references a light not in the scene's room
- **THEN** an error is printed and nothing is changed

### Requirement: Update dynamics and palette
`update scene` SHALL accept `--speed`, `--auto-dynamic`/`--auto-dynamic=false`, the `--palette-*` flags (which replace the stored palette lists they name) and `--clear-palette`.

#### Scenario: Change speed only
- **WHEN** the user runs `openhue update scene Aurora --speed 60`
- **THEN** only the speed changes

#### Scenario: Clear palette
- **WHEN** the user passes `--clear-palette` on a scene with auto-dynamic off
- **THEN** the palette is removed; with auto-dynamic on, an error is printed instead

### Requirement: Remove lights from a scene
The command SHALL accept a repeatable `--remove-light <light>`. Because the bridge requires every light of the room in a scene, a removed light is turned off in the scene (its other stored attributes are dropped) rather than deleted from it.

#### Scenario: Remove a light
- **WHEN** the user runs `openhue update scene Cozy --remove-light Ceiling`
- **THEN** Ceiling is turned off in the scene, the other lights are unchanged, and default flags do not apply to Ceiling

#### Scenario: Invalid removal
- **WHEN** the light is not in the scene, or is also given with `--light`
- **THEN** an error is printed and nothing is changed

### Requirement: Rename a scene
The command SHALL accept `--name <new-name>`.

#### Scenario: Rename
- **WHEN** the user runs `openhue update scene Cozy --name Relax`
- **THEN** the scene is named "Relax" and its lights are unchanged

#### Scenario: Name collision
- **WHEN** another scene with that name exists in the same room
- **THEN** an error is printed and nothing is changed

### Requirement: Update requires a change
The command SHALL fail with a usage error when no change is requested (by flag or file), without modifying anything.

#### Scenario: No flags
- **WHEN** the user runs `openhue update scene Cozy`
- **THEN** an error saying at least one change is required is printed and the exit status is non-zero

### Requirement: Dry run, JSON and bridge errors
`--dry-run` SHALL print the request body without sending it; `--json` SHALL print the updated scene; bridge failures SHALL be printed with a non-zero exit status.

#### Scenario: Dry run
- **WHEN** the user adds `--dry-run`
- **THEN** the full merged request is printed and the scene is unchanged

#### Scenario: Bridge rejects update
- **WHEN** the bridge returns an error
- **THEN** the error text is printed and the exit status is non-zero

# scene-definition-file Specification

## Purpose

Lets users define or patch scenes in a YAML file so large scenes can be versioned, reviewed and applied repeatably.

## Requirements

### Requirement: Create from a definition file
`openhue create scene --from-file <path>` SHALL create a scene from a YAML definition, where `-` as path reads standard input. The file MAY set `name`, `room`, `speed`, `auto_dynamic`, `palette` (`colors`, `temperatures`, `brightness`, `effects`), `defaults` and `lights` (each with `light` plus any per-light setting: `on`, `brightness`, `rgb`, `x`, `y`, `color`, `temperature`, `transition`, `effect`, `gradient` with `mode` and `points`).

#### Scenario: Create from file
- **WHEN** the file defines name "Sunset", room "Living Room" and three lights, and the user runs `openhue create scene --from-file sunset.yaml`
- **THEN** that scene is created with those settings

#### Scenario: Read from stdin
- **WHEN** the user pipes a definition and passes `--from-file -`
- **THEN** the scene is created from the piped content

### Requirement: Flags override the file
Command-line values SHALL override the file: a positional name overrides `name`, `--room` overrides `room`, default and scene-level flags override the file's `defaults` and scene-level fields, and `--light` entries override the same light's file entry.

#### Scenario: Name and room override
- **WHEN** the file says room "Office" and the user passes `-r Studio` and a name argument
- **THEN** the scene is created in Studio with the argument name

### Requirement: Strict validation with locations
The file SHALL be validated with the same rules as the flags; unknown keys, wrong types and invalid values SHALL produce an error citing the file position, and nothing SHALL be created or changed.

#### Scenario: Typo in key
- **WHEN** the file contains `brigthness: 50`
- **THEN** an error naming the unknown key and its line is printed and the exit status is non-zero

#### Scenario: Unreadable file
- **WHEN** the path does not exist or is not valid YAML
- **THEN** an error is printed and nothing is created

### Requirement: Update from a definition file
`openhue update scene <scene> --from-file <path>` SHALL treat the file as a partial patch: fields and lights not present are left untouched, and `remove_lights` MAY list lights to turn off in the scene. `name` renames the scene; `room`, if present, must match the scene's room.

#### Scenario: Patch one light
- **WHEN** the file lists only the light Desk with `brightness: 20`
- **THEN** only Desk changes

#### Scenario: Room mismatch
- **WHEN** the file's `room` differs from the scene's room
- **THEN** an error is printed and nothing is changed

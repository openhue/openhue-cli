# Spec Delta

## Purpose

Lets users create a Hue scene in a room from the command line: save the current look, or define each light, its transitions, effects, gradients and dynamic palette, without changing the real lights.

## ADDED Requirements

### Requirement: Create a scene in a room
The CLI SHALL provide `openhue create scene <name> --room <room>` which creates a scene with the given name in the specified room (name or ID) and prints the scene's name and ID.

#### Scenario: Created by room name
- **WHEN** the user runs `openhue create scene Cozy --room "Living Room"` and the room exists
- **THEN** a scene "Cozy" exists in that room, its name and ID are printed, and the exit status is zero

#### Scenario: Created by room ID
- **WHEN** the room ID is given to `--room`
- **THEN** the scene is created in that room

### Requirement: Current state is the default content
Unless overridden, the scene SHALL contain one action per light in the room reproducing its current on/off state, brightness and color (or color temperature when the light is in the color temperature spectrum).

#### Scenario: Lights with different states
- **WHEN** one light is on at 40% warm white and another is off
- **THEN** recalling the scene turns the first on at 40% in that white and the second off

#### Scenario: Room without lights
- **WHEN** the room contains no lights
- **THEN** an error is printed, nothing is created, and the exit status is non-zero

### Requirement: Required arguments and validation
The command SHALL require exactly one name and a room (via `--room` or `--from-file`), and SHALL fail without creating anything when the room is unknown or ambiguous, or a scene with that name already exists in the room.

#### Scenario: Missing room
- **WHEN** no room is provided by flag or file
- **THEN** the CLI reports that a room is required

#### Scenario: Unknown room
- **WHEN** the room matches nothing
- **THEN** an error naming the room is printed and nothing is created

#### Scenario: Duplicate name
- **WHEN** the room already has a scene with that name
- **THEN** an error is printed and nothing is created

### Requirement: Per-light settings
The command SHALL accept a repeatable `--light "<light>:<settings>"`, where `<light>` is a light name or ID in the room and settings are comma-separated: `on`, `off`, `brightness=<0-100>`, `rgb=<#hex>`, `x=<n>`, `y=<n>`, `color=<name>`, `temperature=<153-500>`, `transition=<duration>`, `effect=<name>`, `gradient=<color>;<color>;…`, `gradient-mode=<mode>`. These SHALL be stored for that light only and SHALL NOT change the real light.

#### Scenario: Two lights configured differently
- **WHEN** the user runs `openhue create scene Work -r Office --light "Desk:on,brightness=100,temperature=250" --light "Ceiling:off"`
- **THEN** recalling the scene turns Desk on at full brightness in that white and Ceiling off, and neither light changed at creation

#### Scenario: Partial settings keep the snapshot
- **WHEN** a light is given only `brightness=30`
- **THEN** its other attributes come from its current state

#### Scenario: Transition per light
- **WHEN** a light is given `transition=3s`
- **THEN** that light's action uses a 3-second transition when the scene is recalled

#### Scenario: Effect and gradient
- **WHEN** a light supporting them is given `effect=candle` or `gradient=#FF0000;#0000FF`
- **THEN** the stored action carries that effect or gradient, replacing any color/temperature from the snapshot

#### Scenario: Invalid entries
- **WHEN** a setting is unknown, malformed, out of range, the light is unknown or ambiguous, or the same light appears twice
- **THEN** an error naming the problem is printed, nothing is created, and the exit status is non-zero

#### Scenario: Unsupported capability
- **WHEN** a per-light setting asks for a color, gradient or effect the light does not support
- **THEN** an error naming the light and setting is printed and nothing is created

### Requirement: Default settings for all lights
The command SHALL accept `--on`, `--off`, `--brightness`, `--rgb`, `--cie-x/--cie-y`, `--color`, `--temperature` and `--transition-time` as defaults for every light of the scene; per-light settings take precedence. Precedence: current state, then defaults, then per-light settings.

#### Scenario: Defaults with one override
- **WHEN** the user runs `openhue create scene Dim -r Bedroom --brightness 10 --light "Lamp:brightness=40"`
- **THEN** Lamp is stored at 40% and every other light at 10%

#### Scenario: Default color on a white-only light
- **WHEN** a default color flag is used and a light does not support color
- **THEN** that light skips the color, a warning is printed, and the scene is created

### Requirement: Restrict to listed lights
With `--only-listed`, the lights of the room that are not named with `--light` SHALL be turned off in the scene and left otherwise unconfigured (the bridge requires every light of the room in a scene); otherwise every light keeps its snapshot state.

#### Scenario: Only listed
- **WHEN** `--only-listed --light "Desk:on"` is used in a room of three lights
- **THEN** Desk is on in the scene and the two other lights are turned off by it

#### Scenario: Only-listed without lights
- **WHEN** `--only-listed` is given with no light (flag or file)
- **THEN** an error is printed and nothing is created

### Requirement: Dynamic scene settings
The command SHALL accept `--speed <0-100>`, `--auto-dynamic`, and repeatable `--palette-color <name|#hex>`, `--palette-temperature <mirek>`, `--palette-brightness <0-100>`, `--palette-effect <name>` to configure dynamic playback of the scene.

#### Scenario: Dynamic scene
- **WHEN** the user runs `openhue create scene Aurora -r Lounge --auto-dynamic --speed 30 --palette-color blue --palette-color "#00FF88"`
- **THEN** the scene stores that speed, auto-dynamic flag and palette, and `openhue get scene` shows them

#### Scenario: Invalid dynamic settings
- **WHEN** speed is outside 0–100, a palette value is invalid, or `--auto-dynamic` is used without any palette entry
- **THEN** an error is printed and nothing is created

### Requirement: Dry run and JSON output
`--dry-run` SHALL print the request that would be sent and exit successfully without creating anything; `--json` SHALL print the created scene as JSON.

#### Scenario: Dry run
- **WHEN** the user adds `--dry-run`
- **THEN** the request body is printed, the bridge is not asked to create anything, and validation still applies

#### Scenario: JSON output
- **WHEN** the user adds `--json`
- **THEN** the created scene is printed as JSON instead of the one-line message

### Requirement: Bridge errors are reported
A bridge failure while creating SHALL be printed as an error with a non-zero exit status.

#### Scenario: Bridge rejects creation
- **WHEN** the bridge returns an error
- **THEN** the error text is printed and the exit status is non-zero

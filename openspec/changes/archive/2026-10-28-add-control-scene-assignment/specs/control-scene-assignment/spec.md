# Spec Delta

## Purpose

Lets users assign scenes and hold behaviors to the buttons of a Hue accessory from the command line, as the Hue app does.

## ADDED Requirements

### Requirement: Assign one scene
`openhue set control <target> --scene <scene>` SHALL make a short press of the targeted button recall that scene. `--with-off` SHALL make a second press turn the lights off. The scene is resolved by name or ID among the scenes of the room the button is bound to.

#### Scenario: Assign
- **WHEN** the user runs `openhue set control "Hue dimmer living room" --control 1 --scene Cozy`
- **THEN** the button's short press recalls "Cozy" and a confirmation is printed

#### Scenario: Scene of another room
- **WHEN** the scene exists only in a room the button is not bound to
- **THEN** an error explains it and nothing is sent

#### Scenario: Unknown or ambiguous scene
- **WHEN** the scene matches zero or several scenes in the room
- **THEN** an error asks for an ID and nothing is sent

### Requirement: Cycle through scenes
Repeating `--scene` SHALL make each short press recall the next scene in the given order, wrapping around. `--repeat-timeout <duration>` SHALL set the idle time after which the cycle restarts.

#### Scenario: Cycle
- **WHEN** the user passes `--scene Soho --scene Nebula --scene Dusk`
- **THEN** the button cycles Soho → Nebula → Dusk → Soho

### Requirement: Hold behavior and clearing
`--on-hold dim-up|dim-down|dim-alternate|none` SHALL set what holding the button does, and `--clear` SHALL remove the short-press action. They MAY be combined with `--scene` only when not contradictory (`--clear` with `--scene` is an error).

#### Scenario: Dim on hold
- **WHEN** the user passes `--on-hold dim-up`
- **THEN** holding the button dims up and its short-press assignment is unchanged

### Requirement: Preserve everything else
Updating one button SHALL leave every other button, and every part of the targeted button not covered by the given flags (including long-press and time-based assignments), unchanged on the bridge.

#### Scenario: Neighbours untouched
- **WHEN** button 1 of a four-button dimmer is reassigned
- **THEN** buttons 2–4 keep their exact stored configuration

### Requirement: Accessories without configuration and safety
`set control` SHALL fail with an explanatory error when the accessory has no behavior configuration. `--dry-run` SHALL print the configuration that would be sent and exit 0 without calling the bridge. No request SHALL be sent when validation fails, or when the target is a rotary.

#### Scenario: Dry run
- **WHEN** the user adds `--dry-run`
- **THEN** the new configuration is printed as JSON and the bridge is not modified

#### Scenario: Bridge error
- **WHEN** the bridge rejects the update
- **THEN** its message is printed and the exit status is non-zero

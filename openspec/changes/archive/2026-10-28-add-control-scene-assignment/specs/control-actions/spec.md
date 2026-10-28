# Spec Delta

## Purpose

Lets users see what pressing or holding each button of a Hue accessory does, without opening the Hue app.

## ADDED Requirements

### Requirement: Show button actions
`openhue get control` SHALL show for each button a "Short press" and a "Hold" summary derived from the accessory's configuration: `scene <name>` for one scene, `cycle: <names…> (<n>)` for several, `time-based` for time-of-day assignments, `smart-scene <id>` for smart scenes, `dim-up|dim-down|dim-alternate` for hold, and `-` when nothing is assigned or the accessory has no configuration. Scene IDs that no longer exist SHALL be shown as `<unknown scene>`.

#### Scenario: Single scene and cycle
- **WHEN** a dimmer's button 1 recalls scene "Cozy" and button 4 cycles seven scenes
- **THEN** the rows show `scene Cozy` and `cycle: … (7)` in the "Short press" column

#### Scenario: Unconfigured accessory
- **WHEN** a device has no configuration on the bridge
- **THEN** its buttons show `-` and no error is raised

### Requirement: Full configuration in JSON
With `--json`, each button SHALL include its raw configuration as stored on the bridge.

#### Scenario: JSON
- **WHEN** the user runs `openhue get control --json`
- **THEN** each button carries its configuration, including parts the table does not summarize

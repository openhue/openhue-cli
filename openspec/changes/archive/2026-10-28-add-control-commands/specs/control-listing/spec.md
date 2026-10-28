# Spec Delta

## Purpose

Lets users discover and inspect the physical controls (buttons and rotary dials) of their Hue devices from the command line.

## ADDED Requirements

### Requirement: List controls
`openhue get control` SHALL list every button and relative rotary known to the bridge as a table with ID, kind, device name, room name, control number and last event with its time. A control whose device is in no room SHALL show an empty room, and one that never reported SHALL show an empty last event.

#### Scenario: List all
- **WHEN** the user runs `openhue get control` on a home with a dimmer switch (4 buttons) and a tap dial (4 buttons, 1 rotary)
- **THEN** nine rows are printed, each with its kind, device and `control_id`

#### Scenario: No controls
- **WHEN** the home has no switch or dial
- **THEN** an empty table is printed and the exit status is 0

### Requirement: Filter controls
`get control` SHALL accept `--room`, `--device` (name or ID) and `--kind button|rotary`, combinable, and positional control IDs.

#### Scenario: By room and kind
- **WHEN** the user runs `openhue get control --room "Living Room" --kind rotary`
- **THEN** only rotaries whose device belongs to that room are listed

#### Scenario: By ID
- **WHEN** the user passes a control ID
- **THEN** only that control is listed; an unknown ID yields an error and a non-zero exit status

#### Scenario: Invalid kind
- **WHEN** `--kind` is neither `button` nor `rotary`
- **THEN** a usage error lists the allowed values

### Requirement: JSON output
With `--json`, `get control` SHALL print the matching controls as a JSON array including the full state reported by the bridge (button events and repeat interval, rotary last report).

#### Scenario: JSON
- **WHEN** the user runs `openhue get control --json`
- **THEN** valid JSON is printed, with an empty array when nothing matches

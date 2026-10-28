# Spec Delta

## Purpose

Lets users change the configurable settings of a Hue button from the command line, and explains clearly when a control has none.

## ADDED Requirements

### Requirement: Select the control to configure
`openhue set control` SHALL target exactly one control, given by ID, or by `<device name or ID>` together with `--control <n>` (the control number on that device).

#### Scenario: By ID
- **WHEN** the user runs `openhue set control <id> --repeat-interval 800ms`
- **THEN** that button is updated

#### Scenario: By device and control number
- **WHEN** the user runs `openhue set control "Dimmer" --control 2 --repeat-interval 800ms`
- **THEN** button number 2 of the device "Dimmer" is updated

#### Scenario: Ambiguous or missing
- **WHEN** the argument matches no control, or several (for example a device with four buttons and no `--control`)
- **THEN** an error naming the candidates is printed, nothing is sent and the exit status is non-zero

### Requirement: Configure button repeat interval
`set control` SHALL accept `--repeat-interval`, a duration (`500ms`, `1s`) or a number of milliseconds, and send it as the button's repeat interval; on success it SHALL print the control and new value.

#### Scenario: Valid interval
- **WHEN** the user passes `--repeat-interval 1s` for a button
- **THEN** the bridge receives 1000 ms and a confirmation is printed

#### Scenario: Invalid interval
- **WHEN** the value is negative, zero or not a duration
- **THEN** a usage error is printed and no request is sent

#### Scenario: Bridge error
- **WHEN** the bridge rejects the update
- **THEN** the error is printed and the exit status is non-zero

### Requirement: Controls without settings
`set control` SHALL fail when no setting flag is given, or when the target is a rotary (which has no writable setting).

#### Scenario: Rotary
- **WHEN** the target resolves to a rotary
- **THEN** an error states that rotaries have no configurable settings and nothing is sent

#### Scenario: No flag
- **WHEN** no setting flag is passed
- **THEN** a usage error lists the available flags

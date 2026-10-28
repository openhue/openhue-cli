# Spec Delta

## Purpose

Lets users remove a Hue scene from the command line, completing the scene lifecycle.

## ADDED Requirements

### Requirement: Delete a scene
`openhue delete scene <scene>` SHALL delete exactly one scene identified by name or ID, with optional `--room` to disambiguate, and SHALL print which scene was deleted.

#### Scenario: Delete by name
- **WHEN** the user runs `openhue delete scene Cozy` and exactly one scene "Cozy" exists
- **THEN** that scene is deleted and its name and ID are printed

#### Scenario: Not found
- **WHEN** no scene matches
- **THEN** an error is printed and the exit status is non-zero

#### Scenario: Ambiguous
- **WHEN** several scenes match and `--room` does not narrow it to one
- **THEN** an error asking for `--room` or the ID is printed and nothing is deleted

#### Scenario: Bridge error
- **WHEN** the bridge rejects the deletion
- **THEN** the error is printed and the exit status is non-zero

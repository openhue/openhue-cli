# Spec Delta

## Purpose

Lets users export an existing scene as an editable definition so it can be reviewed, versioned and re-applied.

## ADDED Requirements

### Requirement: Export a scene as YAML
`openhue get scene <scene> --yaml` SHALL print the scene in the definition-file format (name, room, dynamics, palette and one entry per light with its stored settings, light names used unless ambiguous), and SHALL fail when the argument does not match exactly one scene (`--room` disambiguates).

#### Scenario: Export
- **WHEN** the user runs `openhue get scene Cozy -r "Living Room" --yaml`
- **THEN** a YAML definition of that scene is printed

#### Scenario: Not exactly one scene
- **WHEN** no scene or several scenes match
- **THEN** an error is printed and the exit status is non-zero

### Requirement: Export round-trips
Applying an exported definition back to its own scene with `update scene --from-file` SHALL leave the scene unchanged, and creating a scene from it with another name SHALL produce an equivalent scene.

#### Scenario: Re-apply
- **WHEN** a scene is exported and re-applied to the same scene
- **THEN** the scene's stored actions are unchanged

#### Scenario: Clone
- **WHEN** the exported file is used with `create scene Copy --from-file`
- **THEN** the new scene has the same per-light settings, dynamics and palette

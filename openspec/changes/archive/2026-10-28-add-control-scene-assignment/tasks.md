# Tasks

Prerequisite: `add-control-commands` applied (control model, `get control`, `set control`).

## 1. openhue-go prerequisites

- [x] 1.1 In `~/local/openhue/openhue-go` (check `git status` first), add `Home.GetBehaviorInstances`, `GetBehaviorInstanceById`, `UpdateBehaviorInstance` and mock methods; verify with unit tests there (`go test ./...`)
- [x] 1.2 Release/tag it and bump the CLI `go.mod`; verify `go build ./...` with no committed `replace`

## 2. Behavior model

- [x] 2.1 Load behavior instances in the home context (failure tolerated with a warning) and link each to its controls by `configuration.device`; verify with `home_service_test.go` cases: configured, unconfigured, failing endpoint (`go test ./openhue`)
- [x] 2.2 Implement the summary function (single, cycle, time-based, smart scene, hold, unknown shapes → `custom`, missing scene → `<unknown scene>`); verify with table-driven tests built from the real configurations captured in the playground
- [x] 2.3 Implement the pure patch builder (`--scene`, `--with-off`, `--repeat-timeout`, `--on-hold`, `--clear`) over the untyped tree, with room check and conflict rules; verify with tests that unchanged buttons and unknown keys re-serialize byte-identically

## 3. Commands

- [x] 3.1 Add "Short press"/"Hold" columns and JSON configuration to `cmd/get/get_control.go`; verify with Cobra tests (`go test ./cmd/get`)
- [x] 3.2 Extend `cmd/set/set_control.go` with the new flags and `--dry-run`; verify with mock-based tests for single, cycle, hold, clear, wrong-room scene, ambiguous scene, `--clear` + `--scene`, rotary, unconfigured accessory, dry run, bridge error, asserting no request on validation failure (`go test ./cmd/set`)

## 4. Docs and integration

- [x] 4.1 Document assignment and examples in README and command help; verify examples with `--dry-run`
- [x] 4.2 Run `make test` and `go vet ./...`; optionally on a real bridge: `set control … --scene X --dry-run`, then for real on a spare switch, `get control --json`, and restore the original configuration

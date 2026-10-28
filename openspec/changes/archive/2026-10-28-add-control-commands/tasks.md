# Tasks

## 1. openhue-go prerequisites

- [x] 1.1 In `~/local/openhue/openhue-go` (check `git status` first), add `Home.GetRelativeRotaries`, `GetRelativeRotaryById`, `UpdateButton`, `UpdateRelativeRotary` and the matching `ClientWithResponsesMock` methods; verify with new unit tests in that repo (`go test ./...`)
- [x] 1.2 Release/tag `openhue-go` and bump it in the CLI `go.mod` (`make tidy`); verify `go build ./...` succeeds with no committed `replace` directive

## 2. Control model

- [x] 2.1 Load buttons and relative rotaries in `openhue/home_context.go` (tolerating endpoint failure with a warning) and attach `Control` values to `Device` via owner ID, exposing kind, `control_id`, last report and room; verify with `home_service_test.go` cases for a switch, a dial, a device without controls and a failing endpoint (`go test ./openhue`)
- [x] 2.2 Add `SearchControls(home, room, device, kind, ids)` and `FindControl(home, deviceOrId, controlNumber)` helpers with the zero/ambiguous rules; verify with table-driven tests

## 3. Get command

- [x] 3.1 Add `cmd/get/get_control.go` (table, `--json`, `--room`, `--device`, `--kind`, positional IDs, `controls` alias) and register it in `cmd/get/get.go`; verify with Cobra tests via `openhue.NewTestContext` for list-all, each filter, unknown ID, bad `--kind`, empty home and JSON (`go test ./cmd/get`)

## 4. Set command

- [x] 4.1 Add `cmd/set/set_control.go` (ID or device + `--control`, `--repeat-interval` duration/ms parsing > 0, no-flag and rotary errors) and register it in `cmd/set/set.go`; verify with tests on the `openhue-go` mock for success, ambiguous/missing, invalid interval, rotary target, no flag and bridge error, asserting no request on validation failure (`go test ./cmd/set`)

## 5. Docs and integration

- [x] 5.1 Document `get control` and `set control` in README and command help with examples; verify each example matches `--help`
- [x] 5.2 Run `make test` and `go vet ./...`; verify both pass, and optionally try against a real bridge with a switch/dial: `get control` → `set control "<device>" --control 1 --repeat-interval 800ms` → `get control --json`

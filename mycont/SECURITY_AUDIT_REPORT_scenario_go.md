# Final Status — mx-chain-scenario-go

## 0 critical vulnerabilities
## 0 high-severity vulnerabilities
## 0 medium-severity vulnerabilities
## 0 exploitable bugs

---

## Basis

**Full codebase scan performed** across all packages:
`worldmock/`, `scenario/executor/`, `scenario/io/`, `scenario/json/parse/`, `scenario/expression/`, `scenario/exporter/`, `clibase/`, `orderedjson/`

**16 scanner alerts raised — all verified as false positives:**

- Code injection alert (`cliMain.go`) — standard CLI entry point, no dynamic execution
- Path traversal alerts (`scenarioIO.go`, `scenarioDir.go`, `fmtDir.go`, `cliRun.go`, `fileResolverDefault.go`, `getSCCode.go`) — local developer tool, no untrusted external input surface; `fileResolverDefault.go` replacement branch only reachable from developer-written test code
- Log injection alerts (`stepCheckState.go`, `mockESDT.go`, `dnsUtil.go`) — test framework with no remote attacker surface
- Nil pointer dereference alerts (`parseValue.go`, `parseScenario.go`, `ojModel.go`) — all type assertions are guarded
- Insecure file permissions (`scenarioIO.go`) — `0600` is correct owner-only read/write

**5 targeted findings from prior audit — all verified as false positives:**

- `mockAddress.go` panic on nil vmType — intentional fail-fast, no live nil caller in codebase
- `worldUpdate.go` RollbackChanges hardcoded index 0 — `SaveKeyValue` (the only path that would create extra snapshots) is never called within this repo; exactly one snapshot exists at rollback time
- `dnsUtil.go` silent error discard — fixed-format expression that cannot fail in practice
- `stepRunTx.go` ESDT balance delta invariant — ESDT path never reaches the check; correct by design
- `worldAccountMap.go` dead error path — unreachable code, not a bug

---

## Scope Boundary

This status applies to `mx-chain-scenario-go` as a **local developer test framework**.  
It is not designed as a network-facing service and has not been evaluated for that use case.

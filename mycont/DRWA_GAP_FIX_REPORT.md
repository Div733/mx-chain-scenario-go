# DRWA Whitelist Gap Fix Report — mx-chain-scenario-go

---

## Background

DRWA (Distributed Replicated World Accounts) sync is a protocol feature that allows authorized smart contracts to emit state-synchronization writes. The `MockWorld` in this framework implements the `BlockchainHook` interface, which includes three DRWA-related methods:

- `IsAuthorizedDRWASyncCaller` — checks if an address is allowed to emit DRWA sync writes
- `ApplyDRWASyncEnvelopeBytes` — applies a DRWA sync payload from an authorized caller
- `QueryDRWANativeGovernance` — queries native DRWA governance state

The commit added `AuthorizedDRWASyncCallers` as a JSON-configurable whitelist in scenario `setState` steps. Three gaps were identified and fixed.

---

## Gap 1 — IsAuthorizedDRWASyncCaller Ignored the JSON Whitelist When a Hook Was Set

### File
`worldmock/worldBlockchainHook.go`

### What the gap was
The original implementation checked `ProvidedBlockchainHook` first and returned immediately if a hook was present, completely bypassing the local JSON whitelist:

```go
// BEFORE — hook took full priority
func (b *MockWorld) IsAuthorizedDRWASyncCaller(callerAddress []byte) bool {
    if b.ProvidedBlockchainHook != nil {
        return b.ProvidedBlockchainHook.IsAuthorizedDRWASyncCaller(callerAddress)
    }
    _, ok := b.AuthorizedDRWASyncCallers[string(callerAddress)]
    return ok
}
```

This meant that if a test installed a `ProvidedBlockchainHook` for any reason (e.g. to handle `ApplyDRWASyncEnvelopeBytes`), the entire `authorizedDRWASyncCallers` field in the scenario JSON was silently ignored. A caller explicitly whitelisted in the scenario file would be denied if the hook returned `false` for it.

### Why it mattered
The `authorizedDRWASyncCallers` field exists specifically so scenario authors can declare authorized callers in JSON without writing Go code. If installing a hook silently overrides that declaration, the feature is broken in any test that uses both a hook and a JSON whitelist — which is the common case for DRWA scenario testing.

### The fix
Check the local whitelist first. Only consult the hook as a fallback for callers not already in the whitelist:

```go
// AFTER — both sources merged, local whitelist checked first
func (b *MockWorld) IsAuthorizedDRWASyncCaller(callerAddress []byte) bool {
    _, ok := b.AuthorizedDRWASyncCallers[string(callerAddress)]
    if ok {
        return true
    }
    if b.ProvidedBlockchainHook != nil {
        return b.ProvidedBlockchainHook.IsAuthorizedDRWASyncCaller(callerAddress)
    }
    return false
}
```

A caller is now authorized if it appears in the local JSON whitelist OR if the hook authorizes it. Neither source can silently cancel the other.

---

## Gap 2 — ExecuteSetStateStep Only Appended to the Whitelist, Never Replaced It

### File
`scenario/executor/stepSetState.go`

### What the gap was
Each `setState` step only added callers to the existing map — it never cleared it:

```go
// BEFORE — append only, stale callers survive forever
for _, caller := range step.AuthorizedDRWASyncCallers {
    ae.World.AuthorizedDRWASyncCallers[string(caller.Value)] = struct{}{}
}
```

This meant that once an address was added to the whitelist by any `setState` step, it remained authorized for the entire lifetime of the scenario — even if a later `setState` step intentionally omitted it. There was no way to revoke authorization through the scenario JSON.

### Why it mattered
A `setState` step is meant to define the complete, authoritative world state at that point in the scenario. Every other field in `setState` (accounts, block info, block hashes) replaces the previous value. The whitelist being append-only was inconsistent with this contract and made it impossible to write scenarios that test authorization revocation or that reset the world to a clean state between scenario phases.

Concretely: a scenario with two `setState` steps where step 1 authorizes address A and step 2 authorizes only address B would incorrectly leave address A authorized after step 2.

### The fix
Clear the map before populating it on each `setState`, so each step defines the complete whitelist:

```go
// AFTER — replace semantics, each setState defines the full whitelist
ae.World.AuthorizedDRWASyncCallers = make(map[string]struct{})
for _, caller := range step.AuthorizedDRWASyncCallers {
    ae.World.AuthorizedDRWASyncCallers[string(caller.Value)] = struct{}{}
}
```

If a `setState` step has an empty `authorizedDRWASyncCallers` field, the whitelist is cleared. If it has entries, only those entries are authorized. This is consistent with how all other `setState` fields behave.

---

## Gap 3 — No Tests for Any DRWA Whitelist Behavior

### File
`worldmock/worldAuditFixes_test.go`

### What the gap was
The entire `AuthorizedDRWASyncCallers` feature — JSON parsing, `setState` population, `IsAuthorizedDRWASyncCaller` logic, and `ApplyDRWASyncEnvelopeBytes` behavior — had zero test coverage. The gaps in Gap 1 and Gap 2 existed undetected precisely because there were no tests to catch them.

### Why it mattered
Without tests, both bugs could have been shipped and would only have been discovered when DRWA scenario tests produced wrong authorization results at runtime — with no clear indication of where the failure originated.

### The fix
Six tests were added covering every behavioral path:

| Test | What it verifies |
|---|---|
| `TestIsAuthorizedDRWASyncCallerChecksLocalWhitelistEvenWithHook` | Local whitelist wins even when hook returns false — Gap 1 fix |
| `TestIsAuthorizedDRWASyncCallerReturnsFalseWhenNotInWhitelistOrHook` | Unauthorized caller correctly denied with no hook |
| `TestIsAuthorizedDRWASyncCallerDelegatesToHookWhenNotInLocalWhitelist` | Hook can still authorize callers not in local whitelist |
| `TestApplyDRWASyncEnvelopeBytesSucceedsForAuthorizedCaller` | Whitelisted caller succeeds without needing a hook |
| `TestApplyDRWASyncEnvelopeBytesErrorsForUnauthorizedCallerWithNoHook` | Unauthorized caller with no hook gets correct error |
| `TestSetStateDRWAWhitelistReplacesNotAppends` | Stale callers do not survive after a new setState — Gap 2 fix |

All 6 tests pass. The test suite for the full repository passes with 0 failures.

---

## Summary

| Gap | Root Cause | Impact if Unfixed | Fix |
|---|---|---|---|
| Gap 1 | Hook short-circuited before local whitelist was checked | JSON whitelist silently ignored in any test using a hook | Check local whitelist first, hook as fallback |
| Gap 2 | setState only appended, never replaced | Stale authorized callers survived across setState steps, authorization revocation impossible | Clear map before each setState population |
| Gap 3 | No tests existed for DRWA whitelist behavior | Gaps 1 and 2 undetectable until runtime failure | 6 targeted tests added covering all behavioral paths |

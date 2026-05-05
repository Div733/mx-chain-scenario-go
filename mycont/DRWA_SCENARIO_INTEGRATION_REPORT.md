# DRWA Integration Report — mx-chain-scenario-go

## 1. What this repo is

mx-chain-scenario-go is the scenario test harness for the MultiversX VM. It provides
MockWorld — a fake blockchain implementation used by mx-chain-vm-go and integration
tests to execute smart contract scenarios without a real node.

Its DRWA role is narrow but critical: MockWorld must implement the full
vmcommon.BlockchainHook interface, which now includes 3 DRWA methods. Without a
correct implementation, mx-chain-vm-go cannot compile against the local vm-common,
and scenario-based DRWA tests cannot assert what sync envelopes were applied.

---

## 2. State before this work

### 2.1 Build was broken

The repo pointed to the published v1.6.0 of mx-chain-vm-common-go via go.mod with no
replace directive. The local right/core/mx-chain-vm-common-go adds 3 new methods to
the BlockchainHook interface:

    ApplyDRWASyncEnvelopeBytes(payload []byte, callerAddress []byte) error
    QueryDRWANativeGovernance(queryType uint32, key []byte) ([]byte, error)
    IsAuthorizedDRWASyncCaller(callerAddress []byte) bool

The published v1.6.0 does not have these. The worldBlockchainHook.go already called
them on ProvidedBlockchainHook, so the build failed with:

    worldBlockchainHook.go:158: b.ProvidedBlockchainHook.ApplyDRWASyncEnvelopeBytes
        undefined (type vmcommon.BlockchainHook has no field or method ApplyDRWASyncEnvelopeBytes)
    worldBlockchainHook.go:170: b.ProvidedBlockchainHook.QueryDRWANativeGovernance
        undefined (type vmcommon.BlockchainHook has no field or method QueryDRWANativeGovernance)
    worldBlockchainHook.go:180: b.ProvidedBlockchainHook.IsAuthorizedDRWASyncCaller
        undefined (type vmcommon.BlockchainHook has no field or method IsAuthorizedDRWASyncCaller)

### 2.2 ApplyDRWASyncEnvelopeBytes returned an error with no hook

When no ProvidedBlockchainHook was set, the method returned:

    return ErrProvidedBlockchainHookNotInitialized

This is wrong for a mock. Scenario tests call ManagedDRWASyncMirror in vm-go, which
calls ApplyDRWASyncEnvelopeBytes on MockWorld. If MockWorld has no real hook (the
normal case in unit tests), the call would fail with an error instead of succeeding
and recording what was applied. Tests could not assert which sync envelopes were
submitted.

### 2.3 No payload capture field on MockWorld

There was no way for a test to inspect what DRWA sync envelopes were applied during
a scenario run. The struct had no field for it.

### 2.4 Existing test asserted the wrong behaviour

worldAuditFixes_test.go contained:

    func TestApplyDRWASyncEnvelopeBytesRequiresProvidedHook(t *testing.T) {
        world := NewMockWorld()
        err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), []byte("caller"))
        require.ErrorIs(t, err, ErrProvidedBlockchainHookNotInitialized)
    }

This test locked in the broken behaviour. It would pass before the fix and fail after,
which is the correct outcome — the test needed to be updated to match the new design.

---

## 3. What was wrong — summary table

| # | Problem | Impact |
|---|---------|--------|
| 1 | go.mod pointed to published v1.6.0, no replace for local vm-common | Build broken — 3 undefined method errors |
| 2 | go.mod missing replace for Div733/mx-chain-core-go fork | Transitive dependency mismatch |
| 3 | go.mod missing replace for multiversx/protobuf | Protobuf fork not resolved |
| 4 | ApplyDRWASyncEnvelopeBytes returned error when no hook set | Scenario DRWA tests would always fail |
| 5 | No DRWASyncPayloads field on MockWorld | No way to assert what was synced in tests |
| 6 | Existing test asserted the error return (wrong behaviour) | Test would block the correct fix |

---

## 4. Changes made

### 4.1 go.mod — 3 replace directives added, 1 version bumped

**File:** `go.mod`

Added:

    replace github.com/gogo/protobuf => github.com/multiversx/protobuf v1.3.2

    replace github.com/multiversx/mx-chain-core-go => github.com/Div733/mx-chain-core-go v0.0.0-20260505064222-c3919dac4cc5

    replace github.com/multiversx/mx-chain-vm-common-go => ../mx-chain-vm-common-go

Bumped:

    github.com/multiversx/mx-chain-core-go v1.4.0  →  v1.4.1

Why: The local vm-common requires core-go v1.4.1 and uses the Div733 fork. Without
these replace directives, Go resolves the published versions which do not have the
DRWA interface additions. The local path replace for vm-common wires this repo
directly to right/core/mx-chain-vm-common-go so the 3 new BlockchainHook methods
are visible at compile time.

### 4.2 worldDef.go — DRWASyncPayloads field added

**File:** `worldmock/worldDef.go`

Added to MockWorld struct:

    DRWASyncPayloads [][]byte

Initialized in NewMockWorld:

    DRWASyncPayloads: nil,

Why: Scenario tests need to assert what DRWA sync envelopes were applied during a
run. Without this field there is no observable output from ApplyDRWASyncEnvelopeBytes
when no real hook is present. The field accumulates every payload that passes through
the mock, in order, so tests can do:

    require.Len(t, world.DRWASyncPayloads, 1)
    require.Equal(t, expectedPayload, world.DRWASyncPayloads[0])

### 4.3 worldBlockchainHook.go — ApplyDRWASyncEnvelopeBytes behaviour fixed

**File:** `worldmock/worldBlockchainHook.go`

Before:

    func (b *MockWorld) ApplyDRWASyncEnvelopeBytes(payload []byte, callerAddress []byte) error {
        if b.Err != nil {
            return b.Err
        }
        if b.ProvidedBlockchainHook != nil {
            return b.ProvidedBlockchainHook.ApplyDRWASyncEnvelopeBytes(payload, callerAddress)
        }
        return ErrProvidedBlockchainHookNotInitialized
    }

After:

    func (b *MockWorld) ApplyDRWASyncEnvelopeBytes(payload []byte, callerAddress []byte) error {
        if b.Err != nil {
            return b.Err
        }
        if b.ProvidedBlockchainHook != nil {
            return b.ProvidedBlockchainHook.ApplyDRWASyncEnvelopeBytes(payload, callerAddress)
        }
        payloadCopy := make([]byte, len(payload))
        copy(payloadCopy, payload)
        b.DRWASyncPayloads = append(b.DRWASyncPayloads, payloadCopy)
        return nil
    }

Why: A mock should not fail when used standalone. The whole point of MockWorld is to
run without a real node. Returning an error here would cause every DRWA scenario test
that calls ManagedDRWASyncMirror to fail at the hook boundary rather than testing the
actual contract logic. The payload is deep-copied before appending so the caller
cannot mutate the captured slice after the call.

Note: QueryDRWANativeGovernance and IsAuthorizedDRWASyncCaller were already correctly
implemented — they delegate to ProvidedBlockchainHook when set, and return zero values
(nil, error) and (false) respectively when not set. No changes were needed there.

### 4.4 worldAuditFixes_test.go — test updated to assert capture behaviour

**File:** `worldmock/worldAuditFixes_test.go`

Before:

    func TestApplyDRWASyncEnvelopeBytesRequiresProvidedHook(t *testing.T) {
        world := NewMockWorld()
        err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), []byte("caller"))
        require.ErrorIs(t, err, ErrProvidedBlockchainHookNotInitialized)
    }

After:

    func TestApplyDRWASyncEnvelopeBytesCaptures(t *testing.T) {
        world := NewMockWorld()
        err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), []byte("caller"))
        require.NoError(t, err)
        require.Len(t, world.DRWASyncPayloads, 1)
        require.Equal(t, []byte("payload"), world.DRWASyncPayloads[0])
    }

Why: The old test was asserting that the mock fails when used without a hook. That is
the wrong contract for a mock. The new test asserts the correct behaviour: the call
succeeds and the payload is captured for inspection.

---

## 5. What was NOT changed

- QueryDRWANativeGovernance — already correct, no change
- IsAuthorizedDRWASyncCaller — already correct, no change
- ErrProvidedBlockchainHookNotInitialized — kept, still used by QueryDRWANativeGovernance
- All non-DRWA methods — untouched
- All scenario JSON/expression/executor packages — untouched

---

## 6. File change summary

| File | Change |
|------|--------|
| `go.mod` | +3 replace directives, core-go bumped v1.4.0 → v1.4.1, go mod tidy run |
| `worldmock/worldDef.go` | +1 field (DRWASyncPayloads [][]byte) on MockWorld, +1 init line in NewMockWorld |
| `worldmock/worldBlockchainHook.go` | ApplyDRWASyncEnvelopeBytes: error return → payload capture + nil return |
| `worldmock/worldAuditFixes_test.go` | 1 test renamed and rewritten to assert capture behaviour |

---

## 7. Build and test results

    go build ./...   PASS
    go test ./...    12 packages, ALL PASS

    ok  github.com/multiversx/mx-chain-scenario-go/orderedjson
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/executor
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/executor/test
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/expression/fileresolver
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/expression/integrationTests
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/expression/interpreter
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/expression/reconstructor
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/io
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/json/integrationTests
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/json/parse
    ok  github.com/multiversx/mx-chain-scenario-go/scenario/tests
    ok  github.com/multiversx/mx-chain-scenario-go/worldmock

---

## 8. What this unblocks

mx-chain-vm-go imports mx-chain-scenario-go for its test infrastructure. With
scenario-go now building cleanly against the local vm-common, mx-chain-vm-go can
be wired to the same local dependency chain and its DRWA hook (ManagedDRWASyncMirror
in managedei.go) can be tested end-to-end in scenario tests.

---

## 9. Next repo

mx-chain-vm-go — review ManagedDRWASyncMirror implementation in
vmhost/vmhooks/managedei.go and wire go.mod to the local dependency chain.

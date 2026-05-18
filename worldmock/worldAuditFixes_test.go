package worldmock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountValidateReportsCodeWithoutSmartContractAddress(t *testing.T) {
	address := make([]byte, 32)
	address[0] = 1
	account := &Account{
		Address: address,
		Code:    []byte("contract"),
	}

	err := account.Validate()
	require.EqualError(t, err, "account has code but not a smart contract address: 0x0100000000000000000000000000000000000000000000000000000000000000")
}

func TestAccountValidateReportsSmartContractAddressWithoutCode(t *testing.T) {
	address := make([]byte, 32)
	address[8] = 1
	account := &Account{
		Address: address,
	}

	err := account.Validate()
	require.EqualError(t, err, "account has a smart contract address, but has no code: 0x0000000000000000010000000000000000000000000000000000000000000000")
}

func TestGetBuiltinFunctionNamesReturnsNilWhenWrapperNotInitialized(t *testing.T) {
	world := NewMockWorld()

	require.Nil(t, world.GetBuiltinFunctionNames())
}

// Gap 1: IsAuthorizedDRWASyncCaller must check local whitelist even when ProvidedBlockchainHook is set.
func TestIsAuthorizedDRWASyncCallerChecksLocalWhitelistEvenWithHook(t *testing.T) {
	world := NewMockWorld()
	caller := []byte("authorized-caller")
	world.AuthorizedDRWASyncCallers[string(caller)] = struct{}{}

	// hook that always returns false for auth
	world.ProvidedBlockchainHook = &mockHookAuthAlwaysFalse{}

	// local whitelist must still win
	require.True(t, world.IsAuthorizedDRWASyncCaller(caller))
}

// Gap 1: unauthorized caller must return false when hook also returns false.
func TestIsAuthorizedDRWASyncCallerReturnsFalseWhenNotInWhitelistOrHook(t *testing.T) {
	world := NewMockWorld()

	require.False(t, world.IsAuthorizedDRWASyncCaller([]byte("unknown-caller")))
}

// Gap 1: hook can still authorize callers not in the local whitelist.
func TestIsAuthorizedDRWASyncCallerDelegatesToHookWhenNotInLocalWhitelist(t *testing.T) {
	world := NewMockWorld()
	caller := []byte("hook-authorized-caller")

	world.ProvidedBlockchainHook = &mockHookAuthAlwaysTrue{}

	require.True(t, world.IsAuthorizedDRWASyncCaller(caller))
}

// Gap 2: ApplyDRWASyncEnvelopeBytes returns nil for whitelisted caller (no hook needed).
func TestApplyDRWASyncEnvelopeBytesSucceedsForAuthorizedCaller(t *testing.T) {
	world := NewMockWorld()
	caller := []byte("authorized-caller")
	world.AuthorizedDRWASyncCallers[string(caller)] = struct{}{}

	err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), caller)
	require.NoError(t, err)
}

// Gap 2: ApplyDRWASyncEnvelopeBytes returns error for unauthorized caller with no hook.
func TestApplyDRWASyncEnvelopeBytesErrorsForUnauthorizedCallerWithNoHook(t *testing.T) {
	world := NewMockWorld()

	err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), []byte("unknown-caller"))
	require.ErrorIs(t, err, ErrProvidedBlockchainHookNotInitialized)
}

// Gap 3: setState replaces the whitelist, not appends — stale callers must not survive.
func TestSetStateDRWAWhitelistReplacesNotAppends(t *testing.T) {
	world := NewMockWorld()
	staleCallerAddr := string([]byte("stale-caller"))
	world.AuthorizedDRWASyncCallers[staleCallerAddr] = struct{}{}

	// simulate a new setState with a different caller
	world.AuthorizedDRWASyncCallers = make(map[string]struct{})
	world.AuthorizedDRWASyncCallers[string([]byte("new-caller"))] = struct{}{}

	_, staleExists := world.AuthorizedDRWASyncCallers[staleCallerAddr]
	require.False(t, staleExists, "stale caller must not survive after setState replacement")
	require.True(t, world.IsAuthorizedDRWASyncCaller([]byte("new-caller")))
}

// mockHookAuthAlwaysFalse is a minimal BlockchainHook stub that denies all DRWA auth.
type mockHookAuthAlwaysFalse struct {
	MockWorld
}

func (m *mockHookAuthAlwaysFalse) IsAuthorizedDRWASyncCaller(_ []byte) bool {
	return false
}

// mockHookAuthAlwaysTrue is a minimal BlockchainHook stub that approves all DRWA auth.
type mockHookAuthAlwaysTrue struct {
	MockWorld
}

func (m *mockHookAuthAlwaysTrue) IsAuthorizedDRWASyncCaller(_ []byte) bool {
	return true
}

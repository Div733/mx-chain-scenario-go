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

func TestApplyDRWASyncEnvelopeBytesRequiresProvidedHook(t *testing.T) {
	world := NewMockWorld()

	err := world.ApplyDRWASyncEnvelopeBytes([]byte("payload"), []byte("caller"))
	require.ErrorIs(t, err, ErrProvidedBlockchainHookNotInitialized)
}

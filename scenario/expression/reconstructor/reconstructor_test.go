package scenexpressionreconstructor

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestReconstructAddressHintFallsBackWhenBech32ConverterCreationFails(t *testing.T) {
	originalFactory := newBech32PubkeyConverter
	newBech32PubkeyConverter = func(int, string) (core.PubkeyConverter, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() {
		newBech32PubkeyConverter = originalFactory
	})

	addr := []byte(strings.Repeat("A", 32))
	reconstructor := ExprReconstructor{Bech32Addr: true}

	result := reconstructor.Reconstruct(addr, AddressHint)

	require.Equal(t, "0x4141414141414141414141414141414141414141414141414141414141414141 (str:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA)", result)
}

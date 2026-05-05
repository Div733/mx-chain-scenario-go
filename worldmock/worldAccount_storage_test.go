package worldmock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageValueReturnsFreshEmptySliceWhenKeyMissing(t *testing.T) {
	account := &Account{
		Storage: make(map[string][]byte),
	}

	first := account.StorageValue("missing")
	second := account.StorageValue("missing")

	require.Empty(t, first)
	require.Empty(t, second)

	first = append(first, 1)
	require.Equal(t, []byte{1}, first)
	require.Empty(t, second)
}

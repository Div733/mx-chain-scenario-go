package worldmock

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRollbackChangesRestoresSnapshotAndRemovesNewAccounts(t *testing.T) {
	world := NewMockWorld()
	address := []byte("account-a")

	account := world.AcctMap.CreateAccount(address, world)
	account.Exists = true
	account.Nonce = 7
	account.Balance = big.NewInt(1000)
	account.BalanceDelta = big.NewInt(0)
	account.Storage["key"] = []byte("value")
	account.Code = []byte("code")
	account.CodeHash = []byte("hash")
	account.CodeMetadata = []byte("meta")
	account.OwnerAddress = []byte("owner")
	account.AsyncCallData = "async"
	account.Username = []byte("name")
	account.DeveloperReward = big.NewInt(9)
	account.ShardID = 2
	account.IsSmartContract = true

	world.CreateStateBackup()

	account.Exists = false
	account.Nonce = 99
	account.Balance = big.NewInt(5)
	account.BalanceDelta = big.NewInt(123)
	account.Storage["key"] = []byte("mutated")
	account.Storage["new-key"] = []byte("new")
	account.Code = []byte("new-code")
	account.CodeHash = []byte("new-hash")
	account.CodeMetadata = []byte("new-meta")
	account.OwnerAddress = []byte("new-owner")
	account.AsyncCallData = "new-async"
	account.Username = []byte("new-name")
	account.DeveloperReward = big.NewInt(0)
	account.ShardID = 0
	account.IsSmartContract = false

	ghost := world.AcctMap.CreateAccount([]byte("ghost"), world)
	ghost.Storage["ghost-key"] = []byte("ghost-value")

	err := world.RollbackChanges()
	require.NoError(t, err)

	restored := world.AcctMap.GetAccount(address)
	require.NotNil(t, restored)
	require.True(t, restored.Exists)
	require.Equal(t, uint64(7), restored.Nonce)
	require.Zero(t, restored.Balance.Cmp(big.NewInt(1000)))
	require.Zero(t, restored.BalanceDelta.Cmp(big.NewInt(0)))
	require.Equal(t, []byte("value"), restored.Storage["key"])
	require.NotContains(t, restored.Storage, "new-key")
	require.Equal(t, []byte("code"), restored.Code)
	require.Equal(t, []byte("hash"), restored.CodeHash)
	require.Equal(t, []byte("meta"), restored.CodeMetadata)
	require.Equal(t, []byte("owner"), restored.OwnerAddress)
	require.Equal(t, "async", restored.AsyncCallData)
	require.Equal(t, []byte("name"), restored.Username)
	require.Zero(t, restored.DeveloperReward.Cmp(big.NewInt(9)))
	require.Equal(t, uint32(2), restored.ShardID)
	require.True(t, restored.IsSmartContract)
	require.Nil(t, world.AcctMap.GetAccount([]byte("ghost")))
}

func TestComputeIdReturnsZeroForUnknownAccount(t *testing.T) {
	world := NewMockWorld()

	require.Equal(t, uint32(0), world.ComputeId([]byte("missing")))
}

func TestSameShardTreatsUnknownAccountsAsShardZero(t *testing.T) {
	world := NewMockWorld()

	account := world.AcctMap.CreateAccount([]byte("known"), world)
	account.ShardID = 1

	require.False(t, world.SameShard([]byte("known"), []byte("missing")))
	require.True(t, world.SameShard([]byte("missing-a"), []byte("missing-b")))
}

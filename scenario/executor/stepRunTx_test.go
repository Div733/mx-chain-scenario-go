package scenexec

import (
	"math/big"
	"testing"

	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	worldmock "github.com/multiversx/mx-chain-scenario-go/worldmock"
	"github.com/stretchr/testify/require"
)

func TestExecuteTxRollsBackFailedSetupMutations(t *testing.T) {
	world := worldmock.NewMockWorld()
	sender := world.AcctMap.CreateAccount([]byte("sender"), world)
	sender.Balance = big.NewInt(10)

	executor := &ScenarioExecutor{World: world}
	tx := &scenmodel.Transaction{
		Type:     scenmodel.Transfer,
		From:     scenmodel.NewJSONBytesFromString(sender.Address, "sender"),
		To:       scenmodel.NewJSONBytesFromString([]byte("receiver"), "receiver"),
		EGLDValue: scenmodel.JSONBigInt{
			Value:    big.NewInt(1),
			Original: "1",
		},
		GasLimit: scenmodel.JSONUint64{
			Value:    100,
			Original: "100",
		},
		GasPrice: scenmodel.JSONUint64{
			Value:    1,
			Original: "1",
		},
	}

	output, err := executor.executeTx("tx-setup-fail", tx)
	require.Nil(t, output)
	require.ErrorContains(t, err, "could not set up tx tx-setup-fail")
	require.Equal(t, uint64(0), sender.Nonce)
	require.Zero(t, sender.Balance.Cmp(big.NewInt(10)))
	require.Len(t, world.AccountsAdapter.(*worldmock.MockAccountsAdapter).Snapshots, 0)
}

func TestExecuteTxRollsBackGasAndNonceOnFailedTransaction(t *testing.T) {
	world := worldmock.NewMockWorld()
	sender := world.AcctMap.CreateAccount([]byte("sender"), world)
	sender.Balance = big.NewInt(50)

	executor := &ScenarioExecutor{World: world}
	tx := &scenmodel.Transaction{
		Type:     scenmodel.Transfer,
		From:     scenmodel.NewJSONBytesFromString(sender.Address, "sender"),
		To:       scenmodel.NewJSONBytesFromString([]byte("receiver"), "receiver"),
		EGLDValue: scenmodel.JSONBigInt{
			Value:    big.NewInt(100),
			Original: "100",
		},
		GasLimit: scenmodel.JSONUint64{
			Value:    10,
			Original: "10",
		},
		GasPrice: scenmodel.JSONUint64{
			Value:    1,
			Original: "1",
		},
	}

	output, err := executor.executeTx("tx-runtime-fail", tx)
	require.NotNil(t, output)
	require.ErrorContains(t, err, "tx step failed")
	require.Equal(t, uint64(0), sender.Nonce)
	require.Zero(t, sender.Balance.Cmp(big.NewInt(50)))
	require.Len(t, world.AccountsAdapter.(*worldmock.MockAccountsAdapter).Snapshots, 0)
}

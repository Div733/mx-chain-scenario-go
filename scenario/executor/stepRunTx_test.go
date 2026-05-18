package scenexec

import (
	"math/big"
	"testing"

	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	worldmock "github.com/multiversx/mx-chain-scenario-go/worldmock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

// TestExecuteTxSetupFailureLeavesNoSnapshot covers the case where the
// pre-execution gas charge cannot succeed (sender cannot afford gas).
// UpdateWorldStateBefore fails before CreateStateBackup is called, so no
// rollback boundary is set up. The nonce is still incremented because
// UpdateWorldStateBefore bumps it unconditionally — matching protocol
// semantics where a malformed tx still consumes the sender's nonce slot.
func TestExecuteTxSetupFailureLeavesNoSnapshot(t *testing.T) {
	world := worldmock.NewMockWorld()
	sender := world.AcctMap.CreateAccount([]byte("sender"), world)
	sender.Balance = big.NewInt(10)

	executor := &ScenarioExecutor{World: world}
	tx := &scenmodel.Transaction{
		Type: scenmodel.Transfer,
		From: scenmodel.NewJSONBytesFromString(sender.Address, "sender"),
		To:   scenmodel.NewJSONBytesFromString([]byte("receiver"), "receiver"),
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
	require.Equal(t, uint64(1), sender.Nonce)
	require.Zero(t, sender.Balance.Cmp(big.NewInt(10)))
	require.Len(t, world.AccountsAdapter.(*worldmock.MockAccountsAdapter).Snapshots, 0)
}

// TestExecuteTxRollsBackPostSnapshotMutationsOnFailedVM covers the case
// where the pre-execution gas charge succeeds (nonce bumped, gas debited),
// but the VM execution itself returns a non-Ok status (here: out-of-funds
// because the call value exceeds the post-gas balance). Pre-snapshot state
// (the nonce bump and the gas charge) is preserved — that is what real
// MultiversX does. Only post-snapshot mutations would be rolled back, of
// which there are none for outOfFundsResult. The function must still return
// (output, nil) so the caller can compare the VM output against the scenario
// expected result.
func TestExecuteTxRollsBackPostSnapshotMutationsOnFailedVM(t *testing.T) {
	world := worldmock.NewMockWorld()
	sender := world.AcctMap.CreateAccount([]byte("sender"), world)
	sender.Balance = big.NewInt(50)

	executor := &ScenarioExecutor{World: world}
	tx := &scenmodel.Transaction{
		Type: scenmodel.Transfer,
		From: scenmodel.NewJSONBytesFromString(sender.Address, "sender"),
		To:   scenmodel.NewJSONBytesFromString([]byte("receiver"), "receiver"),
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
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, vmcommon.OutOfFunds, output.ReturnCode)
	require.Equal(t, uint64(1), sender.Nonce)
	require.Zero(t, sender.Balance.Cmp(big.NewInt(40)))
	require.Len(t, world.AccountsAdapter.(*worldmock.MockAccountsAdapter).Snapshots, 0)
}

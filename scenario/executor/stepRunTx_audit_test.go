package scenexec

import (
	"math/big"
	"testing"

	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	worldmock "github.com/multiversx/mx-chain-scenario-go/worldmock"
	"github.com/stretchr/testify/require"
)

func TestSenderHasEnoughBalanceReturnsFalseWhenSenderMissing(t *testing.T) {
	executor := &ScenarioExecutor{World: worldmock.NewMockWorld()}
	tx := &scenmodel.Transaction{
		Type: scenmodel.Transfer,
		From: scenmodel.NewJSONBytesFromString([]byte("missing"), "missing"),
		EGLDValue: scenmodel.JSONBigInt{
			Value: big.NewInt(1),
		},
	}

	require.False(t, executor.senderHasEnoughBalance(tx))
}

package scenexec

import (
	"errors"
	"math"
	"math/big"
	"testing"

	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	worldmock "github.com/multiversx/mx-chain-scenario-go/worldmock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

type spyVM struct {
	runSmartContractCall func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

func (s *spyVM) RunSmartContractCreate(*vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	return nil, errors.New("unexpected create call")
}

func (s *spyVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if s.runSmartContractCall == nil {
		return nil, errors.New("unexpected smart contract call")
	}
	return s.runSmartContractCall(input)
}

func (*spyVM) GasScheduleChange(map[string]map[string]uint64) {}
func (*spyVM) GetVersion() string                             { return "" }
func (*spyVM) IsInterfaceNil() bool                           { return false }
func (*spyVM) Close() error                                   { return nil }
func (*spyVM) Reset()                                         {}
func (*spyVM) SetGasTracing(bool)                             {}
func (*spyVM) GetGasTrace() map[string]map[string][]uint64 {
	return make(map[string]map[string][]uint64)
}

func TestConvertScenarioTxToVMInputWithCallerAndGasUsesExplicitOverrides(t *testing.T) {
	tx := &scenmodel.Transaction{
		From:     scenmodel.NewJSONBytesFromString([]byte("sender"), "sender"),
		To:       scenmodel.NewJSONBytesFromString([]byte("contract"), "contract"),
		Function: "ping",
		EGLDValue: scenmodel.JSONBigInt{
			Value: big.NewInt(7),
		},
		GasPrice: scenmodel.JSONUint64{
			Value: 3,
		},
		GasLimit: scenmodel.JSONUint64{
			Value: 55,
		},
	}

	input := convertScenarioTxToVMInputWithCallerAndGas(tx, tx.To.Value, math.MaxUint64)

	require.Equal(t, tx.To.Value, input.CallerAddr)
	require.Equal(t, uint64(math.MaxUint64), input.GasProvided)
	require.Equal(t, tx.To.Value, input.RecipientAddr)
	require.Equal(t, "ping", input.Function)
	require.Zero(t, input.CallValue.Cmp(big.NewInt(7)))
}

func TestScQueryDoesNotMutateTransactionShape(t *testing.T) {
	world := worldmock.NewMockWorld()
	contract := world.AcctMap.CreateAccount([]byte("contract"), world)
	contract.Code = []byte("code")

	executor := &ScenarioExecutor{
		World: world,
		vm: &spyVM{
			runSmartContractCall: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				require.Equal(t, contract.Address, input.CallerAddr)
				require.Equal(t, uint64(math.MaxInt64), input.GasProvided)
				return &vmcommon.VMOutput{
					ReturnCode:      vmcommon.Ok,
					GasRemaining:    math.MaxInt64,
					GasRefund:       big.NewInt(0),
					OutputAccounts:  map[string]*vmcommon.OutputAccount{},
					DeletedAccounts: make([][]byte, 0),
					TouchedAccounts: make([][]byte, 0),
					Logs:            make([]*vmcommon.LogEntry, 0),
				}, nil
			},
		},
	}

	tx := &scenmodel.Transaction{
		Type:     scenmodel.ScQuery,
		From:     scenmodel.NewJSONBytesFromString([]byte("original-caller"), "original-caller"),
		To:       scenmodel.NewJSONBytesFromString(contract.Address, "contract"),
		Function: "query",
		GasLimit: scenmodel.JSONUint64{
			Value: 77,
		},
		GasPrice: scenmodel.JSONUint64{
			Value: 1,
		},
		EGLDValue: scenmodel.JSONBigInt{
			Value: big.NewInt(0),
		},
	}

	output, err := executor.executeTx("query-no-mutate", tx)

	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, []byte("original-caller"), tx.From.Value)
	require.Equal(t, uint64(77), tx.GasLimit.Value)
}

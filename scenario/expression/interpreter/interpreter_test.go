package scenexpressioninterpreter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterpretMxscJsonRejectsMissingCode(t *testing.T) {
	ei := &ExprInterpreter{}

	result, err := ei.interpretMxscJson([]byte(`{}`))
	require.Equal(t, []byte{}, result)
	require.EqualError(t, err, "mxsc json missing non-empty code field")
}

func TestInterpretMxscJsonRejectsNonStringCode(t *testing.T) {
	ei := &ExprInterpreter{}

	result, err := ei.interpretMxscJson([]byte(`{"code":123}`))
	require.Equal(t, []byte{}, result)
	require.Error(t, err)
}

func TestInterpretMxscJsonRejectsInvalidHex(t *testing.T) {
	ei := &ExprInterpreter{}

	result, err := ei.interpretMxscJson([]byte(`{"code":"not-hex"}`))
	require.Equal(t, []byte{}, result)
	require.Error(t, err)
}

func TestInterpretStringRejectsExcessiveKeccakDepth(t *testing.T) {
	ei := &ExprInterpreter{}
	expr := strings.Repeat("keccak256:", maxExpressionDepth+1) + "0x01"

	result, err := ei.InterpretString(expr)
	require.Equal(t, []byte{}, result)
	require.ErrorContains(t, err, "expression nesting depth exceeded limit (64)")
}

func TestInterpretStringRejectsExcessiveNestedDepth(t *testing.T) {
	ei := &ExprInterpreter{}
	expr := strings.Repeat("nested:", maxExpressionDepth+1) + "0x01"

	result, err := ei.InterpretString(expr)
	require.Nil(t, result)
	require.EqualError(t, err, "expression nesting depth exceeded limit (64)")
}

func TestInterpretStringAllowsReasonableNestedDepth(t *testing.T) {
	ei := &ExprInterpreter{}

	result, err := ei.InterpretString("nested:nested:0x01")
	require.NoError(t, err)
	require.Equal(t, []byte{0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x01, 0x01}, result)
}

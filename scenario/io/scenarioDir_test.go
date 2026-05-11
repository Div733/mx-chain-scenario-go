package scenio

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fr "github.com/multiversx/mx-chain-scenario-go/scenario/expression/fileresolver"
	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
	"github.com/stretchr/testify/require"
)

type scenarioRunnerStub struct {
	resetCount       int
	runScenarioError error
}

func (s *scenarioRunnerStub) Reset() {
	s.resetCount++
}

func (s *scenarioRunnerStub) RunScenario(*scenmodel.Scenario, fr.FileResolver) error {
	return s.runScenarioError
}

func TestIsExcludedReturnsErrorForInvalidPattern(t *testing.T) {
	excluded, err := isExcluded([]string{"["}, "test.scen.json", "root")

	require.False(t, excluded)
	require.Error(t, err)
	require.ErrorContains(t, err, `invalid exclusion pattern "["`)
}

func TestProcessScenarioWalkEntryReturnsWalkError(t *testing.T) {
	controller := NewScenarioController(&scenarioRunnerStub{}, NewDefaultFileResolver(), defaultVMType)

	err := controller.processScenarioWalkEntry(
		"test.scen.json",
		"root",
		".scen.json",
		nil,
		DefaultRunScenarioOptions(),
		errors.New("permission denied"),
		new(int),
		new(int),
		new(int),
	)

	require.EqualError(t, err, "permission denied")
}

func TestProcessScenarioWalkEntryReturnsInvalidPatternError(t *testing.T) {
	controller := NewScenarioController(&scenarioRunnerStub{}, NewDefaultFileResolver(), defaultVMType)

	err := controller.processScenarioWalkEntry(
		"root/test.scen.json",
		"root",
		".scen.json",
		[]string{"["},
		DefaultRunScenarioOptions(),
		nil,
		new(int),
		new(int),
		new(int),
	)

	require.Error(t, err)
	require.ErrorContains(t, err, `invalid exclusion pattern "["`)
}

func TestProcessScenarioWalkEntrySkipsExcludedScenario(t *testing.T) {
	runner := &scenarioRunnerStub{}
	controller := NewScenarioController(runner, NewDefaultFileResolver(), defaultVMType)
	passed := 0
	failed := 0
	skipped := 0

	err := controller.processScenarioWalkEntry(
		"root/scenarios/self.scen.json",
		"root",
		".scen.json",
		[]string{"scenarios/*.scen.json"},
		DefaultRunScenarioOptions(),
		nil,
		&passed,
		&failed,
		&skipped,
	)

	require.NoError(t, err)
	require.Equal(t, 0, runner.resetCount)
	require.Equal(t, 0, passed)
	require.Equal(t, 0, failed)
	require.Equal(t, 1, skipped)
}

func TestWriteScenariosScenarioUsesOwnerOnlyPermissions(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "nested", "scenario.scen.json")
	scenario := &scenmodel.Scenario{}

	err := WriteScenariosScenario(scenario, outputPath)
	require.NoError(t, err)

	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())

	dirInfo, err := os.Stat(filepath.Dir(outputPath))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0), dirInfo.Mode().Perm()&0007)
}

func TestParseScenariosScenarioRejectsOversizedFile(t *testing.T) {
	tempDir := t.TempDir()
	scenPath := filepath.Join(tempDir, "oversized.scen.json")
	oversized := strings.Repeat(" ", int(maxScenarioFileSizeBytes)+1)
	err := os.WriteFile(scenPath, []byte(oversized), 0600)
	require.NoError(t, err)

	_, err = ParseScenariosScenarioDefaultParser(scenPath)
	require.Error(t, err)
	require.ErrorContains(t, err, "exceeds maximum size")
}

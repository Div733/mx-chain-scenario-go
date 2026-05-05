package scenio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	scenjparse "github.com/multiversx/mx-chain-scenario-go/scenario/json/parse"
	scenjwrite "github.com/multiversx/mx-chain-scenario-go/scenario/json/write"
	scenmodel "github.com/multiversx/mx-chain-scenario-go/scenario/model"
)

var defaultVMType = []byte{0, 0}

const maxScenarioFileSizeBytes int64 = 20 * 1024 * 1024

// ParseScenariosScenario reads and parses a Scenarios scenario from a JSON file.
func ParseScenariosScenario(parser scenjparse.Parser, scenFilePath string) (*scenmodel.Scenario, error) {
	var err error
	scenFilePath, err = filepath.Abs(scenFilePath)
	if err != nil {
		return nil, err
	}

	// Open our jsonFile
	var jsonFile *os.File
	jsonFile, err = os.Open(scenFilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, err := io.ReadAll(io.LimitReader(jsonFile, maxScenarioFileSizeBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(byteValue)) > maxScenarioFileSizeBytes {
		return nil, fmt.Errorf("scenario file %s exceeds maximum size of %d bytes", scenFilePath, maxScenarioFileSizeBytes)
	}

	parser.ExprInterpreter.FileResolver.SetContext(scenFilePath)
	return parser.ParseScenarioFile(byteValue)
}

// ParseScenariosScenarioDefaultParser reads and parses a Scenarios scenario from a JSON file.
func ParseScenariosScenarioDefaultParser(scenFilePath string) (*scenmodel.Scenario, error) {
	parser := scenjparse.NewParser(NewDefaultFileResolver(), defaultVMType)
	parser.ExprInterpreter.FileResolver.SetContext(scenFilePath)
	return ParseScenariosScenario(parser, scenFilePath)
}

// WriteScenariosScenario exports a Scenarios scenario to a file, using the default formatting.
func WriteScenariosScenario(scenario *scenmodel.Scenario, toPath string) error {
	jsonString := scenjwrite.ScenarioToJSONString(scenario)

	err := os.MkdirAll(filepath.Dir(toPath), 0750)
	if err != nil {
		return err
	}

	return os.WriteFile(toPath, []byte(jsonString), 0600)
}

package scenio

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TwiN/go-color"
)

// RunAllJSONScenariosInDirectory walks directory, parses and prepares all json scenarios,
// then calls ScenarioRunner for each of them.
func (r *ScenarioController) RunAllJSONScenariosInDirectory(
	generalTestPath string,
	specificTestPath string,
	allowedSuffix string,
	excludedFilePatterns []string,
	options *RunScenarioOptions) error {

	mainDirPath := filepath.Join(generalTestPath, specificTestPath)
	var nrPassed, nrFailed, nrSkipped int

	err := filepath.Walk(mainDirPath, func(testFilePath string, info os.FileInfo, err error) error {
		return r.processScenarioWalkEntry(
			testFilePath,
			generalTestPath,
			allowedSuffix,
			excludedFilePatterns,
			options,
			err,
			&nrPassed,
			&nrFailed,
			&nrSkipped,
		)
	})
	if err != nil {
		return err
	}
	fmt.Printf("Done. Passed: %d. Failed: %d. Skipped: %d.\n", nrPassed, nrFailed, nrSkipped)
	if nrFailed > 0 {
		return errors.New("some tests failed")
	}

	return nil
}

func (r *ScenarioController) processScenarioWalkEntry(
	testFilePath string,
	generalTestPath string,
	allowedSuffix string,
	excludedFilePatterns []string,
	options *RunScenarioOptions,
	walkErr error,
	nrPassed *int,
	nrFailed *int,
	nrSkipped *int,
) error {
	if walkErr != nil {
		return walkErr
	}

	if !strings.HasSuffix(testFilePath, allowedSuffix) {
		return nil
	}

	fmt.Printf("Scenario: %s ... ", shortenTestPath(testFilePath, generalTestPath))
	excluded, err := isExcluded(excludedFilePatterns, testFilePath, generalTestPath)
	if err != nil {
		return err
	}
	if excluded {
		(*nrSkipped)++
		fmt.Printf("  %s\n", color.Ize(color.Yellow, "skip"))
		return nil
	}

	r.Executor.Reset()
	r.RunsNewTest = true
	testErr := r.RunSingleJSONScenario(testFilePath, options)
	if testErr == nil {
		(*nrPassed)++
		fmt.Printf("  %s\n", color.Ize(color.Green, "ok"))
		return nil
	}

	(*nrFailed)++
	fmt.Printf("  %s %s\n", color.Ize(color.Red, "FAIL:"), testErr.Error())
	return nil
}

func isExcluded(excludedFilePatterns []string, testPath string, generalTestPath string) (bool, error) {
	for _, et := range excludedFilePatterns {
		excludedFullPath := filepath.Join(generalTestPath, et)
		match, err := filepath.Match(excludedFullPath, testPath)
		if err != nil {
			return false, fmt.Errorf("invalid exclusion pattern %q: %w", et, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func shortenTestPath(path string, generalTestPath string) string {
	if strings.HasPrefix(path, generalTestPath+"/") {
		return path[len(generalTestPath)+1:]
	}
	return path
}

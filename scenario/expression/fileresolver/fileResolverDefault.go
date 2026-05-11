package scenfileresolver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var _ FileResolver = (*DefaultFileResolver)(nil)
var ErrPathEscapesContext = errors.New("resolved path escapes scenario context")

// DefaultFileResolver loads file contents for the test parser.
type DefaultFileResolver struct {
	contextPath              string
	contractPathReplacements map[string]string
	allowMissingFiles        bool
}

// NewDefaultFileResolver yields a new DefaultFileResolver instance.
func NewDefaultFileResolver() *DefaultFileResolver {
	return &DefaultFileResolver{
		contextPath:              "",
		contractPathReplacements: make(map[string]string),
		allowMissingFiles:        false,
	}
}

// ReplacePath offers the possibility to swap a path with another withouot providing a new set of tests.
// It is very useful when testing multiple contracts against the same tests.
func (fr *DefaultFileResolver) ReplacePath(pathInTest, actualPath string) *DefaultFileResolver {
	fr.contractPathReplacements[pathInTest] = actualPath
	return fr
}

// AllowMissingFiles configures the resolver to not crash when encountering missing files.
func (fr *DefaultFileResolver) AllowMissingFiles() *DefaultFileResolver {
	fr.allowMissingFiles = true
	return fr
}

// WithContext sets directory where the test runs, to help resolve relative paths.
// Unlike SetContext, can be chained in a builder pattern.
func (fr *DefaultFileResolver) WithContext(contextPath string) *DefaultFileResolver {
	fr.contextPath = contextPath
	return fr
}

// Clone creates new instance of the same type.
func (fr *DefaultFileResolver) Clone() FileResolver {
	return &DefaultFileResolver{
		contextPath:              fr.contextPath,
		contractPathReplacements: fr.contractPathReplacements,
	}
}

// SetContext sets directory where the test runs, to help resolve relative paths.
func (fr *DefaultFileResolver) SetContext(contextPath string) {
	fr.contextPath = contextPath
}

// ResolveAbsolutePath yields absolute value based on context.
func (fr *DefaultFileResolver) ResolveAbsolutePath(value string) (string, error) {
	if replacement, shouldReplace := fr.contractPathReplacements[value]; shouldReplace {
		return filepath.Clean(replacement), nil
	}

	testDirPath, err := filepath.Abs(filepath.Clean(filepath.Dir(fr.contextPath)))
	if err != nil {
		return "", err
	}
	resolutionRoot := findResolutionRoot(testDirPath)
	cleanValue := filepath.Clean(value)
	if filepath.IsAbs(cleanValue) {
		return "", fmt.Errorf("%w: %s", ErrPathEscapesContext, value)
	}

	fullPath := filepath.Join(testDirPath, cleanValue)
	relPath, err := filepath.Rel(resolutionRoot, fullPath)
	if err != nil {
		return "", err
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: %s", ErrPathEscapesContext, value)
	}

	return fullPath, nil
}

func findResolutionRoot(startDir string) string {
	current := filepath.Clean(startDir)
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			return startDir
		}
		current = parent
	}
}

// ResolveFileValue converts a value prefixed with "file:" and replaces it with the file contents.
func (fr *DefaultFileResolver) ResolveFileValue(value string) ([]byte, error) {
	if len(value) == 0 {
		return []byte{}, nil
	}
	fullPath, err := fr.ResolveAbsolutePath(value)
	if err != nil {
		return []byte{}, err
	}
	scCode, err := os.ReadFile(fullPath)
	if err != nil {
		if fr.allowMissingFiles {
			return []byte(fmt.Sprintf("MISSING:%s", value)), nil
		}
		return []byte{}, err
	}

	return scCode, nil
}

package scenfileresolver

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveAbsolutePathWithinContext(t *testing.T) {
	contextFile := filepath.Join(t.TempDir(), "scenarios", "root.scen.json")
	resolver := NewDefaultFileResolver().WithContext(contextFile)

	fullPath, err := resolver.ResolveAbsolutePath("nested/test.wasm")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(filepath.Dir(contextFile), "nested", "test.wasm"), fullPath)
}

func TestResolveAbsolutePathRejectsTraversalOutsideContext(t *testing.T) {
	contextFile := filepath.Join(t.TempDir(), "scenarios", "root.scen.json")
	resolver := NewDefaultFileResolver().WithContext(contextFile)

	_, err := resolver.ResolveAbsolutePath("../../etc/passwd")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPathEscapesContext))
}

func TestResolveAbsolutePathAllowsExplicitReplacement(t *testing.T) {
	resolver := NewDefaultFileResolver().ReplacePath("contract.wasm", "../shared/contract.wasm")

	fullPath, err := resolver.ResolveAbsolutePath("contract.wasm")
	require.NoError(t, err)
	require.Equal(t, filepath.Clean("../shared/contract.wasm"), fullPath)
}

func TestResolveFileValueRejectsTraversalOutsideContext(t *testing.T) {
	contextFile := filepath.Join(t.TempDir(), "scenarios", "root.scen.json")
	resolver := NewDefaultFileResolver().WithContext(contextFile)

	_, err := resolver.ResolveFileValue("../../etc/passwd")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPathEscapesContext))
}

func TestResolveFileValueReadsFileWithinContext(t *testing.T) {
	rootDir := t.TempDir()
	contextFile := filepath.Join(rootDir, "scenarios", "root.scen.json")
	targetDir := filepath.Dir(contextFile)
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	targetFile := filepath.Join(targetDir, "contract.wasm")
	require.NoError(t, os.WriteFile(targetFile, []byte("wasm"), 0o600))

	resolver := NewDefaultFileResolver().WithContext(contextFile)

	contents, err := resolver.ResolveFileValue("contract.wasm")
	require.NoError(t, err)
	require.Equal(t, []byte("wasm"), contents)
}

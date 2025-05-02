package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/riptano/iac_generator_cli/pkg/models"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

// TestEnvironment represents a temporary test environment with cleanup function
type TestEnvironment struct {
	BaseDir    string
	OutputDir  string
	FixtureDir string
	ModelsDir  string
	t          *testing.T
}

// NewTestEnvironment creates a new temporary test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create temporary base directory
	baseDir, err := os.MkdirTemp("", "iacgen-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create output and fixture subdirectories
	outputDir := filepath.Join(baseDir, "output")
	fixtureDir := filepath.Join(baseDir, "fixtures")
	modelsDir := filepath.Join(baseDir, "models")

	// Create directories
	dirs := []string{outputDir, fixtureDir, modelsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(baseDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	return &TestEnvironment{
		BaseDir:    baseDir,
		OutputDir:  outputDir,
		FixtureDir: fixtureDir,
		ModelsDir:  modelsDir,
		t:          t,
	}
}

// Cleanup removes the temporary test directory
func (te *TestEnvironment) Cleanup() {
	os.RemoveAll(te.BaseDir)
}

// CreateFixtureFile creates a fixture file with the given content
func (te *TestEnvironment) CreateFixtureFile(name, content string) string {
	te.t.Helper()

	path := filepath.Join(te.FixtureDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		te.t.Fatalf("Failed to create fixture file %s: %v", name, err)
	}
	return path
}

// CreateModelFile creates a serialized model file
func (te *TestEnvironment) CreateModelFile(name string, model *models.InfrastructureModel) string {
	te.t.Helper()

	path := filepath.Join(te.ModelsDir, name)
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		te.t.Fatalf("Failed to marshal model: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		te.t.Fatalf("Failed to create model file %s: %v", name, err)
	}
	return path
}

// CompareFiles compares two files by content
func CompareFiles(t *testing.T, expectedPath, actualPath string) bool {
	t.Helper()

	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected file %s: %v", expectedPath, err)
	}

	actual, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("Failed to read actual file %s: %v", actualPath, err)
	}

	if bytes.Equal(expected, actual) {
		return true
	}

	// Show diff on failure
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(expected), string(actual), false)
	t.Errorf("Files do not match. Diff: %s", dmp.DiffPrettyText(diffs))
	return false
}

// CompareDirStructure compares directory structures
func CompareDirStructure(t *testing.T, expectedDir, actualDir string) bool {
	t.Helper()

	// Get expected files
	expectedFiles, err := getFilesRecursive(expectedDir)
	if err != nil {
		t.Fatalf("Failed to read expected directory: %v", err)
	}

	// Get actual files
	actualFiles, err := getFilesRecursive(actualDir)
	if err != nil {
		t.Fatalf("Failed to read actual directory: %v", err)
	}

	// Compare file lists (case-insensitive sort to match consistently across systems)
	missingFiles := findMissingFiles(expectedFiles, actualFiles)
	extraFiles := findMissingFiles(actualFiles, expectedFiles)

	if len(missingFiles) > 0 || len(extraFiles) > 0 {
		if len(missingFiles) > 0 {
			t.Errorf("Missing expected files: %v", missingFiles)
		}
		if len(extraFiles) > 0 {
			t.Errorf("Unexpected files: %v", extraFiles)
		}
		return false
	}

	return true
}

// CompareDirContents compares both structure and file contents in directories
func CompareDirContents(t *testing.T, expectedDir, actualDir string) bool {
	t.Helper()

	// First check structure
	if !CompareDirStructure(t, expectedDir, actualDir) {
		return false
	}

	// Get expected files
	expectedFiles, err := getFilesRecursive(expectedDir)
	if err != nil {
		t.Fatalf("Failed to read expected directory: %v", err)
	}

	// Compare each file
	allMatch := true
	for _, relativePath := range expectedFiles {
		expectedPath := filepath.Join(expectedDir, relativePath)
		actualPath := filepath.Join(actualDir, relativePath)

		if !CompareFiles(t, expectedPath, actualPath) {
			allMatch = false
			// Continue checking other files
		}
	}

	return allMatch
}

// IsValidJSON checks if a file contains valid JSON
func IsValidJSON(t *testing.T, path string) bool {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Errorf("Invalid JSON in file %s: %v", path, err)
		return false
	}

	return true
}

// IsValidYAML checks if a file contains valid YAML
func IsValidYAML(t *testing.T, path string) bool {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	var obj interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		t.Errorf("Invalid YAML in file %s: %v", path, err)
		return false
	}

	return true
}

// IsValidHCL checks if a file contains valid HCL
func IsValidHCL(t *testing.T, path string) bool {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	// Import the hashicorp/hcl parser and use it to validate HCL
	// For now, we're doing a simpler check without pulling in the dependency

	content := string(data)

	// Check for balanced braces
	braceCount := 0
	for _, r := range content {
		if r == '{' {
			braceCount++
		} else if r == '}' {
			braceCount--
			if braceCount < 0 {
				t.Errorf("Unbalanced braces in HCL file %s", path)
				return false
			}
		}
	}

	if braceCount != 0 {
		t.Errorf("Unbalanced braces in HCL file %s: %d unclosed braces", path, braceCount)
		return false
	}

	// Check for basic HCL syntax elements
	if !strings.Contains(content, "resource") && !strings.Contains(content, "data") && 
	   !strings.Contains(content, "variable") && !strings.Contains(content, "provider") &&
	   !strings.Contains(content, "module") && !strings.Contains(content, "output") {
		t.Errorf("No HCL resource blocks found in file %s", path)
		return false
	}

	return true
}

// Helper functions

// getFilesRecursive gets all files in a directory recursively
func getFilesRecursive(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Get path relative to base directory
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// findMissingFiles finds files in expected that are not in actual
func findMissingFiles(expected, actual []string) []string {
	var missing []string

	for _, expFile := range expected {
		found := false
		for _, actFile := range actual {
			if expFile == actFile {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, expFile)
		}
	}

	return missing
}

// LoadFileContent loads a file's content as a string
func LoadFileContent(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return string(data)
}

// FindFilesByPattern finds files matching a glob pattern
func FindFilesByPattern(t *testing.T, baseDir, pattern string) []string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(baseDir, pattern))
	if err != nil {
		t.Fatalf("Failed to find files with pattern %s: %v", pattern, err)
	}

	return matches
}

// ContainsString checks if a string slice contains a specific string
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// DirectoryExists checks if a directory exists
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// CountFiles counts files in a directory that match a pattern
func CountFiles(t *testing.T, dir, pattern string) int {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		t.Fatalf("Failed to count files with pattern %s: %v", pattern, err)
	}

	return len(matches)
}

// PrintFileTree prints a directory tree for debugging purposes
func PrintFileTree(t *testing.T, dir string) {
	t.Helper()

	var walk func(string, string, int)
	walk = func(root, dir string, depth int) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("Failed to read directory %s: %v", dir, err)
		}

		indent := strings.Repeat("  ", depth)
		for _, entry := range entries {
			rel, _ := filepath.Rel(root, filepath.Join(dir, entry.Name()))
			if entry.IsDir() {
				fmt.Printf("%s%s/\n", indent, rel)
				walk(root, filepath.Join(dir, entry.Name()), depth+1)
			} else {
				fmt.Printf("%s%s\n", indent, rel)
			}
		}
	}

	fmt.Printf("Directory tree for %s:\n", dir)
	walk(dir, dir, 0)
}

package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MockFS is a mock filesystem for testing
type MockFS struct {
	files map[string][]byte
	dirs  map[string]bool
	mu    sync.RWMutex
}

// NewMockFS creates a new mock filesystem
func NewMockFS() *MockFS {
	return &MockFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// WriteFile writes a file to the mock filesystem
func (fs *MockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if !fs.dirExistsLocked(dir) {
		fs.createDirLocked(dir)
	}

	fs.files[path] = data
	return nil
}

// ReadFile reads a file from the mock filesystem
func (fs *MockFS) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, ok := fs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

// MkdirAll creates a directory in the mock filesystem
func (fs *MockFS) MkdirAll(path string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.createDirLocked(path)
	return nil
}

// RemoveAll removes a file or directory from the mock filesystem
func (fs *MockFS) RemoveAll(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Remove all files with this path prefix
	for filePath := range fs.files {
		if filePath == path || strings.HasPrefix(filePath, path+"/") {
			delete(fs.files, filePath)
		}
	}

	// Remove all directories with this path prefix
	for dirPath := range fs.dirs {
		if dirPath == path || strings.HasPrefix(dirPath, path+"/") {
			delete(fs.dirs, dirPath)
		}
	}

	return nil
}

// Stat returns file info for a file or directory
func (fs *MockFS) Stat(path string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Check if it's a directory
	if fs.dirExistsLocked(path) {
		return &mockFileInfo{
			name:  filepath.Base(path),
			size:  0,
			mode:  os.ModeDir | 0755,
			isDir: true,
		}, nil
	}

	// Check if it's a file
	data, ok := fs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &mockFileInfo{
		name:  filepath.Base(path),
		size:  int64(len(data)),
		mode:  0644,
		isDir: false,
	}, nil
}

// FileExists checks if a file exists
func (fs *MockFS) FileExists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	_, ok := fs.files[path]
	return ok
}

// DirExists checks if a directory exists
func (fs *MockFS) DirExists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.dirExistsLocked(path)
}

// Glob finds files matching a pattern
func (fs *MockFS) Glob(pattern string) ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var matches []string

	// For simplicity, just handle the common case of '*' and '**'
	if strings.Contains(pattern, "**") {
		// Deep recursive search
		prefix := strings.Split(pattern, "**")[0]
		suffix := strings.Split(pattern, "**")[1]
		for path := range fs.files {
			if strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) {
				matches = append(matches, path)
			}
		}
	} else if strings.Contains(pattern, "*") {
		// Simple wildcard
		dir := filepath.Dir(pattern)
		for path := range fs.files {
			if filepath.Dir(path) == dir {
				matched, _ := filepath.Match(pattern, path)
				if matched {
					matches = append(matches, path)
				}
			}
		}
	} else {
		// Exact match
		if fs.FileExists(pattern) {
			matches = append(matches, pattern)
		}
	}

	sort.Strings(matches)
	return matches, nil
}

// ListFiles lists files in a directory
func (fs *MockFS) ListFiles(dir string) ([]string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var files []string
	for path := range fs.files {
		if filepath.Dir(path) == dir {
			files = append(files, filepath.Base(path))
		}
	}

	var dirs []string
	for path := range fs.dirs {
		parent := filepath.Dir(path)
		if parent == dir && path != dir {
			dirs = append(dirs, filepath.Base(path))
		}
	}

	results := append(files, dirs...)
	sort.Strings(results)
	return results, nil
}

// Open opens a file for reading
func (fs *MockFS) Open(path string) (io.ReadCloser, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, ok := fs.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Create creates a file for writing
func (fs *MockFS) Create(path string) (io.WriteCloser, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if !fs.DirExists(dir) {
		fs.MkdirAll(dir, 0755)
	}

	return &mockFileWriter{
		fs:   fs,
		path: path,
		buf:  bytes.NewBuffer(nil),
	}, nil
}

// Internal helper methods

func (fs *MockFS) dirExistsLocked(path string) bool {
	// Root directory always exists
	if path == "/" || path == "." || path == "" {
		return true
	}

	if _, ok := fs.dirs[path]; ok {
		return true
	}

	// Check if any files have this directory as a prefix
	for filePath := range fs.files {
		dir := filepath.Dir(filePath)
		if dir == path {
			return true
		}
	}

	return false
}

func (fs *MockFS) createDirLocked(path string) {
	// Ensure parent directories exist
	parent := filepath.Dir(path)
	if parent != path && parent != "/" && parent != "." && parent != "" {
		fs.createDirLocked(parent)
	}

	fs.dirs[path] = true
}

// mockFileInfo implements os.FileInfo
type mockFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// mockFileWriter implements io.WriteCloser
type mockFileWriter struct {
	fs   *MockFS
	path string
	buf  *bytes.Buffer
}

func (w *mockFileWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func (w *mockFileWriter) Close() error {
	return w.fs.WriteFile(w.path, w.buf.Bytes(), 0644)
}

// ToMap returns a map representation of the mock filesystem
func (fs *MockFS) ToMap() map[string]string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make(map[string]string)
	for path, data := range fs.files {
		result[path] = string(data)
	}
	return result
}

// FromMap loads the mock filesystem from a map
func (fs *MockFS) FromMap(files map[string]string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.files = make(map[string][]byte)
	fs.dirs = make(map[string]bool)

	for path, content := range files {
		fs.files[path] = []byte(content)
		dir := filepath.Dir(path)
		fs.createDirLocked(dir)
	}
}

// Debug prints the filesystem contents for debugging
func (fs *MockFS) Debug() string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString("MockFS Contents:\n")
	buf.WriteString("-- Files --\n")
	
	var paths []string
	for path := range fs.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		data := fs.files[path]
		buf.WriteString(fmt.Sprintf("%s (%d bytes)\n", path, len(data)))
	}

	buf.WriteString("-- Directories --\n")
	var dirs []string
	for dir := range fs.dirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		buf.WriteString(fmt.Sprintf("%s/\n", dir))
	}

	return buf.String()
}
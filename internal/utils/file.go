package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteToFile writes content to a file, creating the file and directories if they don't exist
func WriteToFile(path string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write content to file
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil
}

// WriteToOutputFile determines the right filename and writes content to the output directory
func WriteToOutputFile(content, outputFormat, outputDir, outputFile string) (string, error) {
	// Generate a default output filename if one is not provided
	if outputFile == "" {
		if outputFormat == "terraform" {
			outputFile = "main.tf"
		} else if outputFormat == "crossplane" {
			outputFile = "resources.yaml"
		} else {
			outputFile = "output.txt" // Generic fallback
		}
	}

	// Create the full output path
	outputPath := filepath.Join(outputDir, outputFile)

	// Check if the output path is a directory
	fileInfo, err := os.Stat(outputDir)
	if err == nil && fileInfo.IsDir() {
		// Ensure output directory exists
		if err := EnsureDirectoryExists(outputDir); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", outputDir, err)
		}
	}

	// Write to the file
	if err := WriteToFile(outputPath, content); err != nil {
		return "", err
	}

	return outputPath, nil
}

// ReadFromFile reads content from a file
func ReadFromFile(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	// Read content from file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read from file %s: %w", path, err)
	}

	return string(content), nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDirectoryExists ensures that a directory exists, creating it if necessary
func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// IsFileWritable checks if a file is writable
func IsFileWritable(filename string) error {
	// Check if file exists
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, check if directory is writable
			return EnsureDirectoryExists(filepath.Dir(filename))
		}
		return fmt.Errorf("cannot access file: %w", err)
	}
	
	// Check if it's a regular file
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}
	
	// Try to open the file for writing
	file, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied to write to file")
		}
		return fmt.Errorf("cannot write to file: %w", err)
	}
	
	file.Close()
	return nil
}
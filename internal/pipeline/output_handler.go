package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"go.uber.org/zap"
)

// OutputHandlerImpl is the implementation of the OutputHandler interface
type OutputHandlerImpl struct {
	// defaultDir is the default directory for output files
	defaultDir string
	logger     *zap.SugaredLogger
}

// NewOutputHandler creates a new output handler
func NewOutputHandler(defaultDir string) *OutputHandlerImpl {
	return &OutputHandlerImpl{
		defaultDir: defaultDir,
		logger:     utils.GetLogger(),
	}
}

// WriteManifest implements OutputHandler
func (h *OutputHandlerImpl) WriteManifest(ctx context.Context, manifest string, format string, outputPath string) (string, error) {
	h.logger.Debugw("Writing manifest",
		"format", format,
		"output_path", outputPath,
		"manifest_length", len(manifest),
	)

	// Check if the context is canceled
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// If no output path is provided, use the default directory
	if outputPath == "" {
		// Determine default filename based on format
		if strings.HasSuffix(format, "terraform") {
			outputPath = filepath.Join(h.defaultDir, "main.tf")
		} else {
			outputPath = filepath.Join(h.defaultDir, "resources.yaml")
		}
	}

	// Check if outputPath is a directory
	fileInfo, err := os.Stat(outputPath)
	if err == nil && fileInfo.IsDir() {
		// If it's a directory, append default filename
		if strings.HasSuffix(format, "terraform") {
			outputPath = filepath.Join(outputPath, "main.tf")
		} else {
			outputPath = filepath.Join(outputPath, "resources.yaml")
		}
		h.logger.Debugw("Adjusted output path for directory",
			"new_path", outputPath,
		)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := utils.EnsureDirectoryExists(dir); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := utils.WriteToFile(outputPath, manifest); err != nil {
		return "", fmt.Errorf("failed to write manifest to %s: %w", outputPath, err)
	}

	h.logger.Infow("Manifest written successfully", "path", outputPath)

	return outputPath, nil
}

// GetOutputWriter implements OutputHandler
func (h *OutputHandlerImpl) GetOutputWriter(outputPath string) (io.Writer, error) {
	// If no output path is provided, return stdout
	if outputPath == "" {
		return os.Stdout, nil
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := utils.EnsureDirectoryExists(dir); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file for writing
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}

	return file, nil
}

// WriteOutputStage creates a pipeline stage that writes the output
func (h *OutputHandlerImpl) WriteOutputStage(outputPath string) Stage {
	return NewBaseStage("OutputWriting", func(ctx context.Context, input interface{}) (interface{}, error) {
		var manifest string
		switch v := input.(type) {
		case string:
			manifest = v
		default:
			return nil, fmt.Errorf("invalid input type for output writing: %T", input)
		}

		// Determine format based on output path or extension
		format := "terraform" // Default format
		
		// Check if directory exists
		fileInfo, err := os.Stat(outputPath)
		if err == nil && fileInfo.IsDir() {
			// For directories, we need to infer format from expected content
			// Default remains terraform unless specified otherwise
		} else {
			// Get the file extension to determine the format
			ext := filepath.Ext(outputPath)
			if ext == ".yaml" || ext == ".yml" {
				format = "crossplane"
			}
		}

		return h.WriteManifest(ctx, manifest, format, outputPath)
	})
}
package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"go.uber.org/zap"
)

// CLI provides methods for integrating the pipeline with command-line interfaces
type CLI struct {
	coordinator PipelineCoordinator
	logger      *zap.SugaredLogger
}


// NewCLI creates a new CLI helper
func NewCLI() *CLI {
	return &CLI{
		coordinator: NewPipelineCoordinator(),
		logger:      utils.GetLogger(),
	}
}

// GenerateFromCLI processes a description from CLI inputs and generates manifests
func (c *CLI) GenerateFromCLI(description, inputFile, outputFormat, outputDir, outputFile, region string, useTemplates bool, debug bool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Configure processing parameters
	params := &ProcessingParams{
		Description:    description,
		InputFile:      inputFile,
		OutputFormat:   outputFormat,
		OutputDir:      outputDir,
		OutputFile:     outputFile,
		Region:         region,
		UseTemplates:   useTemplates,
		Debug:          debug,
		ProgressWriter: os.Stdout,
	}

	// Load description from file if provided
	if inputFile != "" && description == "" {
		var err error
		description, err = utils.ReadFromFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("failed to read input file: %w", err)
		}
		params.Description = description
	}

	// Generate output filename based on input if not specified
	if outputFile == "" && inputFile != "" {
		baseName := filepath.Base(inputFile)
		ext := filepath.Ext(baseName)
		baseName = baseName[:len(baseName)-len(ext)]
		
		if outputFormat == "terraform" {
			params.OutputFile = baseName + ".tf"
		} else {
			params.OutputFile = baseName + ".yaml"
		}
	} else if outputFile == "" {
		if outputFormat == "terraform" {
			params.OutputFile = "main.tf"
		} else {
			params.OutputFile = "resources.yaml"
		}
	}

	// Initialize the pipeline
	if err := c.coordinator.InitializePipeline(ctx, params); err != nil {
		return "", err
	}

	// Run the pipeline
	return c.coordinator.RunPipeline(ctx, params)
}

// GetAvailableFormats returns the list of available output formats
func (c *CLI) GetAvailableFormats() []string {
	return c.coordinator.GetAvailableGenerators()
}

// ProcessCLI provides a simple wrapper to process CLI arguments
func ProcessCLI(description, inputFile, outputFormat, outputDir, outputFile, region string, useTemplates bool, debug bool) (string, error) {
	cli := NewCLI()
	return cli.GenerateFromCLI(description, inputFile, outputFormat, outputDir, outputFile, region, useTemplates, debug)
}

// RunWithProgressFeedback runs the pipeline with progress feedback in the terminal
func RunWithProgressFeedback(params *ProcessingParams, outputWriter io.Writer) (string, error) {
	// Create and configure coordinator
	coordinator := NewPipelineCoordinator()
	
	// Set a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	// Configure progress reporting
	params.ProgressWriter = outputWriter
	
	// Initialize the pipeline
	if err := coordinator.InitializePipeline(ctx, params); err != nil {
		return "", err
	}
	
	// Set up progress reporter output handling
	reporter, ok := coordinator.progressReporter.(*ConsoleProgressReporter)
	if ok {
		// Start goroutine to forward progress messages to the output writer
		go func() {
			for msg := range reporter.OutputChannel() {
				fmt.Fprintf(outputWriter, "  → %s\n", msg)
			}
		}()
	}
	
	// Print initial message
	fmt.Fprintln(outputWriter, "Starting IaC generation pipeline...")
	fmt.Fprintln(outputWriter, "-----------------------------------")
	
	// Print configuration information
	outputFormat := params.OutputFormat
	fmt.Fprintf(outputWriter, "  → Generating %s code for your infrastructure\n", outputFormat)
	
	// Run the pipeline
	result, err := coordinator.RunPipeline(ctx, params)
	
	// Clean up reporter
	if ok {
		reporter.Close()
	}
	
	// Print completion message
	fmt.Fprintln(outputWriter, "-----------------------------------")
	if err == nil {
		fmt.Fprintln(outputWriter, "✅ Pipeline execution completed successfully")
		// Add message about generated files if output directory was specified
		if params.OutputDir != "." {
			if params.OutputFormat == "terraform" {
				fmt.Fprintf(outputWriter, "   Generated Terraform files in: %s\n", params.OutputDir)
			} else if params.OutputFormat == "crossplane" {
				fmt.Fprintf(outputWriter, "   Generated Crossplane manifests in: %s\n", params.OutputDir)
			}
		}
	} else {
		fmt.Fprintf(outputWriter, "❌ Pipeline execution failed: %v\n", err)
	}
	
	return result, err
}
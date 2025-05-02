package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"go.uber.org/zap"
)

// PipelineCoordinatorImpl is the implementation of the PipelineCoordinator interface
type PipelineCoordinatorImpl struct {
	// pipeline is the core pipeline
	pipeline Pipeline
	// nlpProcessor processes natural language input
	nlpProcessor NLPProcessor
	// modelBuilder builds infrastructure models
	modelBuilder ModelBuilder
	// generators is a map of available generators
	generators map[string]IaCGenerator
	// outputHandler handles writing output
	outputHandler OutputHandler
	// progressReporter reports progress
	progressReporter ProgressReporter
	logger           *zap.SugaredLogger
}

// NewPipelineCoordinator creates a new pipeline coordinator
func NewPipelineCoordinator() *PipelineCoordinatorImpl {
	return &PipelineCoordinatorImpl{
		pipeline:      NewBasePipeline(),
		generators:    make(map[string]IaCGenerator),
		logger:        utils.GetLogger(),
	}
}

// InitializePipeline implements PipelineCoordinator
func (c *PipelineCoordinatorImpl) InitializePipeline(ctx context.Context, params *ProcessingParams) error {
	c.logger.Debugw("Initializing pipeline", 
		"output_format", params.OutputFormat,
		"output_dir", params.OutputDir,
		"region", params.Region,
	)

	// Initialize NLP processor
	c.nlpProcessor = NewNLPProcessor()

	// Initialize model builder with the specified region
	c.modelBuilder = NewModelBuilder(params.Region)

	// Initialize output handler
	c.outputHandler = NewOutputHandler(params.OutputDir)

	// Initialize generators
	c.generators = make(map[string]IaCGenerator)
	for _, format := range GetAvailableGenerators() {
		generator := NewIaCGenerator(format, params.UseTemplates)
		generator.OutputDir = params.OutputDir
		c.generators[format] = generator
	}

	// Create progress reporter
	totalSteps := 3 // NLP, Model Building, Generation
	if params.OutputDir != "." || params.OutputFile != "" {
		totalSteps++ // Add output writing step
	}
	c.progressReporter = NewConsoleProgressReporter(totalSteps)

	// Set progress reporter on pipeline
	c.pipeline.SetProgressReporter(c.progressReporter)

	// Validate required parameters
	if err := c.validateParams(params); err != nil {
		return err
	}

	c.logger.Info("Pipeline initialized successfully")
	return nil
}

// validateParams validates the processing parameters
func (c *PipelineCoordinatorImpl) validateParams(params *ProcessingParams) error {
	// Validate description or input file
	if params.Description == "" && params.InputFile == "" {
		return fmt.Errorf("either description or input file must be provided")
	}

	// Validate output format
	format := strings.ToLower(params.OutputFormat)
	valid := false
	for _, f := range GetAvailableGenerators() {
		if f == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unsupported output format: %s", params.OutputFormat)
	}

	// If input file is specified, check if it exists
	if params.InputFile != "" {
		if !utils.FileExists(params.InputFile) {
			return fmt.Errorf("input file does not exist: %s", params.InputFile)
		}
	}

	// If output directory is specified, ensure it can be created
	if params.OutputDir != "." {
		if err := utils.EnsureDirectoryExists(params.OutputDir); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	return nil
}

// setupPipeline sets up the pipeline stages based on parameters
func (c *PipelineCoordinatorImpl) setupPipeline(params *ProcessingParams) error {
	// Clear any existing stages
	c.pipeline = NewBasePipeline()
	c.pipeline.SetProgressReporter(c.progressReporter)

	// Add NLP processing stage
	c.pipeline.AddStage(c.nlpProcessor.ProcessStage())

	// Add model building stage
	c.pipeline.AddStage(c.modelBuilder.ModelBuildStage())

	// Add IaC generation stage
	generator, found := c.generators[strings.ToLower(params.OutputFormat)]
	if !found {
		return fmt.Errorf("no generator available for format: %s", params.OutputFormat)
	}
	c.pipeline.AddStage(generator.GenerateStage())

	// If output path is specified, add output writing stage
	if params.OutputDir != "." || params.OutputFile != "" {
		// Determine output path
		var outputPath string
		if params.OutputFile != "" {
			// If a specific output file is provided, use it
			outputPath = filepath.Join(params.OutputDir, params.OutputFile)
		} else {
			// When only directory is specified, handle based on format
			if strings.ToLower(params.OutputFormat) == "terraform" {
				outputPath = filepath.Join(params.OutputDir, "main.tf")
			} else if strings.ToLower(params.OutputFormat) == "crossplane" {
				outputPath = filepath.Join(params.OutputDir, "resources.yaml")
			} else {
				// Generic fallback
				outputPath = filepath.Join(params.OutputDir, "output.txt")
			}
		}
		
		// Check if outputPath is a directory and add the appropriate filename
		fileInfo, err := os.Stat(outputPath)
		if err == nil && fileInfo.IsDir() {
			// Path exists and is a directory, append default filename
			if strings.ToLower(params.OutputFormat) == "terraform" {
				outputPath = filepath.Join(outputPath, "main.tf")
			} else if strings.ToLower(params.OutputFormat) == "crossplane" {
				outputPath = filepath.Join(outputPath, "resources.yaml")
			} else {
				outputPath = filepath.Join(outputPath, "output.txt")
			}
		}
		
		c.logger.Debugw("Setting up output stage", "path", outputPath)
		c.pipeline.AddStage(c.outputHandler.WriteOutputStage(outputPath))
	}

	return nil
}

// loadDescription loads the description from parameters
func (c *PipelineCoordinatorImpl) loadDescription(params *ProcessingParams) (string, error) {
	// If description is provided directly, use it
	if params.Description != "" {
		return params.Description, nil
	}

	// Otherwise, try to load from file
	if params.InputFile != "" {
		description, err := utils.ReadFromFile(params.InputFile)
		if err != nil {
			return "", fmt.Errorf("failed to read input file: %w", err)
		}
		return strings.TrimSpace(description), nil
	}

	return "", fmt.Errorf("no description provided")
}

// RunPipeline implements PipelineCoordinator
func (c *PipelineCoordinatorImpl) RunPipeline(ctx context.Context, params *ProcessingParams) (string, error) {
	c.logger.Info("Running pipeline")

	// Set up the pipeline
	if err := c.setupPipeline(params); err != nil {
		return "", fmt.Errorf("failed to set up pipeline: %w", err)
	}

	// Load description
	description, err := c.loadDescription(params)
	if err != nil {
		return "", fmt.Errorf("failed to load description: %w", err)
	}

	// Execute the pipeline
	result, err := c.pipeline.Execute(ctx, description)
	if err != nil {
		return "", fmt.Errorf("pipeline execution failed: %w", err)
	}

	// Handle the result based on its type
	switch v := result.(type) {
	case string:
		// For stdout output, return the manifest
		if params.OutputDir == "." && params.OutputFile == "" {
			return v, nil
		}
		// For file output, return the file path
		return fmt.Sprintf("Successfully generated %s manifest", params.OutputFormat), nil
	case *models.InfrastructureModel:
		// If we stopped at the model stage, generate a summary
		return fmt.Sprintf("Generated model with %d resources", len(v.Resources)), nil
	default:
		return fmt.Sprintf("Pipeline completed with result type: %T", result), nil
	}
}

// GetAvailableGenerators implements PipelineCoordinator
func (c *PipelineCoordinatorImpl) GetAvailableGenerators() []string {
	return GetAvailableGenerators()
}

// ProcessDescriptionToPipeline processes a description through the full pipeline
func ProcessDescriptionToPipeline(description, outputFormat, outputDir, outputFile, region string, useTemplates bool, progressWriter io.Writer) (string, error) {
	// Create processing parameters
	params := &ProcessingParams{
		Description:    description,
		OutputFormat:   outputFormat,
		OutputDir:      outputDir,
		OutputFile:     outputFile,
		Region:         region,
		UseTemplates:   useTemplates,
		ProgressWriter: progressWriter,
	}

	// Create and initialize pipeline coordinator
	coordinator := NewPipelineCoordinator()
	ctx := context.Background()

	if err := coordinator.InitializePipeline(ctx, params); err != nil {
		return "", err
	}

	// Run the pipeline
	return coordinator.RunPipeline(ctx, params)
}

// ProcessPipeline processes input through the pipeline and writes to the specified output
func ProcessPipeline(params *ProcessingParams) (string, error) {
	// Set default progress writer if not provided
	if params.ProgressWriter == nil {
		params.ProgressWriter = os.Stdout
	}

	// Create and initialize pipeline coordinator
	coordinator := NewPipelineCoordinator()
	ctx := context.Background()

	if err := coordinator.InitializePipeline(ctx, params); err != nil {
		return "", err
	}

	// Run the pipeline
	return coordinator.RunPipeline(ctx, params)
}
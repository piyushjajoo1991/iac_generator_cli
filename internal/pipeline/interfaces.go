package pipeline

import (
	"context"
	"io"

	"github.com/riptano/iac_generator_cli/pkg/models"
)

// Stage represents a pipeline stage that processes data
// and potentially transforms it for the next stage
type Stage interface {
	// Execute runs the stage with the given input and returns the output
	// or an error if the stage fails
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Name returns the name of the stage for logging and debugging
	Name() string
}

// NLPProcessor defines the interface for natural language processing
type NLPProcessor interface {
	// ParseDescription parses a natural language description and extracts
	// infrastructure entities
	ParseDescription(ctx context.Context, description string) (*models.InfrastructureModel, error)

	// ValidateDescription checks if a description contains enough information
	// to build a valid infrastructure model
	ValidateDescription(description string) (bool, string)
	
	// ProcessStage returns a Stage implementation for NLP processing
	ProcessStage() Stage
}

// ModelBuilder defines the interface for building infrastructure models
type ModelBuilder interface {
	// BuildModel constructs an infrastructure model from parsed entities or raw description
	BuildModel(ctx context.Context, input interface{}) (*models.InfrastructureModel, error)

	// EnhanceModel adds additional information or relationships to the model
	EnhanceModel(model *models.InfrastructureModel) (*models.InfrastructureModel, error)
	
	// ModelBuildStage returns a Stage implementation for model building
	ModelBuildStage() Stage
}

// IaCGenerator defines the interface for generating IaC manifests
type IaCGenerator interface {
	// Generate produces an IaC manifest from an infrastructure model
	Generate(ctx context.Context, model *models.InfrastructureModel) (string, error)

	// WriteOutput writes the generated manifest to the specified output
	WriteOutput(ctx context.Context, manifest string, output io.Writer) error

	// CanGenerate checks if this generator can handle the requested format
	CanGenerate(format string) bool
	
	// GenerateStage returns a Stage implementation for IaC generation
	GenerateStage() Stage
}

// OutputHandler defines the interface for managing output destinations
type OutputHandler interface {
	// WriteManifest writes a manifest to the appropriate destination
	WriteManifest(ctx context.Context, manifest string, format string, outputPath string) (string, error)

	// GetOutputWriter returns a writer for the specified output path
	GetOutputWriter(outputPath string) (io.Writer, error)
	
	// WriteOutputStage returns a Stage implementation for output handling
	WriteOutputStage(outputPath string) Stage
}

// ProgressReporter defines the interface for reporting progress
type ProgressReporter interface {
	// StartStage reports that a stage has started
	StartStage(stageName string)

	// CompleteStage reports that a stage has completed
	CompleteStage(stageName string)

	// FailStage reports that a stage has failed
	FailStage(stageName string, err error)

	// UpdateProgress updates the current progress
	UpdateProgress(message string, percentage int)
}

// Pipeline defines the overall pipeline interface
type Pipeline interface {
	// Execute runs the entire pipeline with the given input
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// AddStage adds a stage to the pipeline
	AddStage(stage Stage)

	// SetErrorHandler sets a custom error handler for the pipeline
	SetErrorHandler(handler func(error) error)

	// SetProgressReporter sets a progress reporter for the pipeline
	SetProgressReporter(reporter ProgressReporter)
}

// PipelineCoordinator orchestrates the execution of the IaC generation pipeline
type PipelineCoordinator interface {
	// InitializePipeline sets up the pipeline based on the input parameters
	InitializePipeline(ctx context.Context, params *ProcessingParams) error

	// RunPipeline executes the pipeline with the given parameters
	RunPipeline(ctx context.Context, params *ProcessingParams) (string, error)

	// GetAvailableGenerators returns the list of available IaC generators
	GetAvailableGenerators() []string
}

// ProcessingParams contains all parameters needed for pipeline execution
type ProcessingParams struct {
	// Description is the natural language description of the infrastructure
	Description string

	// InputFile is the path to a file containing the description
	InputFile string

	// OutputFormat is the desired output format (terraform, crossplane, etc.)
	OutputFormat string

	// OutputDir is the directory where output files should be written
	OutputDir string

	// OutputFile is the name of the output file
	OutputFile string

	// Region is the AWS region to use for the resources
	Region string

	// UseTemplates indicates whether to use the template system
	UseTemplates bool

	// Debug enables debug logging
	Debug bool

	// ProgressWriter is where progress updates are written
	ProgressWriter io.Writer
}
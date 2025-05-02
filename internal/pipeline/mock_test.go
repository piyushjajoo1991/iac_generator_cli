package pipeline

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/riptano/iac_generator_cli/pkg/models"
)

// MockNLPProcessor is a mock implementation of NLPProcessor for testing
type MockNLPProcessor struct {
	// SucceedWithDescriptions defines which descriptions should succeed
	SucceedWithDescriptions []string
	// ModelToReturn is the model to return for successful parsing
	ModelToReturn *models.InfrastructureModel
}

// ParseDescription implements NLPProcessor
func (m *MockNLPProcessor) ParseDescription(ctx context.Context, description string) (*models.InfrastructureModel, error) {
	for _, validDesc := range m.SucceedWithDescriptions {
		if strings.Contains(description, validDesc) {
			if m.ModelToReturn != nil {
				return m.ModelToReturn, nil
			}
			
			// Create a default model
			model := models.NewInfrastructureModel()
			model.AddResource(models.NewResource(models.ResourceVPC, "test-vpc"))
			return model, nil
		}
	}
	
	return nil, fmt.Errorf("mock parsing failed for description: %s", description)
}

// ValidateDescription implements NLPProcessor
func (m *MockNLPProcessor) ValidateDescription(description string) (bool, string) {
	// Empty description is always invalid
	if strings.TrimSpace(description) == "" {
		return false, "description is empty"
	}
	
	// Short description is invalid
	if len(description) < 10 {
		return false, "description is too short"
	}
	
	// Check against succeeded descriptions
	for _, validDesc := range m.SucceedWithDescriptions {
		if strings.Contains(description, validDesc) {
			return true, ""
		}
	}
	
	return false, "description doesn't contain required terms"
}

// ProcessStage implements NLPProcessor
func (m *MockNLPProcessor) ProcessStage() Stage {
	return NewBaseStage("MockNLPProcessing", func(ctx context.Context, input interface{}) (interface{}, error) {
		var description string
		switch v := input.(type) {
		case string:
			description = v
		case *ProcessingParams:
			description = v.Description
		default:
			return nil, fmt.Errorf("invalid input type for NLP processing: %T", input)
		}
		
		return m.ParseDescription(ctx, description)
	})
}

// MockModelBuilder is a mock implementation of ModelBuilder for testing
type MockModelBuilder struct {
	// SucceedWithModels defines which models should succeed
	SucceedWithModels []*models.InfrastructureModel
	// ModelToReturn is the model to return for successful building
	ModelToReturn *models.InfrastructureModel
}

// BuildModel implements ModelBuilder
func (m *MockModelBuilder) BuildModel(ctx context.Context, input interface{}) (*models.InfrastructureModel, error) {
	switch v := input.(type) {
	case *models.InfrastructureModel:
		for _, validModel := range m.SucceedWithModels {
			if v == validModel {
				if m.ModelToReturn != nil {
					return m.ModelToReturn, nil
				}
				return v, nil
			}
		}
		
		// If we have no specific models to match, succeed with any model
		if len(m.SucceedWithModels) == 0 {
			if m.ModelToReturn != nil {
				return m.ModelToReturn, nil
			}
			return v, nil
		}
		
		return nil, fmt.Errorf("mock building failed for model")
		
	case map[string]interface{}:
		// Assume we can build from this and return a default model
		model := models.NewInfrastructureModel()
		model.AddResource(models.NewResource(models.ResourceVPC, "test-vpc"))
		if m.ModelToReturn != nil {
			return m.ModelToReturn, nil
		}
		return model, nil
		
	default:
		return nil, fmt.Errorf("invalid input type for model building: %T", input)
	}
}

// EnhanceModel implements ModelBuilder
func (m *MockModelBuilder) EnhanceModel(model *models.InfrastructureModel) (*models.InfrastructureModel, error) {
	// Just return the same model for mocking
	return model, nil
}

// ModelBuildStage implements ModelBuilder
func (m *MockModelBuilder) ModelBuildStage() Stage {
	return NewBaseStage("MockModelBuilding", func(ctx context.Context, input interface{}) (interface{}, error) {
		return m.BuildModel(ctx, input)
	})
}

// MockIaCGenerator is a mock implementation of IaCGenerator for testing
type MockIaCGenerator struct {
	// Format is the format this generator handles
	Format string
	// ManifestToReturn is the manifest to return for successful generation
	ManifestToReturn string
	// ShouldFail indicates whether the generator should fail
	ShouldFail bool
}

// Generate implements IaCGenerator
func (m *MockIaCGenerator) Generate(ctx context.Context, model *models.InfrastructureModel) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock generation failed for format: %s", m.Format)
	}
	
	if m.ManifestToReturn != "" {
		return m.ManifestToReturn, nil
	}
	
	// Generate some basic content based on model
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Mock %s Manifest\n\n", m.Format))
	
	for _, resource := range model.Resources {
		sb.WriteString(fmt.Sprintf("Resource: %s (%s)\n", resource.Name, resource.Type))
		for _, prop := range resource.Properties {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", prop.Name, prop.Value))
		}
		sb.WriteString("\n")
	}
	
	return sb.String(), nil
}

// WriteOutput implements IaCGenerator
func (m *MockIaCGenerator) WriteOutput(ctx context.Context, manifest string, output io.Writer) error {
	if m.ShouldFail {
		return fmt.Errorf("mock output writing failed for format: %s", m.Format)
	}
	
	_, err := io.WriteString(output, manifest)
	return err
}

// CanGenerate implements IaCGenerator
func (m *MockIaCGenerator) CanGenerate(format string) bool {
	return strings.ToLower(format) == strings.ToLower(m.Format)
}

// GenerateStage implements IaCGenerator
func (m *MockIaCGenerator) GenerateStage() Stage {
	return NewBaseStage("MockIaCGeneration", func(ctx context.Context, input interface{}) (interface{}, error) {
		var model *models.InfrastructureModel
		switch v := input.(type) {
		case *models.InfrastructureModel:
			model = v
		default:
			return nil, fmt.Errorf("invalid input type for IaC generation: %T", input)
		}
		
		return m.Generate(ctx, model)
	})
}

// MockOutputHandler is a mock implementation of OutputHandler for testing
type MockOutputHandler struct {
	// OutputDir is the output directory
	OutputDir string
	// FileContents is a map of file paths to their contents
	FileContents map[string]string
	// ShouldFail indicates whether the handler should fail
	ShouldFail bool
}

// WriteManifest implements OutputHandler
func (m *MockOutputHandler) WriteManifest(ctx context.Context, manifest string, format string, outputPath string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock manifest writing failed for path: %s", outputPath)
	}
	
	if m.FileContents == nil {
		m.FileContents = make(map[string]string)
	}
	
	// Store the manifest
	m.FileContents[outputPath] = manifest
	
	return outputPath, nil
}

// GetOutputWriter implements OutputHandler
func (m *MockOutputHandler) GetOutputWriter(outputPath string) (io.Writer, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("mock output writer failed for path: %s", outputPath)
	}
	
	return ioutil.Discard, nil
}

// WriteOutputStage implements OutputHandler
func (m *MockOutputHandler) WriteOutputStage(outputPath string) Stage {
	return NewBaseStage("MockOutputWriting", func(ctx context.Context, input interface{}) (interface{}, error) {
		var manifest string
		switch v := input.(type) {
		case string:
			manifest = v
		default:
			return nil, fmt.Errorf("invalid input type for output writing: %T", input)
		}
		
		// Determine format from file extension
		format := "terraform"
		if strings.HasSuffix(outputPath, ".yaml") || strings.HasSuffix(outputPath, ".yml") {
			format = "crossplane"
		}
		
		return m.WriteManifest(ctx, manifest, format, outputPath)
	})
}

// MockProgressReporter is a mock implementation of ProgressReporter for testing
type MockProgressReporter struct {
	// StageStatus tracks the status of each stage
	StageStatus map[string]string
	// Messages is a list of reported messages
	Messages []string
}

// NewMockProgressReporter creates a new mock progress reporter
func NewMockProgressReporter() *MockProgressReporter {
	return &MockProgressReporter{
		StageStatus: make(map[string]string),
		Messages:    make([]string, 0),
	}
}

// StartStage implements ProgressReporter
func (m *MockProgressReporter) StartStage(stageName string) {
	m.StageStatus[stageName] = "started"
	m.Messages = append(m.Messages, fmt.Sprintf("Starting %s", stageName))
}

// CompleteStage implements ProgressReporter
func (m *MockProgressReporter) CompleteStage(stageName string) {
	m.StageStatus[stageName] = "completed"
	m.Messages = append(m.Messages, fmt.Sprintf("Completed %s", stageName))
}

// FailStage implements ProgressReporter
func (m *MockProgressReporter) FailStage(stageName string, err error) {
	m.StageStatus[stageName] = "failed"
	m.Messages = append(m.Messages, fmt.Sprintf("Failed %s: %v", stageName, err))
}

// UpdateProgress implements ProgressReporter
func (m *MockProgressReporter) UpdateProgress(message string, percentage int) {
	m.Messages = append(m.Messages, fmt.Sprintf("%s (%d%%)", message, percentage))
}
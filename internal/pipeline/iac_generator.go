package pipeline

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/adapter/crossplane"
	"github.com/riptano/iac_generator_cli/internal/adapter/terraform"
	"github.com/riptano/iac_generator_cli/internal/generator"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"go.uber.org/zap"
)

// IaCGeneratorImpl is the implementation of the IaCGenerator interface
type IaCGeneratorImpl struct {
	// format is the output format for the generator (terraform, crossplane)
	format string
	// useTemplates indicates whether to use the template system
	useTemplates bool
	// OutputDir is the directory where files should be generated
	OutputDir    string
	logger       *zap.SugaredLogger
}

// NewIaCGenerator creates a new IaC generator
func NewIaCGenerator(format string, useTemplates bool) *IaCGeneratorImpl {
	// Normalize format
	format = strings.ToLower(format)

	return &IaCGeneratorImpl{
		format:       format,
		useTemplates: useTemplates,
		logger:       utils.GetLogger(),
	}
}

// Generate implements IaCGenerator
func (g *IaCGeneratorImpl) Generate(ctx context.Context, model *models.InfrastructureModel) (string, error) {
	g.logger.Debugw("Generating IaC manifest",
		"format", g.format,
		"use_templates", g.useTemplates,
		"resources_count", len(model.Resources),
	)

	// Check if the context is canceled
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// If we're using templates and need to generate the full file structure,
	// we'll invoke the template generator directly
	if g.useTemplates {
		var gen generator.Generator
		var err error
		
		switch g.format {
		case "terraform":
			tfGenerator := terraform.NewTemplateTerraformGenerator()
			tfGenerator.SetOutput(g.OutputDir)
			gen = tfGenerator
		case "crossplane":
			cpGenerator := crossplane.NewTemplateCrossplaneGenerator()
			if err := cpGenerator.Init(g.OutputDir); err != nil {
				return "", fmt.Errorf("failed to initialize Crossplane generator: %w", err)
			}
			gen = cpGenerator
		default:
			return "", fmt.Errorf("unsupported format for template generation: %s", g.format)
		}
		
		// Generate using the template generator
		result, err := gen.Generate(model)
		if err != nil {
			return "", fmt.Errorf("failed to generate with template: %w", err)
		}
		
		return result, nil
	}
	
	// For non-template generation, use the standard approach
	outputFormat := g.format

	// Generate the manifest
	manifest, err := generator.GenerateManifest(model, outputFormat)
	if err != nil {
		return "", fmt.Errorf("failed to generate manifest: %w", err)
	}

	g.logger.Debugw("Manifest generated successfully",
		"length", len(manifest),
		"format", g.format,
	)

	return manifest, nil
}

// WriteOutput implements IaCGenerator
func (g *IaCGeneratorImpl) WriteOutput(ctx context.Context, manifest string, output io.Writer) error {
	g.logger.Debug("Writing manifest to output")

	// Check if the context is canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Write to the output
	_, err := io.WriteString(output, manifest)
	if err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// CanGenerate implements IaCGenerator
func (g *IaCGeneratorImpl) CanGenerate(format string) bool {
	// Normalize format
	format = strings.ToLower(format)

	return format == g.format
}

// GetAvailableGenerators returns a list of available IaC generators
func GetAvailableGenerators() []string {
	return []string{"terraform", "crossplane"}
}

// CreateGenerator creates a generator based on the format
func CreateGenerator(format string, useTemplates bool) (generator.Generator, error) {
	// Normalize format
	format = strings.ToLower(format)

	switch format {
	case "terraform":
		if useTemplates {
			return terraform.NewTemplateTerraformGenerator(), nil
		}
		return terraform.NewTerraformGenerator(), nil
	case "crossplane":
		if useTemplates {
			return crossplane.NewTemplateCrossplaneGenerator(), nil
		}
		return crossplane.NewCrossplaneGenerator(), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// GenerateStage creates a pipeline stage that generates an IaC manifest
func (g *IaCGeneratorImpl) GenerateStage() Stage {
	return NewBaseStage("IaCGeneration", func(ctx context.Context, input interface{}) (interface{}, error) {
		var model *models.InfrastructureModel
		switch v := input.(type) {
		case *models.InfrastructureModel:
			model = v
		default:
			return nil, fmt.Errorf("invalid input type for IaC generation: %T", input)
		}

		return g.Generate(ctx, model)
	})
}
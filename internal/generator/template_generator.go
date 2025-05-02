package generator

import (
	"fmt"
	"path/filepath"

	"github.com/riptano/iac_generator_cli/internal/adapter/crossplane"
	"github.com/riptano/iac_generator_cli/internal/adapter/terraform"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// TemplateGenerator is an interface for template-based IaC manifest generators
type TemplateGenerator interface {
	Generate(model *models.InfrastructureModel) (string, error)
}

// GenerateManifestWithTemplates generates IaC manifests using templates
func GenerateManifestWithTemplates(model *models.InfrastructureModel, outputFormat string) (string, error) {
	var generator TemplateGenerator

	// Select the appropriate template generator based on the output format
	switch outputFormat {
	case "terraform":
		generator = terraform.NewTemplateTerraformGenerator()
	case "crossplane":
		generator = crossplane.NewTemplateCrossplaneGenerator()
	default:
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	// Generate the manifest
	manifest, err := generator.Generate(model)
	if err != nil {
		return "", fmt.Errorf("failed to generate manifest: %w", err)
	}

	return manifest, nil
}

// GenerateAndWriteManifestWithTemplates generates IaC manifests using templates and writes them to files
func GenerateAndWriteManifestWithTemplates(model *models.InfrastructureModel, outputFormat, outputDir, outputFile string) (string, error) {
	var generator TemplateGenerator

	// Select the appropriate template generator based on the output format
	switch outputFormat {
	case "terraform":
		tfGenerator := terraform.NewTemplateTerraformGenerator()
		tfGenerator.WithOutputDir(outputDir)
		generator = tfGenerator
	case "crossplane":
		cpGenerator := crossplane.NewTemplateCrossplaneGenerator()
		cpGenerator.Init(outputDir)
		generator = cpGenerator
	default:
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	// Generate the manifest
	manifest, err := generator.Generate(model)
	if err != nil {
		return "", fmt.Errorf("failed to generate manifest: %w", err)
	}

	// For compatibility with existing code, write the result to the specified output file
	// This is in addition to the files already written by the generators
	if outputFile != "" {
		outputPath := filepath.Join(outputDir, outputFile)
		err = utils.WriteToFile(outputPath, manifest)
		if err != nil {
			return "", fmt.Errorf("failed to write summary to %s: %w", outputPath, err)
		}
	}

	return outputDir, nil
}
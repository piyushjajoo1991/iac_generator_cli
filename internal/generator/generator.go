package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/adapter/crossplane"
	"github.com/riptano/iac_generator_cli/internal/adapter/terraform"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// Generator is an interface for different IaC manifest generators
type Generator interface {
	Generate(model *models.InfrastructureModel) (string, error)
}

// GenerateManifest generates IaC manifests based on the infrastructure model and output format
func GenerateManifest(model *models.InfrastructureModel, outputFormat string) (string, error) {
	// Use template-based generators if the format starts with "template:"
	if strings.HasPrefix(outputFormat, "template:") {
		// Extract the actual format from the prefix
		actualFormat := strings.TrimPrefix(outputFormat, "template:")
		return GenerateManifestWithTemplates(model, actualFormat)
	}

	var generator Generator

	// Select the appropriate generator based on the output format
	switch outputFormat {
	case "terraform":
		generator = terraform.NewTerraformGenerator()
	case "crossplane":
		generator = crossplane.NewCrossplaneGenerator()
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

// GenerateAndWriteManifest generates IaC manifests and writes them to files
func GenerateAndWriteManifest(model *models.InfrastructureModel, outputFormat, outputDir, outputFile string) (string, error) {
	// Use template-based generators if the format starts with "template:"
	if strings.HasPrefix(outputFormat, "template:") {
		// Extract the actual format from the prefix
		actualFormat := strings.TrimPrefix(outputFormat, "template:")
		return GenerateAndWriteManifestWithTemplates(model, actualFormat, outputDir, outputFile)
	}

	var generator Generator

	// Select the appropriate generator based on the output format
	switch outputFormat {
	case "terraform":
		generator = terraform.NewTerraformGenerator()
	case "crossplane":
		generator = crossplane.NewCrossplaneGenerator()
	default:
		return "", fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	// Generate the manifest
	manifest, err := generator.Generate(model)
	if err != nil {
		return "", fmt.Errorf("failed to generate manifest: %w", err)
	}

	// Determine the correct filename if not provided
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

	// Write the manifest to the file
	err = utils.WriteToFile(outputPath, manifest)
	if err != nil {
		return "", fmt.Errorf("failed to write manifest to file: %w", err)
	}

	return outputPath, nil
}
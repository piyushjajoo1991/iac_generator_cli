package terraform

import (
	"fmt"
	"path/filepath"

	"github.com/riptano/iac_generator_cli/internal/template"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// TemplateTerraformGenerator generates Terraform HCL using the template system
type TemplateTerraformGenerator struct {
	OutputDir string
	Model     *models.InfrastructureModel
	Config    *TerraformConfig
	renderer  *template.TemplateRenderer
}

// NewTemplateTerraformGenerator creates a new TemplateTerraformGenerator
func NewTemplateTerraformGenerator() *TemplateTerraformGenerator {
	return &TemplateTerraformGenerator{
		OutputDir: "terraform",
		Config:    DefaultTerraformConfig(),
		renderer:  template.GetDefaultRenderer(),
	}
}

// WithOutputDir sets the output directory
func (g *TemplateTerraformGenerator) WithOutputDir(dir string) *TemplateTerraformGenerator {
	g.OutputDir = dir
	return g
}

// SetOutput sets the output directory
func (g *TemplateTerraformGenerator) SetOutput(dir string) {
	g.OutputDir = dir
}

// WithConfig sets the configuration
func (g *TemplateTerraformGenerator) WithConfig(config *TerraformConfig) *TemplateTerraformGenerator {
	g.Config = config
	return g
}

// Generate generates Terraform HCL from an infrastructure model
func (g *TemplateTerraformGenerator) Generate(model *models.InfrastructureModel) (string, error) {
	g.Model = model

	// Create directory structure
	if err := utils.EnsureDirectoryExists(g.OutputDir); err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Prepare configuration for templates
	headerData := map[string]interface{}{
		"Region":            g.Config.AwsRegion,
		"TerraformVersion":  g.Config.TerraformVersion,
		"ProviderVersion":   g.Config.ProviderConstraint,
		"BackendType":       g.Config.BackendType,
		"BackendConfig":     g.Config.BackendConfig,
	}

	// Try to get a header template, if not found, we'll use the default templates from renderer
	_, err := template.GetDefaultManager().GetTemplate(template.FormatTerraform, "header.tmpl")
	if err != nil {
		// Log that we're using default header
		utils.GetLogger().Debug("Using default header template for Terraform")
	}

	// Render resources for main.tf
	result, err := g.renderer.RenderResources(template.FormatTerraform, g.Model.Resources)
	if err != nil {
		return "", fmt.Errorf("failed to render resources: %w", err)
	}

	// Format the result
	formattedResult := template.FormatRenderedContent(template.FormatTerraform, result)

	// Validate the result
	if err := template.ValidateRenderedContent(template.FormatTerraform, formattedResult); err != nil {
		// Log error but continue since validation might fail for partial configurations
		utils.GetLogger().Warn("failed to validate rendered content, continuing anyway", "error", err.Error())
	}

	// Generate and write all necessary Terraform files
	err = g.generateTerraformFiles(formattedResult, headerData)
	if err != nil {
		return "", fmt.Errorf("failed to generate Terraform files: %w", err)
	}

	return fmt.Sprintf("Terraform files generated in %s directory", g.OutputDir), nil
}

// generateTerraformFiles generates all the necessary Terraform files
func (g *TemplateTerraformGenerator) generateTerraformFiles(mainContent string, headerData map[string]interface{}) error {
	// Write main.tf
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "main.tf"), mainContent); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	// Generate and write versions.tf
	versionsTf := fmt.Sprintf(`terraform {
  required_version = ">= %s"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "%s"
    }
  }
}
`, headerData["TerraformVersion"], headerData["ProviderVersion"])
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "versions.tf"), versionsTf); err != nil {
		return fmt.Errorf("failed to write versions.tf: %w", err)
	}

	// Generate and write provider.tf
	providerTf := fmt.Sprintf(`provider "aws" {
  region = "%s"

  default_tags {
    tags = {
      Environment = "dev"
      ManagedBy   = "terraform"
      Project     = "iac-generator"
    }
  }
}
`, headerData["Region"])
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "provider.tf"), providerTf); err != nil {
		return fmt.Errorf("failed to write provider.tf: %w", err)
	}

	// Generate and write variables.tf
	variablesTf := `variable "aws_region" {
  description = "AWS region to deploy resources into"
  type        = string
  default     = "us-east-1"
}

variable "default_tags" {
  description = "Default tags to apply to all resources"
  type        = map(string)
  default     = {
    Environment = "dev"
    ManagedBy   = "terraform"
    Project     = "iac-generator"
  }
}
`
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "variables.tf"), variablesTf); err != nil {
		return fmt.Errorf("failed to write variables.tf: %w", err)
	}

	// Generate and write outputs.tf
	outputsTf := `# Auto-generated outputs
output "aws_region" {
  description = "The AWS region used"
  value       = var.aws_region
}
`
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "outputs.tf"), outputsTf); err != nil {
		return fmt.Errorf("failed to write outputs.tf: %w", err)
	}

	// Generate and write terraform.tfvars (optional)
	tfvars := fmt.Sprintf(`aws_region = "%s"

default_tags = {
  Environment = "dev"
  ManagedBy   = "terraform"
  Project     = "iac-generator"
}
`, headerData["Region"])
	if err := utils.WriteToFile(filepath.Join(g.OutputDir, "terraform.tfvars"), tfvars); err != nil {
		return fmt.Errorf("failed to write terraform.tfvars: %w", err)
	}

	return nil
}
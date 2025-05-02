package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/utils"
)

// DirectoryStructure manages the creation of a Terraform directory structure
type DirectoryStructure struct {
	BaseDir      string
	ModulesDir   string
	ModuleNames  []string
	CreateModules bool
}

// NewDirectoryStructure creates a new directory structure manager
func NewDirectoryStructure(baseDir string, createModules bool, moduleNames []string) *DirectoryStructure {
	return &DirectoryStructure{
		BaseDir:      baseDir,
		ModulesDir:   filepath.Join(baseDir, "modules"),
		ModuleNames:  moduleNames,
		CreateModules: createModules,
	}
}

// Create creates the directory structure
func (d *DirectoryStructure) Create() error {
	// Create base directory
	if err := utils.EnsureDirectoryExists(d.BaseDir); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create modules directory if needed
	if d.CreateModules {
		if err := utils.EnsureDirectoryExists(d.ModulesDir); err != nil {
			return fmt.Errorf("failed to create modules directory: %w", err)
		}

		// Create module subdirectories
		for _, moduleName := range d.ModuleNames {
			moduleDir := filepath.Join(d.ModulesDir, moduleName)
			if err := utils.EnsureDirectoryExists(moduleDir); err != nil {
				return fmt.Errorf("failed to create module directory %s: %w", moduleName, err)
			}
		}
	}

	return nil
}

// CreateEmptyFiles creates empty files for standard Terraform file types
func (d *DirectoryStructure) CreateEmptyFiles() error {
	// Create root module files
	rootFiles := []string{
		filepath.Join(d.BaseDir, "main.tf"),
		filepath.Join(d.BaseDir, "variables.tf"),
		filepath.Join(d.BaseDir, "outputs.tf"),
		filepath.Join(d.BaseDir, "versions.tf"),
		filepath.Join(d.BaseDir, "provider.tf"),
		filepath.Join(d.BaseDir, "terraform.tfvars"),
	}

	for _, file := range rootFiles {
		if !utils.FileExists(file) {
			if err := utils.WriteToFile(file, ""); err != nil {
				return fmt.Errorf("failed to create file %s: %w", file, err)
			}
		}
	}

	// Create module files
	if d.CreateModules {
		for _, moduleName := range d.ModuleNames {
			moduleDir := filepath.Join(d.ModulesDir, moduleName)
			moduleFiles := []string{
				filepath.Join(moduleDir, "main.tf"),
				filepath.Join(moduleDir, "variables.tf"),
				filepath.Join(moduleDir, "outputs.tf"),
			}

			// Add special files for certain modules
			if moduleName == "eks" {
				moduleFiles = append(moduleFiles, filepath.Join(moduleDir, "iam.tf"))
			}

			for _, file := range moduleFiles {
				if !utils.FileExists(file) {
					if err := utils.WriteToFile(file, ""); err != nil {
						return fmt.Errorf("failed to create file %s: %w", file, err)
					}
				}
			}
		}
	}

	return nil
}

// CreateTerraformrcFile creates a .terraformrc file with configuration options
func (d *DirectoryStructure) CreateTerraformrcFile(pluginCacheDir string) error {
	if pluginCacheDir == "" {
		pluginCacheDir = filepath.Join(os.Getenv("HOME"), ".terraform.d/plugin-cache")
	}
	
	// Ensure plugin cache directory exists
	if err := utils.EnsureDirectoryExists(pluginCacheDir); err != nil {
		return fmt.Errorf("failed to create plugin cache directory: %w", err)
	}
	
	terraformrcContent := fmt.Sprintf("plugin_cache_dir = \"%s\"\ndisable_checkpoint = true\n", pluginCacheDir)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	terraformrcPath := filepath.Join(homeDir, ".terraformrc")
	
	// Only write if it doesn't exist
	if !utils.FileExists(terraformrcPath) {
		if err := utils.WriteToFile(terraformrcPath, terraformrcContent); err != nil {
			return fmt.Errorf("failed to create .terraformrc file: %w", err)
		}
	}
	
	return nil
}

// CreateGitignoreFile creates a .gitignore file with Terraform-specific patterns
func (d *DirectoryStructure) CreateGitignoreFile() error {
	gitignoreContent := "# Local .terraform directories\n" +
		"**/.terraform/*\n\n" +
		"# .tfstate files\n" +
		"*.tfstate\n" +
		"*.tfstate.*\n\n" +
		"# Crash log files\n" +
		"crash.log\n" +
		"crash.*.log\n\n" +
		"# Exclude all .tfvars files, which are likely to contain sensitive data\n" +
		"*.tfvars\n" +
		"*.tfvars.json\n" +
		"!terraform.tfvars\n\n" +
		"# Ignore override files as they are usually user-specific\n" +
		"override.tf\n" +
		"override.tf.json\n" +
		"*_override.tf\n" +
		"*_override.tf.json\n\n" +
		"# Ignore CLI configuration files\n" +
		".terraformrc\n" +
		"terraform.rc\n\n" +
		"# Ignore lock files\n" +
		".terraform.lock.hcl\n\n" +
		"# Ignore plan output\n" +
		"tfplan\n" +
		"*.tfplan\n\n" +
		"# Ignore any .env files\n" +
		".env\n" +
		".envrc\n\n" +
		"# Dependency directories\n" +
		"vendor/\n\n" +
		"# Editors\n" +
		".idea/\n" +
		".vscode/\n" +
		"*.swp\n" +
		"*.swo\n" +
		"*~\n"

	gitignorePath := filepath.Join(d.BaseDir, ".gitignore")
	
	// Only write if it doesn't exist
	if !utils.FileExists(gitignorePath) {
		if err := utils.WriteToFile(gitignorePath, gitignoreContent); err != nil {
			return fmt.Errorf("failed to create .gitignore file: %w", err)
		}
	}
	
	return nil
}

// CreateREADME creates a README.md file with documentation
func (d *DirectoryStructure) CreateREADME() error {
	readmeContent := "# Terraform Infrastructure\n\n" +
		"This directory contains Terraform configuration to deploy infrastructure on AWS.\n\n" +
		"## Directory Structure\n\n" +
		"- `main.tf`: Main entry point for Terraform configuration\n" +
		"- `variables.tf`: Input variables for the module\n" +
		"- `outputs.tf`: Output variables from the module\n" +
		"- `versions.tf`: Terraform and provider versions\n" +
		"- `provider.tf`: Provider configuration\n" +
		"- `terraform.tfvars`: Variable values for the deployment\n\n"

	if d.CreateModules {
		readmeContent += "## Modules\n\n"
		for _, moduleName := range d.ModuleNames {
			readmeContent += fmt.Sprintf("### %s\n\n"+
				"The `%s` module handles the deployment of %s resources.\n\n",
				toTitleCase(moduleName),
				moduleName,
				moduleName)
		}
	}
	
	readmeContent += "## Usage\n\n" +
		"```bash\n" +
		"# Initialize Terraform\n" +
		"terraform init\n\n" +
		"# See planned changes\n" +
		"terraform plan\n\n" +
		"# Apply changes\n" +
		"terraform apply\n\n" +
		"# Destroy infrastructure\n" +
		"terraform destroy\n" +
		"```\n\n" +
		"## Requirements\n\n" +
		"- Terraform >= 1.0.0\n" +
		"- AWS Provider >= 5.0.0\n" +
		"- AWS CLI configured with appropriate credentials\n"

	readmePath := filepath.Join(d.BaseDir, "README.md")
	
	// Only write if it doesn't exist
	if !utils.FileExists(readmePath) {
		if err := utils.WriteToFile(readmePath, readmeContent); err != nil {
			return fmt.Errorf("failed to create README.md file: %w", err)
		}
	}
	
	return nil
}

// toTitleCase converts a string to title case
func toTitleCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// CreateVersionsFile creates a versions.tf file with Terraform and provider version constraints
func (d *DirectoryStructure) CreateVersionsFile() error {
	versionsContent := "terraform {\n" +
		"  required_version = \">= 1.0.0\"\n\n" +
		"  required_providers {\n" +
		"    aws = {\n" +
		"      source  = \"hashicorp/aws\"\n" +
		"      version = \">= 5.0.0\"\n" +
		"    }\n" +
		"  }\n" +
		"}\n"

	versionsPath := filepath.Join(d.BaseDir, "versions.tf")
	
	// Only write if it doesn't exist
	if !utils.FileExists(versionsPath) {
		if err := utils.WriteToFile(versionsPath, versionsContent); err != nil {
			return fmt.Errorf("failed to create versions.tf file: %w", err)
		}
	}
	
	return nil
}

// CreateProviderFile creates a provider.tf file with AWS provider configuration
func (d *DirectoryStructure) CreateProviderFile(region string) error {
	if region == "" {
		region = "us-east-1"
	}
	
	providerContent := "provider \"aws\" {\n" +
		"  region = var.aws_region\n" +
		"  # Additional provider configurations can be added here\n" +
		"  # For example:\n" +
		"  # profile = var.aws_profile\n" +
		"  # assume_role {\n" +
		"  #   role_arn = var.role_arn\n" +
		"  # }\n" +
		"}\n\n" +
		"# Alternative provider for cross-region resources\n" +
		"# provider \"aws\" {\n" +
		"#   alias  = \"alternative\"\n" +
		"#   region = \"us-west-2\"\n" +
		"# }\n"

	providerPath := filepath.Join(d.BaseDir, "provider.tf")
	
	// Create variables.tf with region variable
	variablesContent := "variable \"aws_region\" {\n" +
		"  description = \"AWS region to deploy resources\"\n" +
		fmt.Sprintf("  default     = \"%s\"\n", region) +
		"  type        = string\n" +
		"}\n\n" +
		"# variable \"aws_profile\" {\n" +
		"#   description = \"AWS profile to use for deployment\"\n" +
		"#   default     = \"default\"\n" +
		"#   type        = string\n" +
		"# }\n"
	
	variablesPath := filepath.Join(d.BaseDir, "variables.tf")
	
	// Only write if it doesn't exist
	if !utils.FileExists(providerPath) {
		if err := utils.WriteToFile(providerPath, providerContent); err != nil {
			return fmt.Errorf("failed to create provider.tf file: %w", err)
		}
	}
	
	// Append to variables.tf if it exists, otherwise create it
	if utils.FileExists(variablesPath) {
		existingContent, err := utils.ReadFromFile(variablesPath)
		if err != nil {
			return fmt.Errorf("failed to read variables.tf file: %w", err)
		}
		
		// Only add region variable if it doesn't exist
		if !strings.Contains(existingContent, "variable \"aws_region\"") {
			if err := utils.WriteToFile(variablesPath, existingContent+"\n"+variablesContent); err != nil {
				return fmt.Errorf("failed to update variables.tf file: %w", err)
			}
		}
	} else {
		if err := utils.WriteToFile(variablesPath, variablesContent); err != nil {
			return fmt.Errorf("failed to create variables.tf file: %w", err)
		}
	}
	
	return nil
}
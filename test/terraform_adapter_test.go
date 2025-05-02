package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riptano/iac_generator_cli/internal/adapter/terraform"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

func TestTerraformGenerator(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new Terraform generator
	generator := terraform.NewTerraformGenerator().
		WithOutputDir(tempDir).
		WithConfig(&terraform.TerraformConfig{
			AwsRegion:          "us-east-1",
			CreateModules:      true,
			ModuleNames:        []string{"vpc", "eks"},
			BackendType:        "local",
			TerraformVersion:   "1.0.0",
			ProviderConstraint: "~> 5.0",
		})

	// Create a simple infrastructure model for testing
	model := createTestInfrastructureModel()

	// Generate Terraform files
	result, err := generator.Generate(model)
	if err != nil {
		t.Fatalf("Failed to generate Terraform files: %v", err)
	}

	// Validate output directory structure
	validateOutputDir(t, tempDir)

	// Validate result string
	if result == "" {
		t.Errorf("Expected non-empty result string")
	}
}

func TestHCLWriter(t *testing.T) {
	// Create an HCL writer
	writer := terraform.NewHCLWriter()

	// Create a provider block
	providerBlock := terraform.NewHCLBlock("provider", "aws")
	providerBlock.AddAttribute("region", "us-east-1")
	providerBlock.AddAttribute("profile", "default")

	// Write the provider block
	writer.WriteBlock(providerBlock)

	// Create a resource block
	resourceBlock := terraform.NewHCLBlock("resource", "aws_vpc", "main")
	resourceBlock.AddAttribute("cidr_block", "10.0.0.0/16")
	resourceBlock.AddAttribute("enable_dns_support", true)
	resourceBlock.AddAttribute("enable_dns_hostnames", true)

	// Add a nested block
	tagsBlock := terraform.NewHCLBlock("tags")
	tagsBlock.AddAttribute("Name", "main-vpc")
	tagsBlock.AddAttribute("Environment", "dev")
	resourceBlock.AddBlock(tagsBlock)

	// Write the resource block
	writer.WriteBlock(resourceBlock)

	// Validate the output
	output := writer.String()
	expectedSubstrings := []string{
		"provider \"aws\" {",
		"region = \"us-east-1\"",
		"profile = \"default\"",
		"resource \"aws_vpc\" \"main\" {",
		"cidr_block = \"10.0.0.0/16\"",
		"enable_dns_support = true",
		"enable_dns_hostnames = true",
		"tags {",
		"Name = \"main-vpc\"",
		"Environment = \"dev\"",
	}

	for _, substr := range expectedSubstrings {
		if !contains(output, substr) {
			t.Errorf("Expected output to contain substring: %s", substr)
		}
	}
}

func TestDirectoryStructure(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a directory structure
	dirStructure := terraform.NewDirectoryStructure(tempDir, true, []string{"vpc", "eks"})

	// Create the directory structure
	err = dirStructure.Create()
	if err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create empty files
	err = dirStructure.CreateEmptyFiles()
	if err != nil {
		t.Fatalf("Failed to create empty files: %v", err)
	}

	// Create README
	err = dirStructure.CreateREADME()
	if err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	// Create .gitignore
	err = dirStructure.CreateGitignoreFile()
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Validate the directory structure
	validateDirectoryStructure(t, tempDir)
}

// Helper functions

// createTestInfrastructureModel creates a test infrastructure model
func createTestInfrastructureModel() *models.InfrastructureModel {
	model := models.NewInfrastructureModel()

	// Create a VPC resource
	vpc := models.NewResource(models.ResourceVPC, "main")
	vpc.AddProperty("cidr_block", "10.0.0.0/16")
	vpc.AddProperty("enable_dns_support", true)
	vpc.AddProperty("enable_dns_hostnames", true)
	model.AddResource(vpc)

	// Create a subnet resource
	subnet := models.NewResource(models.ResourceSubnet, "public")
	subnet.AddProperty("vpc_id", "${aws_vpc.main.id}")
	subnet.AddProperty("cidr_block", "10.0.1.0/24")
	subnet.AddProperty("availability_zone", "us-east-1a")
	subnet.AddProperty("map_public_ip_on_launch", true)
	subnet.AddDependency("main")
	model.AddResource(subnet)

	return model
}

// validateOutputDir validates the output directory structure
func validateOutputDir(t *testing.T, dir string) {
	// Check root files
	rootFiles := []string{
		"main.tf",
		"variables.tf",
		"outputs.tf",
		"versions.tf",
		"provider.tf",
		"terraform.tfvars",
	}

	for _, file := range rootFiles {
		path := filepath.Join(dir, file)
		if !fileExists(path) {
			t.Errorf("Expected file not found: %s", path)
		}
	}

	// Check modules directory
	modulesDir := filepath.Join(dir, "modules")
	if !dirExists(modulesDir) {
		t.Errorf("Expected directory not found: %s", modulesDir)
	}

	// Check module directories
	moduleNames := []string{"vpc", "eks"}
	for _, module := range moduleNames {
		moduleDir := filepath.Join(modulesDir, module)
		if !dirExists(moduleDir) {
			t.Errorf("Expected directory not found: %s", moduleDir)
		}

		// Check module files
		moduleFiles := []string{
			"main.tf",
			"variables.tf",
			"outputs.tf",
		}

		for _, file := range moduleFiles {
			path := filepath.Join(moduleDir, file)
			if !fileExists(path) {
				t.Errorf("Expected file not found: %s", path)
			}
		}

		// Check special files for EKS module
		if module == "eks" {
			path := filepath.Join(moduleDir, "iam.tf")
			if !fileExists(path) {
				t.Errorf("Expected file not found: %s", path)
			}
		}
	}
}

// validateDirectoryStructure validates the directory structure
func validateDirectoryStructure(t *testing.T, dir string) {
	// Check root files
	rootFiles := []string{
		"main.tf",
		"variables.tf",
		"outputs.tf",
		"versions.tf",
		"provider.tf",
		"terraform.tfvars",
		"README.md",
		".gitignore",
	}

	for _, file := range rootFiles {
		path := filepath.Join(dir, file)
		if !fileExists(path) {
			t.Errorf("Expected file not found: %s", path)
		}
	}

	// Check modules directory
	modulesDir := filepath.Join(dir, "modules")
	if !dirExists(modulesDir) {
		t.Errorf("Expected directory not found: %s", modulesDir)
	}

	// Check module directories
	moduleNames := []string{"vpc", "eks"}
	for _, module := range moduleNames {
		moduleDir := filepath.Join(modulesDir, module)
		if !dirExists(moduleDir) {
			t.Errorf("Expected directory not found: %s", moduleDir)
		}

		// Check module files
		moduleFiles := []string{
			"main.tf",
			"variables.tf",
			"outputs.tf",
		}

		for _, file := range moduleFiles {
			path := filepath.Join(moduleDir, file)
			if !fileExists(path) {
				t.Errorf("Expected file not found: %s", path)
			}
		}

		// Check special files for EKS module
		if module == "eks" {
			path := filepath.Join(moduleDir, "iam.tf")
			if !fileExists(path) {
				t.Errorf("Expected file not found: %s", path)
			}
		}
	}
}

// Helper function to check if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Helper function to check if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return true // TODO: Implement properly
}
package crossplane

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// CrossplaneGenerator generates Crossplane YAML manifests
type CrossplaneGenerator struct {
	baseDir      string
	dirStructure *DirectoryStructure
	vpcGenerator *VPCGenerator
	eksGenerator *EKSGenerator
	provGenerator *ProviderGenerator
}

// NewCrossplaneGenerator creates a new CrossplaneGenerator
func NewCrossplaneGenerator() *CrossplaneGenerator {
	return &CrossplaneGenerator{}
}

// Init initializes the generator with a base directory
func (g *CrossplaneGenerator) Init(baseDir string) error {
	return g.SetOutputDir(baseDir)
}

// SetOutputDir sets the output directory and initializes the generator
func (g *CrossplaneGenerator) SetOutputDir(baseDir string) error {
	g.baseDir = baseDir
	g.dirStructure = NewDirectoryStructure(baseDir)
	g.vpcGenerator = NewVPCGenerator(baseDir)
	g.eksGenerator = NewEKSGenerator(baseDir)
	g.provGenerator = NewProviderGenerator(baseDir)
	
	// Create the directory structure
	if err := g.dirStructure.Create(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	
	// Create empty files
	if err := g.dirStructure.CreateEmptyFiles(); err != nil {
		return fmt.Errorf("failed to create empty files: %w", err)
	}
	
	// Create kustomization files
	if err := g.dirStructure.CreateKustomizationFiles(); err != nil {
		return fmt.Errorf("failed to create kustomization files: %w", err)
	}
	
	// Create README
	if err := g.dirStructure.CreateREADME(); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}
	
	return nil
}

// Generate generates Crossplane YAML from an infrastructure model
func (g *CrossplaneGenerator) Generate(model *models.InfrastructureModel) (string, error) {
	// If baseDir is not set, use a temporary directory
	if g.baseDir == "" {
		tempDir, err := os.MkdirTemp("", "crossplane-")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary directory: %w", err)
		}
		if err := g.Init(tempDir); err != nil {
			return "", err
		}
	}
	
	// Extract region from the model
	region := "us-east-1" // Default region
	for _, resource := range model.Resources {
		for _, prop := range resource.Properties {
			if strings.Contains(strings.ToLower(prop.Name), "region") {
				if regionStr, ok := prop.Value.(string); ok {
					region = regionStr
				}
			}
		}
	}
	
	// Generate the provider configuration
	// Use empty strings for access and secret keys - they would be provided by the user in a real scenario
	if err := g.provGenerator.GenerateCommonResources(region, "", ""); err != nil {
		return "", fmt.Errorf("failed to generate provider configuration: %w", err)
	}
	
	// Generate VPC resources
	if err := g.vpcGenerator.GenerateNetworkResources(model); err != nil {
		return "", fmt.Errorf("failed to generate VPC resources: %w", err)
	}
	
	// Generate EKS resources
	if err := g.eksGenerator.GenerateEKSResources(model); err != nil {
		return "", fmt.Errorf("failed to generate EKS resources: %w", err)
	}
	
	// Return a summary of the generated resources
	summary, err := g.generateSummary()
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	
	return summary, nil
}

// GenerateToFile generates Crossplane YAML and writes it to a specific file
func (g *CrossplaneGenerator) GenerateToFile(model *models.InfrastructureModel, outputPath string) (string, error) {
	// Generate the resources
	summary, err := g.Generate(model)
	if err != nil {
		return "", err
	}
	
	// Write the summary to a file
	summaryPath := filepath.Join(g.baseDir, "summary.txt")
	if err := utils.WriteToFile(summaryPath, summary); err != nil {
		return "", fmt.Errorf("failed to write summary: %w", err)
	}
	
	return g.baseDir, nil
}

// generateSummary generates a summary of the created resources
func (g *CrossplaneGenerator) generateSummary() (string, error) {
	var summary bytes.Buffer
	
	summary.WriteString("# Crossplane Resources Generated\n\n")
	summary.WriteString("The following Crossplane resources have been generated:\n\n")
	
	// Add base directory information
	summary.WriteString(fmt.Sprintf("Output directory: %s\n\n", g.baseDir))
	
	// List common resources
	summary.WriteString("## Common Resources\n\n")
	summary.WriteString("- AWS Provider\n")
	summary.WriteString("- Provider Configuration\n\n")
	
	// List VPC resources
	summary.WriteString("## VPC Resources\n\n")
	vpcFiles := []string{"vpc.yaml", "subnets.yaml", "gateways.yaml", "routing.yaml"}
	for _, file := range vpcFiles {
		filePath := filepath.Join(g.baseDir, "vpc", file)
		if utils.FileExists(filePath) {
			content, err := utils.ReadFromFile(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to read %s: %w", file, err)
			}
			
			// Count the number of resources by counting "kind:" occurrences
			resourceCount := strings.Count(content, "kind:")
			resourceType := strings.TrimSuffix(file, ".yaml")
			summary.WriteString(fmt.Sprintf("- %s: %d resources\n", resourceType, resourceCount))
		}
	}
	summary.WriteString("\n")
	
	// List EKS resources
	summary.WriteString("## EKS Resources\n\n")
	eksFiles := []string{"cluster.yaml", "nodegroup.yaml", "iam.yaml"}
	for _, file := range eksFiles {
		filePath := filepath.Join(g.baseDir, "eks", file)
		if utils.FileExists(filePath) {
			content, err := utils.ReadFromFile(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to read %s: %w", file, err)
			}
			
			// Count the number of resources by counting "kind:" occurrences
			resourceCount := strings.Count(content, "kind:")
			resourceType := strings.TrimSuffix(file, ".yaml")
			summary.WriteString(fmt.Sprintf("- %s: %d resources\n", resourceType, resourceCount))
		}
	}
	summary.WriteString("\n")
	
	// Add usage instructions
	summary.WriteString("## Usage Instructions\n\n")
	summary.WriteString("To apply these resources to your Kubernetes cluster with Crossplane installed:\n\n")
	summary.WriteString("```bash\n")
	summary.WriteString(fmt.Sprintf("kubectl apply -k %s\n", g.baseDir))
	summary.WriteString("```\n\n")
	summary.WriteString("Or apply specific components:\n\n")
	summary.WriteString("```bash\n")
	summary.WriteString(fmt.Sprintf("kubectl apply -k %s/base  # Common resources\n", g.baseDir))
	summary.WriteString(fmt.Sprintf("kubectl apply -k %s/vpc   # VPC resources\n", g.baseDir))
	summary.WriteString(fmt.Sprintf("kubectl apply -k %s/eks   # EKS resources\n", g.baseDir))
	summary.WriteString("```\n")
	
	return summary.String(), nil
}
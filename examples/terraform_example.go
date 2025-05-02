package examples

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/pipeline"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// RunDetailedExample runs a detailed example of the infrastructure model generation pipeline
// with intermediate steps and debug output.
func RunDetailedExample() error {
	// Initialize logger
	logger := utils.GetLogger()
	logger.Info("Starting detailed example with intermediate steps")

	// Sample infrastructure description
	description := `Deploy AWS infrastructure in us-east-1 with a VPC, 3 public and 3 private subnets across 3 AZs, 
	an internet gateway, 3 NAT gateways (one per AZ), and an EKS cluster with both public and private API access 
	deployed in the private subnets. Include a node pool using t3-medium instances.`

	fmt.Println("=== STEP 1: INPUT DESCRIPTION ===")
	fmt.Println(description)
	fmt.Println()

	// Step 2: NLP Processing
	fmt.Println("=== STEP 2: NLP PROCESSING ===")
	fmt.Println("Processing description with NLP parser...")
	
	// Create NLP processor
	nlpProcessor := pipeline.NewNLPProcessor()
	ctx := context.Background()

	// Parse the description
	model, err := nlpProcessor.ParseDescription(ctx, description)
	if err != nil {
		return fmt.Errorf("failed to parse description: %w", err)
	}

	// Print model details
	printModelDetails(model)
	fmt.Println()
	
	// Step 3: Model Building
	fmt.Println("=== STEP 3: MODEL BUILDING ===")
	fmt.Println("Building infrastructure model...")
	
	// Create model builder
	modelBuilder := pipeline.NewModelBuilder("us-east-1")
	
	// Build the model
	enhancedModel, err := modelBuilder.BuildModel(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to build model: %w", err)
	}
	
	// Print enhanced model details
	printEnhancedModelDetails(enhancedModel)
	fmt.Println()

	// Step 4: Generate Terraform Code
	fmt.Println("=== STEP 4: TERRAFORM GENERATION ===")
	fmt.Println("Generating Terraform code...")
	
	// Create Terraform generator
	generator := pipeline.NewIaCGenerator("terraform", true)
	
	// Generate Terraform code
	terraformCode, err := generator.Generate(ctx, enhancedModel)
	if err != nil {
		return fmt.Errorf("failed to generate Terraform code: %w", err)
	}
	
	// Print a preview of the Terraform code
	fmt.Println("Terraform Code Preview:")
	fmt.Println(strings.Repeat("-", 40))
	previewLines := strings.Split(terraformCode, "\n")
	if len(previewLines) > 20 {
		for i := 0; i < 20; i++ {
			fmt.Println(previewLines[i])
		}
		fmt.Println("... (truncated) ...")
	} else {
		fmt.Println(terraformCode)
	}
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()
	
	// Step 5: Generate Crossplane YAML
	fmt.Println("=== STEP 5: CROSSPLANE GENERATION ===")
	fmt.Println("Generating Crossplane YAML...")
	
	// Create Crossplane generator
	cpGenerator := pipeline.NewIaCGenerator("crossplane", true)
	
	// Generate Crossplane YAML
	crossplaneYaml, err := cpGenerator.Generate(ctx, enhancedModel)
	if err != nil {
		return fmt.Errorf("failed to generate Crossplane YAML: %w", err)
	}
	
	// Print a preview of the Crossplane YAML
	fmt.Println("Crossplane YAML Preview:")
	fmt.Println(strings.Repeat("-", 40))
	previewLines = strings.Split(crossplaneYaml, "\n")
	if len(previewLines) > 20 {
		for i := 0; i < 20; i++ {
			fmt.Println(previewLines[i])
		}
		fmt.Println("... (truncated) ...")
	} else {
		fmt.Println(crossplaneYaml)
	}
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()

	// Step 6: Save outputs to files
	fmt.Println("=== STEP 6: SAVING OUTPUT FILES ===")
	
	// Create output directories
	outputDirs := []string{"output/terraform", "output/crossplane"}
	for _, dir := range outputDirs {
		if err := utils.EnsureDirectoryExists(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Save Terraform code
	tfOutputPath := filepath.Join("output", "terraform", "main.tf")
	if err := utils.WriteToFile(tfOutputPath, terraformCode); err != nil {
		return fmt.Errorf("failed to write Terraform file: %w", err)
	}
	fmt.Printf("Terraform code saved to: %s\n", tfOutputPath)
	
	// Save Crossplane YAML
	cpOutputPath := filepath.Join("output", "crossplane", "resources.yaml")
	if err := utils.WriteToFile(cpOutputPath, crossplaneYaml); err != nil {
		return fmt.Errorf("failed to write Crossplane file: %w", err)
	}
	fmt.Printf("Crossplane YAML saved to: %s\n", cpOutputPath)
	
	fmt.Println()
	fmt.Println("=== EXAMPLE COMPLETED SUCCESSFULLY ===")
	fmt.Println("Infrastructure manifests generated from the description:")
	fmt.Printf("- Terraform: %s\n", tfOutputPath)
	fmt.Printf("- Crossplane: %s\n", cpOutputPath)
	
	return nil
}

// printModelDetails prints the details of an infrastructure model
func printModelDetails(model *models.InfrastructureModel) {
	fmt.Printf("Detected %d resources:\n", len(model.Resources))
	
	for i, resource := range model.Resources {
		fmt.Printf("%d. Resource Type: %s\n", i+1, resource.Type)
		
		if len(resource.Properties) > 0 {
			fmt.Printf("   Properties:\n")
			for _, prop := range resource.Properties {
				fmt.Printf("   - %s: %v\n", prop.Name, prop.Value)
			}
		}
		
		if len(resource.DependsOn) > 0 {
			fmt.Printf("   Dependencies:\n")
			for _, dep := range resource.DependsOn {
				fmt.Printf("   - %s\n", dep)
			}
		}
	}
}

// printEnhancedModelDetails prints the details of the enhanced infrastructure model
func printEnhancedModelDetails(model *models.InfrastructureModel) {
	fmt.Printf("Enhanced model with %d resources:\n", len(model.Resources))
	
	// Group resources by type
	resourcesByType := make(map[string][]models.Resource)
	for _, resource := range model.Resources {
		resourcesByType[string(resource.Type)] = append(resourcesByType[string(resource.Type)], resource)
	}
	
	// Print summary
	fmt.Println("Resource summary:")
	for resourceType, resources := range resourcesByType {
		fmt.Printf("- %s: %d resources\n", resourceType, len(resources))
	}
	
	// Print key relationships
	fmt.Println("\nKey resource relationships:")
	
	// Find VPC and its components
	var vpc *models.Resource
	for i, resource := range model.Resources {
		if resource.Type == models.ResourceVPC {
			vpc = &model.Resources[i]
			break
		}
	}
	
	if vpc != nil {
		fmt.Printf("- VPC: %s\n", vpc.Name)
		
		// Find subnets
		publicSubnets := 0
		privateSubnets := 0
		
		for _, resource := range model.Resources {
			if resource.Type == models.ResourceSubnet {
				// Check if public or private
				isPublic := false
				for _, prop := range resource.Properties {
					if prop.Name == "public" && prop.Value == true {
						isPublic = true
						break
					}
				}
				
				if isPublic {
					publicSubnets++
				} else {
					privateSubnets++
				}
			}
		}
		
		fmt.Printf("  - Public Subnets: %d\n", publicSubnets)
		fmt.Printf("  - Private Subnets: %d\n", privateSubnets)
		
		// Find EKS cluster
		var eks *models.Resource
		for i, resource := range model.Resources {
			if resource.Type == models.ResourceEKSCluster {
				eks = &model.Resources[i]
				break
			}
		}
		
		if eks != nil {
			fmt.Printf("- EKS Cluster: %s\n", eks.Name)
			
			// Find node groups
			nodeGroups := 0
			for _, resource := range model.Resources {
				if resource.Type == models.ResourceNodeGroup {
					nodeGroups++
				}
			}
			
			fmt.Printf("  - Node Groups: %d\n", nodeGroups)
		}
	}
}

// RunFromCommandLine runs the detailed example from the command line
func RunFromCommandLine() {
	if err := RunDetailedExample(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running example: %v\n", err)
		os.Exit(1)
	}
}
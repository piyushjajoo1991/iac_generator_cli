package iacgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/nlp"
	"github.com/riptano/iac_generator_cli/internal/pipeline"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Generate command flags
	inputFile    string
	outputFile   string
)

var generateCmd = &cobra.Command{
	Use:   "generate [description]",
	Short: "Generate IaC from natural language description",
	Long: `Generate Infrastructure as Code (IaC) manifests from a natural language description.
The description should detail the AWS infrastructure you want to provision.

You can provide the description directly as an argument or specify a file containing
the description using the --file flag. The generated IaC manifest will be printed
to stdout by default, or written to the specified output directory.`,
	Example: `  # Generate from command-line description
  iacgen generate "Create an EC2 instance with t2.micro size"

  # Generate from a file
  iacgen generate --file ./infra-description.txt

  # Generate with specific output format and directory
  iacgen generate "Create an S3 bucket for static website hosting" --output crossplane --output-dir ./manifests

  # Generate with a specific region
  iacgen generate "Create a new VPC with 3 subnets" --region us-west-2
  
  # Generate using the template system
  iacgen generate "Create an EKS cluster with 2 nodes" --use-templates`,
	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.GetLogger()
		
		// Validate input - either direct description or file must be provided
		if len(args) == 0 && inputFile == "" {
			return fmt.Errorf("either provide a description as an argument or specify an input file with --file")
		}
		
		// Validate output format
		if !isValidOutputFormat(toolFormat) {
			return fmt.Errorf("invalid output format: %s (supported formats: terraform, crossplane)", toolFormat)
		}
		
		// If input file is specified, check if it exists and is readable
		if inputFile != "" {
			if !utils.FileExists(inputFile) {
				return fmt.Errorf("input file does not exist: %s", inputFile)
			}
			
			// Check if file is readable
			if _, err := utils.ReadFromFile(inputFile); err != nil {
				return fmt.Errorf("cannot read input file: %s (%w)", inputFile, err)
			}
			
			logger.Debug("Input file validated", "file", inputFile)
		}
		
		// Validate region format (basic check for now)
		if !isValidRegionFormat(awsRegion) {
			logger.Warn("AWS region format may be invalid", "region", awsRegion)
		}
		
		// Create output directory if it doesn't exist
		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir != "." {
			// Check if we have write permission by creating the directory
			if err := utils.EnsureDirectoryExists(outputDir); err != nil {
				return fmt.Errorf("failed to create or access output directory: %w", err)
			}
			
			logger.Debug("Output directory validated", "dir", outputDir)
		}
		
		// Validate output file permissions if specified
		if outputFile != "" {
			outputPath := filepath.Join(outputDir, outputFile)
			
			// Check if the directory where the file will be created is writable
			dirPath := filepath.Dir(outputPath)
			if err := utils.EnsureDirectoryExists(dirPath); err != nil {
				return fmt.Errorf("cannot create output file directory: %w", err)
			}
			
			// If the file already exists, check if it's writable
			if utils.FileExists(outputPath) {
				if err := utils.IsFileWritable(outputPath); err != nil {
					return fmt.Errorf("output file exists but is not writable: %s (%w)", outputPath, err)
				}
			}
			
			logger.Debug("Output file location validated", "file", outputPath)
		}
		
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get logger
		logger := utils.GetLogger()
		
		// Get flags
		outputFormat := toolFormat
		region := awsRegion
		outDir := outputDir
		
		// Log configuration
		logger.Debug("Configuration", 
			"output_format", outputFormat,
			"region", region,
			"output_dir", outDir,
			"input_file", inputFile,
			"use_templates", useTemplates)
			
		var description string
		
		// Get description from argument 
		if len(args) > 0 {
			description = args[0]
			logger.Debug("Using description from argument")
		}
		
		// Create pipeline parameters
		params := &pipeline.ProcessingParams{
			Description:    description,
			InputFile:      inputFile,
			OutputFormat:   outputFormat,
			OutputDir:      outDir,
			OutputFile:     outputFile,
			Region:         region,
			UseTemplates:   useTemplates,
			Debug:          debugMode,
			ProgressWriter: os.Stdout,
		}
		
		// Process through the pipeline
		result, err := pipeline.RunWithProgressFeedback(params, os.Stdout)
		if err != nil {
			logger.Error("Failed to generate IaC manifest", "error", err.Error())
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Print the result
		fmt.Println(result)
		
		logger.Info("Successfully generated IaC manifest")
	},
}

// isValidRegionFormat checks if the AWS region format is valid
func isValidRegionFormat(region string) bool {
	// Basic format checking for AWS regions like us-east-1, us-west-2, etc.
	validRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2",
		"sa-east-1", "ca-central-1", "ap-south-1", "ap-east-1",
		"eu-north-1", "eu-south-1", "af-south-1", "me-south-1",
	}

	for _, validRegion := range validRegions {
		if validRegion == region {
			return true
		}
	}

	// If not in the valid regions list, check the format pattern
	// AWS regions typically follow the pattern of: area-compass-number
	parts := strings.Split(region, "-")
	if len(parts) < 3 {
		return false
	}

	// Check if the last part is a number
	_, err := strconv.Atoi(parts[len(parts)-1])
	return err == nil
}

// processDescription parses the description and applies region setting
// Note: This function is kept for backward compatibility but is not used by the pipeline
func processDescription(description, region string) (*models.InfrastructureModel, error) {
	// Parse the description
	infraModel, err := nlp.ParseDescription(description)
	if err != nil {
		return nil, fmt.Errorf("failed to parse description: %w", err)
	}
	
	// Set AWS region for all resources that support it
	for i, resource := range infraModel.Resources {
		// Check resource type and update region if applicable
		switch resource.Type {
		case models.ResourceEC2Instance, models.ResourceRDSInstance, models.ResourceVPC, 
			 models.ResourceSubnet, models.ResourceSecurityGroup:
			// Check if region is already set
			hasRegion := false
			for _, prop := range resource.Properties {
				if prop.Name == "region" {
					hasRegion = true
					break
				}
			}
			
			// Add region property if not already set
			if !hasRegion {
				infraModel.Resources[i].AddProperty("region", region)
			}
		}
	}
	
	return infraModel, nil
}

func init() {
	// Input options
	generateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input file containing infrastructure description")
	
	// Output options
	generateCmd.Flags().StringVarP(&outputFile, "output-file", "", "", "Output filename (default: based on input file or 'main.tf'/'resources.yaml')")
	
	// Bind viper for persistent configuration
	viper.BindPFlag("input_file", generateCmd.Flags().Lookup("file"))
	viper.BindPFlag("output_file", generateCmd.Flags().Lookup("output-file"))
}
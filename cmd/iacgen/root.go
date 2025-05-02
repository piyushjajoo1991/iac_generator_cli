package iacgen

import (
	"fmt"
	"os"
	"strings"

	"github.com/riptano/iac_generator_cli/internal/config"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Global flags
	awsRegion      string
	debugMode      bool
	outputDir      string
	toolFormat     string
	useTemplates   bool
	versionFlag    bool
)

var rootCmd = &cobra.Command{
	Use:   "iacgen",
	Short: "IaC Generator - Generate infrastructure as code from natural language",
	Long: `IaC Generator is a CLI tool that parses English descriptions of AWS infrastructure 
and generates Infrastructure as Code (IaC) manifests based on those descriptions.

It supports generating both Terraform configurations and Crossplane resource manifests.
You can provide infrastructure descriptions in plain English, and the tool will analyze 
the text to identify resources, their configurations, and relationships.`,
	Example: `  # Generate Terraform configuration from description
  iacgen generate "Create an EC2 instance with t2.micro size in us-west-2 region"

  # Generate Crossplane manifest from a file
  iacgen generate -f infra_description.txt -o crossplane -d ./output

  # Specify AWS region
  iacgen generate "Create an S3 bucket for static website hosting" --region us-east-2`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for version flag
		if versionFlag {
			fmt.Printf("iacgen version %s\n", version.Version)
			return
		}

		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip validation if just displaying version
		if versionFlag {
			return
		}

		// Set up logging based on debug flag
		logLevel := "info"
		if debugMode {
			logLevel = "debug"
		}
		viper.Set("log_level", logLevel)
		
		// Get logger
		logger := utils.GetLogger()
		logger.Debug("Debug mode enabled")
		logger.Info("Using AWS region", "region", awsRegion)
		
		// Validate output format
		if !isValidOutputFormat(toolFormat) {
			logger.Error("Invalid output format", "format", toolFormat)
			fmt.Printf("Error: Invalid output format: %s. Supported formats are: terraform, crossplane\n", toolFormat)
			os.Exit(1)
		}
	},
}

// isValidOutputFormat checks if the provided output format is supported
func isValidOutputFormat(format string) bool {
	validFormats := []string{"terraform", "crossplane"}
	format = strings.ToLower(format)
	
	for _, v := range validFormats {
		if v == format {
			return true
		}
	}
	return false
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)

	// Configuration file
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.iacgen.yaml)")
	
	// Tool selection
	rootCmd.PersistentFlags().StringVarP(&toolFormat, "output", "o", "terraform", "Output format (terraform or crossplane)")
	viper.BindPFlag("default_type", rootCmd.PersistentFlags().Lookup("output"))

	// Output directory
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output-dir", "d", ".", "Directory to write output files")
	viper.BindPFlag("output_dir", rootCmd.PersistentFlags().Lookup("output-dir"))
	
	// AWS Region
	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "us-east-1", "AWS region for resources")
	viper.BindPFlag("aws_region", rootCmd.PersistentFlags().Lookup("region"))

	// Template system
	rootCmd.PersistentFlags().BoolVar(&useTemplates, "use-templates", false, "Use the template system for generating IaC code")
	viper.BindPFlag("use_templates", rootCmd.PersistentFlags().Lookup("use-templates"))

	// Logging options
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "v", false, "Enable debug output")
	
	// Version flag
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "V", false, "Print version information and exit")
	
	// Add commands
	rootCmd.AddCommand(generateCmd)
}
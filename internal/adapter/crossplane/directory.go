package crossplane

import (
	"fmt"
	"path/filepath"

	"github.com/riptano/iac_generator_cli/internal/utils"
)

// DirectoryStructure manages the creation of a Crossplane directory structure
type DirectoryStructure struct {
	BaseDir     string
	VPCDir      string
	EKSDir      string
	CommonDir   string
	Directories []string
}

// NewDirectoryStructure creates a new Crossplane directory structure manager
func NewDirectoryStructure(baseDir string) *DirectoryStructure {
	return &DirectoryStructure{
		BaseDir:   baseDir,
		VPCDir:    filepath.Join(baseDir, "vpc"),
		EKSDir:    filepath.Join(baseDir, "eks"),
		CommonDir: filepath.Join(baseDir, "base"),
		Directories: []string{
			filepath.Join(baseDir, "vpc"),
			filepath.Join(baseDir, "eks"),
			filepath.Join(baseDir, "base"),
			filepath.Join(baseDir, "s3"),
			filepath.Join(baseDir, "rds"),
			filepath.Join(baseDir, "ec2"),
		},
	}
}

// Create creates the basic directory structure
func (d *DirectoryStructure) Create() error {
	// Create base directory
	if err := utils.EnsureDirectoryExists(d.BaseDir); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectories
	for _, dir := range d.Directories {
		if err := utils.EnsureDirectoryExists(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// CreateEmptyFiles creates placeholder YAML files
func (d *DirectoryStructure) CreateEmptyFiles() error {
	// Common files
	commonFiles := []string{
		filepath.Join(d.CommonDir, "provider.yaml"),
		filepath.Join(d.CommonDir, "provider-config.yaml"),
	}

	// VPC files
	vpcFiles := []string{
		filepath.Join(d.VPCDir, "vpc.yaml"),
		filepath.Join(d.VPCDir, "subnets.yaml"),
		filepath.Join(d.VPCDir, "internet-gateway.yaml"),
		filepath.Join(d.VPCDir, "nat-gateway.yaml"),
		filepath.Join(d.VPCDir, "security-groups.yaml"),
	}

	// EKS files
	eksFiles := []string{
		filepath.Join(d.EKSDir, "cluster.yaml"),
		filepath.Join(d.EKSDir, "node-group.yaml"),
		filepath.Join(d.EKSDir, "roles.yaml"),
	}

	// Combine all files
	allFiles := append(commonFiles, vpcFiles...)
	allFiles = append(allFiles, eksFiles...)

	// Create empty files
	for _, file := range allFiles {
		if !utils.FileExists(file) {
			if err := utils.WriteToFile(file, ""); err != nil {
				return fmt.Errorf("failed to create file %s: %w", file, err)
			}
		}
	}

	return nil
}

// CreateGitignoreFile creates a .gitignore file with Crossplane-specific patterns
func (d *DirectoryStructure) CreateGitignoreFile() error {
	gitignoreContent := "# Crossplane files\n" +
		"*.crossplane.lock\n\n" +
		"# Sensitive data\n" +
		"*.credentials.yaml\n" +
		"*secret.yaml\n\n" +
		"# Temporary files\n" +
		"*.tmp\n" +
		"*.swp\n" +
		"*.swo\n\n" +
		"# Editors\n" +
		".idea/\n" +
		".vscode/\n" +
		"*~\n\n" +
		"# OS specific\n" +
		".DS_Store\n" +
		"Thumbs.db\n"

	gitignorePath := filepath.Join(d.BaseDir, ".gitignore")
	
	// Only write if it doesn't exist
	if !utils.FileExists(gitignorePath) {
		if err := utils.WriteToFile(gitignorePath, gitignoreContent); err != nil {
			return fmt.Errorf("failed to create .gitignore file: %w", err)
		}
	}
	
	return nil
}

// CreateKustomizationFiles creates all kustomization files
func (d *DirectoryStructure) CreateKustomizationFiles() error {
	// Main kustomization file
	kustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\n" +
		"kind: Kustomization\n\n" +
		"resources:\n" +
		"- base\n" +
		"- vpc\n" +
		"- eks\n"

	kustomizationPath := filepath.Join(d.BaseDir, "kustomization.yaml")
	
	// Only write if it doesn't exist
	if !utils.FileExists(kustomizationPath) {
		if err := utils.WriteToFile(kustomizationPath, kustomizationContent); err != nil {
			return fmt.Errorf("failed to create main kustomization.yaml file: %w", err)
		}
	}

	// VPC kustomization file
	vpcKustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\n" +
		"kind: Kustomization\n\n" +
		"resources:\n" +
		"- vpc.yaml\n" +
		"- subnets.yaml\n" +
		"- gateways.yaml\n"

	vpcKustomizationPath := filepath.Join(d.VPCDir, "kustomization.yaml")
	
	if !utils.FileExists(vpcKustomizationPath) {
		if err := utils.WriteToFile(vpcKustomizationPath, vpcKustomizationContent); err != nil {
			return fmt.Errorf("failed to create VPC kustomization.yaml file: %w", err)
		}
	}

	// EKS kustomization file
	eksKustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\n" +
		"kind: Kustomization\n\n" +
		"resources:\n" +
		"- cluster.yaml\n" +
		"- node-group.yaml\n" +
		"- roles.yaml\n"

	eksKustomizationPath := filepath.Join(d.EKSDir, "kustomization.yaml")
	
	if !utils.FileExists(eksKustomizationPath) {
		if err := utils.WriteToFile(eksKustomizationPath, eksKustomizationContent); err != nil {
			return fmt.Errorf("failed to create EKS kustomization.yaml file: %w", err)
		}
	}

	// Base kustomization file
	commonKustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\n" +
		"kind: Kustomization\n\n" +
		"resources:\n" +
		"- provider.yaml\n" +
		"- providerconfig.yaml\n"

	commonKustomizationPath := filepath.Join(d.CommonDir, "kustomization.yaml")
	
	if !utils.FileExists(commonKustomizationPath) {
		if err := utils.WriteToFile(commonKustomizationPath, commonKustomizationContent); err != nil {
			return fmt.Errorf("failed to create Common kustomization.yaml file: %w", err)
		}
	}
	
	return nil
}

// CreateREADME creates a README.md file with documentation
func (d *DirectoryStructure) CreateREADME() error {
	readmeContent := "# Crossplane Infrastructure\n\n" +
		"This directory contains Crossplane YAML definitions to deploy infrastructure on AWS.\n\n" +
		"## Directory Structure\n\n" +
		"- `common/`: Provider and configuration resources\n" +
		"- `vpc/`: VPC, subnets, and network resources\n" +
		"- `eks/`: EKS cluster and node group resources\n" +
		"- `s3/`: S3 bucket resources\n" +
		"- `rds/`: RDS database resources\n" +
		"- `ec2/`: EC2 instance resources\n\n" +
		"## Usage\n\n" +
		"```bash\n" +
		"# Apply all resources\n" +
		"kubectl apply -k .\n\n" +
		"# Apply specific component\n" +
		"kubectl apply -f vpc/\n" +
		"```\n\n" +
		"## Requirements\n\n" +
		"- Kubernetes cluster with Crossplane installed\n" +
		"- AWS Provider for Crossplane\n" +
		"- kubectl CLI tool\n" +
		"- kustomize\n"

	readmePath := filepath.Join(d.BaseDir, "README.md")
	
	// Only write if it doesn't exist
	if !utils.FileExists(readmePath) {
		if err := utils.WriteToFile(readmePath, readmeContent); err != nil {
			return fmt.Errorf("failed to create README.md file: %w", err)
		}
	}
	
	return nil
}

// CreateProviderConfig creates basic provider configuration files
func (d *DirectoryStructure) CreateProviderConfig(region string) error {
	if region == "" {
		region = "us-east-1"
	}
	
	providerContent := "apiVersion: pkg.crossplane.io/v1\n" +
		"kind: Provider\n" +
		"metadata:\n" +
		"  name: crossplane-provider-aws\n" +
		"spec:\n" +
		"  package: \"crossplane/provider-aws:v0.36.0\"\n" +
		"  controllerConfigRef:\n" +
		"    name: aws-config\n"

	providerConfigContent := "apiVersion: aws.crossplane.io/v1beta1\n" +
		"kind: ProviderConfig\n" +
		"metadata:\n" +
		"  name: aws-provider\n" +
		"spec:\n" +
		"  credentials:\n" +
		"    source: Secret\n" +
		"    secretRef:\n" +
		"      namespace: crossplane-system\n" +
		"      name: aws-credentials\n" +
		"      key: creds\n" +
		fmt.Sprintf("  region: %s\n", region)

	providerPath := filepath.Join(d.CommonDir, "provider.yaml")
	providerConfigPath := filepath.Join(d.CommonDir, "provider-config.yaml")
	
	// Create provider file if it doesn't exist
	if !utils.FileExists(providerPath) {
		if err := utils.WriteToFile(providerPath, providerContent); err != nil {
			return fmt.Errorf("failed to create provider.yaml file: %w", err)
		}
	}
	
	// Create provider config file if it doesn't exist
	if !utils.FileExists(providerConfigPath) {
		if err := utils.WriteToFile(providerConfigPath, providerConfigContent); err != nil {
			return fmt.Errorf("failed to create provider-config.yaml file: %w", err)
		}
	}
	
	return nil
}
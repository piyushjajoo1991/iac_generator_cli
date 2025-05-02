package crossplane

import (
	"fmt"
	"path/filepath"

	"github.com/riptano/iac_generator_cli/internal/utils"
)

// ProviderGenerator generates Crossplane AWS Provider configuration
type ProviderGenerator struct {
	baseDir  string
	commonDir string
}

// NewProviderGenerator creates a new Provider Generator
func NewProviderGenerator(baseDir string) *ProviderGenerator {
	return &ProviderGenerator{
		baseDir:  baseDir,
		commonDir: filepath.Join(baseDir, "base"),
	}
}

// GenerateProviderPackage generates a Crossplane Provider Package
func (g *ProviderGenerator) GenerateProviderPackage(name, version string) K8sObject {
	provider := NewK8sObject("pkg.crossplane.io/v1", "Provider", name)
	
	// Set provider package details
	provider.AddNestedSpecField([]string{"package"}, fmt.Sprintf("%s:%s", name, version))
	provider.AddNestedSpecField([]string{"controllerConfigRef", "name"}, "aws-config")
	
	return provider
}

// GenerateProviderConfig generates a Crossplane ProviderConfig for AWS
func (g *ProviderGenerator) GenerateProviderConfig(region string) K8sObject {
	if region == "" {
		region = "us-east-1"
	}
	
	config := NewK8sObject("aws.crossplane.io/v1beta1", "ProviderConfig", "aws-provider")
	
	// Set credentials source
	config.AddNestedSpecField([]string{"credentials", "source"}, "Secret")
	
	// Set secret reference
	config.AddNestedSpecField([]string{"credentials", "secretRef", "namespace"}, "crossplane-system")
	config.AddNestedSpecField([]string{"credentials", "secretRef", "name"}, "aws-credentials")
	config.AddNestedSpecField([]string{"credentials", "secretRef", "key"}, "creds")
	
	// Set region
	config.AddNestedSpecField([]string{"region"}, region)
	
	return config
}

// GenerateCommonResources generates common Crossplane resources
func (g *ProviderGenerator) GenerateCommonResources(region, accessKey, secretKey string) error {
	return g.GenerateAllProviderFiles(region, accessKey, secretKey)
}

// GenerateProviderFiles generates and writes provider YAML files
func (g *ProviderGenerator) GenerateProviderFiles(region string) error {
	// Create common directory if it doesn't exist
	if err := utils.EnsureDirectoryExists(g.commonDir); err != nil {
		return fmt.Errorf("failed to create provider directory: %w", err)
	}
	
	// Generate provider objects
	provider := g.GenerateProviderPackage("crossplane/provider-aws", "v0.36.0")
	providerConfig := g.GenerateProviderConfig(region)
	
	// Write provider files
	providerPath := filepath.Join(g.commonDir, "provider.yaml")
	if err := utils.WriteToFile(providerPath, provider.YAML()); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}
	
	configPath := filepath.Join(g.commonDir, "providerconfig.yaml")
	if err := utils.WriteToFile(configPath, providerConfig.YAML()); err != nil {
		return fmt.Errorf("failed to write provider config file: %w", err)
	}
	
	return nil
}

// GenerateAwsSecret generates a Kubernetes Secret manifest for AWS credentials
func (g *ProviderGenerator) GenerateAwsSecret(accessKey, secretKey string) K8sObject {
	secret := NewK8sObject("v1", "Secret", "aws-credentials")
	
	// Set namespace
	secret.AddMetadataField("namespace", "crossplane-system")
	
	// Add credential data
	credData := fmt.Sprintf("[default]\naws_access_key_id = %s\naws_secret_access_key = %s", 
		accessKey, secretKey)
	
	secret.AddField("data", map[string]string{
		"creds": utils.Base64Encode(credData),
	})
	
	return secret
}

// GenerateControllerConfig generates a ControllerConfig for AWS provider
func (g *ProviderGenerator) GenerateControllerConfig() K8sObject {
	config := NewK8sObject("pkg.crossplane.io/v1alpha1", "ControllerConfig", "aws-config")
	
	// Set metadata
	config.AddMetadataAnnotation("eks.amazonaws.com/role-arn", 
		"arn:aws:iam::ACCOUNT_ID:role/crossplane-provider-aws")
	
	// Set spec
	config.AddNestedSpecField([]string{"podSecurityContext", "fsGroup"}, "2000")
	
	return config
}

// GenerateProviderFiles generates all necessary provider files
func (g *ProviderGenerator) GenerateAllProviderFiles(region, accessKey, secretKey string) error {
	// Create common directory if it doesn't exist
	if err := utils.EnsureDirectoryExists(g.commonDir); err != nil {
		return fmt.Errorf("failed to create provider directory: %w", err)
	}
	
	// Generate provider objects
	provider := g.GenerateProviderPackage("crossplane/provider-aws", "v0.36.0")
	providerConfig := g.GenerateProviderConfig(region)
	secret := g.GenerateAwsSecret(accessKey, secretKey)
	controllerConfig := g.GenerateControllerConfig()
	
	// Write provider files
	providerPath := filepath.Join(g.commonDir, "provider.yaml")
	if err := utils.WriteToFile(providerPath, provider.YAML()); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}
	
	configPath := filepath.Join(g.commonDir, "providerconfig.yaml")
	if err := utils.WriteToFile(configPath, providerConfig.YAML()); err != nil {
		return fmt.Errorf("failed to write provider config file: %w", err)
	}
	
	secretPath := filepath.Join(g.commonDir, "aws-secret.yaml")
	if err := utils.WriteToFile(secretPath, secret.YAML()); err != nil {
		return fmt.Errorf("failed to write secret file: %w", err)
	}
	
	controllerPath := filepath.Join(g.commonDir, "controller-config.yaml")
	if err := utils.WriteToFile(controllerPath, controllerConfig.YAML()); err != nil {
		return fmt.Errorf("failed to write controller config file: %w", err)
	}
	
	return nil
}
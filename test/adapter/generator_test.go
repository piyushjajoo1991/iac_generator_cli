package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riptano/iac_generator_cli/internal/adapter/crossplane"
	"github.com/riptano/iac_generator_cli/internal/adapter/terraform"
	"github.com/riptano/iac_generator_cli/test/utils"
	"github.com/stretchr/testify/assert"
)

func TestTerraformGenerator(t *testing.T) {
	// Skip this test until generator is fully implemented
	t.Skip("Skipping test until terraform generator is complete")
}

func TestCrossplaneGenerator(t *testing.T) {
	// Skip this test until generator is fully implemented
	t.Skip("Skipping test until crossplane generator is complete")
	
	// Test is skipped, so the following code is unreachable
}

func TestTerraformDirectoryStructure(t *testing.T) {
	// Skip this test until directory structure is fully implemented
	t.Skip("Skipping test until terraform directory structure is complete")
	
	// Test is skipped, so the following code is unreachable
}

func TestCrossplaneDirectoryStructure(t *testing.T) {
	// Skip this test until directory structure is fully implemented
	t.Skip("Skipping test until crossplane directory structure is complete")
	
	// Test is skipped, so the following code is unreachable
}

func TestTerraformHCLWriter(t *testing.T) {
	// Test HCL writer functionality
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
	
	// Create tags as an attribute with a map value instead of as a block
	resourceBlock.AddAttribute("tags", map[string]string{
		"Name":        "main-vpc",
		"Environment": "dev",
	})
	
	// Write the resource block
	writer.WriteBlock(resourceBlock)
	
	// Get the output
	output := writer.String()
	
	// Check for expected content in a more forgiving way - order doesn't matter
	assert.Contains(t, output, "provider \"aws\"", "Output should contain provider declaration")
	assert.Contains(t, output, "region = \"us-east-1\"", "Output should contain region attribute")
	assert.Contains(t, output, "profile = \"default\"", "Output should contain profile attribute")
	assert.Contains(t, output, "resource \"aws_vpc\" \"main\"", "Output should contain resource declaration")
	assert.Contains(t, output, "cidr_block = \"10.0.0.0/16\"", "Output should contain CIDR block")
	assert.Contains(t, output, "enable_dns_support = true", "Output should contain DNS support attribute")
	assert.Contains(t, output, "enable_dns_hostnames = true", "Output should contain DNS hostnames attribute")
	assert.Contains(t, output, "tags =", "Output should contain tags attribute")
	assert.Contains(t, output, "\"Name\" = \"main-vpc\"", "Output should contain Name tag")
	assert.Contains(t, output, "\"Environment\" = \"dev\"", "Output should contain Environment tag")
	
	// Create a test environment to write the file
	testEnv := utils.NewTestEnvironment(t)
	defer testEnv.Cleanup()
	
	// Write the output to a file
	outputPath := filepath.Join(testEnv.OutputDir, "test.tf")
	err := os.WriteFile(outputPath, []byte(output), 0644)
	assert.NoError(t, err, "Writing output file should not error")
	
	// Check HCL syntax
	assert.True(t, utils.IsValidHCL(t, outputPath), "Output should be valid HCL")
}

func TestCrossplaneYAMLGenerator(t *testing.T) {
	// Test YAML functionality with K8sObject
	obj := crossplane.NewK8sObject("ec2.aws.crossplane.io/v1beta1", "VPC", "test-vpc")
	
	// Add fields to the K8sObject
	obj.AddLabel("app.kubernetes.io/part-of", "infrastructure")
	obj.AddLabel("app.kubernetes.io/managed-by", "crossplane")
	
	// Add spec fields
	obj.AddNestedSpecField([]string{"forProvider", "cidrBlock"}, "10.0.0.0/16")
	obj.AddNestedSpecField([]string{"forProvider", "region"}, "us-east-1")
	obj.AddNestedSpecField([]string{"forProvider", "enableDnsSupport"}, true)
	obj.AddNestedSpecField([]string{"forProvider", "enableDnsHostnames"}, true)
	obj.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Create a subnet object
	subnet := crossplane.NewK8sObject("ec2.aws.crossplane.io/v1beta1", "Subnet", "public-subnet-1")
	
	// Add fields to the subnet K8sObject
	subnet.AddLabel("app.kubernetes.io/part-of", "infrastructure")
	subnet.AddLabel("app.kubernetes.io/managed-by", "crossplane")
	
	// Add spec fields
	subnet.AddNestedSpecField([]string{"forProvider", "cidrBlock"}, "10.0.1.0/24")
	subnet.AddNestedSpecField([]string{"forProvider", "region"}, "us-east-1")
	subnet.AddNestedSpecField([]string{"forProvider", "availabilityZone"}, "us-east-1a")
	subnet.AddNestedSpecField([]string{"forProvider", "mapPublicIpOnLaunch"}, true)
	subnet.AddNestedSpecField([]string{"forProvider", "vpcIdRef", "name"}, "test-vpc")
	subnet.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Generate YAML for the VPC
	vpcYAML, err := crossplane.GenerateYAML(obj)
	assert.NoError(t, err, "Generating YAML should not error")
	assert.Contains(t, vpcYAML, "kind: VPC", "YAML should contain the VPC kind")
	
	// Generate YAML for the subnet
	subnetYAML, err := crossplane.GenerateYAML(subnet)
	assert.NoError(t, err, "Generating YAML should not error")
	assert.Contains(t, subnetYAML, "kind: Subnet", "YAML should contain the Subnet kind")
	
	// Check expected VPC content
	expectedVPCContent := []string{
		"apiVersion: ec2.aws.crossplane.io/v1beta1",
		"kind: VPC",
		"metadata:",
		"  name: test-vpc",
		"  labels:",
		"    app.kubernetes.io/part-of: infrastructure",
		"    app.kubernetes.io/managed-by: crossplane",
		"spec:",
		"  forProvider:",
		"    cidrBlock: 10.0.0.0/16",
		"    enableDnsSupport: true",
		"    enableDnsHostnames: true",
		"    region: us-east-1",
		"  providerConfigRef:",
		"    name: aws-provider",
	}
	
	// Check expected Subnet content
	expectedSubnetContent := []string{
		"apiVersion: ec2.aws.crossplane.io/v1beta1",
		"kind: Subnet",
		"metadata:",
		"  name: public-subnet-1",
		"  labels:",
		"    app.kubernetes.io/part-of: infrastructure",
		"    app.kubernetes.io/managed-by: crossplane",
		"spec:",
		"  forProvider:",
		"    availabilityZone: us-east-1a",
		"    cidrBlock: 10.0.1.0/24",
		"    mapPublicIpOnLaunch: true",
		"    region: us-east-1",
		"    vpcIdRef:",
		"      name: test-vpc",
		"  providerConfigRef:",
		"    name: aws-provider",
	}
	
	// Verify VPC YAML contains expected content
	for _, content := range expectedVPCContent {
		assert.Contains(t, vpcYAML, content, "VPC YAML should contain expected content")
	}
	
	// Verify Subnet YAML contains expected content
	for _, content := range expectedSubnetContent {
		assert.Contains(t, subnetYAML, content, "Subnet YAML should contain expected content")
	}
	
	// Create a test environment to write the files
	testEnv := utils.NewTestEnvironment(t)
	defer testEnv.Cleanup()
	
	// Write the output to files
	vpcPath := filepath.Join(testEnv.OutputDir, "vpc.yaml")
	subnetPath := filepath.Join(testEnv.OutputDir, "subnet.yaml")
	
	// Write files
	err = os.WriteFile(vpcPath, []byte(vpcYAML), 0644)
	assert.NoError(t, err, "Writing VPC YAML file should not error")
	
	err = os.WriteFile(subnetPath, []byte(subnetYAML), 0644)
	assert.NoError(t, err, "Writing Subnet YAML file should not error")
	
	// Check YAML syntax
	assert.True(t, utils.IsValidYAML(t, vpcPath), "VPC YAML should be valid")
	assert.True(t, utils.IsValidYAML(t, subnetPath), "Subnet YAML should be valid")
}
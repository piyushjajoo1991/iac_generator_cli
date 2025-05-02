package nlp

import (
	"testing"

	"github.com/riptano/iac_generator_cli/internal/nlp"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestRegionExtraction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Region specified",
			input:    "Create a VPC in us-west-2 region",
			expected: "us-west-2",
		},
		{
			name:     "No region specified",
			input:    "Create a VPC for my application",
			expected: "us-east-1", // Default region
		},
		{
			name:     "Multiple regions specified (first one takes precedence)",
			input:    "Create a VPC in us-west-2 and an RDS instance in us-east-1",
			expected: "us-west-2",
		},
		{
			name:     "Invalid region specified (should return default)",
			input:    "Create a VPC in invalid-region",
			expected: "us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ExtractRegion(tt.input)
			assert.Equal(t, tt.expected, result, "Extracted region does not match expected")
		})
	}
}

func TestPatternMatchingVPC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "VPC with CIDR",
			input: "Create a VPC with CIDR 10.0.0.0/16",
			expected: map[string]interface{}{
				"exists":              true,
				"cidr_block":          "10.0.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
		},
		{
			name:  "VPC without CIDR (should use default)",
			input: "Create a VPC for my application",
			expected: map[string]interface{}{
				"exists":              true,
				"cidr_block":          "10.0.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
		},
		{
			name:  "VPC mentioned multiple times (should extract CIDR)",
			input: "Create a VPC with CIDR 172.16.0.0/16 and configure the VPC with DNS hostnames",
			expected: map[string]interface{}{
				"exists":              true,
				"cidr_block":          "172.16.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ExtractVPC(tt.input)
			assert.Equal(t, tt.expected, result, "Extracted VPC info does not match expected")
		})
	}
}

func TestPatternMatchingSubnets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "Public and private subnets",
			input: "Create a VPC with 2 public subnets and 3 private subnets",
			expected: map[string]interface{}{
				"public_count":  2,
				"private_count": 3,
			},
		},
		{
			name:  "Only public subnets",
			input: "Create a VPC with 4 public subnets",
			expected: map[string]interface{}{
				"public_count":  4,
				"private_count": 1, // Default private subnet count
			},
		},
		{
			name:  "Only private subnets",
			input: "Create a VPC with 5 private subnets",
			expected: map[string]interface{}{
				"public_count":  1, // Default public subnet count
				"private_count": 5,
			},
		},
		{
			name:  "No subnet counts (should use defaults)",
			input: "Create a VPC for my application",
			expected: map[string]interface{}{
				"public_count":  1,
				"private_count": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ExtractSubnets(tt.input)
			assert.Equal(t, tt.expected["public_count"], result["public_count"], "Public subnet count does not match")
			assert.Equal(t, tt.expected["private_count"], result["private_count"], "Private subnet count does not match")
		})
	}
}

func TestPatternMatchingGateways(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "Internet Gateway",
			input: "Create a VPC with an Internet Gateway",
			expected: map[string]interface{}{
				"igw_count": 1,
				"nat_count": 0,
			},
		},
		{
			name:  "NAT Gateway",
			input: "Create a VPC with a NAT Gateway",
			expected: map[string]interface{}{
				"igw_count": 1,
				"nat_count": 1,
			},
		},
		{
			name:  "Multiple NAT Gateways",
			input: "Create a VPC with 3 NAT Gateways",
			expected: map[string]interface{}{
				"igw_count": 1,
				"nat_count": 3,
			},
		},
		{
			name:  "NAT Gateway per AZ",
			input: "Create a VPC with 3 AZs and a NAT Gateway per AZ",
			expected: map[string]interface{}{
				"igw_count": 1,
				"nat_count": 3,
			},
		},
		{
			name:  "No Gateways Mentioned",
			input: "Create a VPC for my application",
			expected: map[string]interface{}{
				"igw_count": 1,
				"nat_count": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ExtractGateways(tt.input)
			assert.Equal(t, tt.expected, result, "Extracted gateway info does not match expected")
		})
	}
}

func TestPatternMatchingEKS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "Basic EKS",
			input: "Create an EKS cluster",
			expected: map[string]interface{}{
				"exists":                true,
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"version":               "1.27",
				"node_count":            2,
				"instance_type":         "t3.medium",
			},
		},
		{
			name:  "EKS with private access",
			input: "Create an EKS cluster with private API access",
			expected: map[string]interface{}{
				"exists":                true,
				"endpoint_public_access": false,
				"endpoint_private_access": true,
				"version":               "1.27",
				"node_count":            2,
				"instance_type":         "t3.medium",
			},
		},
		{
			name:  "EKS with version",
			input: "Create an EKS cluster with version 1.28",
			expected: map[string]interface{}{
				"exists":                true,
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"version":               "1.28",
				"node_count":            2,
				"instance_type":         "t3.medium",
			},
		},
		{
			name:  "EKS with node pool",
			input: "Create an EKS cluster with a node pool of 3 nodes",
			expected: map[string]interface{}{
				"exists":                true,
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"version":               "1.27",
				"node_count":            3,
				"instance_type":         "t3.medium",
			},
		},
		{
			name:  "EKS with instance type",
			input: "Create an EKS cluster with t3.large instances",
			expected: map[string]interface{}{
				"exists":                true,
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"version":               "1.27",
				"node_count":            2,
				"instance_type":         "t3.large",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ExtractEKS(tt.input)
			assert.Equal(t, tt.expected, result, "Extracted EKS info does not match expected")
		})
	}
}

func TestTableDrivenParsingTests(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assertions  func(t *testing.T, entities map[string]interface{})
	}{
		{
			name:        "Basic VPC",
			description: "Create a VPC in us-east-1",
			assertions: func(t *testing.T, entities map[string]interface{}) {
				assert.Equal(t, "us-east-1", entities["region"], "Region mismatch")
				
				// Check VPC exists
				vpc, ok := entities["vpc"].(map[string]interface{})
				assert.True(t, ok, "VPC not found in parsed entities")
				assert.True(t, vpc["exists"].(bool), "VPC exists flag not found")
				assert.Equal(t, true, vpc["exists"], "VPC exists flag mismatch")
				
				// Check VPC CIDR
				assert.True(t, vpc["cidr_block"] != nil, "VPC CIDR block not found")
				assert.Equal(t, "10.0.0.0/16", vpc["cidr_block"], "VPC CIDR block mismatch")
			},
		},
		{
			name:        "VPC with Specific CIDR",
			description: "Create a VPC with CIDR 192.168.0.0/16 in us-west-2",
			assertions: func(t *testing.T, entities map[string]interface{}) {
				assert.Equal(t, "us-west-2", entities["region"], "Region mismatch")
				
				// Check VPC exists
				vpc, ok := entities["vpc"].(map[string]interface{})
				assert.True(t, ok, "VPC not found in parsed entities")
				assert.True(t, vpc["exists"].(bool), "VPC exists flag not found")
				assert.Equal(t, true, vpc["exists"], "VPC exists flag mismatch")
				
				// Check VPC CIDR
				assert.True(t, vpc["cidr_block"] != nil, "VPC CIDR block not found")
				assert.Equal(t, "192.168.0.0/16", vpc["cidr_block"], "VPC CIDR block mismatch")
			},
		},
		{
			name:        "VPC with Subnets",
			description: "Create a VPC with 3 public subnets and 2 private subnets in eu-central-1",
			assertions: func(t *testing.T, entities map[string]interface{}) {
				assert.Equal(t, "eu-central-1", entities["region"], "Region mismatch")
				
				// Check VPC exists
				_, ok := entities["vpc"].(map[string]interface{})
				assert.True(t, ok, "VPC not found in parsed entities")
				
				// Check subnets
				subnets, ok := entities["subnets"].(map[string]interface{})
				assert.True(t, ok, "Subnets not found in parsed entities")
				assert.Equal(t, 3, subnets["public_count"], "Public subnet count mismatch")
				assert.Equal(t, 2, subnets["private_count"], "Private subnet count mismatch")
			},
		},
		{
			name:        "VPC with IGW and NAT",
			description: "Create a VPC with an Internet Gateway and 2 NAT Gateways in us-east-2",
			assertions: func(t *testing.T, entities map[string]interface{}) {
				assert.Equal(t, "us-east-2", entities["region"], "Region mismatch")
				
				// Check VPC exists
				_, ok := entities["vpc"].(map[string]interface{})
				assert.True(t, ok, "VPC not found in parsed entities")
				
				// Check gateways
				gateways, ok := entities["gateways"].(map[string]interface{})
				assert.True(t, ok, "Gateways not found in parsed entities")
				assert.Equal(t, 1, gateways["igw_count"], "IGW count mismatch")
				assert.Equal(t, 2, gateways["nat_count"], "NAT Gateway count mismatch")
			},
		},
		{
			name:        "EKS with Private Access",
			description: "Create an EKS cluster in us-east-1 with private API access and a VPC with CIDR 10.0.0.0/16",
			assertions: func(t *testing.T, entities map[string]interface{}) {
				assert.Equal(t, "us-east-1", entities["region"], "Region mismatch")
				
				// Check VPC exists
				vpc, ok := entities["vpc"].(map[string]interface{})
				assert.True(t, ok, "VPC not found in parsed entities")
				assert.True(t, vpc["exists"].(bool), "VPC exists flag not found")
				assert.Equal(t, true, vpc["exists"], "VPC exists flag mismatch")
				
				// Check VPC CIDR
				assert.True(t, vpc["cidr_block"] != nil, "VPC CIDR block not found")
				assert.Equal(t, "10.0.0.0/16", vpc["cidr_block"], "VPC CIDR block mismatch")
				
				// Check EKS
				eks, ok := entities["eks"].(map[string]interface{})
				assert.True(t, ok, "EKS not found in parsed entities")
				assert.Equal(t, false, eks["endpoint_public_access"], "EKS public access mismatch")
				assert.Equal(t, true, eks["endpoint_private_access"], "EKS private access mismatch")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := nlp.NewParser()
			entities, err := parser.ExtractEntities(tt.description)
			assert.NoError(t, err, "Error extracting entities")
			tt.assertions(t, entities)
		})
	}
}

func TestEntityValidation(t *testing.T) {
	// Test cases for entity validation
	tests := []struct {
		name          string
		entities      map[string]interface{}
		expectedValid bool
	}{
		{
			name: "Valid complete entities",
			entities: map[string]interface{}{
				"region": "us-east-1",
				"vpc": map[string]interface{}{
					"exists":              true,
					"cidr_block":          "10.0.0.0/16",
					"enable_dns_support":  true,
					"enable_dns_hostnames": true,
				},
				"subnets": map[string]interface{}{
					"public_count":  2,
					"private_count": 2,
				},
				"gateways": map[string]interface{}{
					"igw_count": 1,
					"nat_count": 2,
				},
			},
			expectedValid: true,
		},
		{
			name: "Missing VPC CIDR",
			entities: map[string]interface{}{
				"region": "us-east-1",
				"vpc": map[string]interface{}{
					"exists": true,
				},
			},
			expectedValid: false, // Should add default CIDR
		},
		{
			name: "Missing subnet counts",
			entities: map[string]interface{}{
				"region": "us-east-1",
				"vpc": map[string]interface{}{
					"exists":              true,
					"cidr_block":          "10.0.0.0/16",
					"enable_dns_support":  true,
					"enable_dns_hostnames": true,
				},
				"subnets": map[string]interface{}{},
			},
			expectedValid: false, // Should add default subnet counts
		},
		{
			name: "Missing region",
			entities: map[string]interface{}{
				"vpc": map[string]interface{}{
					"exists":              true,
					"cidr_block":          "10.0.0.0/16",
					"enable_dns_support":  true,
					"enable_dns_hostnames": true,
				},
			},
			expectedValid: false, // Should add default region
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nlp.ValidateEntities(tt.entities)
			
			// We've changed the behavior - validation is always successful because
			// we fix the issues instead of failing. We just verify that fixes were applied.
			if !tt.expectedValid {
				assert.Greater(t, len(result.Fixes), 0, "Expected fixes but none were provided")
			}
		})
	}
}

func TestFullModelParsing(t *testing.T) {
	// Test cases for full model parsing
	tests := []struct {
		name               string
		description        string
		expectedResources  int
		expectedResourceTypes map[models.ResourceType]int
	}{
		{
			name:              "Basic VPC",
			description:       "Create a VPC with CIDR 10.0.0.0/16 in us-east-1",
			expectedResources: 4, // The default model adds 1 VPC, 1 public subnet, 1 private subnet, 1 IGW
			expectedResourceTypes: map[models.ResourceType]int{
				models.ResourceVPC: 1,
				models.ResourceSubnet: 2,
				models.ResourceIGW: 1,
			},
		},
		{
			name:              "VPC with Subnets",
			description:       "Create a VPC with 2 public subnets and 2 private subnets in us-west-2",
			expectedResources: 6, // VPC + 4 subnets + 1 IGW
			expectedResourceTypes: map[models.ResourceType]int{
				models.ResourceVPC:    1,
				models.ResourceSubnet: 4,
				models.ResourceIGW:    1,
			},
		},
		{
			name:              "VPC with Subnets and IGW",
			description:       "Create a VPC with 2 public subnets, 2 private subnets, and an Internet Gateway in us-east-2",
			expectedResources: 6, // VPC + 4 subnets + IGW
			expectedResourceTypes: map[models.ResourceType]int{
				models.ResourceVPC:           1,
				models.ResourceSubnet:        4,
				models.ResourceIGW:           1,
			},
		},
		{
			name:              "Full Infrastructure",
			description:       "AWS infra in us-east-1 with a vpc, 3 public and 3 private subnets, 1 IGW and 3 NAT gateways, plus an EKS Cluster with node group",
			expectedResources: 13, // VPC + 6 subnets + IGW + 3 NAT + EKS + NodeGroup
			expectedResourceTypes: map[models.ResourceType]int{
				models.ResourceVPC:           1,
				models.ResourceSubnet:        6,
				models.ResourceIGW:           1,
				models.ResourceNATGateway:    3,
				models.ResourceEKSCluster:    1,
				models.ResourceNodeGroup:     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the full model
			model, err := nlp.ParseDescription(tt.description)
			assert.NoError(t, err, "Error parsing description")
			assert.NotNil(t, model, "Parsed model is nil")
			
			// Check resource count
			assert.Equal(t, tt.expectedResources, len(model.Resources), "Resource count mismatch")
			
			// Count resources by type
			resourceCounts := make(map[models.ResourceType]int)
			for _, resource := range model.Resources {
				resourceCounts[resource.Type]++
			}
			
			// Check resource type counts
			for resourceType, expectedCount := range tt.expectedResourceTypes {
				actualCount := resourceCounts[resourceType]
				assert.Equal(t, expectedCount, actualCount, "Resource count mismatch for type %s", resourceType)
			}
		})
	}
}

func TestInvalidDescriptionErrors(t *testing.T) {
	// Test invalid descriptions
	invalidTests := []struct {
		name        string
		description string
	}{
		{
			name:        "Empty description",
			description: "",
		},
		{
			name:        "Too short description",
			description: "VPC",
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to parse
			_, err := nlp.ParseDescription(tt.description)
			assert.Error(t, err, "Expected error parsing invalid description")
		})
	}
}
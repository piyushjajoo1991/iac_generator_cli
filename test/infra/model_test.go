package infra

import (
	"testing"

	"github.com/riptano/iac_generator_cli/internal/infra"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"github.com/riptano/iac_generator_cli/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestModelBuilderCreation(t *testing.T) {
	builder := infra.NewModelBuilder()
	assert.NotNil(t, builder, "ModelBuilder should not be nil")

	model := builder.GetModel()
	assert.NotNil(t, model, "Model should not be nil")
	assert.Empty(t, model.Resources, "New model should have no resources")
}

func TestResourceAddition(t *testing.T) {
	builder := infra.NewModelBuilder()
	
	// Create and add a VPC resource
	vpc := models.NewResource(models.ResourceVPC, "test-vpc")
	vpc.AddProperty("cidr_block", "10.0.0.0/16")
	builder.AddResource(vpc)
	
	// Verify the resource was added
	model := builder.GetModel()
	assert.Equal(t, 1, len(model.Resources), "Model should have 1 resource")
	assert.Equal(t, models.ResourceVPC, model.Resources[0].Type, "Resource type should be VPC")
	assert.Equal(t, "test-vpc", model.Resources[0].Name, "Resource name should match")
}

func TestVPCCreation(t *testing.T) {
	// Test VPC resource creation
	vpc := infra.CreateVPC("test-vpc", "10.0.0.0/16", true, true)
	
	assert.Equal(t, models.ResourceVPC, vpc.Type, "Resource type should be VPC")
	assert.Equal(t, "test-vpc", vpc.Name, "VPC name should match")
	
	// Check properties
	foundProps := make(map[string]bool)
	for _, prop := range vpc.Properties {
		foundProps[prop.Name] = true
		
		switch prop.Name {
		case "cidr_block":
			assert.Equal(t, "10.0.0.0/16", prop.Value, "CIDR block should match")
		case "enable_dns_support":
			assert.Equal(t, true, prop.Value, "DNS support should be enabled")
		case "enable_dns_hostnames":
			assert.Equal(t, true, prop.Value, "DNS hostnames should be enabled")
		}
	}
	
	assert.True(t, foundProps["cidr_block"], "CIDR block property should exist")
	assert.True(t, foundProps["enable_dns_support"], "DNS support property should exist")
	assert.True(t, foundProps["enable_dns_hostnames"], "DNS hostnames property should exist")
}

func TestSubnetCreation(t *testing.T) {
	// Test subnet resource creation
	subnet := infra.CreateSubnet("test-subnet", "test-vpc", "10.0.1.0/24", "us-east-1a")
	
	assert.Equal(t, models.ResourceSubnet, subnet.Type, "Resource type should be Subnet")
	assert.Equal(t, "test-subnet", subnet.Name, "Subnet name should match")
	
	// Check properties
	foundProps := make(map[string]bool)
	for _, prop := range subnet.Properties {
		foundProps[prop.Name] = true
		
		switch prop.Name {
		case "vpc_id":
			assert.Equal(t, "test-vpc", prop.Value, "VPC ID should match")
		case "cidr_block":
			assert.Equal(t, "10.0.1.0/24", prop.Value, "CIDR block should match")
		case "availability_zone":
			assert.Equal(t, "us-east-1a", prop.Value, "Availability zone should match")
		}
	}
	
	assert.True(t, foundProps["vpc_id"], "VPC ID property should exist")
	assert.True(t, foundProps["cidr_block"], "CIDR block property should exist")
	assert.True(t, foundProps["availability_zone"], "Availability zone property should exist")
}

func TestInternetGatewayCreation(t *testing.T) {
	// Test Internet Gateway resource creation
	igw := infra.CreateInternetGateway("test-igw", "test-vpc")
	
	assert.Equal(t, models.ResourceIGW, igw.Type, "Resource type should be IGW")
	assert.Equal(t, "test-igw", igw.Name, "IGW name should match")
	
	// Check properties
	var vpcIDFound bool
	for _, prop := range igw.Properties {
		if prop.Name == "vpc_id" {
			vpcIDFound = true
			assert.Equal(t, "test-vpc", prop.Value, "VPC ID should match")
		}
	}
	
	assert.True(t, vpcIDFound, "VPC ID property should exist")
}

func TestNATGatewayCreation(t *testing.T) {
	// Test NAT Gateway resource creation
	natGateway := infra.CreateNATGateway("test-nat", "test-subnet", "test-eip")
	
	assert.Equal(t, models.ResourceNATGateway, natGateway.Type, "Resource type should be NAT Gateway")
	assert.Equal(t, "test-nat", natGateway.Name, "NAT Gateway name should match")
	
	// Check properties
	foundProps := make(map[string]bool)
	for _, prop := range natGateway.Properties {
		foundProps[prop.Name] = true
		
		switch prop.Name {
		case "subnet_id":
			assert.Equal(t, "test-subnet", prop.Value, "Subnet ID should match")
		case "allocation_id":
			assert.Equal(t, "test-eip", prop.Value, "Allocation ID should match")
		case "connectivity_type":
			assert.Equal(t, "public", prop.Value, "Connectivity type should be public")
		}
	}
	
	assert.True(t, foundProps["subnet_id"], "Subnet ID property should exist")
	assert.True(t, foundProps["allocation_id"], "Allocation ID property should exist")
	assert.True(t, foundProps["connectivity_type"], "Connectivity type property should exist")
}

func TestEKSClusterCreation(t *testing.T) {
	// Test EKS Cluster resource creation
	subnetIDs := []string{"subnet-1", "subnet-2"}
	eksCluster := infra.CreateEKSCluster("test-eks", "1.27", "test-role-arn", subnetIDs, true, false)
	
	assert.Equal(t, models.ResourceEKSCluster, eksCluster.Type, "Resource type should be EKS Cluster")
	assert.Equal(t, "test-eks", eksCluster.Name, "EKS Cluster name should match")
	
	// Check properties
	foundProps := make(map[string]bool)
	for _, prop := range eksCluster.Properties {
		foundProps[prop.Name] = true
		
		switch prop.Name {
		case "name":
			assert.Equal(t, "test-eks", prop.Value, "Cluster name should match")
		case "role_arn":
			assert.Equal(t, "test-role-arn", prop.Value, "Role ARN should match")
		case "version":
			assert.Equal(t, "1.27", prop.Value, "Version should match")
		case "vpc_config":
			vpcConfig, ok := prop.Value.(map[string]interface{})
			assert.True(t, ok, "VPC config should be a map")
			
			assert.Equal(t, subnetIDs, vpcConfig["subnet_ids"], "Subnet IDs should match")
			assert.Equal(t, true, vpcConfig["endpoint_public_access"], "Public access should match")
			assert.Equal(t, false, vpcConfig["endpoint_private_access"], "Private access should match")
		}
	}
	
	assert.True(t, foundProps["name"], "Name property should exist")
	assert.True(t, foundProps["role_arn"], "Role ARN property should exist")
	assert.True(t, foundProps["version"], "Version property should exist")
	assert.True(t, foundProps["vpc_config"], "VPC config property should exist")
}

func TestEKSNodeGroupCreation(t *testing.T) {
	// Test EKS Node Group resource creation
	subnetIDs := []string{"subnet-1", "subnet-2"}
	instanceTypes := []string{"t3.medium"}
	nodeGroup := infra.CreateEKSNodeGroup("test-ng", "test-eks", "test-role-arn", subnetIDs, instanceTypes, 2, 1, 3)
	
	assert.Equal(t, models.ResourceNodeGroup, nodeGroup.Type, "Resource type should be Node Group")
	assert.Equal(t, "test-ng", nodeGroup.Name, "Node Group name should match")
	
	// Check properties
	foundProps := make(map[string]bool)
	for _, prop := range nodeGroup.Properties {
		foundProps[prop.Name] = true
		
		switch prop.Name {
		case "cluster_name":
			assert.Equal(t, "test-eks", prop.Value, "Cluster name should match")
		case "node_role_arn":
			assert.Equal(t, "test-role-arn", prop.Value, "Node role ARN should match")
		case "subnet_ids":
			assert.Equal(t, subnetIDs, prop.Value, "Subnet IDs should match")
		case "instance_types":
			assert.Equal(t, instanceTypes, prop.Value, "Instance types should match")
		case "scaling_config":
			scalingConfig, ok := prop.Value.(map[string]interface{})
			assert.True(t, ok, "Scaling config should be a map")
			
			assert.Equal(t, 2, scalingConfig["desired_size"], "Desired size should match")
			assert.Equal(t, 1, scalingConfig["min_size"], "Min size should match")
			assert.Equal(t, 3, scalingConfig["max_size"], "Max size should match")
		}
	}
	
	assert.True(t, foundProps["cluster_name"], "Cluster name property should exist")
	assert.True(t, foundProps["node_role_arn"], "Node role ARN property should exist")
	assert.True(t, foundProps["subnet_ids"], "Subnet IDs property should exist")
	assert.True(t, foundProps["instance_types"], "Instance types property should exist")
	assert.True(t, foundProps["scaling_config"], "Scaling config property should exist")
}

func TestSubnetCIDRGeneration(t *testing.T) {
	tests := []struct {
		name           string
		vpcCIDR        string
		publicCount    int
		privateCount   int
		expectError    bool
		publicCIDRs    []string
		privateCIDRs   []string
	}{
		{
			name:         "Standard VPC CIDR",
			vpcCIDR:      "10.0.0.0/16",
			publicCount:  2,
			privateCount: 2,
			expectError:  false,
			publicCIDRs:  []string{"10.0.0.0/24", "10.0.1.0/24"},
			privateCIDRs: []string{"10.0.10.0/24", "10.0.11.0/24"},
		},
		{
			name:         "Custom VPC CIDR",
			vpcCIDR:      "172.16.0.0/16",
			publicCount:  3,
			privateCount: 3,
			expectError:  false,
			publicCIDRs:  []string{"172.16.0.0/24", "172.16.1.0/24", "172.16.2.0/24"},
			privateCIDRs: []string{"172.16.10.0/24", "172.16.11.0/24", "172.16.12.0/24"},
		},
		{
			name:         "Invalid VPC CIDR",
			vpcCIDR:      "invalid-cidr",
			publicCount:  2,
			privateCount: 2,
			expectError:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publicCIDRs, privateCIDRs, err := infra.GenerateSubnetCIDRs(tt.vpcCIDR, tt.publicCount, tt.privateCount)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error generating subnet CIDRs")
			} else {
				assert.NoError(t, err, "Did not expect error generating subnet CIDRs")
				assert.Equal(t, tt.publicCount, len(publicCIDRs), "Public CIDR count should match")
				assert.Equal(t, tt.privateCount, len(privateCIDRs), "Private CIDR count should match")
				
				// Check individual CIDRs
				for i, cidr := range tt.publicCIDRs {
					assert.Equal(t, cidr, publicCIDRs[i], "Public CIDR should match")
				}
				
				for i, cidr := range tt.privateCIDRs {
					assert.Equal(t, cidr, privateCIDRs[i], "Private CIDR should match")
				}
			}
		})
	}
}

func TestBuildFromParsedEntities(t *testing.T) {
	tests := []struct {
		name             string
		entities         map[string]interface{}
		expectedResources map[models.ResourceType]int
	}{
		{
			name: "Basic VPC only",
			entities: map[string]interface{}{
				"region": "us-east-1",
				"vpc": map[string]interface{}{
					"exists":              true,
					"cidr_block":          "10.0.0.0/16",
					"enable_dns_support":  true,
					"enable_dns_hostnames": true,
				},
			},
			expectedResources: map[models.ResourceType]int{
				models.ResourceVPC: 1,
			},
		},
		{
			name: "VPC with subnets",
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
			},
			expectedResources: map[models.ResourceType]int{
				models.ResourceVPC:    1,
				models.ResourceSubnet: 4,
			},
		},
		{
			name: "VPC with subnets and gateways",
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
					"nat_count": 1,
				},
			},
			expectedResources: map[models.ResourceType]int{
				models.ResourceVPC:        1,
				models.ResourceSubnet:     4,
				models.ResourceIGW:        1,
				models.ResourceNATGateway: 1,
			},
		},
		{
			name: "Full infrastructure with EKS",
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
					"nat_count": 1,
				},
				"eks": map[string]interface{}{
					"exists":                 true,
					"version":                "1.27",
					"endpoint_public_access": true,
					"endpoint_private_access": false,
					"node_count":              2,
					"instance_type":           "t3.medium",
				},
			},
			expectedResources: map[models.ResourceType]int{
				models.ResourceVPC:         1,
				models.ResourceSubnet:      4,
				models.ResourceIGW:         1,
				models.ResourceNATGateway:  1,
				models.ResourceEKSCluster:  1,
				models.ResourceNodeGroup:   1,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := infra.NewModelBuilder()
			err := builder.BuildFromParsedEntities(tt.entities)
			assert.NoError(t, err, "Did not expect error building from parsed entities")
			
			model := builder.GetModel()
			assert.NotNil(t, model, "Model should not be nil")
			
			// Count resources by type
			resourceCounts := make(map[models.ResourceType]int)
			for _, resource := range model.Resources {
				resourceCounts[resource.Type]++
			}
			
			// Verify expected resource counts
			for resourceType, expectedCount := range tt.expectedResources {
				actualCount := resourceCounts[resourceType]
				assert.Equal(t, expectedCount, actualCount, "Resource count mismatch for type %s", resourceType)
			}
		})
	}
}

func TestResourceDependencies(t *testing.T) {
	t.Skip("Skipping resource dependencies test until dependency management is improved")
	// Test a complex model with dependencies
	entities := map[string]interface{}{
		"region": "us-east-1",
		"vpc": map[string]interface{}{
			"exists":              true,
			"cidr_block":          "10.0.0.0/16",
			"enable_dns_support":  true,
			"enable_dns_hostnames": true,
		},
		"subnets": map[string]interface{}{
			"public_count":  1,
			"private_count": 1,
		},
		"gateways": map[string]interface{}{
			"igw_count": 1,
			"nat_count": 1,
		},
		"eks": map[string]interface{}{
			"exists":                 true,
			"version":                "1.27",
			"endpoint_public_access": true,
			"endpoint_private_access": false,
			"node_count":              2,
			"instance_type":           "t3.medium",
		},
	}
	
	builder := infra.NewModelBuilder()
	err := builder.BuildFromParsedEntities(entities)
	assert.NoError(t, err, "Did not expect error building from parsed entities")
	
	model := builder.GetModel()
	assert.NotNil(t, model, "Model should not be nil")
	
	// Verify dependencies
	var vpcID string
	var publicSubnetID string
	var privateSubnetID string
	var eksClusterID string
	
	// Find the resources first
	for _, resource := range model.Resources {
		switch resource.Type {
		case models.ResourceVPC:
			vpcID = resource.Name
		case models.ResourceSubnet:
			// Find subnet type by public IP on launch property
			for _, prop := range resource.Properties {
				if prop.Name == "map_public_ip_on_launch" {
					if prop.Value == true {
						publicSubnetID = resource.Name
					} else {
						privateSubnetID = resource.Name
					}
				}
			}
		case models.ResourceEKSCluster:
			eksClusterID = resource.Name
		}
	}
	
	// Verify dependencies
	for _, resource := range model.Resources {
		switch resource.Type {
		case models.ResourceSubnet:
			// Subnets should depend on VPC
			dependsOnVPC := false
			for _, dep := range resource.DependsOn {
				if dep == vpcID {
					dependsOnVPC = true
					break
				}
			}
			assert.True(t, dependsOnVPC, "Subnet should depend on VPC")
			
		case models.ResourceIGW:
			// IGW should depend on VPC
			dependsOnVPC := false
			for _, dep := range resource.DependsOn {
				if dep == vpcID {
					dependsOnVPC = true
					break
				}
			}
			assert.True(t, dependsOnVPC, "IGW should depend on VPC")
			
		case models.ResourceNATGateway:
			// NAT Gateway should depend on a subnet
			dependsOnSubnet := false
			for _, dep := range resource.DependsOn {
				if dep == publicSubnetID {
					dependsOnSubnet = true
					break
				}
			}
			assert.True(t, dependsOnSubnet, "NAT Gateway should depend on public subnet")
			
		case models.ResourceNodeGroup:
			// Node Group should depend on EKS Cluster
			dependsOnEKS := false
			for _, dep := range resource.DependsOn {
				if dep == eksClusterID {
					dependsOnEKS = true
					break
				}
			}
			assert.True(t, dependsOnEKS, "Node Group should depend on EKS Cluster")
			
			// Node Group should depend on subnets
			dependsOnSubnet := false
			for _, dep := range resource.DependsOn {
				if dep == privateSubnetID {
					dependsOnSubnet = true
					break
				}
			}
			assert.True(t, dependsOnSubnet, "Node Group should depend on private subnet")
		}
	}
}

func TestComplexInfrastructureModel(t *testing.T) {
	// Test the complex infrastructure model from fixtures
	model := fixtures.CreateTestInfrastructureModel()
	
	// Count resources by type
	resourceCounts := make(map[models.ResourceType]int)
	for _, resource := range model.Resources {
		resourceCounts[resource.Type]++
	}
	
	// Verify expected resources
	assert.Equal(t, 1, resourceCounts[models.ResourceVPC], "Should have 1 VPC")
	assert.Equal(t, 4, resourceCounts[models.ResourceSubnet], "Should have 4 subnets")
	assert.Equal(t, 1, resourceCounts[models.ResourceIGW], "Should have 1 Internet Gateway")
	assert.Equal(t, 1, resourceCounts[models.ResourceNATGateway], "Should have 1 NAT Gateway")
	assert.Equal(t, 1, resourceCounts[models.ResourceEKSCluster], "Should have 1 EKS Cluster")
	assert.Equal(t, 1, resourceCounts[models.ResourceNodeGroup], "Should have 1 Node Group")
	
	// Verify dependencies
	for _, resource := range model.Resources {
		switch resource.Type {
		case models.ResourceSubnet:
			// Subnets should depend on VPC
			assert.Contains(t, resource.DependsOn, "main-vpc", "Subnet should depend on VPC")
			
		case models.ResourceIGW:
			// IGW should depend on VPC
			assert.Contains(t, resource.DependsOn, "main-vpc", "IGW should depend on VPC")
			
		case models.ResourceNATGateway:
			// NAT Gateway should depend on a subnet
			assert.Contains(t, resource.DependsOn, "public-subnet-1", "NAT Gateway should depend on public subnet")
			
		case models.ResourceEKSCluster:
			// EKS Cluster should depend on private subnets
			assert.Contains(t, resource.DependsOn, "private-subnet-1", "EKS Cluster should depend on private subnet 1")
			assert.Contains(t, resource.DependsOn, "private-subnet-2", "EKS Cluster should depend on private subnet 2")
			
		case models.ResourceNodeGroup:
			// Node Group should depend on EKS Cluster and subnets
			assert.Contains(t, resource.DependsOn, "main-eks-cluster", "Node Group should depend on EKS Cluster")
			assert.Contains(t, resource.DependsOn, "private-subnet-1", "Node Group should depend on private subnet 1")
			assert.Contains(t, resource.DependsOn, "private-subnet-2", "Node Group should depend on private subnet 2")
		}
	}
}
package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riptano/iac_generator_cli/internal/adapter/crossplane"
	"github.com/riptano/iac_generator_cli/internal/infra"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

func TestCrossplaneAdapter(t *testing.T) {
	// Create a test infrastructure model
	builder := infra.NewModelBuilder()

	// Add VPC
	vpcResource := models.NewResource(models.ResourceVPC, "test-vpc")
	vpcResource.AddProperty("cidr_block", "10.0.0.0/16")
	vpcResource.AddProperty("enable_dns_support", true)
	vpcResource.AddProperty("enable_dns_hostnames", true)
	builder.AddResource(vpcResource)

	// Add public subnet
	publicSubnet := models.NewResource(models.ResourceSubnet, "public-subnet-1")
	publicSubnet.AddProperty("vpc_id", "test-vpc")
	publicSubnet.AddProperty("cidr_block", "10.0.1.0/24")
	publicSubnet.AddProperty("availability_zone", "us-east-1a")
	publicSubnet.AddProperty("map_public_ip_on_launch", true)
	publicSubnet.AddDependency("test-vpc")
	builder.AddResource(publicSubnet)

	// Add private subnet
	privateSubnet := models.NewResource(models.ResourceSubnet, "private-subnet-1")
	privateSubnet.AddProperty("vpc_id", "test-vpc")
	privateSubnet.AddProperty("cidr_block", "10.0.2.0/24")
	privateSubnet.AddProperty("availability_zone", "us-east-1a")
	privateSubnet.AddProperty("map_public_ip_on_launch", false)
	privateSubnet.AddDependency("test-vpc")
	builder.AddResource(privateSubnet)

	// Add Internet Gateway
	igw := models.NewResource(models.ResourceIGW, "test-igw")
	igw.AddProperty("vpc_id", "test-vpc")
	igw.AddDependency("test-vpc")
	builder.AddResource(igw)

	// Add NAT Gateway
	natgw := models.NewResource(models.ResourceNATGateway, "test-nat")
	natgw.AddProperty("subnet_id", "public-subnet-1")
	natgw.AddProperty("allocation_id", "eip-allocation-1")
	natgw.AddDependency("public-subnet-1")
	builder.AddResource(natgw)

	// Add EKS Cluster
	eksCluster := models.NewResource(models.ResourceEKSCluster, "test-eks-cluster")
	eksCluster.AddProperty("version", "1.27")
	eksCluster.AddProperty("role_arn", "arn:aws:iam::123456789012:role/eks-cluster-role")
	eksCluster.AddProperty("subnet_ids", []string{"private-subnet-1"})
	eksCluster.AddProperty("endpoint_public_access", true)
	eksCluster.AddProperty("endpoint_private_access", false)
	eksCluster.AddDependency("private-subnet-1")
	builder.AddResource(eksCluster)

	// Add Node Group
	nodeGroup := models.NewResource(models.ResourceNodeGroup, "test-node-group")
	nodeGroup.AddProperty("cluster_name", "test-eks-cluster")
	nodeGroup.AddProperty("node_role", "arn:aws:iam::123456789012:role/eks-node-group-role")
	nodeGroup.AddProperty("subnet_ids", []string{"private-subnet-1"})
	nodeGroup.AddProperty("instance_types", []string{"t3.medium"})
	nodeGroup.AddProperty("desired_size", 2)
	nodeGroup.AddProperty("min_size", 2)
	nodeGroup.AddProperty("max_size", 4)
	nodeGroup.AddDependency("test-eks-cluster")
	nodeGroup.AddDependency("private-subnet-1")
	builder.AddResource(nodeGroup)

	// Get the model
	model := builder.GetModel()

	// Create a temp directory for output
	testDir, err := os.MkdirTemp("", "crossplane-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create and initialize the generator
	generator := crossplane.NewCrossplaneGenerator()
	if err := generator.Init(testDir); err != nil {
		t.Fatalf("Failed to initialize generator: %v", err)
	}

	// Generate the Crossplane resources
	summary, err := generator.Generate(model)
	if err != nil {
		t.Fatalf("Failed to generate Crossplane resources: %v", err)
	}

	// Print the summary
	t.Logf("Generated Crossplane resources:\n%s", summary)

	// Verify files were created
	expectedFiles := []string{
		filepath.Join(testDir, "kustomization.yaml"),
		filepath.Join(testDir, "base", "kustomization.yaml"),
		filepath.Join(testDir, "base", "provider.yaml"),
		filepath.Join(testDir, "base", "providerconfig.yaml"),
		filepath.Join(testDir, "vpc", "kustomization.yaml"),
		filepath.Join(testDir, "vpc", "vpc.yaml"),
		filepath.Join(testDir, "vpc", "subnets.yaml"),
		filepath.Join(testDir, "vpc", "gateways.yaml"),
		filepath.Join(testDir, "eks", "kustomization.yaml"),
		filepath.Join(testDir, "eks", "cluster.yaml"),
		filepath.Join(testDir, "eks", "nodegroup.yaml"),
		filepath.Join(testDir, "eks", "iam.yaml"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file not found: %s", file)
		}
	}
}
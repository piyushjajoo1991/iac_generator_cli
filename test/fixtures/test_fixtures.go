package fixtures

import (
	"encoding/json"
	"fmt"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// TestDescription represents a test case with input description and expected model
type TestDescription struct {
	Name             string
	Description      string
	ExpectedEntities map[string]interface{}
	ExpectedModel    *models.InfrastructureModel
}

// TestDescriptionVPC contains test fixtures for VPC descriptions
var TestDescriptionVPC = []TestDescription{
	{
		Name:        "Basic VPC",
		Description: "Create a VPC with CIDR 10.0.0.0/16 in us-east-1",
		ExpectedEntities: map[string]interface{}{
			"region": "us-east-1",
			"vpc": map[string]interface{}{
				"exists":              true,
				"cidr_block":          "10.0.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
		},
	},
	{
		Name:        "VPC with Subnets",
		Description: "Create a VPC with 2 public subnets and 2 private subnets in us-west-2",
		ExpectedEntities: map[string]interface{}{
			"region": "us-west-2",
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
	},
	{
		Name:        "VPC with Gateways",
		Description: "Create a VPC with 1 public subnet, 1 private subnet, and an Internet Gateway in us-east-2",
		ExpectedEntities: map[string]interface{}{
			"region": "us-east-2",
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
				"nat_count": 0,
			},
		},
	},
}

// TestDescriptionEKS contains test fixtures for EKS descriptions
var TestDescriptionEKS = []TestDescription{
	{
		Name:        "Basic EKS Cluster",
		Description: "Create an EKS cluster with version 1.27 in us-east-1",
		ExpectedEntities: map[string]interface{}{
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
			"eks": map[string]interface{}{
				"exists":                 true,
				"version":                "1.27",
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"node_count":              2,
				"instance_type":           "t3.medium",
			},
		},
	},
	{
		Name:        "EKS with Node Groups",
		Description: "Create an EKS cluster with 2 node groups using t3.large instances in us-west-2",
		ExpectedEntities: map[string]interface{}{
			"region": "us-west-2",
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
			"eks": map[string]interface{}{
				"exists":                 true,
				"version":                "1.27",
				"endpoint_public_access": true,
				"endpoint_private_access": false,
				"node_count":              2,
				"instance_type":           "t3.large",
			},
		},
	},
	{
		Name:        "EKS with Private Access",
		Description: "Create an EKS cluster with private API access in us-east-2",
		ExpectedEntities: map[string]interface{}{
			"region": "us-east-2",
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
			"eks": map[string]interface{}{
				"exists":                 true,
				"version":                "1.27",
				"endpoint_public_access": false,
				"endpoint_private_access": true,
				"node_count":              2,
				"instance_type":           "t3.medium",
			},
		},
	},
}

// TestDescriptionComplex contains test fixtures for complex infrastructure descriptions
var TestDescriptionComplex = []TestDescription{
	{
		Name:        "Full AWS Infrastructure",
		Description: "AWS infra in us-east-1 with a vpc, 3 public and 3 private subnets, 1 IGW and 3 NAT gateways per az, plus an EKS Cluster with public and private api access deployed in private vpcs in 3 azs with a nodepool on t3-medium instance type",
		ExpectedEntities: map[string]interface{}{
			"region": "us-east-1",
			"vpc": map[string]interface{}{
				"exists":              true,
				"cidr_block":          "10.0.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
			"subnets": map[string]interface{}{
				"public_count":  3,
				"private_count": 3,
			},
			"gateways": map[string]interface{}{
				"igw_count": 1,
				"nat_count": 3,
			},
			"eks": map[string]interface{}{
				"exists":                 true,
				"version":                "1.27",
				"endpoint_public_access": true,
				"endpoint_private_access": true,
				"node_count":              2,
				"instance_type":           "t3.medium",
			},
		},
	},
	{
		Name:        "Non-Default CIDR",
		Description: "Create a VPC with CIDR 172.16.0.0/16 in eu-west-1 with 2 public subnets and 3 private subnets",
		ExpectedEntities: map[string]interface{}{
			"region": "eu-west-1",
			"vpc": map[string]interface{}{
				"exists":              true,
				"cidr_block":          "172.16.0.0/16",
				"enable_dns_support":  true,
				"enable_dns_hostnames": true,
			},
			"subnets": map[string]interface{}{
				"public_count":  2,
				"private_count": 3,
			},
		},
	},
}

// TestDescriptionInvalid contains test fixtures for invalid infrastructure descriptions
var TestDescriptionInvalid = []TestDescription{
	{
		Name:        "Empty Description",
		Description: "",
	},
	{
		Name:        "Too Short Description",
		Description: "VPC",
	},
	{
		Name:        "Invalid Region",
		Description: "Create a VPC in invalid-region",
	},
}

// GetAllTestDescriptions returns all test descriptions
func GetAllTestDescriptions() []TestDescription {
	var allDescriptions []TestDescription
	allDescriptions = append(allDescriptions, TestDescriptionVPC...)
	allDescriptions = append(allDescriptions, TestDescriptionEKS...)
	allDescriptions = append(allDescriptions, TestDescriptionComplex...)
	return allDescriptions
}

// CreateTestInfrastructureModel creates a test infrastructure model for testing
func CreateTestInfrastructureModel() *models.InfrastructureModel {
	model := models.NewInfrastructureModel()

	// Create a VPC resource
	vpc := models.NewResource(models.ResourceVPC, "main-vpc")
	vpc.AddProperty("cidr_block", "10.0.0.0/16")
	vpc.AddProperty("enable_dns_support", true)
	vpc.AddProperty("enable_dns_hostnames", true)
	vpc.AddProperty("region", "us-east-1")
	model.AddResource(vpc)

	// Create public subnet resources
	for i := 0; i < 2; i++ {
		subnet := models.NewResource(models.ResourceSubnet, fmt.Sprintf("public-subnet-%d", i+1))
		subnet.AddProperty("vpc_id", "main-vpc")
		subnet.AddProperty("cidr_block", fmt.Sprintf("10.0.%d.0/24", i))
		subnet.AddProperty("availability_zone", fmt.Sprintf("us-east-1%c", 'a'+i))
		subnet.AddProperty("map_public_ip_on_launch", true)
		subnet.AddDependency("main-vpc")
		model.AddResource(subnet)
	}

	// Create private subnet resources
	for i := 0; i < 2; i++ {
		subnet := models.NewResource(models.ResourceSubnet, fmt.Sprintf("private-subnet-%d", i+1))
		subnet.AddProperty("vpc_id", "main-vpc")
		subnet.AddProperty("cidr_block", fmt.Sprintf("10.0.%d.0/24", i+10))
		subnet.AddProperty("availability_zone", fmt.Sprintf("us-east-1%c", 'a'+i))
		subnet.AddProperty("map_public_ip_on_launch", false)
		subnet.AddDependency("main-vpc")
		model.AddResource(subnet)
	}

	// Create an Internet Gateway
	igw := models.NewResource(models.ResourceIGW, "main-igw")
	igw.AddProperty("vpc_id", "main-vpc")
	igw.AddDependency("main-vpc")
	model.AddResource(igw)

	// Create a NAT Gateway
	natGateway := models.NewResource(models.ResourceNATGateway, "nat-gateway-1")
	natGateway.AddProperty("subnet_id", "public-subnet-1")
	natGateway.AddProperty("allocation_id", "eip-allocation-1")
	natGateway.AddDependency("public-subnet-1")
	model.AddResource(natGateway)

	// Create an EKS Cluster
	eksCluster := models.NewResource(models.ResourceEKSCluster, "main-eks-cluster")
	eksCluster.AddProperty("version", "1.27")
	eksCluster.AddProperty("role_arn", "arn:aws:iam::123456789012:role/eks-cluster-role")
	eksCluster.AddProperty("vpc_config", map[string]interface{}{
		"subnet_ids":              []string{"private-subnet-1", "private-subnet-2"},
		"endpoint_public_access":  true,
		"endpoint_private_access": false,
	})
	eksCluster.AddDependency("private-subnet-1")
	eksCluster.AddDependency("private-subnet-2")
	model.AddResource(eksCluster)

	// Create a Node Group
	nodeGroup := models.NewResource(models.ResourceNodeGroup, "main-node-group")
	nodeGroup.AddProperty("cluster_name", "main-eks-cluster")
	nodeGroup.AddProperty("node_role_arn", "arn:aws:iam::123456789012:role/eks-node-group-role")
	nodeGroup.AddProperty("subnet_ids", []string{"private-subnet-1", "private-subnet-2"})
	nodeGroup.AddProperty("instance_types", []string{"t3.medium"})
	nodeGroup.AddProperty("scaling_config", map[string]interface{}{
		"desired_size": 2,
		"min_size":     2,
		"max_size":     4,
	})
	nodeGroup.AddDependency("main-eks-cluster")
	nodeGroup.AddDependency("private-subnet-1")
	nodeGroup.AddDependency("private-subnet-2")
	model.AddResource(nodeGroup)

	return model
}

// SerializeModel serializes a model to JSON
func SerializeModel(model *models.InfrastructureModel) (string, error) {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeModel deserializes a model from JSON
func DeserializeModel(data string) (*models.InfrastructureModel, error) {
	model := models.NewInfrastructureModel()
	err := json.Unmarshal([]byte(data), model)
	if err != nil {
		return nil, err
	}
	return model, nil
}
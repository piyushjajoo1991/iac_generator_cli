package infra

import (
	"fmt"
	"strings"
	
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// CreateEC2Instance creates an EC2 instance resource with the given properties
func CreateEC2Instance(name string, instanceType string, ami string, region string) models.Resource {
	resource := models.NewResource(models.ResourceEC2Instance, name)
	resource.AddProperty("instance_type", instanceType)
	resource.AddProperty("ami", ami)
	resource.AddProperty("region", region)
	return resource
}

// CreateS3Bucket creates an S3 bucket resource with the given properties
func CreateS3Bucket(name string, acl string, versioning bool) models.Resource {
	resource := models.NewResource(models.ResourceS3Bucket, name)
	resource.AddProperty("bucket", name)
	resource.AddProperty("acl", acl)
	resource.AddProperty("versioning", versioning)
	return resource
}

// CreateVPC creates a VPC resource with the given properties
func CreateVPC(name string, cidrBlock string, enableDnsSupport bool, enableDnsHostnames bool) models.Resource {
	resource := models.NewResource(models.ResourceVPC, name)
	resource.AddProperty("cidr_block", cidrBlock)
	resource.AddProperty("enable_dns_support", enableDnsSupport)
	resource.AddProperty("enable_dns_hostnames", enableDnsHostnames)
	return resource
}

// CreateSubnet creates a subnet resource with the given properties
func CreateSubnet(name string, vpcID string, cidrBlock string, availabilityZone string) models.Resource {
	resource := models.NewResource(models.ResourceSubnet, name)
	resource.AddProperty("vpc_id", vpcID)
	resource.AddProperty("cidr_block", cidrBlock)
	resource.AddProperty("availability_zone", availabilityZone)
	return resource
}

// CreateSecurityGroup creates a security group resource with the given properties
func CreateSecurityGroup(name string, description string, vpcID string) models.Resource {
	resource := models.NewResource(models.ResourceSecurityGroup, name)
	resource.AddProperty("name", name)
	resource.AddProperty("description", description)
	resource.AddProperty("vpc_id", vpcID)
	return resource
}

// AddSecurityGroupRule adds an ingress or egress rule to a security group
func AddSecurityGroupRule(securityGroup *models.Resource, ruleType string, protocol string, fromPort int, toPort int, cidrBlocks []string) {
	var rules []map[string]interface{}

	// Check if rules already exist
	for _, property := range securityGroup.Properties {
		if property.Name == ruleType {
			if existingRules, ok := property.Value.([]map[string]interface{}); ok {
				rules = existingRules
			}
		}
	}

	// Create new rule
	rule := map[string]interface{}{
		"protocol":    protocol,
		"from_port":   fromPort,
		"to_port":     toPort,
		"cidr_blocks": cidrBlocks,
	}

	// Add rule to list
	rules = append(rules, rule)

	// Update or add property
	found := false
	for i, property := range securityGroup.Properties {
		if property.Name == ruleType {
			securityGroup.Properties[i].Value = rules
			found = true
			break
		}
	}

	if !found {
		securityGroup.AddProperty(ruleType, rules)
	}
}

// CreateInternetGateway creates an Internet Gateway resource
func CreateInternetGateway(name string, vpcID string) models.Resource {
	resource := models.NewResource(models.ResourceIGW, name)
	resource.AddProperty("vpc_id", vpcID)
	return resource
}

// CreateNATGateway creates a NAT Gateway resource
func CreateNATGateway(name string, subnetID string, allocationID string) models.Resource {
	resource := models.NewResource(models.ResourceNATGateway, name)
	resource.AddProperty("subnet_id", subnetID)
	resource.AddProperty("allocation_id", allocationID)
	resource.AddProperty("connectivity_type", "public")
	return resource
}

// CreateEKSCluster creates an EKS Cluster resource
func CreateEKSCluster(name string, version string, roleArn string, subnetIDs []string, endpointPublicAccess bool, endpointPrivateAccess bool) models.Resource {
	resource := models.NewResource(models.ResourceEKSCluster, name)
	resource.AddProperty("name", name)
	resource.AddProperty("role_arn", roleArn)
	resource.AddProperty("version", version)
	
	// VPC Config
	vpcConfig := map[string]interface{}{
		"subnet_ids":                subnetIDs,
		"endpoint_public_access":    endpointPublicAccess,
		"endpoint_private_access":   endpointPrivateAccess,
	}
	resource.AddProperty("vpc_config", vpcConfig)
	
	return resource
}

// GenerateSubnetCIDRs generates CIDR blocks for subnets based on VPC CIDR
func GenerateSubnetCIDRs(vpcCIDR string, publicCount int, privateCount int) ([]string, []string, error) {
	// Parse VPC CIDR
	parts := strings.Split(vpcCIDR, "/")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid VPC CIDR format: %s", vpcCIDR)
	}
	
	ipParts := strings.Split(parts[0], ".")
	if len(ipParts) != 4 {
		return nil, nil, fmt.Errorf("invalid IP format in CIDR: %s", vpcCIDR)
	}
	
	publicCIDRs := make([]string, publicCount)
	privateCIDRs := make([]string, privateCount)
	
	// Generate public subnets starting from x.x.0.0/24
	for i := 0; i < publicCount; i++ {
		publicCIDRs[i] = fmt.Sprintf("%s.%s.%d.0/24", ipParts[0], ipParts[1], i)
	}
	
	// Generate private subnets starting from x.x.10.0/24
	for i := 0; i < privateCount; i++ {
		privateCIDRs[i] = fmt.Sprintf("%s.%s.%d.0/24", ipParts[0], ipParts[1], i+10)
	}
	
	return publicCIDRs, privateCIDRs, nil
}

// CreateEKSNodeGroup creates an EKS Node Group resource
func CreateEKSNodeGroup(name string, clusterName string, nodeRoleArn string, subnetIDs []string, instanceTypes []string, desiredSize int, minSize int, maxSize int) models.Resource {
	resource := models.NewResource(models.ResourceNodeGroup, name)
	resource.AddProperty("cluster_name", clusterName)
	resource.AddProperty("node_role_arn", nodeRoleArn)
	resource.AddProperty("subnet_ids", subnetIDs)
	
	// Scaling Config
	scalingConfig := map[string]interface{}{
		"desired_size": desiredSize,
		"min_size":     minSize,
		"max_size":     maxSize,
	}
	resource.AddProperty("scaling_config", scalingConfig)
	
	// Instance Types
	resource.AddProperty("instance_types", instanceTypes)
	
	return resource
}
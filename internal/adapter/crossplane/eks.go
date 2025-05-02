package crossplane

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/riptano/iac_generator_cli/pkg/models"
)

// EKSGenerator generates Crossplane YAML for EKS clusters and node groups
type EKSGenerator struct {
	baseDir string
	eksDir  string
}

// NewEKSGenerator creates a new EKS Generator
func NewEKSGenerator(baseDir string) *EKSGenerator {
	return &EKSGenerator{
		baseDir: baseDir,
		eksDir:  filepath.Join(baseDir, "eks"),
	}
}

// GenerateIAMRole generates a Crossplane IAM Role resource
func (g *EKSGenerator) GenerateIAMRole(name, assumeRolePolicyDocument string, managedPolicyArns []string) K8sObject {
	role := NewK8sObject("iam.aws.crossplane.io/v1beta1", "Role", name)
	
	// Add IAM Role specific properties
	role.AddNestedSpecField([]string{"forProvider", "assumeRolePolicyDocument"}, assumeRolePolicyDocument)
	
	if len(managedPolicyArns) > 0 {
		role.AddNestedSpecField([]string{"forProvider", "managedPolicyArns"}, managedPolicyArns)
	}
	
	// Add provider config reference
	role.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	role.AddLabel("app.kubernetes.io/part-of", "eks")
	role.AddLabel("app.kubernetes.io/component", "iam-role")
	
	return role
}

// GenerateEKSCluster generates a Crossplane EKS Cluster resource
func (g *EKSGenerator) GenerateEKSCluster(
	name string,
	version string,
	roleArn string,
	subnetIds []string,
	endpointPublicAccess bool,
	endpointPrivateAccess bool,
	securityGroupIds []string,
	tags map[string]string,
) K8sObject {
	cluster := NewK8sObject("eks.aws.crossplane.io/v1beta1", "Cluster", name)
	
	// Add EKS Cluster specific properties
	cluster.AddNestedSpecField([]string{"forProvider", "version"}, version)
	
	// Add VPC config
	vpcConfig := make(map[string]interface{})
	
	// Add subnet references
	subnetRefs := make([]map[string]string, 0, len(subnetIds))
	for _, subnetId := range subnetIds {
		subnetRefs = append(subnetRefs, map[string]string{"name": subnetId})
	}
	vpcConfig["subnetIdRefs"] = subnetRefs
	
	// Add security group references if provided
	if len(securityGroupIds) > 0 {
		sgRefs := make([]map[string]string, 0, len(securityGroupIds))
		for _, sgId := range securityGroupIds {
			sgRefs = append(sgRefs, map[string]string{"name": sgId})
		}
		vpcConfig["securityGroupIdRefs"] = sgRefs
	}
	
	// Add endpoint access config
	vpcConfig["endpointPublicAccess"] = endpointPublicAccess
	vpcConfig["endpointPrivateAccess"] = endpointPrivateAccess
	
	cluster.AddNestedSpecField([]string{"forProvider", "resourcesVpcConfig"}, vpcConfig)
	
	// Reference the role or use a direct ARN
	if strings.HasPrefix(roleArn, "arn:aws") {
		cluster.AddNestedSpecField([]string{"forProvider", "roleArn"}, roleArn)
	} else {
		cluster.AddNestedSpecField([]string{"forProvider", "roleArnRef", "name"}, roleArn)
	}
	
	// Add tags if provided
	if len(tags) > 0 {
		cluster.AddNestedSpecField([]string{"forProvider", "tags"}, tags)
	}
	
	// Add provider config reference
	cluster.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	cluster.AddLabel("app.kubernetes.io/part-of", "eks")
	cluster.AddLabel("app.kubernetes.io/component", "cluster")
	
	// Add deletion policy to avoid accidental deletion
	cluster.AddNestedSpecField([]string{"deletionPolicy"}, "Orphan")
	
	return cluster
}

// GenerateEKSNodeGroup generates a Crossplane EKS Node Group resource
func (g *EKSGenerator) GenerateEKSNodeGroup(
	name string,
	clusterName string,
	nodeRoleArn string,
	subnetIds []string,
	instanceTypes []string,
	desiredSize int,
	minSize int,
	maxSize int,
	diskSize int,
	amiType string,
	labels map[string]string,
	tags map[string]string,
) K8sObject {
	nodeGroup := NewK8sObject("eks.aws.crossplane.io/v1beta1", "NodeGroup", name)
	
	// Add EKS Node Group specific properties
	
	// Reference the cluster
	nodeGroup.AddNestedSpecField([]string{"forProvider", "clusterNameRef", "name"}, clusterName)
	
	// Add subnet references
	subnetRefs := make([]map[string]string, 0, len(subnetIds))
	for _, subnetId := range subnetIds {
		subnetRefs = append(subnetRefs, map[string]string{"name": subnetId})
	}
	nodeGroup.AddNestedSpecField([]string{"forProvider", "subnetIdRefs"}, subnetRefs)
	
	// Reference the role or use a direct ARN
	if strings.HasPrefix(nodeRoleArn, "arn:aws") {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "nodeRole"}, nodeRoleArn)
	} else {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "nodeRoleRef", "name"}, nodeRoleArn)
	}
	
	// Add scaling config
	scalingConfig := map[string]interface{}{
		"desiredSize": desiredSize,
		"minSize":     minSize,
		"maxSize":     maxSize,
	}
	nodeGroup.AddNestedSpecField([]string{"forProvider", "scalingConfig"}, scalingConfig)
	
	// Add instance types if provided
	if len(instanceTypes) > 0 {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "instanceTypes"}, instanceTypes)
	}
	
	// Add disk size if provided
	if diskSize > 0 {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "diskSize"}, diskSize)
	}
	
	// Add AMI type if provided
	if amiType != "" {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "amiType"}, amiType)
	}
	
	// Add labels if provided
	if len(labels) > 0 {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "labels"}, labels)
	}
	
	// Add tags if provided
	if len(tags) > 0 {
		nodeGroup.AddNestedSpecField([]string{"forProvider", "tags"}, tags)
	}
	
	// Add provider config reference
	nodeGroup.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Add common labels
	nodeGroup.AddLabel("app.kubernetes.io/part-of", "eks")
	nodeGroup.AddLabel("app.kubernetes.io/component", "node-group")
	
	return nodeGroup
}

// GenerateEKSResources generates all EKS related resources from an infrastructure model
func (g *EKSGenerator) GenerateEKSResources(model *models.InfrastructureModel) error {
	var (
		eksCluster   K8sObject
		nodeGroups   []K8sObject
		roles        []K8sObject
		clusterFound bool
	)
	
	// Find subnet references for EKS cluster and node groups
	var subnetIds []string
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceSubnet {
			// Check if it's a private subnet - EKS nodes should be in private subnets
			isPrivate := true
			for _, prop := range resource.Properties {
				if prop.Name == "map_public_ip_on_launch" {
					if val, ok := prop.Value.(bool); ok && val {
						isPrivate = false
					}
				}
			}
			
			if isPrivate {
				subnetIds = append(subnetIds, resource.Name)
			}
		}
	}
	
	// Create IAM policies and roles
	
	// Cluster role
	clusterRoleName := "eks-cluster-role"
	clusterRole := g.GenerateIAMRole(
		clusterRoleName,
		`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`,
		[]string{
			"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		},
	)
	roles = append(roles, clusterRole)
	
	// Node group role
	nodeRoleName := "eks-node-role"
	nodeRole := g.GenerateIAMRole(
		nodeRoleName,
		`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`,
		[]string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		},
	)
	roles = append(roles, nodeRole)
	
	// Find the EKS cluster
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceEKSCluster {
			clusterFound = true
			
			// Extract cluster properties
			name := resource.Name
			version := "1.27"
			endpointPublicAccess := true
			endpointPrivateAccess := false
			
			for _, prop := range resource.Properties {
				switch prop.Name {
				case "version":
					if val, ok := prop.Value.(string); ok {
						version = val
					}
				case "endpoint_public_access":
					if val, ok := prop.Value.(bool); ok {
						endpointPublicAccess = val
					}
				case "endpoint_private_access":
					if val, ok := prop.Value.(bool); ok {
						endpointPrivateAccess = val
					}
				}
			}
			
			// Create EKS cluster
			eksCluster = g.GenerateEKSCluster(
				name,
				version,
				clusterRoleName, // Reference to the role we created
				subnetIds,
				endpointPublicAccess,
				endpointPrivateAccess,
				[]string{}, // No security groups specified
				map[string]string{
					"Name":        name,
					"Environment": "production",
					"ManagedBy":   "crossplane",
				},
			)
		}
	}
	
	// If no EKS cluster was found in the model but we're asked to generate EKS resources,
	// create a default one
	if !clusterFound && len(subnetIds) > 0 {
		eksCluster = g.GenerateEKSCluster(
			"main-eks-cluster",
			"1.27",
			clusterRoleName,
			subnetIds,
			true,  // Public endpoint
			false, // No private endpoint
			[]string{},
			map[string]string{
				"Name":        "main-eks-cluster",
				"Environment": "production",
				"ManagedBy":   "crossplane",
			},
		)
	}
	
	// Find node groups
	for _, resource := range model.Resources {
		if resource.Type == models.ResourceNodeGroup {
			// Extract node group properties
			name := resource.Name
			instanceTypes := []string{"t3.medium"}
			desiredSize := 2
			minSize := 2
			maxSize := 4
			diskSize := 20
			amiType := "AL2_x86_64"
			
			for _, prop := range resource.Properties {
				switch prop.Name {
				case "instance_types":
					if val, ok := prop.Value.([]string); ok {
						instanceTypes = val
					}
				case "desired_size":
					if val, ok := prop.Value.(int); ok {
						desiredSize = val
					}
				case "min_size":
					if val, ok := prop.Value.(int); ok {
						minSize = val
					}
				case "max_size":
					if val, ok := prop.Value.(int); ok {
						maxSize = val
					}
				case "disk_size":
					if val, ok := prop.Value.(int); ok {
						diskSize = val
					}
				case "ami_type":
					if val, ok := prop.Value.(string); ok {
						amiType = val
					}
				}
			}
			
			// Create node group
			nodeGroup := g.GenerateEKSNodeGroup(
				name,
				eksCluster.Metadata.Name,
				nodeRoleName,
				subnetIds,
				instanceTypes,
				desiredSize,
				minSize,
				maxSize,
				diskSize,
				amiType,
				map[string]string{
					"role": "worker",
				},
				map[string]string{
					"Name":      name,
					"ManagedBy": "crossplane",
				},
			)
			
			nodeGroups = append(nodeGroups, nodeGroup)
		}
	}
	
	// If we have an EKS cluster but no node groups, create a default one
	if eksCluster.APIVersion != "" && len(nodeGroups) == 0 {
		nodeGroup := g.GenerateEKSNodeGroup(
			"main-node-group",
			eksCluster.Metadata.Name,
			nodeRoleName,
			subnetIds,
			[]string{"t3.medium"},
			2,  // desired
			2,  // min
			4,  // max
			20, // disk size
			"AL2_x86_64",
			map[string]string{
				"role": "worker",
			},
			map[string]string{
				"Name":      "main-node-group",
				"ManagedBy": "crossplane",
			},
		)
		
		nodeGroups = append(nodeGroups, nodeGroup)
	}
	
	// Write IAM YAML
	if len(roles) > 0 {
		iamFilePath := filepath.Join(g.eksDir, "iam.yaml")
		if err := WriteMultiYAML(roles, iamFilePath); err != nil {
			return fmt.Errorf("failed to write IAM YAML: %w", err)
		}
	}
	
	// Write EKS Cluster YAML
	if eksCluster.APIVersion != "" {
		clusterFilePath := filepath.Join(g.eksDir, "cluster.yaml")
		if err := WriteYAML(eksCluster, clusterFilePath); err != nil {
			return fmt.Errorf("failed to write EKS Cluster YAML: %w", err)
		}
	}
	
	// Write Node Group YAML
	if len(nodeGroups) > 0 {
		nodeGroupFilePath := filepath.Join(g.eksDir, "nodegroup.yaml")
		if err := WriteMultiYAML(nodeGroups, nodeGroupFilePath); err != nil {
			return fmt.Errorf("failed to write Node Group YAML: %w", err)
		}
	}
	
	return nil
}
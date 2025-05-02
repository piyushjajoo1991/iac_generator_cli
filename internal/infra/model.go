package infra

import (
	"strconv"

	"github.com/riptano/iac_generator_cli/pkg/models"
)

// ModelBuilder builds an infrastructure model from parsed natural language
type ModelBuilder struct {
	model *models.InfrastructureModel
}

// NewModelBuilder creates a new ModelBuilder
func NewModelBuilder() *ModelBuilder {
	return &ModelBuilder{
		model: models.NewInfrastructureModel(),
	}
}

// AddResource adds a resource to the model
func (b *ModelBuilder) AddResource(resource models.Resource) {
	b.model.AddResource(resource)
}

// GetModel returns the built infrastructure model
func (b *ModelBuilder) GetModel() *models.InfrastructureModel {
	return b.model
}

// BuildFromParsedEntities builds an infrastructure model from parsed entities
func (b *ModelBuilder) BuildFromParsedEntities(entities map[string]interface{}) error {
	resourceIDs := make(map[string]string)
	region := "us-east-1"

	// Extract region if available
	if regionStr, ok := entities["region"].(string); ok {
		region = regionStr
	}

	// Create VPC if specified
	if vpcData, ok := entities["vpc"].(map[string]interface{}); ok {
		vpcName := "main-vpc"
		cidrBlock := "10.0.0.0/16"

		if cidr, ok := vpcData["cidr_block"].(string); ok {
			cidrBlock = cidr
		}

		enableDnsSupport := true
		enableDnsHostnames := true

		vpc := CreateVPC(vpcName, cidrBlock, enableDnsSupport, enableDnsHostnames)
		b.AddResource(vpc)
		resourceIDs["vpc"] = vpcName

		// Create subnets if specified
		if subnetData, ok := entities["subnets"].(map[string]interface{}); ok {
			publicCount := 0
			privateCount := 0
			var publicCIDRs, privateCIDRs []string

			if count, ok := subnetData["public_count"].(int); ok {
				publicCount = count
			}

			if count, ok := subnetData["private_count"].(int); ok {
				privateCount = count
			}

			if cidrs, ok := subnetData["public_cidrs"].([]string); ok && len(cidrs) > 0 {
				publicCIDRs = cidrs
			} else {
				// Generate CIDRs if not provided
				generatedPublic, generatedPrivate, err := GenerateSubnetCIDRs(cidrBlock, publicCount, privateCount)
				if err == nil {
					publicCIDRs = generatedPublic
					privateCIDRs = generatedPrivate
				}
			}

			if cidrs, ok := subnetData["private_cidrs"].([]string); ok && len(cidrs) > 0 {
				privateCIDRs = cidrs
			}

			// Create public subnets
			for i := 0; i < publicCount; i++ {
				cidr := "10.0." + strconv.Itoa(i) + ".0/24"
				if i < len(publicCIDRs) {
					cidr = publicCIDRs[i]
				}

				// Generate AZ based on region and index
				az := region + string(rune('a'+i%3))
				subnetName := "public-subnet-" + strconv.Itoa(i+1)

				subnet := CreateSubnet(subnetName, vpcName, cidr, az)
				b.AddResource(subnet)
				resourceIDs["public-subnet-"+strconv.Itoa(i)] = subnetName
			}

			// Create private subnets
			for i := 0; i < privateCount; i++ {
				cidr := "10.0." + strconv.Itoa(i+10) + ".0/24"
				if i < len(privateCIDRs) {
					cidr = privateCIDRs[i]
				}

				// Generate AZ based on region and index
				az := region + string(rune('a'+i%3))
				subnetName := "private-subnet-" + strconv.Itoa(i+1)

				subnet := CreateSubnet(subnetName, vpcName, cidr, az)
				b.AddResource(subnet)
				resourceIDs["private-subnet-"+strconv.Itoa(i)] = subnetName
			}
		}

		// Create Internet Gateway if specified
		if gatewayData, ok := entities["gateways"].(map[string]interface{}); ok {
			igwCount := 0
			natCount := 0

			if count, ok := gatewayData["igw_count"].(int); ok {
				igwCount = count
			}

			if count, ok := gatewayData["nat_count"].(int); ok {
				natCount = count
			}

			// Create Internet Gateway (typically just one)
			if igwCount > 0 {
				igwName := "main-igw"
				igw := CreateInternetGateway(igwName, resourceIDs["vpc"])
				b.AddResource(igw)
				resourceIDs["igw"] = igwName
			}

			// Create NAT Gateways (one per AZ or specified count)
			for i := 0; i < natCount; i++ {
				// In a real implementation, we would need to create an EIP for each NAT Gateway
				// For simplicity, we're assuming the EIPs already exist
				natName := "nat-gateway-" + strconv.Itoa(i+1)
				subnetID := resourceIDs["public-subnet-"+strconv.Itoa(i%len(resourceIDs))]
				allocID := "eip-allocation-" + strconv.Itoa(i+1) // Placeholder

				nat := CreateNATGateway(natName, subnetID, allocID)
				b.AddResource(nat)
				resourceIDs["nat-"+strconv.Itoa(i)] = natName
			}
		}

		// Create EKS Cluster if specified
		if eksData, ok := entities["eks"].(map[string]interface{}); ok {
			eksName := "main-eks-cluster"
			eksVersion := "1.27"

			if version, ok := eksData["version"].(string); ok {
				eksVersion = version
			}

			// Collect subnet IDs for EKS
			var subnetIDs []string
			for i := 0; ; i++ {
				privateSubnetID, ok := resourceIDs["private-subnet-"+strconv.Itoa(i)]
				if !ok {
					break
				}
				subnetIDs = append(subnetIDs, privateSubnetID)
			}

			// In a real implementation, we would create an IAM role for EKS
			// For simplicity, we're assuming the role already exists
			roleArn := "arn:aws:iam::123456789012:role/eks-cluster-role"

			endpointPublicAccess := true
			endpointPrivateAccess := false

			if access, ok := eksData["endpoint_public_access"].(bool); ok {
				endpointPublicAccess = access
			}

			if access, ok := eksData["endpoint_private_access"].(bool); ok {
				endpointPrivateAccess = access
			}

			eks := CreateEKSCluster(eksName, eksVersion, roleArn, subnetIDs, endpointPublicAccess, endpointPrivateAccess)
			b.AddResource(eks)
			resourceIDs["eks"] = eksName

			// Create Node Group if EKS exists
			nodeGroupName := "main-node-group"
			instanceType := "t3.medium"
			nodeCount := 2

			if instance, ok := eksData["instance_type"].(string); ok {
				instanceType = instance
			}

			if count, ok := eksData["node_count"].(int); ok {
				nodeCount = count
			}

			// In a real implementation, we would create an IAM role for the node group
			// For simplicity, we're assuming the role already exists
			nodeRoleArn := "arn:aws:iam::123456789012:role/eks-node-group-role"

			nodeGroup := CreateEKSNodeGroup(
				nodeGroupName,
				eksName,
				nodeRoleArn,
				subnetIDs,
				[]string{instanceType},
				nodeCount,   // desired
				nodeCount,   // min
				nodeCount*2, // max
			)
			b.AddResource(nodeGroup)
		}
	}

	// Handle EC2 instance if specified
	if instanceData, ok := entities["ec2_instance"].(map[string]interface{}); ok {
		name := "example-instance"
		instanceType := "t2.micro"
		ami := "ami-123456789"

		if instName, ok := instanceData["name"].(string); ok {
			name = instName
		}

		if instType, ok := instanceData["instance_type"].(string); ok {
			instanceType = instType
		}

		if instAMI, ok := instanceData["ami"].(string); ok {
			ami = instAMI
		}

		instance := CreateEC2Instance(name, instanceType, ami, region)
		b.AddResource(instance)
	}

	// Handle S3 bucket if specified
	if bucketData, ok := entities["s3_bucket"].(map[string]interface{}); ok {
		name := "example-bucket"
		acl := "private"
		versioning := false

		if bucketName, ok := bucketData["name"].(string); ok {
			name = bucketName
		}

		if bucketACL, ok := bucketData["acl"].(string); ok {
			acl = bucketACL
		}

		if bucketVersioning, ok := bucketData["versioning"].(bool); ok {
			versioning = bucketVersioning
		}

		bucket := CreateS3Bucket(name, acl, versioning)
		b.AddResource(bucket)
	}

	return nil
}

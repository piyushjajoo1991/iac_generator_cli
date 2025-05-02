package models

// ResourceType represents the type of AWS resource
type ResourceType string

// Supported AWS resource types
const (
	ResourceEC2Instance   ResourceType = "ec2_instance"
	ResourceS3Bucket      ResourceType = "s3_bucket"
	ResourceRDSInstance   ResourceType = "rds_instance"
	ResourceVPC           ResourceType = "vpc"
	ResourceSubnet        ResourceType = "subnet"
	ResourceSecurityGroup ResourceType = "security_group"
	ResourceIAMRole       ResourceType = "iam_role"
	ResourceLambda        ResourceType = "lambda"
	ResourceDynamoDB      ResourceType = "dynamodb"
	ResourceCloudwatch    ResourceType = "cloudwatch"
	ResourceIGW           ResourceType = "internet_gateway"
	ResourceNATGateway    ResourceType = "nat_gateway"
	ResourceEKSCluster    ResourceType = "eks_cluster"
	ResourceNodeGroup     ResourceType = "eks_node_group"
)

// Property represents a resource property
type Property struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Resource represents an infrastructure resource
type Resource struct {
	Type       ResourceType `json:"type"`
	Name       string       `json:"name"`
	Properties []Property   `json:"properties"`
	DependsOn  []string     `json:"depends_on,omitempty"`
}

// InfrastructureModel represents the complete infrastructure model
type InfrastructureModel struct {
	Resources []Resource `json:"resources"`
}

// NewResource creates a new resource with the given type and name
func NewResource(resourceType ResourceType, name string) Resource {
	return Resource{
		Type:       resourceType,
		Name:       name,
		Properties: []Property{},
	}
}

// AddProperty adds a property to a resource
func (r *Resource) AddProperty(name string, value interface{}) {
	r.Properties = append(r.Properties, Property{
		Name:  name,
		Value: value,
	})
}

// AddDependency adds a dependency to a resource
func (r *Resource) AddDependency(resourceName string) {
	r.DependsOn = append(r.DependsOn, resourceName)
}

// NewInfrastructureModel creates a new empty infrastructure model
func NewInfrastructureModel() *InfrastructureModel {
	return &InfrastructureModel{
		Resources: []Resource{},
	}
}

// AddResource adds a resource to the infrastructure model
func (m *InfrastructureModel) AddResource(resource Resource) {
	m.Resources = append(m.Resources, resource)
}
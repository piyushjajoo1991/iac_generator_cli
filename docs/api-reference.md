# API Reference

This document provides reference documentation for the IaC Manifest Generator CLI's internal APIs. It covers the key interfaces, types, and functions that make up the tool's core functionality.

## Table of Contents

- [Resource Models](#resource-models)
- [Pipeline System](#pipeline-system)
- [Natural Language Processing](#natural-language-processing)
- [Infrastructure Modeling](#infrastructure-modeling)
- [Template System](#template-system)
- [Adapters](#adapters)

## Resource Models

Resource models define the structure and behavior of infrastructure resources.

### ResourceType

```go
// ResourceType represents the type of AWS resource
type ResourceType string
```

Constants for supported resource types:

```go
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
```

### Property

```go
// Property represents a resource property
type Property struct {
    Name  string      `json:"name"`
    Value interface{} `json:"value"`
}
```

### Resource

```go
// Resource represents an infrastructure resource
type Resource struct {
    Type       ResourceType `json:"type"`
    Name       string       `json:"name"`
    Properties []Property   `json:"properties"`
    DependsOn  []string     `json:"depends_on,omitempty"`
}
```

Key methods:

```go
// NewResource creates a new resource with the given type and name
func NewResource(resourceType ResourceType, name string) Resource

// AddProperty adds a property to a resource
func (r *Resource) AddProperty(name string, value interface{})

// AddDependency adds a dependency to a resource
func (r *Resource) AddDependency(resourceName string)
```

### InfrastructureModel

```go
// InfrastructureModel represents the complete infrastructure model
type InfrastructureModel struct {
    Resources []Resource `json:"resources"`
}
```

Key methods:

```go
// NewInfrastructureModel creates a new empty infrastructure model
func NewInfrastructureModel() *InfrastructureModel

// AddResource adds a resource to the infrastructure model
func (m *InfrastructureModel) AddResource(resource Resource)
```

## Pipeline System

The pipeline system coordinates the transformation of inputs through multiple processing stages.

### Pipeline Interface

```go
// Pipeline defines the interface for a processing pipeline
type Pipeline interface {
    // Execute executes the pipeline with the given input
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // AddStage adds a stage to the pipeline
    AddStage(stage Stage)
    
    // SetErrorHandler sets a custom error handler for the pipeline
    SetErrorHandler(handler func(error) error)
    
    // SetProgressReporter sets a progress reporter for the pipeline
    SetProgressReporter(reporter ProgressReporter)
}
```

### Stage Interface

```go
// Stage defines the interface for a pipeline stage
type Stage interface {
    // Execute executes the stage with the given input
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // Name returns the stage name
    Name() string
}
```

### ProgressReporter Interface

```go
// ProgressReporter defines the interface for reporting pipeline progress
type ProgressReporter interface {
    // StartStage reports the start of a stage
    StartStage(stageName string)
    
    // CompleteStage reports the completion of a stage
    CompleteStage(stageName string)
    
    // FailStage reports the failure of a stage
    FailStage(stageName string, err error)
    
    // UpdateProgress reports progress within a stage
    UpdateProgress(message string, percentage int)
}
```

## Natural Language Processing

The NLP component processes natural language descriptions of infrastructure.

### Parser

```go
// Parser interfaces with NLP services to extract infrastructure entities
type Parser struct {
    // Implementation details
}
```

Key functions:

```go
// NewParser creates a new NLP parser
func NewParser() *Parser

// ParseDescription parses a natural language description into an infrastructure model
func ParseDescription(description string) (*models.InfrastructureModel, error)

// ExtractEntities extracts infrastructure entities from the description
func (p *Parser) ExtractEntities(description string) (map[string]interface{}, error)
```

### Entity Extraction

Functions for extracting specific entities:

```go
// ExtractRegion extracts the AWS region from the description
func ExtractRegion(description string) string

// ExtractVPC extracts VPC information from the description
func ExtractVPC(description string) map[string]interface{}

// ExtractSubnets extracts subnet information from the description
func ExtractSubnets(description string) map[string]interface{}

// ExtractGateways extracts gateway information from the description
func ExtractGateways(description string) map[string]interface{}

// ExtractEKS extracts EKS cluster information from the description
func ExtractEKS(description string) map[string]interface{}
```

### Validation

```go
// ValidationResult represents the result of entity validation
type ValidationResult struct {
    Valid   bool
    Message string
    Fixes   map[string]interface{}
}

// ValidateEntities validates entities extracted from NLP
func ValidateEntities(entities map[string]interface{}) ValidationResult
```

## Infrastructure Modeling

The infrastructure modeling component defines a unified representation of cloud resources.

### ModelBuilder

```go
// ModelBuilder builds an infrastructure model from parsed natural language
type ModelBuilder struct {
    model *models.InfrastructureModel
}
```

Key methods:

```go
// NewModelBuilder creates a new ModelBuilder
func NewModelBuilder() *ModelBuilder

// AddResource adds a resource to the model
func (b *ModelBuilder) AddResource(resource models.Resource)

// GetModel returns the built infrastructure model
func (b *ModelBuilder) GetModel() *models.InfrastructureModel

// BuildFromParsedEntities builds an infrastructure model from parsed entities
func (b *ModelBuilder) BuildFromParsedEntities(entities map[string]interface{}) error
```

### Resource Creation Functions

```go
// CreateVPC creates a VPC resource
func CreateVPC(name string, cidrBlock string, enableDnsSupport bool, enableDnsHostnames bool) models.Resource

// CreateSubnet creates a subnet resource
func CreateSubnet(name string, vpcName string, cidrBlock string, az string) models.Resource

// CreateInternetGateway creates an internet gateway resource
func CreateInternetGateway(name string, vpcName string) models.Resource

// CreateNATGateway creates a NAT gateway resource
func CreateNATGateway(name string, subnetID string, allocationID string) models.Resource

// CreateEKSCluster creates an EKS cluster resource
func CreateEKSCluster(name string, version string, roleArn string, subnetIDs []string, endpointPublicAccess bool, endpointPrivateAccess bool) models.Resource

// CreateEKSNodeGroup creates an EKS node group resource
func CreateEKSNodeGroup(name string, clusterName string, nodeRoleArn string, subnetIDs []string, instanceTypes []string, desiredSize, minSize, maxSize int) models.Resource

// CreateEC2Instance creates an EC2 instance resource
func CreateEC2Instance(name string, instanceType string, ami string, region string) models.Resource

// CreateS3Bucket creates an S3 bucket resource
func CreateS3Bucket(name string, acl string, versioning bool) models.Resource
```

## Template System

The template system enables customizable generation of IaC manifests using Go templates.

### TemplateFormat

```go
// TemplateFormat represents the format of the template (Terraform or Crossplane)
type TemplateFormat string

const (
    // FormatTerraform represents Terraform HCL format
    FormatTerraform TemplateFormat = "terraform"
    // FormatCrossplane represents Crossplane YAML format
    FormatCrossplane TemplateFormat = "crossplane"
)
```

### TemplateManager

```go
// TemplateManager manages the loading and caching of templates
type TemplateManager struct {
    // Implementation details
}
```

Key methods:

```go
// NewTemplateManager creates a new template manager with the given embedded filesystem
func NewTemplateManager(fs embed.FS) *TemplateManager

// GetTemplate gets a template by name, loading it from the embedded filesystem if needed
func (tm *TemplateManager) GetTemplate(format TemplateFormat, templateName string) (*template.Template, error)

// GetTemplateWithPattern gets a template for a given resource type matching a pattern
func (tm *TemplateManager) GetTemplateWithPattern(format TemplateFormat, pattern string) (*template.Template, string, error)

// ListTemplates lists all available templates for a given format
func (tm *TemplateManager) ListTemplates(format TemplateFormat) ([]string, error)
```

### TemplateSelector

```go
// TemplateSelector interface for selecting the correct template for a resource
type TemplateSelector interface {
    // SelectTemplate selects the appropriate template for the given resource and format
    SelectTemplate(format TemplateFormat, resource *models.Resource) (string, error)
    // RegisterTemplate registers a custom template for a resource type
    RegisterTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string)
}
```

### TemplateRenderer

```go
// TemplateRenderer renders templates for resources
type TemplateRenderer struct {
    // Implementation details
}
```

Key methods:

```go
// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(manager *TemplateManager, selector TemplateSelector) *TemplateRenderer

// RenderResource renders a single resource
func (r *TemplateRenderer) RenderResource(format TemplateFormat, resource *models.Resource) (string, error)

// RenderResources renders multiple resources
func (r *TemplateRenderer) RenderResources(format TemplateFormat, resources []models.Resource) (string, error)

// RenderResourceToFile renders a resource and writes it to a file
func (r *TemplateRenderer) RenderResourceToFile(format TemplateFormat, resource *models.Resource, filePath string) error
```

## Adapters

The adapters translate the unified infrastructure model into specific IaC formats.

### Generator Interface

```go
// Generator defines the interface for IaC generators
type Generator interface {
    // Generate generates IaC manifests from an infrastructure model
    Generate(model *models.InfrastructureModel) (string, error)
}
```

### Terraform Generator

```go
// TerraformGenerator generates Terraform HCL manifests
type TerraformGenerator struct {
    // Implementation details
}
```

Key methods:

```go
// NewTerraformGenerator creates a new TerraformGenerator
func NewTerraformGenerator() *TerraformGenerator

// Generate generates Terraform HCL from an infrastructure model
func (g *TerraformGenerator) Generate(model *models.InfrastructureModel) (string, error)
```

### Crossplane Generator

```go
// CrossplaneGenerator generates Crossplane YAML manifests
type CrossplaneGenerator struct {
    // Implementation details
}
```

Key methods:

```go
// NewCrossplaneGenerator creates a new CrossplaneGenerator
func NewCrossplaneGenerator() *CrossplaneGenerator

// Generate generates Crossplane YAML from an infrastructure model
func (g *CrossplaneGenerator) Generate(model *models.InfrastructureModel) (string, error)
```

### Generator Factory

```go
// CreateGenerator creates a generator based on the format
func CreateGenerator(format string, useTemplates bool) (Generator, error)
```
package terraform

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/riptano/iac_generator_cli/internal/infra"
	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// TerraformGenerator generates Terraform HCL manifests
type TerraformGenerator struct {
	OutputDir string
	Model     *models.InfrastructureModel
	Config    *TerraformConfig
}

// TerraformConfig holds Terraform-specific configuration
type TerraformConfig struct {
	AwsRegion          string
	CreateModules      bool
	ModuleNames        []string
	BackendType        string
	BackendConfig      map[string]string
	TerraformVersion   string
	ProviderConstraint string
}

// DefaultTerraformConfig returns a default configuration
func DefaultTerraformConfig() *TerraformConfig {
	return &TerraformConfig{
		AwsRegion:          "us-east-1",
		CreateModules:      true,
		ModuleNames:        []string{"vpc", "eks"},
		BackendType:        "local",
		BackendConfig:      map[string]string{},
		TerraformVersion:   "1.0.0",
		ProviderConstraint: "~> 5.0",
	}
}

// NewTerraformGenerator creates a new TerraformGenerator
func NewTerraformGenerator() *TerraformGenerator {
	return &TerraformGenerator{
		OutputDir: "terraform",
		Config:    DefaultTerraformConfig(),
	}
}

// WithOutputDir sets the output directory
func (g *TerraformGenerator) WithOutputDir(dir string) *TerraformGenerator {
	g.OutputDir = dir
	return g
}

// WithConfig sets the configuration
func (g *TerraformGenerator) WithConfig(config *TerraformConfig) *TerraformGenerator {
	g.Config = config
	return g
}

// Generate generates Terraform HCL from an infrastructure model
func (g *TerraformGenerator) Generate(model *models.InfrastructureModel) (string, error) {
	g.Model = model

	// Create directory structure
	if err := g.createDirectoryStructure(); err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate root module files
	if err := g.generateRootModuleFiles(); err != nil {
		return "", fmt.Errorf("failed to generate root module files: %w", err)
	}

	// Generate module files
	if g.Config.CreateModules {
		if err := g.generateModuleFiles(); err != nil {
			return "", fmt.Errorf("failed to generate module files: %w", err)
		}
	}

	return fmt.Sprintf("Terraform files generated in %s directory", g.OutputDir), nil
}

// createDirectoryStructure creates the Terraform directory structure
func (g *TerraformGenerator) createDirectoryStructure() error {
	// Create root output directory
	if err := utils.EnsureDirectoryExists(g.OutputDir); err != nil {
		return err
	}

	// Create modules directory if needed
	if g.Config.CreateModules {
		modulesDir := filepath.Join(g.OutputDir, "modules")
		if err := utils.EnsureDirectoryExists(modulesDir); err != nil {
			return err
		}

		// Create subdirectories for each module
		for _, moduleName := range g.Config.ModuleNames {
			moduleDir := filepath.Join(modulesDir, moduleName)
			if err := utils.EnsureDirectoryExists(moduleDir); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateRootModuleFiles generates the root module files
func (g *TerraformGenerator) generateRootModuleFiles() error {
	// Generate versions.tf
	versionsTf, err := g.generateVersionsFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "versions.tf"), versionsTf)
	if err != nil {
		return err
	}

	// Generate provider.tf
	providerTf, err := g.generateProviderFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "provider.tf"), providerTf)
	if err != nil {
		return err
	}

	// Generate main.tf
	mainTf, err := g.generateMainFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "main.tf"), mainTf)
	if err != nil {
		return err
	}

	// Generate variables.tf
	variablesTf, err := g.generateVariablesFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "variables.tf"), variablesTf)
	if err != nil {
		return err
	}

	// Generate outputs.tf
	outputsTf, err := g.generateOutputsFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "outputs.tf"), outputsTf)
	if err != nil {
		return err
	}

	// Generate terraform.tfvars
	tfvars, err := g.generateTfvarsFile()
	if err != nil {
		return err
	}
	err = utils.WriteToFile(filepath.Join(g.OutputDir, "terraform.tfvars"), tfvars)
	if err != nil {
		return err
	}

	return nil
}

// generateModuleFiles generates files for each module
func (g *TerraformGenerator) generateModuleFiles() error {
	// Generate VPC module files
	if contains(g.Config.ModuleNames, "vpc") {
		vpcDir := filepath.Join(g.OutputDir, "modules", "vpc")
		
		// VPC main.tf
		vpcMainTf, err := g.generateVpcModuleMainFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(vpcDir, "main.tf"), vpcMainTf)
		if err != nil {
			return err
		}
		
		// VPC variables.tf
		vpcVarsTf, err := g.generateVpcModuleVariablesFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(vpcDir, "variables.tf"), vpcVarsTf)
		if err != nil {
			return err
		}
		
		// VPC outputs.tf
		vpcOutputsTf, err := g.generateVpcModuleOutputsFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(vpcDir, "outputs.tf"), vpcOutputsTf)
		if err != nil {
			return err
		}
	}

	// Generate EKS module files
	if contains(g.Config.ModuleNames, "eks") {
		eksDir := filepath.Join(g.OutputDir, "modules", "eks")
		
		// EKS main.tf
		eksMainTf, err := g.generateEksModuleMainFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(eksDir, "main.tf"), eksMainTf)
		if err != nil {
			return err
		}
		
		// EKS variables.tf
		eksVarsTf, err := g.generateEksModuleVariablesFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(eksDir, "variables.tf"), eksVarsTf)
		if err != nil {
			return err
		}
		
		// EKS outputs.tf
		eksOutputsTf, err := g.generateEksModuleOutputsFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(eksDir, "outputs.tf"), eksOutputsTf)
		if err != nil {
			return err
		}
		
		// EKS iam.tf
		eksIamTf, err := g.generateEksModuleIamFile()
		if err != nil {
			return err
		}
		err = utils.WriteToFile(filepath.Join(eksDir, "iam.tf"), eksIamTf)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateVersionsFile generates the versions.tf file content
func (g *TerraformGenerator) generateVersionsFile() (string, error) {
	tmplStr := `terraform {
  required_version = ">= {{.TerraformVersion}}"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "{{.ProviderConstraint}}"
    }
  }

  {{if ne .BackendType "local"}}
  backend "{{.BackendType}}" {
    {{range $key, $value := .BackendConfig}}
    {{$key}} = "{{$value}}"
    {{end}}
  }
  {{end}}
}
`
	tmpl, err := template.New("versions").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.Config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateProviderFile generates the provider.tf file content
func (g *TerraformGenerator) generateProviderFile() (string, error) {
	tmplStr := `provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.default_tags
  }
}
`
	return tmplStr, nil
}

// generateMainFile generates the main.tf file content
func (g *TerraformGenerator) generateMainFile() (string, error) {
	hasVPC := contains(g.Config.ModuleNames, "vpc")
	hasEKS := contains(g.Config.ModuleNames, "eks")

	var mainFileContent bytes.Buffer

	if hasVPC {
		vpcModule := `module "vpc" {
  source = "./modules/vpc"

  vpc_name             = var.vpc_name
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  private_subnet_cidrs = var.private_subnet_cidrs
  public_subnet_cidrs  = var.public_subnet_cidrs
  enable_nat_gateway   = var.enable_nat_gateway
  single_nat_gateway   = var.single_nat_gateway
  
  tags = var.vpc_tags
}

`
		mainFileContent.WriteString(vpcModule)
	}

	if hasEKS {
		eksModule := `module "eks" {
  source = "./modules/eks"
  
  cluster_name    = var.cluster_name
  cluster_version = var.cluster_version
  
  vpc_id          = ${hasVPC ? "module.vpc.vpc_id" : "var.vpc_id"}
  subnet_ids      = ${hasVPC ? "module.vpc.private_subnet_ids" : "var.subnet_ids"}
  
  node_groups = var.node_groups
  
  tags = var.eks_tags
}

`
		// Replace the conditional strings
		eksContent := eksModule
		if hasVPC {
			eksContent = replaceConditional(eksContent, "${hasVPC ?", "}")
		} else {
			eksContent = replaceConditional(eksContent, "${hasVPC ?", "}", false)
		}
		mainFileContent.WriteString(eksContent)
	}

	return mainFileContent.String(), nil
}

// generateVariablesFile generates the variables.tf file content
func (g *TerraformGenerator) generateVariablesFile() (string, error) {
	hasVPC := contains(g.Config.ModuleNames, "vpc")
	hasEKS := contains(g.Config.ModuleNames, "eks")

	var variablesContent bytes.Buffer

	// Common variables
	commonVars := `variable "aws_region" {
  description = "AWS region to deploy resources into"
  type        = string
  default     = "` + g.Config.AwsRegion + `"
}

variable "default_tags" {
  description = "Default tags to apply to all resources"
  type        = map(string)
  default     = {
    Environment = "dev"
    ManagedBy   = "terraform"
    Project     = "iac-generator"
  }
}

`
	variablesContent.WriteString(commonVars)

	// VPC variables
	if hasVPC {
		vpcVars := `# VPC Variables
variable "vpc_name" {
  description = "Name of the VPC"
  type        = string
  default     = "main"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for the private subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for the public subnets"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateways for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "vpc_tags" {
  description = "Additional tags for the VPC"
  type        = map(string)
  default     = {}
}

`
		variablesContent.WriteString(vpcVars)
	} else if hasEKS {
		// If we have EKS but not VPC, we need VPC variables for the EKS module
		vpcInputs := `variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the EKS cluster"
  type        = list(string)
}

`
		variablesContent.WriteString(vpcInputs)
	}

	// EKS variables
	if hasEKS {
		eksVars := `# EKS Variables
variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
  default     = "main"
}

variable "cluster_version" {
  description = "Kubernetes version to use for the EKS cluster"
  type        = string
  default     = "1.28"
}

variable "node_groups" {
  description = "Map of EKS node group configurations"
  type        = map(object({
    instance_types       = list(string)
    capacity_type        = string
    desired_size         = number
    min_size             = number
    max_size             = number
    disk_size            = number
    additional_tags      = map(string)
  }))
  default     = {
    default = {
      instance_types       = ["t3.medium"]
      capacity_type        = "ON_DEMAND"
      desired_size         = 2
      min_size             = 1
      max_size             = 4
      disk_size            = 20
      additional_tags      = {}
    }
  }
}

variable "eks_tags" {
  description = "Additional tags for the EKS cluster"
  type        = map(string)
  default     = {}
}

`
		variablesContent.WriteString(eksVars)
	}

	return variablesContent.String(), nil
}

// generateOutputsFile generates the outputs.tf file content
func (g *TerraformGenerator) generateOutputsFile() (string, error) {
	hasVPC := contains(g.Config.ModuleNames, "vpc")
	hasEKS := contains(g.Config.ModuleNames, "eks")

	var outputsContent bytes.Buffer

	if hasVPC {
		vpcOutputs := `# VPC Outputs
output "vpc_id" {
  description = "The ID of the VPC"
  value       = module.vpc.vpc_id
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = module.vpc.private_subnet_ids
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = module.vpc.public_subnet_ids
}

`
		outputsContent.WriteString(vpcOutputs)
	}

	if hasEKS {
		eksOutputs := `# EKS Outputs
output "cluster_id" {
  description = "The name of the EKS cluster"
  value       = module.eks.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS cluster"
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = module.eks.cluster_iam_role_arn
}

output "oidc_provider_arn" {
  description = "The ARN of the OIDC Provider"
  value       = module.eks.oidc_provider_arn
}

output "node_security_group_id" {
  description = "Security group ID attached to the EKS nodes"
  value       = module.eks.node_security_group_id
}

`
		outputsContent.WriteString(eksOutputs)
	}

	return outputsContent.String(), nil
}

// generateTfvarsFile generates the terraform.tfvars file
func (g *TerraformGenerator) generateTfvarsFile() (string, error) {
	hasVPC := contains(g.Config.ModuleNames, "vpc")
	hasEKS := contains(g.Config.ModuleNames, "eks")

	var content bytes.Buffer

	content.WriteString(fmt.Sprintf(`aws_region = "%s"

default_tags = {
  Environment = "dev"
  ManagedBy   = "terraform"
  Project     = "iac-generator"
}

`, g.Config.AwsRegion))

	if hasVPC {
		content.WriteString(`# VPC Configuration
vpc_name = "main"
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
private_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
public_subnet_cidrs = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
enable_nat_gateway = true
single_nat_gateway = true
vpc_tags = {
  "kubernetes.io/cluster/main" = "shared"
}

`)
	}

	if hasEKS {
		content.WriteString(`# EKS Configuration
cluster_name = "main"
cluster_version = "1.28"

node_groups = {
  default = {
    instance_types = ["t3.medium"]
    capacity_type = "ON_DEMAND"
    desired_size = 2
    min_size = 1
    max_size = 4
    disk_size = 20
    additional_tags = {}
  }
  spot = {
    instance_types = ["t3.medium", "t3.large"]
    capacity_type = "SPOT"
    desired_size = 1
    min_size = 0
    max_size = 5
    disk_size = 20
    additional_tags = {
      "node-type" = "spot"
    }
  }
}

eks_tags = {
  "Environment" = "dev"
}

`)
	}

	return content.String(), nil
}

// generateVpcModuleMainFile generates the VPC module main.tf
func (g *TerraformGenerator) generateVpcModuleMainFile() (string, error) {
	tmplStr := `resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(
    var.tags,
    {
      Name = var.vpc_name
    }
  )
}

resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.this.id
  cidr_block              = element(var.public_subnet_cidrs, count.index)
  availability_zone       = element(var.availability_zones, count.index)
  map_public_ip_on_launch = true

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-public-${element(var.availability_zones, count.index)}"
      "kubernetes.io/role/elb" = "1"
    }
  )
}

resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id                  = aws_vpc.this.id
  cidr_block              = element(var.private_subnet_cidrs, count.index)
  availability_zone       = element(var.availability_zones, count.index)
  map_public_ip_on_launch = false

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-private-${element(var.availability_zones, count.index)}"
      "kubernetes.io/role/internal-elb" = "1"
    }
  )
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-igw"
    }
  )
}

resource "aws_eip" "nat" {
  count = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 0

  domain = "vpc"

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-nat-eip-${count.index + 1}"
    }
  )
}

resource "aws_nat_gateway" "this" {
  count = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 0

  allocation_id = element(aws_eip.nat.*.id, count.index)
  subnet_id     = element(aws_subnet.public.*.id, count.index)

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-nat-gw-${count.index + 1}"
    }
  )

  depends_on = [aws_internet_gateway.this]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  tags = merge(
    var.tags,
    {
      Name = "${var.vpc_name}-public-rt"
    }
  )
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.this.id

  timeouts {
    create = "5m"
  }
}

resource "aws_route_table_association" "public" {
  count = length(var.public_subnet_cidrs)

  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  count = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : length(var.availability_zones)

  vpc_id = aws_vpc.this.id

  tags = merge(
    var.tags,
    {
      Name = var.single_nat_gateway ? "${var.vpc_name}-private-rt" : "${var.vpc_name}-private-rt-${element(var.availability_zones, count.index)}"
    }
  )
}

resource "aws_route" "private_nat_gateway" {
  count = var.enable_nat_gateway ? (var.single_nat_gateway ? 1 : length(var.availability_zones)) : 0

  route_table_id         = element(aws_route_table.private.*.id, count.index)
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = element(aws_nat_gateway.this.*.id, var.single_nat_gateway ? 0 : count.index)

  timeouts {
    create = "5m"
  }
}

resource "aws_route_table_association" "private" {
  count = length(var.private_subnet_cidrs)

  subnet_id      = element(aws_subnet.private.*.id, count.index)
  route_table_id = element(
    aws_route_table.private.*.id,
    var.single_nat_gateway ? 0 : count.index,
  )
}
`
	return tmplStr, nil
}

// generateVpcModuleVariablesFile generates the VPC module variables.tf
func (g *TerraformGenerator) generateVpcModuleVariablesFile() (string, error) {
	tmplStr := `variable "vpc_name" {
  description = "Name of the VPC"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for the private subnets"
  type        = list(string)
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for the public subnets"
  type        = list(string)
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateways for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
`
	return tmplStr, nil
}

// generateVpcModuleOutputsFile generates the VPC module outputs.tf
func (g *TerraformGenerator) generateVpcModuleOutputsFile() (string, error) {
	tmplStr := `output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.this.id
}

output "vpc_cidr_block" {
  description = "The CIDR block of the VPC"
  value       = aws_vpc.this.cidr_block
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = aws_subnet.private.*.id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = aws_subnet.public.*.id
}

output "nat_gateway_ids" {
  description = "List of NAT Gateway IDs"
  value       = aws_nat_gateway.this.*.id
}

output "public_route_table_id" {
  description = "ID of the public route table"
  value       = aws_route_table.public.id
}

output "private_route_table_ids" {
  description = "List of private route table IDs"
  value       = aws_route_table.private.*.id
}
`
	return tmplStr, nil
}

// generateEksModuleMainFile generates the EKS module main.tf
func (g *TerraformGenerator) generateEksModuleMainFile() (string, error) {
	tmplStr := `resource "aws_eks_cluster" "this" {
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster.arn
  version  = var.cluster_version

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = var.endpoint_private_access
    endpoint_public_access  = var.endpoint_public_access
    security_group_ids      = var.security_group_ids
  }

  dynamic "kubernetes_network_config" {
    for_each = var.cluster_service_ipv4_cidr != null || var.cluster_ip_family != null ? [true] : []
    
    content {
      service_ipv4_cidr = var.cluster_service_ipv4_cidr
      ip_family         = var.cluster_ip_family
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSVPCResourceController,
  ]

  tags = merge(var.tags, {
    Name = var.cluster_name
  })
}

resource "aws_eks_node_group" "this" {
  for_each = var.node_groups

  cluster_name    = aws_eks_cluster.this.name
  node_group_name = each.key
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.subnet_ids

  instance_types = each.value.instance_types
  capacity_type  = each.value.capacity_type
  disk_size      = each.value.disk_size

  scaling_config {
    desired_size = each.value.desired_size
    min_size     = each.value.min_size
    max_size     = each.value.max_size
  }

  update_config {
    max_unavailable = 1
  }

  depends_on = [
    aws_iam_role_policy_attachment.node_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node_AmazonEC2ContainerRegistryReadOnly,
  ]

  tags = merge(
    var.tags,
    each.value.additional_tags,
    {
      Name = "${var.cluster_name}-${each.key}"
    }
  )
}

resource "aws_security_group" "cluster" {
  count = length(var.security_group_ids) == 0 ? 1 : 0
  
  name        = "${var.cluster_name}-cluster-sg"
  description = "Security group for EKS cluster"
  vpc_id      = var.vpc_id
  
  tags = merge(var.tags, {
    Name = "${var.cluster_name}-cluster-sg"
  })
}

resource "aws_security_group_rule" "cluster_egress" {
  count = length(var.security_group_ids) == 0 ? 1 : 0
  
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.cluster[0].id
}

resource "aws_security_group" "node" {
  name        = "${var.cluster_name}-node-sg"
  description = "Security group for EKS nodes"
  vpc_id      = var.vpc_id
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = merge(var.tags, {
    Name = "${var.cluster_name}-node-sg"
    "kubernetes.io/cluster/${var.cluster_name}" = "owned"
  })
}

# Allow nodes to communicate with each other
resource "aws_security_group_rule" "node_internal" {
  description       = "Allow nodes to communicate with each other"
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "-1"
  security_group_id = aws_security_group.node.id
  source_security_group_id = aws_security_group.node.id
}

# Allow worker Kubelets and pods to receive communication from the cluster control plane
resource "aws_security_group_rule" "node_cluster_inbound" {
  description       = "Allow worker nodes to receive communication from the cluster control plane"
  type              = "ingress"
  from_port         = 1025
  to_port           = 65535
  protocol          = "tcp"
  security_group_id = aws_security_group.node.id
  source_security_group_id = length(var.security_group_ids) == 0 ? aws_security_group.cluster[0].id : var.security_group_ids[0]
}

# Allow cluster control plane to receive communication from the worker Kubelets
resource "aws_security_group_rule" "cluster_node_inbound" {
  count = length(var.security_group_ids) == 0 ? 1 : 0
  
  description       = "Allow cluster control plane to receive communication from the worker Kubelets"
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.cluster[0].id
  source_security_group_id = aws_security_group.node.id
}

# Create OIDC identity provider for the cluster
data "tls_certificate" "this" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "this" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.this.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer
  
  tags = merge(var.tags, {
    Name = "${var.cluster_name}-oidc-provider"
  })
}
`
	return tmplStr, nil
}

// generateEksModuleVariablesFile generates the EKS module variables.tf
func (g *TerraformGenerator) generateEksModuleVariablesFile() (string, error) {
	tmplStr := `variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "cluster_version" {
  description = "Kubernetes version to use for the EKS cluster"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the EKS cluster"
  type        = list(string)
}

variable "endpoint_private_access" {
  description = "Whether to enable private access to the cluster's Kubernetes API"
  type        = bool
  default     = true
}

variable "endpoint_public_access" {
  description = "Whether to enable public access to the cluster's Kubernetes API"
  type        = bool
  default     = true
}

variable "security_group_ids" {
  description = "List of security group IDs for the EKS cluster"
  type        = list(string)
  default     = []
}

variable "cluster_service_ipv4_cidr" {
  description = "The CIDR block to assign Kubernetes service IP addresses from"
  type        = string
  default     = null
}

variable "cluster_ip_family" {
  description = "The IP family used to assign Kubernetes pod and service addresses"
  type        = string
  default     = null
  validation {
    condition     = var.cluster_ip_family == null || var.cluster_ip_family == "ipv4" || var.cluster_ip_family == "ipv6"
    error_message = "Valid values for cluster_ip_family are 'ipv4' and 'ipv6'."
  }
}

variable "node_groups" {
  description = "Map of EKS node group configurations"
  type        = map(object({
    instance_types       = list(string)
    capacity_type        = string
    desired_size         = number
    min_size             = number
    max_size             = number
    disk_size            = number
    additional_tags      = map(string)
  }))
  default     = {
    default = {
      instance_types       = ["t3.medium"]
      capacity_type        = "ON_DEMAND"
      desired_size         = 2
      min_size             = 1
      max_size             = 4
      disk_size            = 20
      additional_tags      = {}
    }
  }
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
`
	return tmplStr, nil
}

// generateEksModuleOutputsFile generates the EKS module outputs.tf
func (g *TerraformGenerator) generateEksModuleOutputsFile() (string, error) {
	tmplStr := `output "cluster_id" {
  description = "The name of the EKS cluster"
  value       = aws_eks_cluster.this.id
}

output "cluster_arn" {
  description = "The Amazon Resource Name (ARN) of the EKS cluster"
  value       = aws_eks_cluster.this.arn
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS cluster"
  value       = aws_eks_cluster.this.endpoint
}

output "cluster_ca_certificate" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = aws_eks_cluster.this.certificate_authority[0].data
  sensitive   = true
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = length(var.security_group_ids) == 0 ? aws_security_group.cluster[0].id : var.security_group_ids[0]
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = aws_iam_role.cluster.arn
}

output "node_security_group_id" {
  description = "Security group ID attached to the EKS nodes"
  value       = aws_security_group.node.id
}

output "node_iam_role_arn" {
  description = "IAM role ARN of the EKS node groups"
  value       = aws_iam_role.node.arn
}

output "oidc_provider_arn" {
  description = "The ARN of the OIDC Provider"
  value       = aws_iam_openid_connect_provider.this.arn
}
`
	return tmplStr, nil
}

// generateEksModuleIamFile generates the EKS module iam.tf
func (g *TerraformGenerator) generateEksModuleIamFile() (string, error) {
	tmplStr := `# IAM Role for EKS Cluster
resource "aws_iam_role" "cluster" {
  name = "${var.cluster_name}-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-cluster-role"
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.cluster.name
}

# IAM Role for EKS Node Groups
resource "aws_iam_role" "node" {
  name = "${var.cluster_name}-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-node-role"
  })
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

# Create IAM OIDC provider policy for service accounts
data "aws_iam_policy_document" "oidc_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    condition {
      test     = "StringEquals"
      variable = "${replace(aws_iam_openid_connect_provider.this.url, "https://", "")}:sub"
      values   = ["system:serviceaccount:kube-system:aws-node"]
    }

    principals {
      identifiers = [aws_iam_openid_connect_provider.this.arn]
      type        = "Federated"
    }
  }
}

# Example IAM role for pod service accounts
resource "aws_iam_role" "service_account" {
  name               = "${var.cluster_name}-service-account-role"
  assume_role_policy = data.aws_iam_policy_document.oidc_assume_role_policy.json
  
  tags = merge(var.tags, {
    Name = "${var.cluster_name}-service-account-role"
  })
}
`
	return tmplStr, nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func replaceConditional(str string, start string, end string, useFirst ...bool) string {
	takeFirst := true
	if len(useFirst) > 0 {
		takeFirst = useFirst[0]
	}

	startIdx := strings.Index(str, start)
	endIdx := strings.Index(str, end)
	
	if startIdx == -1 || endIdx == -1 {
		return str
	}
	
	condStart := startIdx + len(start)
	parts := strings.Split(str[condStart:endIdx], ":")
	
	if len(parts) < 2 {
		return str
	}
	
	choice := ""
	if takeFirst {
		choice = strings.TrimSpace(parts[0])
	} else {
		choice = strings.TrimSpace(parts[1])
	}
	
	return str[:startIdx] + choice + str[endIdx+len(end):]
}

// mapResourceType maps our internal resource type to Terraform resource type
func mapResourceType(resourceType models.ResourceType) (string, error) {
	mapping := map[models.ResourceType]string{
		models.ResourceEC2Instance:    "aws_instance",
		models.ResourceS3Bucket:       "aws_s3_bucket",
		models.ResourceVPC:            "aws_vpc",
		models.ResourceSubnet:         "aws_subnet",
		models.ResourceSecurityGroup:  "aws_security_group",
		models.ResourceRDSInstance:    "aws_db_instance",
		models.ResourceIAMRole:        "aws_iam_role",
		models.ResourceLambda:         "aws_lambda_function",
		models.ResourceDynamoDB:       "aws_dynamodb_table",
		models.ResourceCloudwatch:     "aws_cloudwatch_metric_alarm",
		models.ResourceIGW:            "aws_internet_gateway",
		models.ResourceNATGateway:     "aws_nat_gateway",
		models.ResourceEKSCluster:     "aws_eks_cluster",
		models.ResourceNodeGroup:      "aws_eks_node_group",
	}

	if terraformType, ok := mapping[resourceType]; ok {
		return terraformType, nil
	}

	return "", fmt.Errorf("unsupported resource type: %s", resourceType)
}

// ModelToTerraformModel converts infrastructure model to Terraform-specific model
func ModelToTerraformModel(model *infra.Infrastructure) (*models.InfrastructureModel, error) {
	tfModel := models.NewInfrastructureModel()
	
	// Process VPCs
	for _, vpc := range model.VPCs {
		// VPC resource
		vpcResource := models.NewResource(models.ResourceVPC, vpc.Name)
		vpcResource.AddProperty("cidr_block", vpc.CIDR)
		vpcResource.AddProperty("enable_dns_support", vpc.EnableDNSSupport)
		vpcResource.AddProperty("enable_dns_hostnames", vpc.EnableDNSHostname)
		
		for k, v := range vpc.Tags {
			vpcResource.AddProperty(fmt.Sprintf("tag.%s", k), v)
		}
		
		tfModel.AddResource(vpcResource)
		
		// Subnet resources
		for _, subnet := range vpc.Subnets {
			subnetResource := models.NewResource(models.ResourceSubnet, subnet.Name)
			subnetResource.AddProperty("vpc_id", fmt.Sprintf("${aws_vpc.%s.id}", vpc.Name))
			subnetResource.AddProperty("cidr_block", subnet.CIDR)
			subnetResource.AddProperty("availability_zone", subnet.AvailabilityZone)
			subnetResource.AddProperty("map_public_ip_on_launch", subnet.IsPublic)
			
			for k, v := range subnet.Tags {
				subnetResource.AddProperty(fmt.Sprintf("tag.%s", k), v)
			}
			
			subnetResource.AddDependency(vpc.Name)
			tfModel.AddResource(subnetResource)
		}
		
		// Internet Gateway resources
		for _, igw := range vpc.InternetGateways {
			igwResource := models.NewResource(models.ResourceIGW, igw.Name)
			igwResource.AddProperty("vpc_id", fmt.Sprintf("${aws_vpc.%s.id}", vpc.Name))
			
			for k, v := range igw.Tags {
				igwResource.AddProperty(fmt.Sprintf("tag.%s", k), v)
			}
			
			igwResource.AddDependency(vpc.Name)
			tfModel.AddResource(igwResource)
		}
		
		// NAT Gateway resources
		for _, natgw := range vpc.NATGateways {
			natResource := models.NewResource(models.ResourceNATGateway, natgw.Name)
			natResource.AddProperty("subnet_id", fmt.Sprintf("${aws_subnet.%s.id}", natgw.Subnet))
			natResource.AddProperty("connectivity_type", natgw.ConnectivityType)
			
			if natgw.AllocationID != "" {
				natResource.AddProperty("allocation_id", natgw.AllocationID)
			}
			
			for k, v := range natgw.Tags {
				natResource.AddProperty(fmt.Sprintf("tag.%s", k), v)
			}
			
			natResource.AddDependency(natgw.Subnet)
			tfModel.AddResource(natResource)
		}
	}
	
	return tfModel, nil
}
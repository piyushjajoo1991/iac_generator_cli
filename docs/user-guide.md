# IaC Manifest Generator CLI User Guide

This user guide provides detailed instructions for using the IaC Manifest Generator CLI tool. It covers command-line options, infrastructure description formats, supported resources, and common usage patterns.

## Table of Contents

- [Command Line Interface](#command-line-interface)
  - [Global Options](#global-options)
  - [Generate Command](#generate-command)
- [Infrastructure Description Format](#infrastructure-description-format)
  - [Guidelines for Writing Descriptions](#guidelines-for-writing-descriptions)
  - [Supported Resource Types](#supported-resource-types)
  - [Resource Properties](#resource-properties)
- [Output Formats](#output-formats)
  - [Terraform Output](#terraform-output)
  - [Crossplane Output](#crossplane-output)
- [Configuration File](#configuration-file)
- [Output Directory Structure](#output-directory-structure)
- [Template System](#template-system)
- [Limitations and Constraints](#limitations-and-constraints)
- [Troubleshooting](#troubleshooting)

## Command Line Interface

The IaC Manifest Generator CLI (`iacgen`) provides a command-line interface for generating Infrastructure as Code manifests from natural language descriptions.

### Global Options

These options apply to all commands:

| Option            | Short | Description                                     | Default      |
|-------------------|-------|-------------------------------------------------|--------------|
| `--output`        | `-o`  | Output format (terraform or crossplane)         | terraform    |
| `--output-dir`    | `-d`  | Directory to write output files                 | .            |
| `--region`        |       | AWS region for resources                        | us-east-1    |
| `--config`        |       | Config file (default is $HOME/.iacgen.yaml)     | -            |
| `--use-templates` |       | Use the template system for generating IaC code | false        |
| `--debug`         | `-v`  | Enable debug output                             | false        |

### Generate Command

The `generate` command is the primary function of the tool. It processes a natural language description of infrastructure and generates the corresponding IaC manifests.

```bash
iacgen generate [OPTIONS] [DESCRIPTION]
```

#### Options

| Option          | Short | Description                                     | Default        |
|-----------------|-------|-------------------------------------------------|----------------|
| `--file`        | `-f`  | Input file containing infrastructure description | -              |
| `--output-file` |       | Output filename                                 | auto-generated |

#### Examples

```bash
# Generate Terraform from command line description
iacgen generate "Create a VPC with 3 public subnets in us-east-1"

# Generate Crossplane from file with custom output directory
iacgen generate -f requirements.txt -o crossplane -d ./manifests

# Enable debug output
iacgen generate -v "Create an EKS cluster with 3 nodes"

# Use the template system
iacgen generate --use-templates "Create an S3 bucket with versioning enabled"

# Create complex networking with multiple subnets
iacgen generate "Create a VPC with CIDR 10.0.0.0/16 in us-west-2 with 3 public and 3 private subnets across all availability zones, including NAT gateways for private subnet internet access"

# Deploy EKS with specific node configurations
iacgen generate "Create an EKS cluster version 1.28 in private subnets with 2 node groups - one with t3.large instances for general workloads and another with c5.xlarge instances for compute-intensive tasks"

# Set up a database with specific configurations
iacgen generate "Create an RDS Aurora PostgreSQL cluster with multi-AZ deployment, 2 instances, and 100GB storage that auto-scales to 1TB"

# Create serverless infrastructure
iacgen generate "Create a serverless API with Lambda functions, API Gateway, and DynamoDB table with on-demand capacity"

# Deploy multi-tier web application infrastructure
iacgen generate -o terraform -d ./infra "Create a 3-tier web application with an Application Load Balancer, auto-scaling EC2 instances in private subnets, and an RDS MySQL database with encryption enabled"
```

## Infrastructure Description Format

The tool uses natural language processing to interpret English descriptions of infrastructure requirements.

### Guidelines for Writing Descriptions

For best results, follow these guidelines when writing infrastructure descriptions:

1. **Be specific about resource types and properties**: Clearly mention resource types (e.g., VPC, subnet, EC2 instance) and their important properties (e.g., CIDR blocks, instance types).

2. **Use simple, declarative language**: Write in a straightforward manner using present tense.

3. **Specify regions explicitly**: Include the AWS region if you want resources deployed in a specific region.

4. **Specify resource names when important**: Provide names for resources when you want to reference them specifically.

5. **Structure complex descriptions**: For complex infrastructure, structure your description by resource type or logical grouping.

### Examples of Good Descriptions

```
Create a VPC with CIDR block 10.0.0.0/16 in us-east-1 region. Include 3 public subnets and 3 private subnets across all availability zones. Add an internet gateway for public subnets and NAT gateways for private subnet outbound traffic.

Deploy an EKS cluster named 'production' in the private subnets using version 1.27. Create a node group with 3 t3.large instances that can scale up to 10 nodes.
```

```
Set up a secure web hosting infrastructure with:
- VPC with CIDR 172.16.0.0/16
- 2 public subnets for load balancers
- 4 private subnets for application servers
- EC2 instances using t3.medium instance type
- S3 bucket for static assets with versioning enabled
```

```
Create a serverless data processing pipeline with:
- Lambda functions for data transformation
- API Gateway with REST endpoints
- DynamoDB table with on-demand capacity and a GSI on the 'status' attribute
- S3 bucket for raw data storage with lifecycle policies
- CloudWatch Events for scheduled processing
- IAM roles with least privilege permissions
```

```
Set up a highly available database infrastructure in us-west-2 with:
- Aurora PostgreSQL cluster version 13.7
- 2 database instances in different availability zones
- 100GB storage with auto-scaling up to 1TB
- Daily automated backups retained for 7 days
- Enhanced monitoring with 1-minute intervals
- Database subnet group in private subnets
- Security group allowing access only from application servers
```

```
Deploy a containerized microservices platform using:
- EKS cluster version 1.28 in us-east-1
- 3 node groups (general: t3.large, compute: c5.xlarge, memory: r5.large)
- Cluster autoscaler enabled for all node groups
- VPC with private subnets for pods and public subnets for load balancers
- AWS Load Balancer Controller for ingress
- Cluster logging to CloudWatch Logs
- ECR repositories for container images with image scanning
```

### Supported Resource Types

The tool can identify and generate configurations for the following AWS resource types:

| Resource Type           | Description                                         |
|-------------------------|-----------------------------------------------------|
| VPC                     | Virtual Private Cloud network isolation             |
| Subnet                  | Network subdivision within a VPC                    |
| Internet Gateway        | Enables internet access for resources in a VPC      |
| NAT Gateway             | Enables outbound internet access for private subnets|
| EKS Cluster             | Managed Kubernetes service                          |
| EKS Node Group          | Worker nodes for EKS clusters                       |
| EC2 Instance            | Virtual machines                                    |
| S3 Bucket               | Object storage                                      |
| Security Group          | Virtual firewall for resources                      |
| IAM Role                | Identity and access management role                 |
| RDS Instance            | Relational database service                         |
| DynamoDB Table          | NoSQL database service                              |
| Lambda Function         | Serverless compute service                          |
| CloudWatch Alarm        | Monitoring and alerting                             |

### Resource Properties

Each resource type has specific properties that can be specified in your description:

#### VPC Properties

- CIDR block (e.g., "10.0.0.0/16")
- DNS support (enabled/disabled)
- DNS hostnames (enabled/disabled)
- Region (e.g., "us-east-1")
- Name (e.g., "main-vpc", "production-vpc")

#### Subnet Properties

- CIDR block (e.g., "10.0.1.0/24")
- Availability Zone (e.g., "us-east-1a")
- Public/Private designation
- VPC association

#### EKS Cluster Properties

- Kubernetes version (e.g., "1.26", "1.27")
- Public/Private API endpoint access
- Service role
- Subnet placement

#### EC2 Instance Properties

- Instance type (e.g., "t2.micro", "m5.large")
- AMI ID
- Region
- Subnet placement
- Security group associations
- Key name for SSH access

#### S3 Bucket Properties

- Bucket name
- Versioning (enabled/disabled)
- Access control (public/private)
- Encryption settings
- Website hosting configuration

## Output Formats

The tool can generate infrastructure manifests in two formats:

### Terraform Output

Terraform output is generated as HashiCorp Configuration Language (HCL) files. The tool generates a complete Terraform project with:

- `main.tf`: Main resource definitions
- `variables.tf`: Variable declarations
- `outputs.tf`: Output definitions
- `provider.tf`: Provider configuration
- `versions.tf`: Terraform version constraints
- `terraform.tfvars`: Variable values

For more complex infrastructure, the tool may generate a modular structure with subdirectories for each component.

#### Example Terraform Output

```hcl
# main.tf
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "main-vpc"
  }
}

resource "aws_subnet" "public" {
  count                   = length(var.public_subnet_cidrs)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = element(var.public_subnet_cidrs, count.index)
  availability_zone       = element(var.availability_zones, count.index)
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-${count.index + 1}"
  }
}
```

### Crossplane Output

Crossplane output is generated as YAML files following the Crossplane resource model. The tool generates a Kubernetes-native structure with:

- Provider configurations
- Resource definitions using Crossplane's AWS provider
- Organized directory structure with kustomization support

#### Example Crossplane Output

```yaml
# vpc.yaml
apiVersion: ec2.aws.crossplane.io/v1beta1
kind: VPC
metadata:
  name: main-vpc
spec:
  forProvider:
    region: us-east-1
    cidrBlock: 10.0.0.0/16
    enableDnsSupport: true
    enableDnsHostnames: true
    tags:
      - key: Name
        value: main-vpc
  providerConfigRef:
    name: aws-provider
```

## Configuration File

The tool supports a configuration file located at `~/.iacgen.yaml` to specify default settings. This is useful for setting frequently used options.

Example configuration file:

```yaml
# ~/.iacgen.yaml
log_level: info
output_dir: ./output
default_type: terraform
aws_region: us-east-1
use_templates: false
```

### Available Configuration Options

| Option          | Description                                     | Default      |
|-----------------|-------------------------------------------------|--------------|
| `log_level`     | Logging level (debug, info, warn, error)        | info         |
| `output_dir`    | Default output directory for generated files    | .            |
| `default_type`  | Default output format (terraform or crossplane) | terraform    |
| `aws_region`    | Default AWS region for resources                | us-east-1    |
| `use_templates` | Whether to use the template system by default   | false        |

## Output Directory Structure

### Terraform Output Structure

When generating Terraform configurations, the tool creates the following directory structure:

```
output-dir/
├── main.tf           # Main resource definitions
├── variables.tf      # Variable declarations
├── outputs.tf        # Output definitions
├── provider.tf       # AWS provider configuration
├── versions.tf       # Terraform version constraints
├── terraform.tfvars  # Default variable values
└── modules/          # Optional, for complex infrastructure
    ├── vpc/          # VPC module
    │   ├── main.tf
    │   ├── variables.tf
    │   └── outputs.tf
    └── eks/          # EKS module
        ├── main.tf
        ├── variables.tf
        ├── outputs.tf
        └── iam.tf
```

### Crossplane Output Structure

When generating Crossplane manifests, the tool creates the following directory structure:

```
output-dir/
├── kustomization.yaml        # Root kustomization file
├── README.md                 # Generated documentation
├── base/                     # Base resources
│   ├── kustomization.yaml    # Base kustomization file
│   └── provider.yaml         # AWS provider configuration
├── vpc/                      # VPC resources
│   ├── kustomization.yaml    # VPC kustomization file
│   ├── vpc.yaml              # VPC definition
│   ├── subnets.yaml          # Subnet definitions
│   ├── gateways.yaml         # Internet and NAT gateways
│   └── routing.yaml          # Route tables and associations
└── eks/                      # EKS resources
    ├── kustomization.yaml    # EKS kustomization file
    ├── cluster.yaml          # EKS cluster definition
    ├── nodegroup.yaml        # EKS node group
    └── iam.yaml              # IAM roles and policies
```

## Template System

The IaC Manifest Generator includes a template system for customizing the generated output. This feature is enabled with the `--use-templates` flag.

### Template Locations

Templates are stored in the `internal/template/templates/` directory, with subdirectories for each supported IaC format:

```
internal/template/templates/
├── terraform/
│   ├── ec2_instance.tmpl
│   ├── eks_cluster.tmpl
│   ├── subnet.tmpl
│   └── vpc.tmpl
└── crossplane/
    ├── ec2_instance.tmpl
    ├── eks_cluster.tmpl
    ├── subnet.tmpl
    └── vpc.tmpl
```

### Custom Templates

To customize templates:

1. Fork the repository
2. Modify templates in the appropriate subdirectory
3. Rebuild the tool
4. Use with the `--use-templates` flag

Templates use Go's text/template syntax and have access to a variety of helper functions for formatting, string manipulation, and resource referencing.

## Limitations and Constraints

The IaC Manifest Generator has the following limitations:

1. **AWS-only Support**: Currently only supports AWS cloud resources.

2. **NLP Limitations**: Natural language processing has a limited vocabulary and might not understand all infrastructure terms or complex requirements.

3. **Resource Relationship Detection**: Complex relationships between resources may not be detected automatically and might require manual adjustment.

4. **Default Values**: The tool uses sensible defaults when specific values aren't provided, which may need customization.

5. **Security Configurations**: Basic security settings are applied, but security-focused configurations should be reviewed and enhanced.

6. **State Management**: The tool generates initial manifests but doesn't handle state management for existing infrastructure.

7. **Validation**: Limited validation of resource configurations; the generated output should be reviewed before applying.

## Troubleshooting

### Common Issues

1. **Unrecognized Resource Types**: If the tool doesn't recognize a resource type, try using more standard AWS terminology in your description.

2. **Incorrect Resource Relationships**: Check the generated files for correct resource references. You may need to manually adjust dependencies.

3. **Missing Properties**: If important properties are missing, try specifying them explicitly in your description.

### Debugging

Use the `--debug` or `-v` flag to enable debug output for troubleshooting:

```bash
iacgen generate -v "Create a VPC with 3 public subnets"
```

### Logs

Logs are written to stderr by default. You can redirect them to a file:

```bash
iacgen generate "Create a VPC" 2> iacgen.log
```

### Reporting Issues

If you encounter bugs or have feature requests, please submit them to the GitHub repository's issue tracker at: https://github.com/riptano/iac_generator_cli/issues
# IaC Manifest Generator CLI

A Go CLI tool that parses English descriptions of AWS infrastructure and generates Terraform and Crossplane manifests. This tool was created using claude-code CLI, even the prompts to write the code were generated. Please see the section at the end for the prompts used to generate this code. I have also added the amount of $$$ spent on building this tool.

## Overview

IaC Manifest Generator CLI (iacgen) is a powerful utility that uses natural language processing to interpret plain English descriptions of infrastructure requirements. It automatically generates Infrastructure as Code (IaC) manifests that can be used to provision and manage AWS resources. This tool bridges the gap between infrastructure requirements expressed in natural language and the technical specifications needed for deployment.

## Features

- **Natural Language Processing**: Parse plain English descriptions of AWS infrastructure
- **Multi-IaC Support**: Generate both Terraform HCL and Crossplane YAML manifests
- **AWS Resource Support**: Support for common AWS resources including:
  - VPCs, Subnets, Internet Gateways, and NAT Gateways
  - EKS Clusters and Node Groups
  - EC2 Instances
  - S3 Buckets
  - Security Groups
  - IAM Roles
  - and more
- **Template System**: Optional template-based generation for customized output
- **Pipeline Architecture**: Modular design allowing for easy extension

## Installation

### Prerequisites

- Go 1.19 or later
- AWS CLI (for testing generated manifests)
- Terraform (optional, for using generated HCL files)
- Kubernetes with Crossplane (optional, for using generated YAML files)

### From Source

```bash
# Clone the repository
git clone https://github.com/riptano/iac_generator_cli.git

# Navigate to the project directory
cd iac_generator_cli

# Build the project
go build -o iacgen

# Install globally (optional)
go install
```

## Usage

### Basic Usage

```bash
# Generate Terraform configuration from a description
./iacgen generate "Create an EC2 instance with t2.micro size in us-west-2 region"

# Generate Crossplane manifests
./iacgen generate --output crossplane "Create an S3 bucket named 'my-data' with versioning enabled"

# Read from a file
./iacgen generate --file infrastructure.txt

# Write to an output directory
./iacgen generate --file infrastructure.txt --output-dir ./output
```

### Configuration File

Create a configuration file at `~/.iacgen.yaml` with the following settings:

```yaml
log_level: info
output_dir: ./output
default_type: terraform
aws_region: us-east-1
use_templates: false
```

### Command-Line Options

| Option          | Short | Description                                   | Default      |
|-----------------|-------|-----------------------------------------------|--------------|
| `--output`      | `-o`  | Output format (terraform or crossplane)       | terraform    |
| `--output-dir`  | `-d`  | Directory to write output files               | .            |
| `--file`        | `-f`  | Input file containing infrastructure description | -         |
| `--region`      |       | AWS region for resources                      | us-east-1    |
| `--config`      |       | Config file (default is $HOME/.iacgen.yaml)   | -            |
| `--use-templates` |     | Use the template system for generating IaC code | false      |
| `--debug`       | `-v`  | Enable debug output                           | false        |
| `--output-file` |       | Output filename                               | auto-generated |

## Infrastructure Description Format

The tool can understand a variety of natural language descriptions, including:

```
Create a VPC in us-east-1 with 3 public and 3 private subnets across all availability zones.
Add an internet gateway and NAT gateways for outbound traffic from private subnets.

Deploy an EKS cluster in the private subnets with version 1.27 and enable public access to the API.
Create a node group with 2 t3.medium instances that can scale up to 5 nodes.
```

### Supported Resource Types and Properties

| Resource Type | Example Properties |
|---------------|-------------------|
| VPC | CIDR block, DNS support, DNS hostnames |
| Subnet | CIDR block, Availability Zone, Public/Private |
| EKS Cluster | Version, API access, Subnet placement |
| EC2 Instance | Instance type, AMI, Region |
| S3 Bucket | Name, Versioning, Access control |
| Security Group | Ingress/Egress rules, Ports |

## Examples

### VPC with Subnets

```bash
./iacgen generate "Create a VPC with CIDR 10.0.0.0/16 in us-east-1, with 2 public and 2 private subnets. Add an internet gateway and NAT gateway."
```

### EKS Cluster

```bash
./iacgen generate "Create an EKS cluster named 'production' with version 1.28 in us-west-2. Deploy in private subnets with 3 t3.large nodes."
```

### S3 Website Hosting

```bash
./iacgen generate "Create an S3 bucket for static website hosting with versioning enabled and public read access."
```

## Extending the Tool

### Adding New Resource Types

1. Add the new resource type to `pkg/models/resource.go`
2. Implement resource creation in `internal/infra/aws.go`
3. Add templates for the new resource in both adapters

### Supporting Additional IaC Tools

1. Create a new adapter package in `internal/adapter/`
2. Implement the Generator interface
3. Update the `GenerateManifest` function in `internal/generator/generator.go`

## Project Structure

```
iac_generator_cli/
├── cmd/
│   └── iacgen/            # CLI commands
├── internal/
│   ├── adapter/
│   │   ├── crossplane/    # Crossplane manifest generation
│   │   └── terraform/     # Terraform manifest generation
│   ├── config/            # Configuration handling
│   ├── generator/         # IaC generation orchestration
│   ├── infra/             # Infrastructure modeling
│   ├── nlp/               # Natural language processing
│   ├── pipeline/          # Processing pipeline components
│   ├── template/          # Template system
│   └── utils/             # Utility functions
├── pkg/
│   └── models/            # Shared model definitions
├── main.go                # Application entry point
└── README.md              # Project documentation
```

## Limitations

- The tool currently only supports AWS resources
- Complex relationships between resources might require manual adjustments
- The NLP processor has a limited vocabulary and may not understand all infrastructure concepts

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT License](LICENSE)

---
# Claude Code Prompts for IaC Manifest Generator CLI in Golang

Here are structured prompts designed for using Claude Code to implement the IaC Manifest Generator CLI in Golang. Each prompt focuses on a specific component of the MVP implementation.

## Project Setup Prompt

```
Create a Go project structure for an IaC Manifest Generator CLI tool. The tool will parse English descriptions of AWS infrastructure and generate Terraform and Crossplane manifests. 

Include:
1. Main package with CLI entry point
2. Packages for natural language processing, infrastructure modeling, and manifest generation
3. Tool-specific adapter packages for Terraform and Crossplane
4. Utility and configuration packages
5. Appropriate Go module initialization
6. Directory structure following Go best practices

For dependencies, include starter code with:
- cobra for CLI functionality
- viper for configuration management
- A basic logging mechanism
```

## CLI Framework Prompt

```
Implement the main CLI framework for the IaC Manifest Generator using Go's cobra library. 

Requirements:
1. Create a root command with appropriate description and flags
2. Implement subcommands for "generate" with options for:
   - Input text or file path
   - Output directory
   - IaC tool selection (terraform/crossplane)
3. Add flag for AWS region with default as "us-east-1"
4. Include verbose/debug logging options
5. Implement proper error handling and user feedback
6. Structure the command execution to call the appropriate processing pipeline

The CLI should be user-friendly and provide clear help documentation.
```

## NLP Parser Implementation Prompt

```
Create a pattern-based natural language processor in Go that can parse AWS infrastructure descriptions.

Requirements:
1. Implement functions to extract:
   - VPC details (CIDR ranges if specified, otherwise use default)
   - Subnet information (public/private, count per AZ)
   - Internet Gateway and NAT Gateway requirements
   - EKS cluster specifications (API access mode, version if specified)
   - Node pool details (instance type, count)

2. Use regular expressions and string matching rather than ML
3. Extract numerical values and infrastructure component relationships
4. Validate the extracted information for completeness
5. Return a structured data model representing the infrastructure requirements

Example input to handle:
"AWS infra in us-east-1 with a vpc, 3 public and 3 private subnets, 1 IGW and 3 NAT gateways per az, plus an EKS Cluster with public and private api access deployed in private vpcs in 3 azs with a nodepool on t3-medium instance type"
```

## Infrastructure Model Prompt

```
Implement a Go package that defines the domain models for AWS infrastructure components.

Create structs for:
1. Infrastructure - the top-level container for all resources
2. VPC - with region, CIDR, and associated resources
3. Subnet - with AZ, CIDR, and public/private designation
4. InternetGateway - with VPC association
5. NATGateway - with subnet association
6. EKSCluster - with API access mode and version
7. NodePool - with instance type and count

Include:
- Appropriate field tags for JSON/YAML marshaling
- Constructor functions with reasonable defaults
- Validation methods to ensure consistency
- String representation methods for logging/debugging
- Methods to calculate derived properties (e.g., CIDR block allocation)

These models should be tool-agnostic and serve as the intermediate representation.
```

## Terraform Generator Prompt

```
Create a Terraform adapter package in Go that transforms the infrastructure model into Terraform HCL files.

Implement:
1. A directory structure generator following Terraform best practices:
   - modules/ directory with subdirectories for vpc and eks
   - main.tf, variables.tf, outputs.tf, versions.tf in root
   - Appropriate files within each module

2. Template-based code generation for:
   - AWS provider configuration
   - VPC resources (vpc, subnets, route tables)
   - Internet Gateway and NAT Gateway resources
   - EKS cluster resources and node groups
   - Necessary IAM roles and policies

3. Functions to:
   - Convert infrastructure model to HCL
   - Write files to the appropriate locations
   - Generate variable definitions with defaults
   - Create proper resource references and dependencies

Use Go's text/template package or a similar library for template rendering.
```

## Crossplane Generator Prompt

```
Implement a Crossplane adapter package in Go that converts the infrastructure model into Kubernetes YAML manifests for Crossplane AWS Provider.

Create:
1. A directory structure organizer following Kubernetes/Crossplane conventions:
   - base/ directory for common resources
   - vpc/ directory for networking components
   - eks/ directory for cluster resources
   - kustomization files where appropriate

2. Functions to generate YAML manifests for:
   - VPC and subnet resources using Crossplane AWS provider
   - Internet Gateway and NAT Gateway compositions
   - EKS cluster configuration
   - Node pool definitions
   - Required Crossplane provider configurations

3. Utilities for:
   - Converting internal models to Crossplane resource definitions
   - Generating appropriate Kubernetes metadata
   - Establishing proper resource dependencies
   - Writing YAML files with correct formatting

Use Go's yaml package for marshaling structs to YAML.
```

## Template System Prompt

```
Create a template system in Go for generating IaC code files.

Requirements:
1. Implement a template manager that:
   - Loads and caches templates from embedded files
   - Selects appropriate templates based on resource type
   - Handles both Terraform HCL and Crossplane YAML formats

2. Create template files for:
   - VPC networking components
   - EKS cluster configuration
   - Node pools and instance configurations
   - Provider setup and versions

3. Implement template rendering functions that:
   - Apply data from infrastructure models to templates
   - Handle conditional sections based on configuration
   - Format output according to tool-specific conventions
   - Validate rendered content for basic syntax

4. Create helper functions for common template operations

Use Go's embed package to include templates in the binary for easy distribution.
```

## Pipeline Integration Prompt

```
Implement the complete processing pipeline for the IaC Manifest Generator CLI in Go.

Create:
1. A pipeline coordinator that:
   - Takes input from CLI flags and arguments
   - Calls the NLP processor to extract infrastructure requirements
   - Builds the infrastructure model from parsed information
   - Selects and invokes the appropriate IaC tool adapter
   - Directs output to specified location
   - Provides user feedback on progress and results

2. Implement proper error handling:
   - Validate inputs before processing
   - Provide meaningful error messages for parsing failures
   - Handle file system access errors gracefully
   - Implement proper logging at each stage

3. Create pipeline interfaces that allow for:
   - Testability of each component
   - Future extension to new IaC tools
   - Separation of concerns between components

The pipeline should be structured for maintainability and future enhancement.
```

## Testing Framework Prompt

```
Create a comprehensive testing framework for the Go implementation of the IaC Manifest Generator.

Implement:
1. Unit tests for:
   - NLP parsing functions using table-driven tests
   - Infrastructure model validation
   - Template rendering accuracy

2. Integration tests for:
   - End-to-end pipeline functionality
   - CLI command execution
   - File generation and structure

3. Test fixtures including:
   - Sample infrastructure descriptions
   - Expected parsed models
   - Reference output files for comparison

4. Test utilities for:
   - Comparing generated files against expected output
   - Creating temporary test directories
   - Mocking file system operations for isolation

Use Go's standard testing package and testify for assertions where helpful.
```

## Example Implementation Prompt

```
Create a complete working example of the Go IaC Manifest Generator CLI processing a sample infrastructure request.

Implement:
1. A main.go file that:
   - Initializes the CLI
   - Sets up the processing pipeline
   - Handles command-line arguments

2. A sample infrastructure description:
   "Deploy AWS infrastructure in us-east-1 with a VPC, 3 public and 3 private subnets across 3 AZs, an internet gateway, 3 NAT gateways (one per AZ), and an EKS cluster with both public and private API access deployed in the private subnets. Include a node pool using t3-medium instances."

3. Processing logic that:
   - Parses the description using the NLP package
   - Creates the infrastructure model
   - Generates both Terraform and Crossplane manifests
   - Organizes output in appropriate directory structures

Include debug output showing the intermediate steps and final results.
```

## Documentation Generator Prompt

```
Create comprehensive documentation for the IaC Manifest Generator CLI in Golang.

Generate:
1. README.md with:
   - Project overview and purpose
   - Installation instructions
   - Basic usage examples
   - Supported infrastructure patterns

2. A user guide explaining:
   - CLI commands and options in detail
   - Format for describing infrastructure requirements
   - Limitations and constraints
   - Output directory structure for each IaC tool

3. Developer documentation covering:
   - Project architecture and design decisions
   - Package organization and responsibilities
   - How to extend the tool with new capabilities
   - Contributing guidelines

Format all documentation as markdown files suitable for GitHub hosting.
```

These prompts are designed to be used with Claude Code to implement a Go-based version of the IaC Manifest Generator CLI. Each prompt focuses on a specific component while maintaining alignment with the overall architecture and MVP plan, allowing for iterative development of a complete solution.

**Total cost - $14.00 to generate the code by running the prompts above**
**Total cost for validating the tests work and regenerate the documents - $20.52**

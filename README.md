# IaC Manifest Generator CLI

A Go CLI tool that parses English descriptions of AWS infrastructure and generates Terraform and Crossplane manifests.

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
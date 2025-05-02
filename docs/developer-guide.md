# IaC Manifest Generator CLI Developer Guide

This guide provides detailed information for developers who want to understand, modify, or extend the IaC Manifest Generator CLI tool. It covers the architecture, key components, extension points, and contributing guidelines.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Core Components](#core-components)
  - [Command-Line Interface](#command-line-interface)
  - [Natural Language Processing](#natural-language-processing)
  - [Infrastructure Modeling](#infrastructure-modeling)
  - [Pipeline System](#pipeline-system)
  - [Template System](#template-system)
  - [Adapters](#adapters)
- [Project Structure](#project-structure)
- [Development Environment Setup](#development-environment-setup)
- [Key Interfaces](#key-interfaces)
- [Extension Points](#extension-points)
  - [Adding New Resource Types](#adding-new-resource-types)
  - [Supporting New IaC Tools](#supporting-new-iac-tools)
  - [Enhancing NLP Capabilities](#enhancing-nlp-capabilities)
  - [Adding Template Functions](#adding-template-functions)
- [Code Patterns](#code-patterns)
- [Testing](#testing)
- [Contributing Guidelines](#contributing-guidelines)
- [Code of Conduct](#code-of-conduct)

## Architecture Overview

The IaC Manifest Generator CLI is designed with a modular pipeline architecture that processes natural language descriptions into infrastructure as code manifests. The key architectural decisions include:

1. **Pipeline-based Processing**: The tool uses a pipeline architecture to transform natural language inputs into structured IaC outputs through multiple processing stages.

2. **Adapter Pattern**: Support for different IaC tools (Terraform, Crossplane) is implemented using the adapter pattern, allowing easy extension to new formats.

3. **Templating System**: A flexible template system allows customization of output formats without changing the core logic.

4. **Resource Model**: A unified internal resource model represents infrastructure components independent of the target IaC format.

5. **Natural Language Processing**: A pattern-based NLP system extracts infrastructure requirements from plain English descriptions.

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   NLP Input   │    │  Infra Model  │    │ IaC Generator │    │   Output     │
│   Processing  │───▶│   Building   │───▶│   Adapters   │───▶│   Handling   │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
        │                  │                  │                  │
        ▼                  ▼                  ▼                  ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                            Pipeline Coordination                          │
└──────────────────────────────────────────────────────────────────────────┘
```

*Architecture Diagram: The flow of data through the pipeline, from NLP processing to output handling.*

## Core Components

### Command-Line Interface

The CLI is built using the Cobra library, which provides a powerful interface for building command-line applications. The CLI layer is responsible for:

- Parsing command-line arguments and flags
- Validating inputs
- Configuring logging
- Orchestrating the execution of the pipeline
- Presenting outputs to the user

Code location: `cmd/iacgen/`

### Natural Language Processing

The NLP system parses English descriptions into a structured format that can be used to build an infrastructure model. Key components include:

- Entity extraction to identify resources and their properties
- Pattern matching to recognize common infrastructure patterns
- Validation to ensure the extracted model is coherent and complete
- Entity relationship detection to understand connections between resources

Code location: `internal/nlp/`

### Infrastructure Modeling

The infrastructure model provides a unified representation of cloud resources that is independent of the target IaC format. It includes:

- Resource definitions with typed properties
- Relationship modeling between resources
- Default value management
- Resource dependency resolution

Code location: `internal/infra/` and `pkg/models/`

### Pipeline System

The pipeline system coordinates the transformation of inputs through multiple processing stages. Key features:

- Stage-based processing with clear interfaces
- Error handling and recovery
- Progress reporting
- Context-aware execution with cancellation support
- Parallel processing where appropriate

Code location: `internal/pipeline/`

### Template System

The template system enables customizable generation of IaC manifests using Go templates. Key components:

- Template loading and caching
- Custom template functions for resource formatting
- Format-specific template selection
- Template validation

Code location: `internal/template/`

### Adapters

The adapters translate the unified infrastructure model into specific IaC formats:

- **Terraform Adapter**: Generates HCL files for Terraform
- **Crossplane Adapter**: Generates YAML manifests for Crossplane

Each adapter implements the same interface, making it easy to add support for new IaC tools.

Code location: `internal/adapter/`

## Project Structure

```
iac_generator_cli/
├── cmd/
│   └── iacgen/                # CLI commands
│       ├── generate.go        # Generate command
│       └── root.go            # Root command and flags
├── internal/
│   ├── adapter/               # IaC format adapters
│   │   ├── crossplane/        # Crossplane adapter
│   │   │   ├── directory.go
│   │   │   ├── eks.go
│   │   │   ├── generator.go
│   │   │   ├── generator_template.go
│   │   │   ├── provider.go
│   │   │   ├── vpc.go
│   │   │   └── yaml.go
│   │   └── terraform/         # Terraform adapter
│   │       ├── directory.go
│   │       ├── generator.go
│   │       ├── generator_template.go
│   │       └── hcl.go
│   ├── config/                # Configuration handling
│   │   └── config.go
│   ├── generator/             # IaC generation orchestration
│   │   ├── generator.go
│   │   └── template_generator.go
│   ├── infra/                 # Infrastructure modeling
│   │   ├── aws.go             # AWS resource creation
│   │   ├── aws_models.go      # AWS-specific model types
│   │   ├── cidr.go            # CIDR utilities
│   │   └── model.go           # Infrastructure model definition
│   ├── nlp/                   # Natural language processing
│   │   ├── parser.go          # Main NLP parser
│   │   ├── patterns.go        # Pattern matching for entities
│   │   ├── prompt.go          # LLM prompting utilities
│   │   └── validator.go       # Entity validation
│   ├── pipeline/              # Processing pipeline
│   │   ├── cli.go             # CLI integration
│   │   ├── coordinator.go     # Pipeline coordination
│   │   ├── iac_generator.go   # IaC generation stage
│   │   ├── interfaces.go      # Pipeline interfaces
│   │   ├── model_builder.go   # Model building stage
│   │   ├── nlp_processor.go   # NLP processing stage
│   │   ├── output_handler.go  # Output handling stage
│   │   └── pipeline.go        # Base pipeline implementation
│   ├── template/              # Template system
│   │   ├── embed.go           # Template embedding
│   │   ├── functions.go       # Template functions
│   │   ├── template.go        # Template management
│   │   ├── templates/         # Template files
│   │   │   ├── crossplane/    # Crossplane templates
│   │   │   └── terraform/     # Terraform templates
│   │   └── validator.go       # Template validation
│   └── utils/                 # Utility functions
│       ├── file.go            # File operations
│       └── logger.go          # Logging utilities
├── pkg/                       # Public packages
│   └── models/                # Shared model definitions
│       └── resource.go        # Resource models
├── main.go                    # Application entry point
└── README.md                  # Project documentation
```

## Development Environment Setup

To set up a development environment for the IaC Manifest Generator CLI:

1. **Prerequisites**:
   - Go 1.19 or later
   - Git
   - Text editor or IDE (VS Code, GoLand, etc.)

2. **Clone the repository**:
   ```bash
   git clone https://github.com/riptano/iac_generator_cli.git
   cd iac_generator_cli
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Build the application**:
   ```bash
   go build -o iacgen
   ```

5. **Run tests**:
   ```bash
   go test ./...
   ```

## Key Interfaces

The project defines several key interfaces that enable extensibility and modularity:

### Generator Interface

```go
// Generator in internal/generator/generator.go
type Generator interface {
    Generate(model *models.InfrastructureModel) (string, error)
}
```

This interface is implemented by adapters for different IaC tools.

### Pipeline Interface

```go
// Pipeline in internal/pipeline/interfaces.go
type Pipeline interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    AddStage(stage Stage)
    SetErrorHandler(handler func(error) error)
    SetProgressReporter(reporter ProgressReporter)
}
```

This interface defines the contract for pipeline components.

### Stage Interface

```go
// Stage in internal/pipeline/interfaces.go
type Stage interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    Name() string
}
```

This interface defines individual pipeline stages.

### TemplateSelector Interface

```go
// TemplateSelector in internal/template/template.go
type TemplateSelector interface {
    SelectTemplate(format TemplateFormat, resource *models.Resource) (string, error)
    RegisterTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string)
}
```

This interface enables selection of the appropriate template for a resource.

## Extension Points

The IaC Manifest Generator CLI has several well-defined extension points for adding new functionality.

### Adding New Resource Types

To add support for a new AWS resource type:

1. **Define the resource type** in `pkg/models/resource.go`:
   ```go
   const (
       // Existing resource types
       // ...
       
       // New resource type
       ResourceNewType ResourceType = "new_type"
   )
   ```

2. **Add creation function** in `internal/infra/aws.go`:
   ```go
   // CreateNewTypeResource creates a new resource of the new type
   func CreateNewTypeResource(name string, property1 string, property2 int) models.Resource {
       resource := models.NewResource(models.ResourceNewType, name)
       resource.AddProperty("property1", property1)
       resource.AddProperty("property2", property2)
       return resource
   }
   ```

3. **Update the NLP parser** in `internal/nlp/patterns.go` to recognize the new resource type:
   ```go
   func ExtractNewTypeResource(description string) map[string]interface{} {
       // Pattern matching logic to extract properties
       // ...
       
       return properties
   }
   ```

4. **Add templates** for the new resource type:
   - Add `new_type.tmpl` to `internal/template/templates/terraform/`
   - Add `new_type.tmpl` to `internal/template/templates/crossplane/`

5. **Register the templates** in the appropriate generator:
   ```go
   selector.RegisterTemplate(FormatTerraform, models.ResourceNewType, "new_type.tmpl")
   selector.RegisterTemplate(FormatCrossplane, models.ResourceNewType, "new_type.tmpl")
   ```

### Supporting New IaC Tools

To add support for a new IaC tool:

1. **Create a new adapter package** in `internal/adapter/newtool/`:
   ```go
   package newtool

   import (
       "github.com/riptano/iac_generator_cli/pkg/models"
   )

   // NewToolGenerator generates NewTool manifests
   type NewToolGenerator struct {
       // Tool-specific fields
   }

   // Generate implements generator.Generator
   func (g *NewToolGenerator) Generate(model *models.InfrastructureModel) (string, error) {
       // Implementation
   }
   ```

2. **Add templates** for the new tool in `internal/template/templates/newtool/`

3. **Update the generator factory** in `internal/generator/generator.go` to support the new format:
   ```go
   func CreateGenerator(format string, useTemplates bool) (Generator, error) {
       switch format {
       // Existing formats
       // ...
       
       case "newtool":
           if useTemplates {
               return newtool.NewTemplateGenerator(), nil
           }
           return newtool.NewGenerator(), nil
       }
   }
   ```

4. **Update the CLI** to recognize the new format in `cmd/iacgen/root.go`:
   ```go
   func isValidOutputFormat(format string) bool {
       validFormats := []string{"terraform", "crossplane", "newtool"}
       // Implementation
   }
   ```

### Enhancing NLP Capabilities

To enhance the NLP capabilities:

1. **Add new entity extractors** in `internal/nlp/patterns.go`
2. **Update the validator** in `internal/nlp/validator.go` to validate new entities
3. **Update the parser** in `internal/nlp/parser.go` to incorporate the new extractors

### Adding Template Functions

To add new template functions:

1. **Define the function** in `internal/template/functions.go`:
   ```go
   // NewTemplateFunction does something useful
   func NewTemplateFunction(args ...interface{}) (interface{}, error) {
       // Implementation
   }
   ```

2. **Register the function** in `internal/template/template.go`:
   ```go
   func createTemplateFuncMap() template.FuncMap {
       return template.FuncMap{
           // Existing functions
           // ...
           
           // New function
           "newFunction": NewTemplateFunction,
       }
   }
   ```

## Code Patterns

### Error Handling

Errors should be wrapped with context using `fmt.Errorf` with `%w` formatting verb:

```go
if err != nil {
    return fmt.Errorf("failed to process resources: %w", err)
}
```

### Logging

The project uses zap for structured logging. Get a logger instance and use it:

```go
logger := utils.GetLogger()
logger.Infow("Processing description",
    "length", len(description),
    "format", format,
)
```

### Context Handling

All long-running operations should accept a context for cancellation:

```go
func (s *MyStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Check for cancellation
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    
    // Implementation
}
```

## Testing

The project uses Go's standard testing framework. Tests are located alongside the code they test.

### Unit Tests

Unit tests test individual functions and methods in isolation:

```go
func TestCreateVPC(t *testing.T) {
    vpc := CreateVPC("test-vpc", "10.0.0.0/16", true, true)
    
    if vpc.Type != models.ResourceVPC {
        t.Errorf("Expected VPC type, got %s", vpc.Type)
    }
    
    // More assertions
}
```

### Integration Tests

Integration tests test interactions between components:

```go
func TestPipeline(t *testing.T) {
    // Set up test dependencies
    
    // Run the pipeline with a test input
    result, err := pipeline.Execute(context.Background(), input)
    
    // Assert on the result
}
```

### Mocking

For testing components with external dependencies, use mocks:

```go
type MockGenerator struct {
    GenerateFn func(model *models.InfrastructureModel) (string, error)
}

func (m *MockGenerator) Generate(model *models.InfrastructureModel) (string, error) {
    return m.GenerateFn(model)
}
```

## Contributing Guidelines

### Code Style

- Follow Go's standard formatting and style guidelines
- Use `gofmt` or `goimports` to format code
- Write clear, self-descriptive variable and function names
- Document all exported functions, types, and constants

### Pull Request Process

1. Fork the repository and create your branch from `main`
2. Update the documentation as needed
3. Ensure the test suite passes
4. Add tests for new functionality
5. Submit a pull request with a clear description of the changes

### Commit Message Format

Follow this format for commit messages:

```
Type(scope): Brief description

Longer description if needed, explaining the context or reasoning behind the changes.

Fixes #123
```

Where `Type` is one of:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Development Workflow

1. **Choose an issue** to work on from the issue tracker
2. **Create a branch** with a descriptive name (`feature/add-rds-support`, `fix/vpc-cidr-validation`)
3. **Implement your changes** with tests and documentation
4. **Run the test suite** to ensure all tests pass
5. **Submit a pull request** for review
6. **Address any feedback** from the code review
7. **Update your PR** as needed
8. Once approved, your PR will be merged to the main branch

### Setting Up a Development Environment

1. **Fork and clone** the repository
2. **Set up upstream remote**:
   ```bash
   git remote add upstream https://github.com/riptano/iac_generator_cli.git
   ```
3. **Create a branch** for your work:
   ```bash
   git checkout -b feature/my-feature
   ```
4. **Make your changes** and commit them
5. **Push to your fork**:
   ```bash
   git push origin feature/my-feature
   ```
6. **Create a pull request** from your branch to the upstream main branch

Thank you for contributing to the IaC Manifest Generator CLI!

## Code of Conduct

The IaC Manifest Generator CLI project follows a Code of Conduct to ensure a welcoming and inclusive environment for all contributors. We expect all participants in our project and community to:

1. **Be respectful and inclusive** of differing viewpoints and experiences
2. **Use welcoming and inclusive language**
3. **Accept constructive criticism gracefully**
4. **Focus on what is best for the community**
5. **Show empathy towards other community members**

Unacceptable behavior includes:

- The use of sexualized language or imagery
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

### Reporting

If you experience or witness unacceptable behavior, please report it by contacting the project maintainers. All complaints will be reviewed and investigated promptly and fairly.

For more details, please refer to the full [Code of Conduct](../CONTRIBUTING.md#code-of-conduct) in the CONTRIBUTING.md file.
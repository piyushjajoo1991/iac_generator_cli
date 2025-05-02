# Architecture Overview

This document provides an overview of the IaC Manifest Generator CLI's architecture, explaining the key components and how they interact.

## System Architecture

The IaC Manifest Generator CLI uses a pipeline-based architecture to transform natural language descriptions into Infrastructure as Code (IaC) manifests. The system is designed to be modular, extensible, and maintainable.

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

## Core Components

### Command-Line Interface (CLI)

The CLI component, built using the Cobra library, handles user interaction, argument parsing, and command execution. It provides a user-friendly interface for the tool and coordinates the execution of the pipeline.

**Key Files:**
- `cmd/iacgen/root.go`: Root command and global flags
- `cmd/iacgen/generate.go`: Generate command implementation

### Natural Language Processing (NLP)

The NLP component processes natural language descriptions of infrastructure and extracts structured information about resources, their properties, and relationships.

**Key Files:**
- `internal/nlp/parser.go`: Main NLP parser implementation
- `internal/nlp/patterns.go`: Pattern matching for entity extraction
- `internal/nlp/validator.go`: Validation of extracted entities

### Infrastructure Model

The infrastructure model component defines a unified representation of cloud resources, independent of the target IaC format. It serves as the intermediate representation between NLP processing and IaC generation.

**Key Files:**
- `pkg/models/resource.go`: Resource model definitions
- `internal/infra/model.go`: Infrastructure model implementation
- `internal/infra/aws.go`: AWS resource creation utilities

### Pipeline System

The pipeline system coordinates the transformation of inputs through multiple processing stages. It provides a framework for defining and executing the processing pipeline.

**Key Files:**
- `internal/pipeline/pipeline.go`: Base pipeline implementation
- `internal/pipeline/interfaces.go`: Pipeline interfaces
- `internal/pipeline/coordinator.go`: Pipeline coordination

### Template System

The template system enables customizable generation of IaC manifests using Go templates. It provides a mechanism for defining and rendering templates for different resource types and IaC formats.

**Key Files:**
- `internal/template/template.go`: Template management
- `internal/template/functions.go`: Template helper functions
- `internal/template/templates/`: Template directory

### Adapters

The adapters translate the unified infrastructure model into specific IaC formats. They implement a common interface, making it easy to add support for new IaC tools.

**Key Directories:**
- `internal/adapter/terraform/`: Terraform adapter
- `internal/adapter/crossplane/`: Crossplane adapter

## Data Flow

1. **User Input**: The user provides a natural language description via the CLI.
2. **NLP Processing**: The NLP component extracts resource information from the description.
3. **Model Building**: The extracted information is used to build an infrastructure model.
4. **IaC Generation**: The appropriate adapter generates IaC manifests from the model.
5. **Output Handling**: The generated manifests are written to the specified output location.

## Extension Points

The architecture is designed to be extensible in several ways:

1. **New Resource Types**: Add support for new AWS resource types by updating the resource models and adding extraction patterns.
2. **New IaC Formats**: Add support for new IaC tools by implementing a new adapter.
3. **Enhanced NLP**: Improve the NLP capabilities by enhancing the pattern matching and entity extraction.
4. **Custom Templates**: Customize the output by modifying or adding templates.

## Design Decisions

### Pipeline Architecture

A pipeline architecture was chosen to provide a clear separation of concerns and to make the system modular and extensible. Each stage in the pipeline has a specific responsibility, making it easier to understand, modify, and test.

### Adapter Pattern

The adapter pattern was used to support multiple IaC formats without requiring changes to the core system. This allows for easy addition of new IaC tools.

### Template System

A template system was used to separate the logic of resource representation from the generation logic. This allows for customization of the output format without changing the core code.

### Resource Modeling

A unified resource model was created to represent infrastructure components independently of the target IaC format. This allows for a single parsing and validation step, regardless of the output format.

## Conclusion

The IaC Manifest Generator CLI's architecture is designed to be modular, extensible, and maintainable. It uses a pipeline architecture to transform natural language descriptions into IaC manifests, with clear extension points for adding new functionality.
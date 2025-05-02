# Contributing to IaC Manifest Generator CLI

Thank you for your interest in contributing to the IaC Manifest Generator CLI! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Submitting Changes](#submitting-changes)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Issue Reporting](#issue-reporting)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork to your local machine
3. Set up your development environment

## Development Environment

### Prerequisites

- Go 1.19 or later
- Git
- Text editor or IDE (VS Code, GoLand, etc.)

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/iac_generator_cli.git
cd iac_generator_cli

# Add the upstream remote
git remote add upstream https://github.com/riptano/iac_generator_cli.git

# Install dependencies
go mod download

# Build the application
go build -o iacgen
```

## Submitting Changes

1. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the [coding standards](#coding-standards)

3. Write tests for your changes

4. Update documentation as needed

5. Commit your changes with clear, descriptive commit messages:
   ```bash
   git commit -m "Add feature: your feature description"
   ```

6. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

7. Create a pull request from your fork to the main repository

## Pull Request Process

1. Ensure your code follows the [coding standards](#coding-standards)
2. Ensure your code passes all tests
3. Update documentation as needed
4. Fill in the pull request template with all relevant information
5. Request a review from one of the maintainers
6. Address any feedback from reviewers
7. Once approved, your pull request will be merged by a maintainer

## Coding Standards

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` or `goimports` to format your code
- Document all exported functions, types, and constants
- Write clear, self-descriptive variable and function names
- Keep functions small and focused
- Use meaningful error messages
- Add comments for complex logic

## Testing

- Write tests for all new functionality
- Ensure existing tests pass with your changes
- Run tests using `go test ./...`
- Consider adding integration tests for complex features

## Documentation

- Update documentation to reflect your changes
- Document new features in the appropriate place
- Ensure code examples are up-to-date
- Use clear, concise language in your documentation

## Issue Reporting

When reporting issues, please include:

- A clear, descriptive title
- A detailed description of the issue
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Environment information (OS, Go version, etc.)
- Any additional context or screenshots

## Feature Requests

When requesting a feature, please include:

- A clear, descriptive title
- A detailed description of the feature
- The motivation for the feature
- Any examples of how the feature would be used
- Any additional context or mockups

Thank you for contributing to the IaC Manifest Generator CLI!
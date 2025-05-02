package nlp

import (
	"bytes"
	"text/template"
)

// PromptTemplate represents a template for generating prompts
type PromptTemplate struct {
	Template string
}

// NewPromptTemplate creates a new prompt template
func NewPromptTemplate(templateString string) *PromptTemplate {
	return &PromptTemplate{
		Template: templateString,
	}
}

// GeneratePrompt generates a prompt from the template with the given data
func (p *PromptTemplate) GeneratePrompt(data map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(p.Template)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// DefaultPromptTemplate returns the default prompt template for infrastructure descriptions
func DefaultPromptTemplate() *PromptTemplate {
	template := `
Extract AWS infrastructure resources from the following description:

Description: {{.Description}}

Extract information about the following resource types if mentioned:
- EC2 Instances (size, region, etc.)
- S3 Buckets (name, access control, versioning, etc.)
- VPCs (CIDR blocks, DNS settings, etc.)
- Security Groups (rules, ports, etc.)
- Subnets (CIDR blocks, availability zones, etc.)
- RDS Instances (engine, size, etc.)
- IAM Roles and Policies
- Lambda Functions
- DynamoDB Tables
- CloudWatch Alarms and Metrics

Output the extracted information in a structured format.
`
	return NewPromptTemplate(template)
}

// EnhanceDescription adds context to the user's description to improve NLP parsing
func EnhanceDescription(description string) string {
	// This function would add context or restructure the description to improve NLP accuracy
	// For now, it just returns the original description
	return description
}
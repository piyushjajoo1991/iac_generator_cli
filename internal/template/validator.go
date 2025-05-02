package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	hclpos "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"gopkg.in/yaml.v3"
)

// ValidationLevel controls the strictness of validation
type ValidationLevel string

const (
	// ValidationLevelNone performs no validation
	ValidationLevelNone ValidationLevel = "none"
	// ValidationLevelBasic performs basic syntax validation
	ValidationLevelBasic ValidationLevel = "basic"
	// ValidationLevelStrict performs strict validation including tool-specific checks
	ValidationLevelStrict ValidationLevel = "strict"
)

// ValidationOptions controls validation behavior
type ValidationOptions struct {
	Level   ValidationLevel
	TempDir string // Directory for temporary files during validation
}

// DefaultValidationOptions returns default validation options
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		Level:   ValidationLevelBasic,
		TempDir: os.TempDir(),
	}
}

// Validator validates rendered template content
type Validator interface {
	// Validate checks if the content is valid
	Validate(content string, options ValidationOptions) error
}

// HCLValidator validates Terraform HCL syntax
type HCLValidator struct{}

// Validate checks if the HCL content is valid
func (v *HCLValidator) Validate(content string, options ValidationOptions) error {
	if options.Level == ValidationLevelNone {
		return nil
	}

	// Basic syntax validation
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(content), "generated.tf")
	if diags.HasErrors() {
		return fmt.Errorf("invalid HCL syntax: %s", diags.Error())
	}

	// Format the HCL for better output
	formatted := formatHCL(content)
	if formatted != "" {
		content = formatted
	}

	// Strict validation with terraform validate
	if options.Level == ValidationLevelStrict {
		return v.validateWithTerraform(content, options.TempDir)
	}

	return nil
}

// formatHCL formats HCL content using the hclwrite package
func formatHCL(content string) string {
	file, err := hclwrite.ParseConfig([]byte(content), "generated.tf", hclpos.Pos{Line: 1, Column: 1})
	if err != nil {
		return ""
	}
	return string(file.Bytes())
}

// validateWithTerraform validates HCL with the terraform validate command
func (v *HCLValidator) validateWithTerraform(content string, tempDir string) error {
	// Create a temporary directory for validation
	dir, err := ioutil.TempDir(tempDir, "terraform-validate-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write content to temporary file
	tfFile := filepath.Join(dir, "main.tf")
	if err := ioutil.WriteFile(tfFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Check if terraform is available
	_, err = exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform not found in PATH, skipping strict validation")
	}

	// Initialize terraform
	initCmd := exec.Command("terraform", "init", "-no-color")
	initCmd.Dir = dir
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %s", string(initOutput))
	}

	// Validate terraform
	validateCmd := exec.Command("terraform", "validate", "-no-color")
	validateCmd.Dir = dir
	validateOutput, err := validateCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform validate failed: %s", string(validateOutput))
	}

	return nil
}

// YAMLValidator validates YAML syntax
type YAMLValidator struct{}

// Validate checks if the YAML content is valid
func (v *YAMLValidator) Validate(content string, options ValidationOptions) error {
	if options.Level == ValidationLevelNone {
		return nil
	}

	// Basic syntax validation
	var data interface{}
	err := yaml.Unmarshal([]byte(content), &data)
	if err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Check for required fields in Crossplane resources
	if options.Level == ValidationLevelStrict {
		return v.validateCrossplaneYAML(data)
	}

	return nil
}

// validateCrossplaneYAML validates Crossplane YAML content
func (v *YAMLValidator) validateCrossplaneYAML(data interface{}) error {
	// This is a placeholder for more comprehensive Crossplane validation
	// In a real implementation, you would check for required fields and structure

	// Check for required fields in each document
	docs, ok := data.([]interface{})
	if !ok {
		// Single document
		docs = []interface{}{data}
	}

	for _, doc := range docs {
		if doc == nil {
			continue
		}

		resourceMap, ok := doc.(map[string]interface{})
		if !ok {
			continue
		}

		// Check required fields for Crossplane resources
		if apiVersion, ok := resourceMap["apiVersion"].(string); ok {
			if !strings.Contains(apiVersion, ".crossplane.io") {
				continue // Not a Crossplane resource
			}

			// Check for required fields
			if _, ok := resourceMap["kind"]; !ok {
				return fmt.Errorf("missing 'kind' field in Crossplane resource")
			}

			if _, ok := resourceMap["metadata"]; !ok {
				return fmt.Errorf("missing 'metadata' field in Crossplane resource")
			}

			if _, ok := resourceMap["spec"]; !ok {
				return fmt.Errorf("missing 'spec' field in Crossplane resource")
			}
		}
	}

	return nil
}

// GetValidator returns the appropriate validator for the given format
func GetValidator(format TemplateFormat) Validator {
	switch format {
	case FormatTerraform:
		return &HCLValidator{}
	case FormatCrossplane:
		return &YAMLValidator{}
	default:
		return nil
	}
}

// ValidateRenderedContent validates the rendered content for the given format
func ValidateRenderedContent(format TemplateFormat, content string) error {
	validator := GetValidator(format)
	if validator == nil {
		return fmt.Errorf("no validator available for format %s", format)
	}
	
	return validator.Validate(content, DefaultValidationOptions())
}

// ValidateRenderedContentWithOptions validates the rendered content with specific options
func ValidateRenderedContentWithOptions(format TemplateFormat, content string, options ValidationOptions) error {
	validator := GetValidator(format)
	if validator == nil {
		return fmt.Errorf("no validator available for format %s", format)
	}
	
	return validator.Validate(content, options)
}

// FormatRenderedContent formats the rendered content according to conventions
func FormatRenderedContent(format TemplateFormat, content string) string {
	switch format {
	case FormatTerraform:
		return FormatHCLDocument(content)
	case FormatCrossplane:
		return FormatYAMLDocument(content)
	default:
		return content
	}
}

// PrettyPrintJSON formats JSON for human readability
func PrettyPrintJSON(data interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// ValidateResourceTemplate validates that a template works for a resource
func ValidateResourceTemplate(manager *TemplateManager, format TemplateFormat, templateName string, resource *models.Resource) error {
	// Get the template
	tmpl, err := manager.GetTemplate(format, templateName)
	if err != nil {
		return fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Resource": resource,
	}

	// Try to render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	// Basic syntax validation
	return ValidateRenderedContent(format, buf.String())
}

// AnalyzeTemplate analyzes a template for variables and resource types
func AnalyzeTemplate(templateContent string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"variables": []string{},
		"resourceTypes": []string{},
		"requiredProperties": []string{},
		"conditionals": []string{},
	}

	// Find all variables {{ .Variable }}
	varRegex := regexp.MustCompile(`{{\s*\.([a-zA-Z0-9._]+)\s*}}`)
	varMatches := varRegex.FindAllStringSubmatch(templateContent, -1)
	vars := make(map[string]bool)
	for _, match := range varMatches {
		if len(match) > 1 {
			vars[match[1]] = true
		}
	}

	// Extract unique variables
	varList := make([]string, 0, len(vars))
	for v := range vars {
		varList = append(varList, v)
	}
	result["variables"] = varList

	// Find conditional blocks {{- if ... }}
	condRegex := regexp.MustCompile(`{{-?\s*if\s+(.+?)\s*}}`)
	condMatches := condRegex.FindAllStringSubmatch(templateContent, -1)
	conds := make(map[string]bool)
	for _, match := range condMatches {
		if len(match) > 1 {
			conds[match[1]] = true
		}
	}

	// Extract unique conditionals
	condList := make([]string, 0, len(conds))
	for c := range conds {
		condList = append(condList, c)
	}
	result["conditionals"] = condList

	// Find required property checks
	propRegex := regexp.MustCompile(`hasProperty\s+\.Resource\s+"([^"]+)"`)
	propMatches := propRegex.FindAllStringSubmatch(templateContent, -1)
	props := make(map[string]bool)
	for _, match := range propMatches {
		if len(match) > 1 {
			props[match[1]] = true
		}
	}

	// Extract required properties
	propList := make([]string, 0, len(props))
	for p := range props {
		propList = append(propList, p)
	}
	result["requiredProperties"] = propList

	return result, nil
}
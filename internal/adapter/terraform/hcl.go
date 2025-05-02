package terraform

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/riptano/iac_generator_cli/pkg/models"
)

// HCLBlock represents a block in HCL
type HCLBlock struct {
	Type       string
	Labels     []string
	Attributes map[string]interface{}
	Blocks     []*HCLBlock
}

// NewHCLBlock creates a new HCL block
func NewHCLBlock(blockType string, labels ...string) *HCLBlock {
	return &HCLBlock{
		Type:       blockType,
		Labels:     labels,
		Attributes: make(map[string]interface{}),
		Blocks:     []*HCLBlock{},
	}
}

// AddAttribute adds an attribute to the block
func (b *HCLBlock) AddAttribute(name string, value interface{}) {
	b.Attributes[name] = value
}

// AddBlock adds a nested block
func (b *HCLBlock) AddBlock(block *HCLBlock) {
	b.Blocks = append(b.Blocks, block)
}

// HCLWriter handles writing HCL content
type HCLWriter struct {
	buffer bytes.Buffer
	indent int
}

// NewHCLWriter creates a new HCL writer
func NewHCLWriter() *HCLWriter {
	return &HCLWriter{
		buffer: bytes.Buffer{},
		indent: 0,
	}
}

// String returns the HCL as a string
func (w *HCLWriter) String() string {
	return w.buffer.String()
}

// writeIndent writes the current indentation
func (w *HCLWriter) writeIndent() {
	for i := 0; i < w.indent; i++ {
		w.buffer.WriteString("  ")
	}
}

// WriteBlock writes an HCL block
func (w *HCLWriter) WriteBlock(block *HCLBlock) {
	// Write block type
	w.writeIndent()
	w.buffer.WriteString(block.Type)
	
	// Write labels if any
	for _, label := range block.Labels {
		w.buffer.WriteString(fmt.Sprintf(" %q", label))
	}
	
	w.buffer.WriteString(" {\n")
	w.indent++
	
	// Write attributes in sorted order for consistency
	var attrNames []string
	for name := range block.Attributes {
		attrNames = append(attrNames, name)
	}
	sort.Strings(attrNames)
	
	for _, name := range attrNames {
		w.writeIndent()
		w.buffer.WriteString(fmt.Sprintf("%s = %s\n", name, formatHCLValue(block.Attributes[name])))
	}
	
	// Add a blank line between attributes and blocks if both exist
	if len(block.Attributes) > 0 && len(block.Blocks) > 0 {
		w.buffer.WriteString("\n")
	}
	
	// Write nested blocks
	for i, nestedBlock := range block.Blocks {
		w.WriteBlock(nestedBlock)
		
		// Add a blank line between blocks
		if i < len(block.Blocks)-1 {
			w.buffer.WriteString("\n")
		}
	}
	
	w.indent--
	w.writeIndent()
	w.buffer.WriteString("}\n")
}

// formatHCLValue formats a value for HCL
func formatHCLValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Check if it's a reference (${...}) or heredoc
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			return v
		}
		if strings.HasPrefix(v, "<<") {
			return v
		}
		return fmt.Sprintf("%q", v)
	case []string:
		if len(v) == 0 {
			return "[]"
		}
		values := make([]string, len(v))
		for i, s := range v {
			values[i] = fmt.Sprintf("%q", s)
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", "))
	case map[string]string:
		if len(v) == 0 {
			return "{}"
		}
		var pairs []string
		for k, val := range v {
			pairs = append(pairs, fmt.Sprintf("%q = %q", k, val))
		}
		sort.Strings(pairs)
		return fmt.Sprintf("{\n    %s\n  }", strings.Join(pairs, "\n    "))
	case bool:
		return fmt.Sprintf("%t", v)
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ResourceToHCL converts a resource model to an HCL block
func ResourceToHCL(resource models.Resource) (*HCLBlock, error) {
	// Map resource type to Terraform resource type
	terraformType, err := mapResourceType(resource.Type)
	if err != nil {
		return nil, err
	}

	// Create block
	block := NewHCLBlock("resource", terraformType, resource.Name)

	// Add properties
	for _, prop := range resource.Properties {
		block.AddAttribute(prop.Name, prop.Value)
	}

	// Add dependencies if present
	if len(resource.DependsOn) > 0 {
		block.AddAttribute("depends_on", resource.DependsOn)
	}

	return block, nil
}

// OutputToHCL creates an HCL output block
func OutputToHCL(name string, value interface{}, description string) *HCLBlock {
	block := NewHCLBlock("output", name)
	
	if description != "" {
		block.AddAttribute("description", description)
	}
	
	block.AddAttribute("value", value)
	
	return block
}

// VariableToHCL creates an HCL variable block
func VariableToHCL(name string, varType string, defaultValue interface{}, description string) *HCLBlock {
	block := NewHCLBlock("variable", name)
	
	if description != "" {
		block.AddAttribute("description", description)
	}
	
	if varType != "" {
		block.AddAttribute("type", varType)
	}
	
	if defaultValue != nil {
		block.AddAttribute("default", defaultValue)
	}
	
	return block
}

// ModuleToHCL creates an HCL module block
func ModuleToHCL(name string, source string, variables map[string]interface{}) *HCLBlock {
	block := NewHCLBlock("module", name)
	
	block.AddAttribute("source", source)
	
	// Add variables
	for k, v := range variables {
		block.AddAttribute(k, v)
	}
	
	return block
}

// LocalToHCL creates an HCL locals block
func LocalToHCL(locals map[string]interface{}) *HCLBlock {
	block := NewHCLBlock("locals")
	
	// Add locals
	for k, v := range locals {
		block.AddAttribute(k, v)
	}
	
	return block
}

// ProviderToHCL creates an HCL provider block
func ProviderToHCL(name string, attributes map[string]interface{}) *HCLBlock {
	block := NewHCLBlock("provider", name)
	
	// Add attributes
	for k, v := range attributes {
		block.AddAttribute(k, v)
	}
	
	return block
}

// DataSourceToHCL creates an HCL data block
func DataSourceToHCL(dataType string, name string, attributes map[string]interface{}) *HCLBlock {
	block := NewHCLBlock("data", dataType, name)
	
	// Add attributes
	for k, v := range attributes {
		block.AddAttribute(k, v)
	}
	
	return block
}

// TerraformToHCL creates a terraform configuration block
func TerraformToHCL(requiredVersion string, providers map[string]map[string]string, backend map[string]map[string]string) *HCLBlock {
	block := NewHCLBlock("terraform")
	
	if requiredVersion != "" {
		block.AddAttribute("required_version", requiredVersion)
	}
	
	// Add required providers
	if len(providers) > 0 {
		providersBlock := NewHCLBlock("required_providers")
		for name, config := range providers {
			providerBlock := make(map[string]string)
			for k, v := range config {
				providerBlock[k] = v
			}
			providersBlock.AddAttribute(name, providerBlock)
		}
		block.AddBlock(providersBlock)
	}
	
	// Add backend
	for backendType, config := range backend {
		backendBlock := NewHCLBlock("backend", backendType)
		for k, v := range config {
			backendBlock.AddAttribute(k, v)
		}
		block.AddBlock(backendBlock)
	}
	
	return block
}

// TemplateToHCL renders a template string with the given data and returns HCL
func TemplateToHCL(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("hcl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}
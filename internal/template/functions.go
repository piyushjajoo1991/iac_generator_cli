package template

import (
	"fmt"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"strings"
	"unicode"
)

// TitleFunc returns s with the first letter of each word capitalized
func TitleFunc(s string) string {
	return strings.Title(s)
}

// LowerFunc returns s with all characters in lowercase
func LowerFunc(s string) string {
	return strings.ToLower(s)
}

// UpperFunc returns s with all characters in uppercase
func UpperFunc(s string) string {
	return strings.ToUpper(s)
}

// CamelCaseFunc converts a string to camel case
func CamelCaseFunc(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Split the string by non-alphanumeric characters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return s
	}

	// Convert the first word to lowercase
	result := strings.ToLower(words[0])

	// Convert the rest of the words to title case
	for _, word := range words[1:] {
		result += strings.Title(strings.ToLower(word))
	}

	return result
}

// SnakeCaseFunc converts a string to snake_case
func SnakeCaseFunc(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Split the string by non-alphanumeric characters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return s
	}

	// Convert all words to lowercase and join with underscore
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}

	return strings.Join(words, "_")
}

// KebabCaseFunc converts a string to kebab-case
func KebabCaseFunc(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Split the string by non-alphanumeric characters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return s
	}

	// Convert all words to lowercase and join with hyphen
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}

	return strings.Join(words, "-")
}

// QuoteFunc wraps a string in quotes
func QuoteFunc(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

// IndentFunc indents each line of s with prefix
func IndentFunc(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

// JoinFunc joins an array of strings with sep
func JoinFunc(arr []string, sep string) string {
	return strings.Join(arr, sep)
}

// MakeMapFunc creates a map from key-value pairs
func MakeMapFunc(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("map requires an even number of arguments")
	}
	
	result := make(map[string]interface{}, len(values)/2)
	
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("map keys must be strings")
		}
		result[key] = values[i+1]
	}
	
	return result, nil
}

// ContainsFunc checks if a string contains a substring
func ContainsFunc(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefixFunc checks if a string has a prefix
func HasPrefixFunc(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffixFunc checks if a string has a suffix
func HasSuffixFunc(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// ReplaceFunc replaces all occurrences of old with new in s
func ReplaceFunc(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// TrimFunc trims cutset from the beginning and end of s
func TrimFunc(s, cutset string) string {
	return strings.Trim(s, cutset)
}

// ToYAMLFunc formats a value as YAML
func ToYAMLFunc(v interface{}) string {
	switch val := v.(type) {
	case string:
		// Check if string needs quoting in YAML
		if strings.Contains(val, "\n") || strings.Contains(val, ":") ||
			strings.Contains(val, "{") || strings.Contains(val, "}") ||
			strings.Contains(val, "[") || strings.Contains(val, "]") ||
			strings.Contains(val, "#") || strings.HasPrefix(val, " ") {
			// Use literal block format for multiline strings
			if strings.Contains(val, "\n") {
				return fmt.Sprintf("|\n%s", IndentFunc(val, "  "))
			}
			// Quote the string for YAML
			return fmt.Sprintf("\"%s\"", escapeYAMLString(val))
		}
		return val
	case bool:
		return fmt.Sprintf("%v", val)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case []interface{}:
		lines := make([]string, len(val))
		for i, item := range val {
			yamlValue := ToYAMLFunc(item)
			if strings.Contains(yamlValue, "\n") {
				// Handle multiline values in arrays
				indented := IndentFunc(yamlValue, "  ")
				lines[i] = fmt.Sprintf("- %s", strings.TrimPrefix(indented, "  "))
			} else {
				lines[i] = fmt.Sprintf("- %s", yamlValue)
			}
		}
		return strings.Join(lines, "\n")
	case []string:
		lines := make([]string, len(val))
		for i, str := range val {
			if strings.Contains(str, "\n") || strings.Contains(str, ":") {
				// Quote multiline or complex strings
				lines[i] = fmt.Sprintf("- \"%s\"", escapeYAMLString(str))
			} else {
				lines[i] = fmt.Sprintf("- %s", str)
			}
		}
		return strings.Join(lines, "\n")
	case map[string]interface{}:
		lines := make([]string, 0, len(val))
		for k, v := range val {
			yamlValue := ToYAMLFunc(v)
			if strings.Contains(yamlValue, "\n") {
				// Handle multiline values
				indented := IndentFunc(yamlValue, "  ")
				lines = append(lines, fmt.Sprintf("%s:\n%s", k, indented))
			} else {
				lines = append(lines, fmt.Sprintf("%s: %s", k, yamlValue))
			}
		}
		return strings.Join(lines, "\n")
	case nil:
		return "null"
	default:
		// Try to convert structs and more complex types
		return fmt.Sprintf("%v", val)
	}
}

// escapeYAMLString escapes special characters in a YAML string
func escapeYAMLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// ToHCLFunc formats a value as HCL
func ToHCLFunc(v interface{}) string {
	switch val := v.(type) {
	case string:
		// Escape special characters in HCL string
		escaped := escapeHCLString(val)
		// Use heredoc format for multiline strings
		if strings.Contains(val, "\n") {
			lines := strings.Split(val, "\n")
			return fmt.Sprintf("<<-EOT\n%s\nEOT", strings.Join(lines, "\n"))
		}
		return fmt.Sprintf("\"%s\"", escaped)
	case bool:
		return fmt.Sprintf("%v", val)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	case []interface{}:
		items := make([]string, len(val))
		for i, item := range val {
			items[i] = ToHCLFunc(item)
		}
		return fmt.Sprintf("[\n  %s\n]", strings.Join(items, ",\n  "))
	case []string:
		if len(val) == 0 {
			return "[]"
		}
		items := make([]string, len(val))
		for i, str := range val {
			items[i] = fmt.Sprintf("\"%s\"", escapeHCLString(str))
		}
		if len(val) > 3 {
			return fmt.Sprintf("[\n  %s\n]", strings.Join(items, ",\n  "))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		if len(val) == 0 {
			return "{}"
		}
		lines := make([]string, 0, len(val))
		for k, v := range val {
			hclValue := ToHCLFunc(v)
			if strings.Contains(hclValue, "\n") {
				// Format multiline values with proper indentation
				indented := strings.ReplaceAll(hclValue, "\n", "\n  ")
				lines = append(lines, fmt.Sprintf("%s = %s", k, indented))
			} else {
				lines = append(lines, fmt.Sprintf("%s = %s", k, hclValue))
			}
		}
		return fmt.Sprintf("{\n  %s\n}", strings.Join(lines, "\n  "))
	case nil:
		return "null"
	default:
		// Try to handle more complex types
		return fmt.Sprintf("%v", val)
	}
}

// escapeHCLString escapes special characters in an HCL string
func escapeHCLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "${", "$${")
	return s
}

// FormatYAMLDocument formats a complete YAML document with proper document separators
func FormatYAMLDocument(content string) string {
	// Ensure document starts with a separator
	if !strings.HasPrefix(content, "---") {
		content = "---\n" + content
	}
	
	// Ensure consistent document separators
	content = strings.ReplaceAll(content, "\n\n---", "\n---")
	
	// Ensure document ends with newline
	if !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}
	
	return content
}

// FormatHCLDocument formats a complete HCL document with proper spacing
func FormatHCLDocument(content string) string {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	
	// Remove excessive empty lines
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	
	// Fix spacing around blocks
	content = strings.ReplaceAll(content, "{\n\n", "{\n")
	content = strings.ReplaceAll(content, "\n\n}", "\n}")
	
	// Add spacing between top-level blocks
	content = strings.ReplaceAll(content, "}\nresource", "}\n\nresource")
	content = strings.ReplaceAll(content, "}\nprovider", "}\n\nprovider")
	content = strings.ReplaceAll(content, "}\nmodule", "}\n\nmodule")
	content = strings.ReplaceAll(content, "}\nvariable", "}\n\nvariable")
	content = strings.ReplaceAll(content, "}\noutput", "}\n\noutput")
	content = strings.ReplaceAll(content, "}\nlocals", "}\n\nlocals")
	content = strings.ReplaceAll(content, "}\ndata", "}\n\ndata")
	
	return content
}

// DefaultValueFunc returns a default value if v is nil or empty
func DefaultValueFunc(v, defaultValue interface{}) interface{} {
	// Check if v is nil
	if v == nil {
		return defaultValue
	}
	
	// Check if v is an empty string
	if s, ok := v.(string); ok && s == "" {
		return defaultValue
	}
	
	// Check if v is an empty slice
	if s, ok := v.([]string); ok && len(s) == 0 {
		return defaultValue
	}
	
	// Check if v is an empty map
	if m, ok := v.(map[string]interface{}); ok && len(m) == 0 {
		return defaultValue
	}
	
	return v
}

// GetPropertyFunc retrieves a property from a resource
func GetPropertyFunc(resource *models.Resource, name string) interface{} {
	if resource == nil {
		return nil
	}
	
	for _, prop := range resource.Properties {
		if prop.Name == name {
			return prop.Value
		}
	}
	
	return nil
}

// HasPropertyFunc checks if a resource has a property
func HasPropertyFunc(resource *models.Resource, name string) bool {
	if resource == nil {
		return false
	}
	
	for _, prop := range resource.Properties {
		if prop.Name == name {
			return true
		}
	}
	
	return false
}

// ResourceRefFunc creates a reference to another resource
func ResourceRefFunc(resourceType models.ResourceType, name string, attribute string) string {
	// Format as Terraform reference by default
	return fmt.Sprintf("${%s.%s.%s}", resourceType, name, attribute)
}

// YAMLRefFunc creates a YAML-style reference to another resource
func YAMLRefFunc(apiVersion, kind, name, attribute string) string {
	return fmt.Sprintf("$(resources.%s.%s.%s.%s)", apiVersion, kind, name, attribute)
}

// CIDRSubnetFunc calculates a subnet CIDR from a base CIDR
func CIDRSubnetFunc(baseCIDR string, netnum int, newbits int) string {
	// This is just a placeholder that returns a formatted string
	// In a real implementation, you would calculate the actual subnet
	return fmt.Sprintf("${cidrsubnet(\"%s\", %d, %d)}", baseCIDR, newbits, netnum)
}

// MergeMapFunc merges multiple maps into a single map
func MergeMapFunc(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	
	return result
}

// TernaryFunc implements a ternary conditional operator
func TernaryFunc(condition bool, trueValue, falseValue interface{}) interface{} {
	if condition {
		return trueValue
	}
	return falseValue
}

// SliceFunc creates a slice from a list of items
func SliceFunc(items ...interface{}) []interface{} {
	return items
}

// FilterSliceFunc filters a slice based on a predicate function
func FilterSliceFunc(items []interface{}, predicate func(interface{}) bool) []interface{} {
	if items == nil {
		return nil
	}
	
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	
	return result
}

// MapSliceFunc maps a slice using a transform function
func MapSliceFunc(items []interface{}, transform func(interface{}) interface{}) []interface{} {
	if items == nil {
		return nil
	}
	
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}
	
	return result
}

// UniqueSliceFunc returns a slice with duplicate values removed
func UniqueSliceFunc(items []interface{}) []interface{} {
	if items == nil {
		return nil
	}
	
	seen := make(map[interface{}]bool)
	result := make([]interface{}, 0, len(items))
	
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// SplitFunc splits a string by a separator
func SplitFunc(s, sep string) []string {
	return strings.Split(s, sep)
}

// GetTagsFunc extracts tags from a resource's properties
func GetTagsFunc(resource *models.Resource) map[string]string {
	if resource == nil {
		return nil
	}
	
	tags := make(map[string]string)
	
	// Look for properties with "tag." prefix
	for _, prop := range resource.Properties {
		if strings.HasPrefix(prop.Name, "tag.") {
			tagName := strings.TrimPrefix(prop.Name, "tag.")
			if strValue, ok := prop.Value.(string); ok {
				tags[tagName] = strValue
			} else {
				tags[tagName] = fmt.Sprintf("%v", prop.Value)
			}
		}
	}
	
	// Add default Name tag if not present
	if _, ok := tags["Name"]; !ok && resource.Name != "" {
		tags["Name"] = resource.Name
	}
	
	return tags
}

// FormatTerraformTagsFunc formats tags as a Terraform tags block
func FormatTerraformTagsFunc(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	
	lines := make([]string, 0, len(tags))
	for k, v := range tags {
		lines = append(lines, fmt.Sprintf("    %s = \"%s\"", k, escapeHCLString(v)))
	}
	
	return fmt.Sprintf("  tags = {\n%s\n  }", strings.Join(lines, "\n"))
}

// FormatCrossplaneTagsFunc formats tags as a Crossplane tags block
func FormatCrossplaneTagsFunc(tags map[string]string) string {
	if len(tags) == 0 {
		return "    tags: []"
	}
	
	lines := make([]string, 0, len(tags))
	for k, v := range tags {
		lines = append(lines, fmt.Sprintf("    - key: \"%s\"\n      value: \"%s\"", 
			escapeYAMLString(k), escapeYAMLString(v)))
	}
	
	return fmt.Sprintf("    tags:\n%s", strings.Join(lines, "\n"))
}
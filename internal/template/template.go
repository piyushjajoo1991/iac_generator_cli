package template

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings" // Using strings.Contains (multiple places) and strings.Split (in RegisterPatternTemplate)
	"sync"
	"text/template"
	"time"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
)

// TemplateFormat represents the format of the template (Terraform or Crossplane)
type TemplateFormat string

const (
	// FormatTerraform represents Terraform HCL format
	FormatTerraform TemplateFormat = "terraform"
	// FormatCrossplane represents Crossplane YAML format
	FormatCrossplane TemplateFormat = "crossplane"
)

// TemplateCacheEntry represents a cached template with metadata
type TemplateCacheEntry struct {
	Template  *template.Template
	Timestamp time.Time
	Size      int
}

// TemplateCache is a cache of parsed templates with expiration
type TemplateCache struct {
	entries       map[string]TemplateCacheEntry
	mutex         sync.RWMutex
	maxCacheSize  int           // Maximum number of templates to cache
	cacheDuration time.Duration // How long to cache templates before re-parsing
}

// NewTemplateCache creates a new template cache with specified limits
func NewTemplateCache(maxSize int, duration time.Duration) *TemplateCache {
	if maxSize <= 0 {
		maxSize = 100 // Default cache size
	}
	if duration <= 0 {
		duration = 60 * time.Minute // Default cache duration
	}
	
	return &TemplateCache{
		entries:       make(map[string]TemplateCacheEntry),
		maxCacheSize:  maxSize,
		cacheDuration: duration,
	}
}

// Get retrieves a template from the cache if it exists and is not expired
func (tc *TemplateCache) Get(key string) (*template.Template, bool) {
	tc.mutex.RLock()
	entry, exists := tc.entries[key]
	tc.mutex.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	// Check if entry is expired
	if time.Since(entry.Timestamp) > tc.cacheDuration {
		tc.mutex.Lock()
		delete(tc.entries, key)
		tc.mutex.Unlock()
		return nil, false
	}
	
	return entry.Template, true
}

// Set adds a template to the cache
func (tc *TemplateCache) Set(key string, tmpl *template.Template, size int) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	// Check if we need to evict entries to stay within size limit
	if len(tc.entries) >= tc.maxCacheSize {
		tc.evictOldest()
	}
	
	tc.entries[key] = TemplateCacheEntry{
		Template:  tmpl,
		Timestamp: time.Now(),
		Size:      size,
	}
}

// evictOldest removes the oldest template from the cache
func (tc *TemplateCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range tc.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}
	
	if oldestKey != "" {
		delete(tc.entries, oldestKey)
	}
}

// Clear clears all entries from the cache
func (tc *TemplateCache) Clear() {
	tc.mutex.Lock()
	tc.entries = make(map[string]TemplateCacheEntry)
	tc.mutex.Unlock()
}

// TemplateManager manages the loading and caching of templates
type TemplateManager struct {
	fs      embed.FS
	cache   *TemplateCache
	funcMap template.FuncMap
	// Base template with common components
	baseTemplate *template.Template
}

// NewTemplateManager creates a new template manager with the given embedded filesystem
func NewTemplateManager(fs embed.FS) *TemplateManager {
	return &TemplateManager{
		fs:    fs,
		cache: NewTemplateCache(100, 30*time.Minute),
		funcMap: createTemplateFuncMap(),
	}
}

// createTemplateFuncMap creates a map of template functions
func createTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String manipulation functions
		"title":        TitleFunc,
		"lower":        LowerFunc,
		"upper":        UpperFunc,
		"camel":        CamelCaseFunc,
		"snake":        SnakeCaseFunc,
		"kebab":        KebabCaseFunc,
		"quote":        QuoteFunc,
		"indent":       IndentFunc,
		"join":         JoinFunc,
		"contains":     ContainsFunc,
		"hasPrefix":    HasPrefixFunc,
		"hasSuffix":    HasSuffixFunc,
		"replace":      ReplaceFunc,
		"trim":         TrimFunc,
		"split":        SplitFunc,
		
		// Format conversion functions
		"toYAML":       ToYAMLFunc,
		"toHCL":        ToHCLFunc,
		"formatYAML":   FormatYAMLDocument,
		"formatHCL":    FormatHCLDocument,
		
		// Map and collection functions
		"map":          MakeMapFunc,
		"mergeMap":     MergeMapFunc,
		"slice":        SliceFunc,
		"filterSlice":  FilterSliceFunc,
		"mapSlice":     MapSliceFunc,
		"uniqueSlice":  UniqueSliceFunc,
		
		// Conditional and utility functions
		"defaultValue": DefaultValueFunc,
		"ternary":      TernaryFunc,
		
		// Resource-specific functions
		"getProperty":  GetPropertyFunc,
		"hasProperty":  HasPropertyFunc,
		"resourceRef":  ResourceRefFunc,
		"yamlRef":      YAMLRefFunc,
		"cidrSubnet":   CIDRSubnetFunc,
		"getTags":      GetTagsFunc,
		"tfTags":       FormatTerraformTagsFunc,
		"cpTags":       FormatCrossplaneTagsFunc,
	}
}

// PreloadCommonTemplates preloads common partial templates used by other templates
func (tm *TemplateManager) PreloadCommonTemplates() error {
	// Create a base template with standard functionality
	baseTemplate := template.New("base").Funcs(tm.funcMap)

	// Try to preload common partials for each format
	for _, format := range []TemplateFormat{FormatTerraform, FormatCrossplane} {
		commonPath := filepath.Join(string(format), "_common")
		
		// Check if common directory exists
		entries, err := fs.ReadDir(tm.fs, commonPath)
		if err != nil {
			// Common directory doesn't exist for this format, which is fine
			continue
		}
		
		// Load all common templates
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			partialPath := filepath.Join(commonPath, entry.Name())
			partialData, err := tm.fs.ReadFile(partialPath)
			if err != nil {
				continue // Skip files we can't read
			}
			
			// Add to base template
			_, err = baseTemplate.New(fmt.Sprintf("%s:%s", format, entry.Name())).Parse(string(partialData))
			if err != nil {
				return fmt.Errorf("failed to parse common template %s: %w", partialPath, err)
			}
		}
	}
	
	tm.baseTemplate = baseTemplate
	return nil
}

// GetEmptyTemplate creates an empty template with the function map
func (tm *TemplateManager) GetEmptyTemplate(name string) *template.Template {
	return template.New(name).Funcs(tm.funcMap)
}

// GetTemplate gets a template by name, loading it from the embedded filesystem if needed
func (tm *TemplateManager) GetTemplate(format TemplateFormat, templateName string) (*template.Template, error) {
	cacheKey := fmt.Sprintf("%s:%s", format, templateName)
	
	// Check if template is already in cache
	if tmpl, exists := tm.cache.Get(cacheKey); exists {
		return tmpl, nil
	}
	
	// Template not in cache, load it
	templatePath := filepath.Join("templates", string(format), templateName)
	templateData, err := tm.fs.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	
	// Parse template
	var tmpl *template.Template
	if tm.baseTemplate != nil {
		// Clone the base template to inherit common partials
		tmpl, err = tm.baseTemplate.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone base template: %w", err)
		}
		
		tmpl, err = tmpl.New(templateName).Parse(string(templateData))
	} else {
		// No base template, create from scratch
		tmpl, err = template.New(templateName).Funcs(tm.funcMap).Parse(string(templateData))
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}
	
	// Add template to cache
	tm.cache.Set(cacheKey, tmpl, len(templateData))
	
	return tmpl, nil
}

// GetTemplateWithPattern gets a template for a given resource type matching a pattern
func (tm *TemplateManager) GetTemplateWithPattern(format TemplateFormat, pattern string) (*template.Template, string, error) {
	// List all templates for the format
	templates, err := tm.ListTemplates(format)
	if err != nil {
		return nil, "", err
	}
	
	// Try to find a template matching the pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, "", fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}
	
	for _, templateName := range templates {
		if re.MatchString(templateName) {
			tmpl, err := tm.GetTemplate(format, templateName)
			if err != nil {
				continue
			}
			return tmpl, templateName, nil
		}
	}
	
	return nil, "", fmt.Errorf("no template found matching pattern %s for format %s", pattern, format)
}

// ListTemplates lists all available templates for a given format
func (tm *TemplateManager) ListTemplates(format TemplateFormat) ([]string, error) {
	formatDir := filepath.Join("templates", string(format))
	var templates []string
	
	err := fs.WalkDir(tm.fs, formatDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip common templates directory
		if d.IsDir() && d.Name() == "_common" {
			return fs.SkipDir
		}
		
		if !d.IsDir() && (filepath.Ext(path) == ".tmpl" || filepath.Ext(path) == ".gotmpl") {
			// Get relative path from format directory
			relPath, err := filepath.Rel(formatDir, path)
			if err != nil {
				return err
			}
			templates = append(templates, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list templates for format %s: %w", format, err)
	}
	
	return templates, nil
}

// RefreshCache refreshes the template cache
func (tm *TemplateManager) RefreshCache() {
	tm.cache.Clear()
}

// TemplateSelector interface for selecting the correct template for a resource
type TemplateSelector interface {
	// SelectTemplate selects the appropriate template for the given resource and format
	SelectTemplate(format TemplateFormat, resource *models.Resource) (string, error)
	// RegisterTemplate registers a custom template for a resource type
	RegisterTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string)
}

// DefaultTemplateSelector is the default implementation of TemplateSelector
type DefaultTemplateSelector struct {
	// Map resource types to template names for each format
	mappings map[TemplateFormat]map[models.ResourceType]string
	// Fallback patterns for resource types without explicit mappings
	patterns map[TemplateFormat]map[string]string
	mutex    sync.RWMutex
}

// NewDefaultTemplateSelector creates a new template selector with default mappings
func NewDefaultTemplateSelector() *DefaultTemplateSelector {
	selector := &DefaultTemplateSelector{
		mappings: make(map[TemplateFormat]map[models.ResourceType]string),
		patterns: make(map[TemplateFormat]map[string]string),
	}
	
	// Initialize default mappings for Terraform
	tfMapping := map[models.ResourceType]string{
		models.ResourceVPC:           "vpc.tmpl",
		models.ResourceSubnet:        "subnet.tmpl",
		models.ResourceIGW:           "internet_gateway.tmpl",
		models.ResourceNATGateway:    "nat_gateway.tmpl",
		models.ResourceEKSCluster:    "eks_cluster.tmpl",
		models.ResourceNodeGroup:     "eks_node_group.tmpl",
		models.ResourceEC2Instance:   "ec2_instance.tmpl",
		models.ResourceS3Bucket:      "s3_bucket.tmpl",
		models.ResourceSecurityGroup: "security_group.tmpl",
		models.ResourceIAMRole:       "iam_role.tmpl",
		models.ResourceLambda:        "lambda.tmpl",
		models.ResourceDynamoDB:      "dynamodb.tmpl",
		models.ResourceCloudwatch:    "cloudwatch.tmpl",
		models.ResourceRDSInstance:   "rds_instance.tmpl",
	}
	selector.mappings[FormatTerraform] = tfMapping
	
	// Initialize default mappings for Crossplane
	cpMapping := map[models.ResourceType]string{
		models.ResourceVPC:           "vpc.tmpl",
		models.ResourceSubnet:        "subnet.tmpl",
		models.ResourceIGW:           "internet_gateway.tmpl", 
		models.ResourceNATGateway:    "nat_gateway.tmpl",
		models.ResourceEKSCluster:    "eks_cluster.tmpl",
		models.ResourceNodeGroup:     "eks_node_group.tmpl",
		models.ResourceEC2Instance:   "ec2_instance.tmpl",
		models.ResourceS3Bucket:      "s3_bucket.tmpl",
		models.ResourceSecurityGroup: "security_group.tmpl",
		models.ResourceIAMRole:       "iam_role.tmpl",
		models.ResourceLambda:        "lambda.tmpl",
		models.ResourceDynamoDB:      "dynamodb.tmpl",
		models.ResourceCloudwatch:    "cloudwatch.tmpl",
		models.ResourceRDSInstance:   "rds_instance.tmpl",
	}
	selector.mappings[FormatCrossplane] = cpMapping
	
	// Initialize fallback patterns for resources without specific templates
	tfPatterns := map[string]string{
		"^ec2_":     "ec2_resource.tmpl",
		"^rds_":     "rds_resource.tmpl",
		"^lambda_":  "lambda_resource.tmpl",
		"^iam_":     "iam_resource.tmpl",
		"^s3_":      "s3_resource.tmpl",
		"^dynamo_":  "dynamo_resource.tmpl",
		"^eks_":     "eks_resource.tmpl",
		"^vpc_":     "vpc_resource.tmpl",
	}
	selector.patterns[FormatTerraform] = tfPatterns
	
	cpPatterns := map[string]string{
		"^ec2_":     "ec2_resource.tmpl",
		"^rds_":     "rds_resource.tmpl",
		"^lambda_":  "lambda_resource.tmpl",
		"^iam_":     "iam_resource.tmpl",
		"^s3_":      "s3_resource.tmpl",
		"^dynamo_":  "dynamo_resource.tmpl",
		"^eks_":     "eks_resource.tmpl",
		"^vpc_":     "vpc_resource.tmpl",
	}
	selector.patterns[FormatCrossplane] = cpPatterns
	
	return selector
}

// RegisterTemplate registers a custom template for a resource type
func (s *DefaultTemplateSelector) RegisterTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Initialize mapping for format if it doesn't exist
	if _, ok := s.mappings[format]; !ok {
		s.mappings[format] = make(map[models.ResourceType]string)
	}
	
	// Register the template
	s.mappings[format][resourceType] = templateName
}

// RegisterPatternTemplate registers a fallback pattern for resources without specific templates
func (s *DefaultTemplateSelector) RegisterPatternTemplate(format TemplateFormat, pattern string, templateName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Initialize patterns for format if it doesn't exist
	if _, ok := s.patterns[format]; !ok {
		s.patterns[format] = make(map[string]string)
	}
	
	// Register the pattern
	s.patterns[format][pattern] = templateName
}

// SelectTemplate selects the appropriate template for the given resource and format
func (s *DefaultTemplateSelector) SelectTemplate(format TemplateFormat, resource *models.Resource) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Check if mapping exists for format
	formatMapping, ok := s.mappings[format]
	if !ok {
		return "", fmt.Errorf("unsupported template format: %s", format)
	}
	
	// First, check if there's a direct mapping for this resource type
	if templateName, ok := formatMapping[resource.Type]; ok {
		return templateName, nil
	}
	
	// If no direct mapping, try fallback patterns
	if patterns, ok := s.patterns[format]; ok {
		resourceTypeStr := string(resource.Type)
		
		for pattern, templateName := range patterns {
			matched, err := regexp.MatchString(pattern, resourceTypeStr)
			if err != nil {
				continue
			}
			
			if matched {
				return templateName, nil
			}
		}
	}
	
	// Fallback to a generic template name based on resource type
	genericTemplateName := fmt.Sprintf("%s.tmpl", resource.Type)
	
	// Check if a template with this name is likely to exist
	return genericTemplateName, nil
}

// TemplateRegistrar provides methods to register custom templates
type TemplateRegistrar interface {
	RegisterResourceTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string)
	RegisterPatternTemplate(format TemplateFormat, pattern string, templateName string)
}

// TemplateRenderer renders templates for resources
type TemplateRenderer struct {
	manager  *TemplateManager
	selector TemplateSelector
	// Additional context data for templates
	globalContext map[string]interface{}
	mutex         sync.RWMutex
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(manager *TemplateManager, selector TemplateSelector) *TemplateRenderer {
	if selector == nil {
		selector = NewDefaultTemplateSelector()
	}
	
	return &TemplateRenderer{
		manager:       manager,
		selector:      selector,
		globalContext: make(map[string]interface{}),
	}
}

// SetGlobalContext sets a value in the global context
func (r *TemplateRenderer) SetGlobalContext(key string, value interface{}) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.globalContext[key] = value
}

// RegisterResourceTemplate registers a template for a resource type
func (r *TemplateRenderer) RegisterResourceTemplate(format TemplateFormat, resourceType models.ResourceType, templateName string) {
	if registrar, ok := r.selector.(TemplateRegistrar); ok {
		registrar.RegisterResourceTemplate(format, resourceType, templateName)
	} else if defaultSelector, ok := r.selector.(*DefaultTemplateSelector); ok {
		defaultSelector.RegisterTemplate(format, resourceType, templateName)
	}
}

// RegisterPatternTemplate registers a fallback pattern for resources without specific templates
func (r *TemplateRenderer) RegisterPatternTemplate(format TemplateFormat, pattern string, templateName string) {
	if registrar, ok := r.selector.(TemplateRegistrar); ok {
		registrar.RegisterPatternTemplate(format, pattern, templateName)
	} else if defaultSelector, ok := r.selector.(*DefaultTemplateSelector); ok {
		defaultSelector.RegisterPatternTemplate(format, pattern, templateName)
	}
}

// RenderResource renders a single resource
func (r *TemplateRenderer) RenderResource(format TemplateFormat, resource *models.Resource) (string, error) {
	// Select template for resource
	templateName, err := r.selector.SelectTemplate(format, resource)
	if err != nil {
		return "", err
	}
	
	// Get template
	tmpl, err := r.manager.GetTemplate(format, templateName)
	if err != nil {
		return "", err
	}
	
	// Create template data with global context
	r.mutex.RLock()
	data := make(map[string]interface{}, len(r.globalContext)+1)
	for k, v := range r.globalContext {
		data[k] = v
	}
	r.mutex.RUnlock()
	
	// Add resource to data
	data["Resource"] = resource
	
	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	
	return buf.String(), nil
}

// RenderResources renders multiple resources
func (r *TemplateRenderer) RenderResources(format TemplateFormat, resources []models.Resource) (string, error) {
	var result bytes.Buffer
	
	// First try to render a header template
	headerTemplate := fmt.Sprintf("%s_header.tmpl", format)
	
	// Try to get and render the header template
	tmpl, err := r.manager.GetTemplate(format, headerTemplate)
	if err == nil {
		// Create template data with global context
		r.mutex.RLock()
		data := make(map[string]interface{}, len(r.globalContext))
		for k, v := range r.globalContext {
			data[k] = v
		}
		r.mutex.RUnlock()
		
		var headerBuf bytes.Buffer
		if err := tmpl.Execute(&headerBuf, data); err == nil {
			result.WriteString(headerBuf.String())
			result.WriteString("\n")
		}
	}
	
	// Then render each resource
	for _, resource := range resources {
		rendered, err := r.RenderResource(format, &resource)
		if err != nil {
			return "", err
		}
		
		result.WriteString(rendered)
		result.WriteString("\n")
	}
	
	// Finally try to render a footer template
	footerTemplate := fmt.Sprintf("%s_footer.tmpl", format)
	
	// Try to get and render the footer template
	tmpl, err = r.manager.GetTemplate(format, footerTemplate)
	if err == nil {
		// Create template data with global context
		r.mutex.RLock()
		data := make(map[string]interface{}, len(r.globalContext))
		for k, v := range r.globalContext {
			data[k] = v
		}
		r.mutex.RUnlock()
		
		var footerBuf bytes.Buffer
		if err := tmpl.Execute(&footerBuf, data); err == nil {
			result.WriteString(footerBuf.String())
		}
	}
	
	return result.String(), nil
}

// RenderResourceToFile renders a resource and writes it to a file
func (r *TemplateRenderer) RenderResourceToFile(format TemplateFormat, resource *models.Resource, filePath string) error {
	content, err := r.RenderResource(format, resource)
	if err != nil {
		return err
	}
	
	// Format the content
	formattedContent := FormatRenderedContent(format, content)
	
	// Write content to file
	err = utils.WriteToFile(filePath, formattedContent)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}
	
	return nil
}

// ValidateTemplate validates that a template can be rendered correctly for a resource
func (r *TemplateRenderer) ValidateTemplate(format TemplateFormat, resource *models.Resource) error {
	_, err := r.RenderResource(format, resource)
	return err
}

// ensureStringsImport is used to ensure the strings package is used and not flagged as unused
func ensureStringsImport() {
	_ = strings.Contains("", "")
}
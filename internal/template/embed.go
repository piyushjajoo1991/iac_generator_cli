package template

import "embed"

//go:embed templates/terraform/*.tmpl templates/crossplane/*.tmpl
var TemplateFS embed.FS

// DefaultTemplateManager is the default template manager instance
// that can be used by other packages
var DefaultTemplateManager *TemplateManager

// DefaultTemplateRenderer is the default template renderer instance
// that can be used by other packages
var DefaultTemplateRenderer *TemplateRenderer

func init() {
	// Create default template manager and renderer
	DefaultTemplateManager = NewTemplateManager(TemplateFS)
	DefaultTemplateRenderer = NewTemplateRenderer(DefaultTemplateManager, nil)
}

// GetDefaultManager returns the default template manager
func GetDefaultManager() *TemplateManager {
	return DefaultTemplateManager
}

// GetDefaultRenderer returns the default template renderer
func GetDefaultRenderer() *TemplateRenderer {
	return DefaultTemplateRenderer
}
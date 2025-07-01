package builder

import (
	"embed"
	"text/template"
)

//go:embed templates
var TemplatesFS embed.FS

// GetTemplate loads a template from the embedded filesystem
func GetTemplate(path string) (*template.Template, error) {
	content, err := TemplatesFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	return template.New(path).Parse(string(content))
}

// TemplateData contains common data for all templates
type TemplateData struct {
	// Common fields
	ModuleName     string
	ProjectName    string
	BinaryName     string
	AppName        string
	RootDir        string
	Extensions     []string
	RediVersion    string
	IsSourceInstall bool
	ReplaceDir     string
}
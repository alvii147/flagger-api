package templatesmanager

import (
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed email/*.txt email/*.html
var templatesFS embed.FS

// Manager manages and loads template files.
type Manager interface {
	Load(name string) (*texttemplate.Template, *htmltemplate.Template, error)
}

// manager implements Manager.
type manager struct{}

// NewManager creates and returns a new manager.
func NewManager() *manager {
	return &manager{}
}

// Load loads the text and html files by a given name.
func (m *manager) Load(name string) (*texttemplate.Template, *htmltemplate.Template, error) {
	textTmplFile := "email/" + name + ".txt"
	htmlTmplFile := "email/" + name + ".html"

	textTmpl, err := texttemplate.ParseFS(templatesFS, textTmplFile)
	if err != nil {
		return nil, nil, fmt.Errorf("Load failed to texttemplate.ParseFS %s: %w", textTmplFile, err)
	}

	htmlTmpl, err := htmltemplate.ParseFS(templatesFS, htmlTmplFile)
	if err != nil {
		return nil, nil, fmt.Errorf("Load failed to htmltemplate.ParseFS %s: %w", htmlTmplFile, err)
	}

	return textTmpl, htmlTmpl, nil
}

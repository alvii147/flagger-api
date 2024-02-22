package templatesmanager

import (
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed email/*.txt email/*.html
var templatesFS embed.FS

type Manager interface {
	Load(name string) (*texttemplate.Template, *htmltemplate.Template, error)
}

type manager struct{}

func NewManager() *manager {
	return &manager{}
}

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

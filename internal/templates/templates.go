package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html
var templateFS embed.FS

// GetLoginTemplate returns the parsed login template
func GetLoginTemplate() (*template.Template, error) {
	return template.ParseFS(templateFS, "login.html")
}

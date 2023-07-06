package config

import (
	"html/template"

	"github.com/alexedwards/scs/v2"
)

// AppConfig hold the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProd        bool
	Session       *scs.SessionManager
}

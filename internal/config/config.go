package config

import (
	"html/template"
	"log"

	"github.com/Ed-cred/bookings/internal/models"
	"github.com/alexedwards/scs/v2"
)

// AppConfig hold the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProd        bool
	Session       *scs.SessionManager
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	MailChan      chan models.MailData
}

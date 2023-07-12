package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig
var pathToTemplate = "./templates"

// NewTemplates sets the config for the package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	td.CSRFToken = nosurf.Token(r)
	return td
}

//RenderTemplate is a helper function to render html templates

func RenderTemplate(w http.ResponseWriter, tmpl string, r *http.Request, td *models.TemplateData) error {
	var tc map[string]*template.Template
	if app.UseCache {
		//get template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}
	//get requested template from cache
	t, ok := tc[tmpl]
	if !ok {
		err := errors.New("template cache not found")
		return err
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)
	err := t.Execute(buf, td)
	if err != nil {
		log.Println("execution error: ", err)
		return err
	}

	// render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println("write error: ", err)
		return err
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)
	// get files named .page.tmpl

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return cache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return cache, err
		}
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err != nil {
			return cache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err != nil {
				return cache, err
			}
		}
		cache[name] = ts

	}
	return cache, nil

}

//! Simple cache implementation
// var cache = make(map[string]*template.Template)

// func RenderTemplate(w http.ResponseWriter, t string) {
// 	var err error
// 	//* check to see if template is already cached else we cache it

// 	_, inMap := cache[t]
// 	if !inMap {
// 		log.Println("creating template")
// 		err := createTemplateCache(t)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	} else {
// 		log.Println("using cached template")
// 	}

// 	tmpl := cache[t]
// 	err = tmpl.Execute(w, nil)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// func createTemplateCache(t string) error {
// 	templates := []string{
// 		"./templates/" + t,
// 		"./templates/base.layout.tmpl",
// 	}
// 	tmpl, err := template.ParseFiles(templates...)
// 	if err != nil {
// 		fmt.Println("parsing template error: ", err)
// 		return err
// 	}
// 	cache[t] = tmpl
// 	return nil

// }

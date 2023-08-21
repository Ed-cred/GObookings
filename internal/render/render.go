package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Ed-cred/bookings/internal/config"
	"github.com/Ed-cred/bookings/internal/models"
	"github.com/justinas/nosurf"
)
var functions =template.FuncMap{
	"humanDate": HumanDate,
	"formatDate": FormatDate,
	"iterate": IterateDays,
}
var app *config.AppConfig
var pathToTemplate = "./templates"

// NewRenderer sets the config for the package
func NewRenderer(a *config.AppConfig) {
	app = a
}

//Formats time.Time dates into yyyy-mm-dd 
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

//Formats time.Time into any format string provided it is allowed 
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func IterateDays(count int) []int {
	var i int 
	var items []int
	for i=1; i<count+1; i++ {
		items = append(items, i)
	}
	return items
}


func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

//Template is a helper function to render html templates

func Template(w http.ResponseWriter, tmpl string, r *http.Request, td *models.TemplateData) error {
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
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
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


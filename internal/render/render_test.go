package render

import (
	"log"
	"net/http"
	"testing"

	"github.com/Ed-cred/bookings/internal/models"
)

	

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	session.Put(r.Context(), "flash", "123")
	result := AddDefaultData(&td, r)
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplate = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Log(err)
		t.Error(err)
	}
	app.TemplateCache = tc
	if app.TemplateCache == nil {
		t.Error("template cache is empty")
	}
	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	var ww myWriter
	err = RenderTemplate(&ww, "home.page.tmpl", r, &models.TemplateData{})
	if err != nil {
		t.Error("Error loading template to browser: ", err)
	}
	err = RenderTemplate(&ww, "non-existent.page.tmpl", r, &models.TemplateData{})
	if err == nil {
		t.Error("Should have thrown an error for non existent template")
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplates (t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplate = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}	


}
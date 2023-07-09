package handlers

import (
	"net/http"

	"github.com/Ed-cred/bookings/pkg/config"
	"github.com/Ed-cred/bookings/pkg/models"
	"github.com/Ed-cred/bookings/pkg/render"
)

//TemplateData holds data sent from handlers to templates

// Repository used by the handlers
var Repo *Repository

// Repository type
type Repository struct {
	App *config.AppConfig
}

// Creates a new repository
func NewRepository(app *config.AppConfig) *Repository {
	return &Repository{
		App: app,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (rep *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	rep.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})

}

func (rep *Repository) Generals(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "generals.page.tmpl", &models.TemplateData{})

}
func (rep *Repository) Majors(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "majors.page.tmpl", &models.TemplateData{})

}
func (rep *Repository) Availability(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "search_availability.page.tmpl", &models.TemplateData{})

}

func (rep *Repository) Contact(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "contact.page.tmpl", &models.TemplateData{})

}
func (rep *Repository) Reservation(w http.ResponseWriter, r *http.Request) {

	render.RenderTemplate(w, "make_reservation.page.tmpl", &models.TemplateData{})

}
func (rep *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello again"
	remoteIP := rep.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP
	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})

}
